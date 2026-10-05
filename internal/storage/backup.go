package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

// Backup creates a point-in-time consistent SQLite backup using the native VACUUM INTO statement.
// This safely handles active WAL files and concurrent readers.
func Backup(ctx context.Context, db *sql.DB, destPath string) error {
	if destPath == "" {
		return fmt.Errorf("backup destination path cannot be empty")
	}

	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory %q: %w", destDir, err)
	}

	// SQLite VACUUM INTO requires the target file to not exist prior to invocation
	if _, err := os.Stat(destPath); err == nil {
		if err := os.Remove(destPath); err != nil {
			return fmt.Errorf("failed to remove existing target backup file %q: %w", destPath, err)
		}
	}

	// Execute SQLite VACUUM INTO
	query := "VACUUM INTO ?"
	if _, err := db.ExecContext(ctx, query, destPath); err != nil {
		return fmt.Errorf("VACUUM INTO failed: %w", err)
	}

	// Verify backup was created and is non-empty
	fi, err := os.Stat(destPath)
	if err != nil {
		return fmt.Errorf("backup verification failed: %w", err)
	}
	if fi.Size() == 0 {
		return fmt.Errorf("backup file created at %q has 0 bytes", destPath)
	}

	return nil
}
