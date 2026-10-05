package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// Open opens and initializes a SQLite database connection with required pragmas.
func Open(path string) (*sql.DB, error) {
	if path == "" {
		return nil, fmt.Errorf("database path cannot be empty")
	}

	isMemory := path == ":memory:" || strings.HasPrefix(path, "file::memory:") || strings.Contains(path, "mode=memory")
	if !isMemory {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %q: %w", dir, err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database %q: %w", path, err)
	}

	// SQLite connection pooling: single connection avoids database locking issues
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Execute pragmas
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout = 5000;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	if !isMemory {
		if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL;"); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
		}
		if _, err := db.ExecContext(ctx, "PRAGMA synchronous = NORMAL;"); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("failed to set synchronous normal: %w", err)
		}
	}

	return db, nil
}

// OpenMemory creates an in-memory SQLite database for unit and integration testing.
func OpenMemory() (*sql.DB, error) {
	return Open(":memory:")
}
