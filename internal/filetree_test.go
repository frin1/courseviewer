package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildFileTree(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "courseviewer-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFiles := []string{
		"file1.txt",
		"file2.html",
		"dir1/file3.txt",
		"dir1/dir2/file4.mp4",
	}

	for _, tf := range testFiles {
		path := filepath.Join(tmpDir, tf)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if f, err := os.Create(path); err != nil {
			t.Fatal(err)
		} else {
			f.Close()
		}
	}

	config := Config{
		BasePath:      tmpDir,
		HiddenFileExt: []string{".srt"},
	}

	tree := buildFileTree(config, "")

	if tree.Name != filepath.Base(tmpDir) {
		t.Errorf("wrong root name, got %s, want %s", tree.Name, filepath.Base(tmpDir))
	}
	if !tree.IsDir {
		t.Error("root should be a directory")
	}
	if len(tree.Children) != 3 {
		t.Errorf("wrong number of children, got %d, want 3", len(tree.Children))
	}
}
