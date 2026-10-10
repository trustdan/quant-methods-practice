package storage

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationsFreshDatabase(t *testing.T) {
	db, err := OpenMemory()
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	backupCalled := false
	backupFn := func() error {
		backupCalled = true
		return nil
	}

	if err := RunMigrations(ctx, db, backupFn); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	if backupCalled {
		t.Fatalf("expected backup to NOT be called on fresh database with no prior migrations")
	}

	applied, err := GetAppliedMigrations(ctx, db)
	if err != nil {
		t.Fatalf("GetAppliedMigrations failed: %v", err)
	}

	if len(applied) == 0 {
		t.Fatalf("expected at least 1 applied migration, got %d", len(applied))
	}

	m1, ok := applied[1]
	if !ok {
		t.Fatalf("expected migration version 1 to be applied")
	}
	if m1.Name != "initial_schema" {
		t.Errorf("expected migration name 'initial_schema', got %q", m1.Name)
	}
	if m1.Checksum == "" {
		t.Errorf("expected non-empty migration checksum")
	}

	// Verify required tables exist
	requiredTables := []string{
		"schema_migrations",
		"settings",
		"sessions",
		"question_instances",
		"drill_stage_states",
		"attempts",
		"assistance_events",
		"session_drafts",
		"command_idempotency",
		"mastery_projections",
		"candidate_questions",
		"content_approval_events",
		"saved_explanations",
		"tutor_drafts",
		"exam_responses",
		"arcade_scores",
	}

	for _, table := range requiredTables {
		var name string
		err := db.QueryRowContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("required table %q not found in schema: %v", table, err)
		}
	}
}

func TestMigrationsIdempotency(t *testing.T) {
	db, err := OpenMemory()
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := RunMigrations(ctx, db, nil); err != nil {
		t.Fatalf("initial RunMigrations failed: %v", err)
	}

	// Running migrations a second time should be a no-op
	if err := RunMigrations(ctx, db, nil); err != nil {
		t.Fatalf("second RunMigrations failed: %v", err)
	}

	applied, err := GetAppliedMigrations(ctx, db)
	if err != nil {
		t.Fatalf("GetAppliedMigrations failed: %v", err)
	}
	all, err := LoadMigrations()
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != len(all) {
		t.Errorf("expected %d applied migrations, got %d", len(all), len(applied))
	}
}

func TestMigrationsChecksumMismatch(t *testing.T) {
	db, err := OpenMemory()
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := EnsureSchemaMigrationsTable(ctx, db); err != nil {
		t.Fatalf("EnsureSchemaMigrationsTable failed: %v", err)
	}

	// Insert a tampered record for version 1 with a fake checksum
	_, err = db.ExecContext(ctx,
		"INSERT INTO schema_migrations (version, name, checksum, applied_at) VALUES (1, 'initial_schema', 'tampered_checksum_12345', '2026-10-05T00:00:00Z')",
	)
	if err != nil {
		t.Fatalf("failed to insert mock migration: %v", err)
	}

	err = RunMigrations(ctx, db, nil)
	if err == nil {
		t.Fatalf("expected checksum mismatch error, got nil")
	}
}

func TestMigrationBackupTriggerOnUpgrade(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "upgrade.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := EnsureSchemaMigrationsTable(ctx, db); err != nil {
		t.Fatalf("failed to ensure migrations table: %v", err)
	}

	// Pretend version 0 was applied
	_, err = db.ExecContext(ctx, "INSERT INTO schema_migrations (version, name, checksum, applied_at) VALUES (0, 'baseline', 'none', '2026-10-05T00:00:00Z')")
	if err != nil {
		t.Fatalf("failed to insert baseline: %v", err)
	}

	backupCalled := false
	backupFn := func() error {
		backupCalled = true
		return Backup(ctx, db, filepath.Join(tmpDir, "upgrade_backup.db"))
	}

	if err := RunMigrations(ctx, db, backupFn); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	if !backupCalled {
		t.Errorf("expected backup to be called before applying new migrations on an existing schema")
	}

	// Verify backup file exists
	if _, err := os.Stat(filepath.Join(tmpDir, "upgrade_backup.db")); err != nil {
		t.Errorf("expected backup file to exist: %v", err)
	}
}

func TestTransactionalRollbackOnFailure(t *testing.T) {
	db, err := OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := EnsureSchemaMigrationsTable(ctx, db); err != nil {
		t.Fatalf("failed to ensure migrations table: %v", err)
	}

	// Simulate running an invalid migration in a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	_, _ = tx.ExecContext(ctx, "CREATE TABLE test_rollback (id INTEGER PRIMARY KEY);")
	// Cause a syntax error
	_, err = tx.ExecContext(ctx, "INVALID SQL SYNTAX HERE;")
	if err == nil {
		t.Fatalf("expected error on invalid SQL")
	}
	_ = tx.Rollback()

	// Verify test_rollback table was not committed
	var name string
	err = db.QueryRowContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name='test_rollback'").Scan(&name)
	if err != sql.ErrNoRows {
		t.Errorf("expected table to not exist after rollback, got err: %v", err)
	}
}
