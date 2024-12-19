// Global state
let currentlyPlaying = null;
let readPaths = new Set();
let lastReadPath = null;
let rootPath = '';

// Core initialization
document.addEventListener('DOMContentLoaded', async () => {
    await loadReadStatus();
});

// State management
async function loadReadStatus() {
    try {
        const response = await fetch('/api/read-status');
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const data = await response.json();
        readPaths = new Set(data.paths);

        await loadFileTree();
        findRootPath();
        if (data.lastRead?.path) {
            lastReadPath = `${rootPath}/${data.lastRead.path}`;
            expandPath(lastReadPath);
            loadContent(data.lastRead.path);
        }
    } catch (error) {
        console.error('Error in loadReadStatus:', error);
    }
}

async function markAsRead(path) {
    try {
        await fetch(`/api/mark-read/${encodeURIComponent(path)}`, {
            method: 'POST'
        });
        readPaths.add(path);
        lastReadPath = path;
        updateTreeReadStatus();
    } catch (error) {
        console.error('Error marking as read:', error);
    }
}

// Tree management
async function loadFileTree() {
    try {
        const response = await fetch('/api/tree');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const tree = await response.json();
        const sidebar = document.getElementById('sidebar');
        sidebar.innerHTML = '<h2>CourseViewer</h2>';
        renderTree(tree, sidebar, '');
        updateTreeReadStatus();
    } catch (error) {
        console.error('Error loading file tree:', error);
    }
}

function renderTree(file, container, fullPath) {
    const item = document.createElement('div');
    const currentPath = fullPath ? `${fullPath}/${file.name}` : file.name;
    
    item.className = `tree-item ${file.isDir ? 'folder' : 'file'}`;
    item.setAttribute('data-path', currentPath);
    item.setAttribute('title', file.name);
    
    if (!file.isDir && readPaths.has(currentPath)) {
        item.classList.add('read');
    }
    if (currentPath === lastReadPath) {
        item.classList.add('last-read');
    }

    const nameSpan = document.createElement('span');
    nameSpan.textContent = file.name;
    item.appendChild(nameSpan);

    if (file.isDir) {
        renderFolder(file, item, currentPath, nameSpan);
    } else {
        renderFile(item, nameSpan, file.path);
    }

    container.appendChild(item);
}

function renderFolder(file, item, currentPath, nameSpan) {
    const folderContent = document.createElement('div');
    folderContent.className = 'children';
    
    nameSpan.addEventListener('click', () => toggleFolder(item));
    
    if (file.children?.length > 0) {
        file.children.forEach(child => renderTree(child, folderContent, currentPath));
        item.appendChild(folderContent);
    }
}

function renderFile(item, nameSpan, path) {
    nameSpan.addEventListener('click', () => {
        loadContent(path);
        highlightSelectedFile(item);
    });
}

// Tree utilities
function findRootPath() {
    const firstFolder = document.querySelector('.tree-item.folder');
    if (firstFolder) {
        rootPath = firstFolder.getAttribute('data-path');
    }
}

function updateTreeReadStatus() {
    findRootPath();
    document.querySelectorAll('.tree-item.file').forEach(item => {
        const path = item.getAttribute('data-path');
        const shortenedPath = path.replace(`${rootPath}/`, '');
        
        if (readPaths.has(shortenedPath)) {
            item.classList.add('read');
        }
        if (shortenedPath === lastReadPath || path === lastReadPath) {
            item.classList.add('last-read');
        }
    });
}

function expandPath(path) {
    if (!path) return;
    
    const pathParts = path.split('/');
    let currentPath = pathParts[0];
    
    for (let i = 1; i < pathParts.length; i++) {
        const folderElement = document.querySelector(`[data-path="${currentPath}"]`);
        if (folderElement?.classList.contains('folder')) {
            folderElement.classList.add('expanded');
            const children = folderElement.querySelector('.children');
            if (children) {
                children.classList.add('visible');
            }
        }
        currentPath += '/' + pathParts[i];
    }
}

function toggleFolder(folderElement) {
    folderElement.classList.toggle('expanded');
    const children = folderElement.querySelector('.children');
    if (children) {
        children.classList.toggle('visible');
    }
}

function highlightSelectedFile(element) {
    const previouslySelected = document.querySelector('.tree-item.selected');
    if (previouslySelected) {
        previouslySelected.classList.remove('selected');
    }
    element.classList.add('selected');
}

// Content loading
async function loadContent(path) {
    const contentDiv = document.getElementById('content');
    const contentFrame = document.getElementById('content-frame');
    const fileExtension = path.split('.').pop().toLowerCase();
    
    try {
        if (fileExtension === 'html') {
            return loadHtmlContent(contentDiv, contentFrame, path);
        }
        
        contentFrame.classList.remove('visible');
        contentDiv.classList.remove('hidden');
        contentDiv.innerHTML = '';

        // Add file path header
        const header = document.createElement('h5');
        header.style.fontWeight = 'bold';
        header.style.marginBottom = '1rem';
        header.textContent = path;
        contentDiv.appendChild(header);
        
        if (fileExtension === 'mp4') {
            return loadVideoContent(contentDiv, path);
        }

        return loadTextContent(contentDiv, path, fileExtension);
    } catch (error) {
        console.error('Error loading content:', error);
        contentDiv.innerHTML = `Error loading content: ${error.message}`;
    }
}

function loadHtmlContent(contentDiv, contentFrame, path) {
    contentDiv.classList.add('hidden');
    contentFrame.classList.add('visible');
    contentFrame.src = encodeURI(`/content/${path}`);
}

function loadVideoContent(contentDiv, path) {
    if (currentlyPlaying) {
        currentlyPlaying.pause();
    }
    
    const video = createVideoElement(path);
    const readButton = createReadButton(path);
    
    contentDiv.appendChild(readButton);
    contentDiv.appendChild(video);
    currentlyPlaying = video;
}

function createVideoElement(path) {
    const video = document.createElement('video');
    video.controls = true;
    video.autoplay = true;
    video.style.width = '100%';
    video.style.maxHeight = '80vh';
    video.preload = 'auto';
    
    const source = document.createElement('source');
    source.src = encodeURI(`/content/${path}`);
    source.type = 'video/mp4';
    
    video.appendChild(source);
    return video;
}

async function loadTextContent(contentDiv, path, fileExtension) {
    if (fileExtension === 'pdf') {
        await loadPdfContent(contentDiv, path);
        return;
    }

    const response = await fetch(encodeURI(`/content/${path}`));
    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const text = await response.text();
    const readButton = createReadButton(path);
    contentDiv.appendChild(readButton);
    
    switch (fileExtension) {
        case 'md':
            contentDiv.innerHTML += marked.parse(text);
            break;
        case 'txt':
            const pre = document.createElement('pre');
            pre.textContent = text;
            contentDiv.appendChild(pre);
            break;
        default:
            contentDiv.innerHTML += text;
    }
}

async function loadPdfContent(contentDiv, path) {
    const pdfjsLib = window['pdfjs-dist/build/pdf'];
    const loadingTask = pdfjsLib.getDocument(encodeURI(`/content/${path}`));
    const pdf = await loadingTask.promise;

    contentDiv.innerHTML = ''; // Clear previous content

    for (let pageNum = 1; pageNum <= pdf.numPages; pageNum++) {
        const page = await pdf.getPage(pageNum);
        const viewport = page.getViewport({ scale: 1.5 });

        const canvas = document.createElement('canvas');
        const context = canvas.getContext('2d');
        canvas.height = viewport.height;
        canvas.width = viewport.width;

        const renderContext = {
            canvasContext: context,
            viewport: viewport
        };

        await page.render(renderContext).promise;
        contentDiv.appendChild(canvas);
    }
}

function createReadButton(path) {
    const readButton = document.createElement('button');
    readButton.className = 'read-button';
    readButton.textContent = readPaths.has(path) ? 'Markert som lest' : 'Marker som lest';
    readButton.onclick = async () => {
        await markAsRead(path);
        readButton.textContent = 'Markert som lest';
    };
    return readButton;
}