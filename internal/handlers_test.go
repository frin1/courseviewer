package internal

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func TestContentHandler(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "courseviewer-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testContent := "test content"
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	config := Config{
		BasePath: tmpDir,
		DB:       &sql.DB{},
	}

	r := mux.NewRouter()
	r.HandleFunc("/content/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fullPath := filepath.Join(config.BasePath, vars["path"])

		file, err := os.Open(fullPath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer file.Close()

		http.ServeContent(w, r, vars["path"], time.Now(), file)
	})

	req := httptest.NewRequest("GET", "/content/test.txt", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("wrong status code, got %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != testContent {
		t.Errorf("wrong content, got %s, want %s", w.Body.String(), testContent)
	}
}

func TestMarkAsRead(t *testing.T) {
	db, err := InitDB(":memory:", true)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	config := Config{DB: db}
	r := mux.NewRouter()

	r.HandleFunc("/api/mark-read/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if _, err := config.DB.Exec("INSERT OR REPLACE INTO read_status (path) VALUES (?)", vars["path"]); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/api/mark-read/test/path.txt", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("wrong status code, got %d, want %d", w.Code, http.StatusOK)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM read_status WHERE path = ?", "test/path.txt").Scan(&count); err != nil {
		t.Errorf("failed to query read status: %v", err)
	}
	if count != 1 {
		t.Errorf("wrong read count, got %d, want 1", count)
	}
}
