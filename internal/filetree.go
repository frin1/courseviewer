package internal

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func buildFileTree(config Config, currentPath string) File {
	fullPath := filepath.Join(config.BasePath, currentPath)
	info, err := os.Stat(fullPath)
	if err != nil {
		log.Fatal(err)
	}

	if !info.IsDir() && shouldHideFile(info.Name(), config.HiddenFileExt) {
		return File{}
	}

	relativePath := currentPath
	if currentPath == "" {
		relativePath = info.Name()
	}

	file := File{
		Name:  info.Name(),
		Path:  relativePath,
		IsDir: info.IsDir(),
	}

	if !info.IsDir() {
		return file
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if !shouldHideFile(entry.Name(), config.HiddenFileExt) {
			childPath := filepath.Join(currentPath, entry.Name())
			childFile := buildFileTree(config, childPath)
			if childFile.Name != "" {
				file.Children = append(file.Children, childFile)
			}
		}
	}

	return file
}

func shouldHideFile(filename string, hiddenExts []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, hidden := range hiddenExts {
		if strings.TrimPrefix(hidden, ".") == strings.TrimPrefix(ext, ".") {
			return true
		}
	}
	return false
}
