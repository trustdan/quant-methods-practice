package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDBPragmasAndOptions(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_pragmas.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Check foreign_keys pragma
	var fk int
	if err := db.QueryRowContext(ctx, "PRAGMA foreign_keys;").Scan(&fk); err != nil {
		t.Fatalf("failed to query foreign_keys pragma: %v", err)
	}
	if fk != 1 {
		t.Errorf("expected foreign_keys = 1, got %d", fk)
	}

	// Check journal_mode pragma (WAL for file databases)
	var jmode string
	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode;").Scan(&jmode); err != nil {
		t.Fatalf("failed to query journal_mode pragma: %v", err)
	}
	if jmode != "wal" {
		t.Errorf("expected journal_mode 'wal', got %q", jmode)
	}
}

func TestResolveDBPath(t *testing.T) {
	// Custom db flag overrides everything
	customDB := filepath.Join("custom", "path", "quant.db")
	resolved, err := ResolveDBPath("", customDB)
	if err != nil {
		t.Fatalf("ResolveDBPath failed: %v", err)
	}
	if resolved != customDB {
		t.Errorf("expected %q, got %q", customDB, resolved)
	}

	// Custom data dir flag
	customDir := filepath.Join("custom", "dir")
	resolved, err = ResolveDBPath(customDir, "")
	if err != nil {
		t.Fatalf("ResolveDBPath failed: %v", err)
	}
	expected := filepath.Join(customDir, "quant-methods.db")
	if resolved != expected {
		t.Errorf("expected %q, got %q", expected, resolved)
	}

	// Environment variable override
	t.Setenv("QUANT_DATA_DIR", customDir)
	resolved, err = ResolveDBPath("", "")
	if err != nil {
		t.Fatalf("ResolveDBPath failed: %v", err)
	}
	if resolved != expected {
		t.Errorf("expected %q, got %q", expected, resolved)
	}
}

func TestDefaultDataDir(t *testing.T) {
	// Unset environment variable
	_ = os.Unsetenv("QUANT_DATA_DIR")

	dir, err := DefaultDataDir()
	if err != nil {
		t.Fatalf("DefaultDataDir failed: %v", err)
	}
	if dir == "" {
		t.Fatalf("expected non-empty default data directory")
	}
	if filepath.Base(dir) != AppName {
		t.Errorf("expected directory base %q, got %q", AppName, filepath.Base(dir))
	}
}
