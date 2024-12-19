package internal

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, config Config, content embed.FS) {
	r.HandleFunc("/api/tree", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Building file tree from path: %s", config.BasePath)
		tree := buildFileTree(config, "")
		if err := json.NewEncoder(w).Encode(tree); err != nil {
			log.Printf("Error encoding tree: %v", err)
		}
	}).Methods("GET")

	r.HandleFunc("/content/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fullPath := filepath.Join(config.BasePath, vars["path"])

		log.Printf("Accessing file: %s", fullPath)

		file, err := os.Open(fullPath)
		if err != nil {
			log.Printf("Error opening file: %v", err)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ext := strings.ToLower(filepath.Ext(fullPath))
		switch ext {
		case ".md":
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		case ".txt":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		case ".mp4":
			w.Header().Set("Content-Type", "video/mp4")
		default:
			w.Header().Set("Content-Type", "application/octet-stream")
		}

		w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))

		if r.Header.Get("Range") != "" {
			rangeHeader := r.Header.Get("Range")
			rangeStart := strings.Split(strings.Split(rangeHeader, "=")[1], "-")[0]
			start, err := strconv.ParseInt(rangeStart, 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, info.Size()-1, info.Size()))
			w.Header().Set("Accept-Ranges", "bytes")
			w.WriteHeader(http.StatusPartialContent)
			file.Seek(start, io.SeekStart)
		}

		_, err = io.Copy(w, file)
		if err != nil {
			log.Printf("Error copying file: %v", err)
		}
	}).Methods("GET")

	r.HandleFunc("/api/mark-read/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		_, err := config.DB.Exec("INSERT OR REPLACE INTO read_status (path) VALUES (?)", vars["path"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}).Methods("POST")

	r.HandleFunc("/api/read-status", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Fetching read status...")
		rows, err := config.DB.Query("SELECT path FROM read_status")
		if err != nil {
			log.Printf("Error querying read status: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var paths []string
		for rows.Next() {
			var path string
			if err := rows.Scan(&path); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			paths = append(paths, path)
		}

		var lastRead ReadStatus
		err = config.DB.QueryRow("SELECT path, read_at FROM read_status ORDER BY read_at DESC LIMIT 1").Scan(&lastRead.Path, &lastRead.Timestamp)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := ReadStatusResponse{
			Paths:    paths,
			LastRead: lastRead,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if config.DevMode {
			tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
			tmpl.Execute(w, nil)
		} else {
			tmplContent, err := content.ReadFile("web/templates/index.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tmpl := template.Must(template.New("index.html").Parse(string(tmplContent)))
			tmpl.Execute(w, nil)
		}
	})
}
