package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/frin1/courseviewer/internal"

	"github.com/gorilla/mux"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed web/static/* web/templates/*
var content embed.FS

func main() {
	config := internal.Config{WebRoot: "web"}
	flag.StringVar(&config.BasePath, "path", ".", "Base path for course content")
	flag.BoolVar(&config.DevMode, "dev", false, "Development mode - use local files")
	hiddenExts := flag.String("hide", ".srt", "Comma-separated list of file extensions to hide")
	dbType := flag.String("db", "memory", "Database type: 'memory' for in-memory, 'file' for file-based")
	dbPath := flag.String("dbpath", "content.db", "Path to the database file (used only if db type is 'file')")
	port := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	config.HiddenFileExt = strings.Split(*hiddenExts, ",")

	// Determine if the database should be in-memory or file-based
	inMemory := *dbType == "memory"

	// Database initialization
	db, err := internal.InitDB(*dbPath, inMemory)
	if err != nil {
		log.Fatal(err)
	}
	config.DB = db
	if !inMemory {
		defer db.Close()
	}

	r := mux.NewRouter()

	if config.DevMode {
		log.Println("Running in development mode - using local files")
		r.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
			http.FileServer(http.Dir(filepath.Join(config.WebRoot, "static")))))
	} else {
		log.Println("Running in production mode - using embedded files")
		// Create a sub-filesystem for static files
		staticFS, err := fs.Sub(content, "web/static")
		if err != nil {
			log.Fatal(err)
		}
		r.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
			http.FileServer(http.FS(staticFS))))
	}

	internal.RegisterRoutes(r, config, content)

	log.Printf("Server starting on http://localhost:%d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), r))
}
