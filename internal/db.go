package internal

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dbPath string, inMemory bool) (*sql.DB, error) {
	dsn := dbPath
	if inMemory {
		dsn = ":memory:"
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := setupSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func setupSchema(db *sql.DB) error {
	schema := `
        CREATE TABLE IF NOT EXISTS read_status (
            path TEXT PRIMARY KEY,
            read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        CREATE INDEX IF NOT EXISTS idx_read_at ON read_status(read_at DESC);
    `

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("schema setup failed: %w", err)
	}
	return nil
}
