package storage

import (
	"context"
	"path/filepath"
	"testing"
)

func TestConsistentSQLiteBackup(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "source.db")
	backupPath := filepath.Join(tmpDir, "backup", "source_snapshot.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := RunMigrations(ctx, db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Insert test data
	store := NewStore(db, nil)
	if err := store.SaveSettings(ctx, "test_key", `{"theme":"dark"}`); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	// Run backup
	if err := Backup(ctx, db, backupPath); err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Open the backup file and verify data
	backupDB, err := Open(backupPath)
	if err != nil {
		t.Fatalf("failed to open backup database: %v", err)
	}
	defer backupDB.Close()

	backupStore := NewStore(backupDB, nil)
	val, err := backupStore.GetSettings(ctx, "test_key")
	if err != nil {
		t.Fatalf("failed to read settings from backup: %v", err)
	}
	if val != `{"theme":"dark"}` {
		t.Errorf("expected backup to contain settings, got %q", val)
	}
	_ = backupDB.Close()

	// Test overwriting existing backup safely
	if err := Backup(ctx, db, backupPath); err != nil {
		t.Fatalf("subsequent Backup overwriting target failed: %v", err)
	}
}
