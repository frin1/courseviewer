package internal

import (
	"database/sql"
	"time"
)

type File struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsDir    bool   `json:"isDir"`
	Children []File `json:"children,omitempty"`
}

type Config struct {
	BasePath      string
	HiddenFileExt []string
	DB            *sql.DB
	DevMode       bool
	WebRoot       string
}

type ReadStatus struct {
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
}

type ReadStatusResponse struct {
	Paths    []string   `json:"paths"`
	LastRead ReadStatus `json:"lastRead"`
}
