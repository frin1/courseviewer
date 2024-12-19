package internal

import (
	"os"
	"testing"
)

func TestInitDB(t *testing.T) {
	tests := []struct {
		name     string
		dbPath   string
		inMemory bool
		wantErr  bool
	}{
		{
			name:     "In-memory database",
			dbPath:   "",
			inMemory: true,
			wantErr:  false,
		},
		{
			name:     "File database",
			dbPath:   "test.db",
			inMemory: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := InitDB(tt.dbPath, tt.inMemory)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if db == nil {
				t.Error("expected database connection, got nil")
				return
			}

			var count int
			if err := db.QueryRow("SELECT count(*) FROM read_status").Scan(&count); err != nil {
				t.Errorf("failed to query table: %v", err)
			}

			db.Close()
			if !tt.inMemory {
				os.Remove(tt.dbPath)
			}
		})
	}
}
