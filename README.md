# CourseViewer

CourseViewer is a lightweight web server that provides a clean interface for browsing and tracking progress through locally stored content like courses, documentation, or media libraries.

## Features

- File system browsing with expandable directory tree
- Support for various content types (HTML, Markdown, text, video)
- Progress tracking for viewed content
- In-memory or file-based SQLite database
- Embedded static assets in binary
- Mobile-friendly responsive design

## Installation

```bash
go install github.com/yourusername/CourseViewer@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/courseviewer
cd courseviewer
go build
```

## Usage

```bash
# Basic usage with in-memory database
courseviewer --path /path/to/content

# Hide specific file extensions
courseviewer --path /path/to/content --hide ".srt,.tmp,.DS_Store"

# Use file-based database
courseviewer --path /path/to/content --db file --dbpath ./progress.db

# Development mode (uses local static files)
courseviewer --path /path/to/content --dev
```

Access the content via `http://localhost:8080`

## Options

- `--path`: Root directory containing content (required)
- `--hide`: Comma-separated list of file extensions to hide (default: ".srt")
- `--db`: Database type ("memory" or "file", default: "memory")
- `--dbpath`: Database file location (used with --db file)
- `--dev`: Development mode using local static files

## Development

Requirements:
- Go 1.23 or higher
- SQLite3

Run tests:
```bash
go test -v ./...
```
