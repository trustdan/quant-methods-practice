package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Migration represents a versioned schema change.
type Migration struct {
	Version  int
	Name     string
	SQL      string
	Checksum string
}

// AppliedMigration records a previously executed migration.
type AppliedMigration struct {
	Version   int
	Name      string
	Checksum  string
	AppliedAt time.Time
}

// LoadMigrations parses all embedded migration SQL files in ascending order.
func LoadMigrations() ([]Migration, error) {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid migration filename format %q (expected <version>_<name>.sql)", entry.Name())
		}

		ver, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid version prefix in migration %q: %w", entry.Name(), err)
		}

		content, err := migrationFS.ReadFile(path.Join("migrations", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %q: %w", entry.Name(), err)
		}

		rawSQL := string(content)
		hash := sha256.Sum256([]byte(rawSQL))
		checksum := hex.EncodeToString(hash[:])

		name := strings.TrimSuffix(parts[1], ".sql")
		migrations = append(migrations, Migration{
			Version:  ver,
			Name:     name,
			SQL:      rawSQL,
			Checksum: checksum,
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// EnsureSchemaMigrationsTable initializes the migration tracking table.
func EnsureSchemaMigrationsTable(ctx context.Context, db *sql.DB) error {
	const query = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    checksum TEXT NOT NULL,
    applied_at TEXT NOT NULL
);`
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}
	return nil
}

// GetAppliedMigrations retrieves all applied migrations mapped by version.
func GetAppliedMigrations(ctx context.Context, db *sql.DB) (map[int]AppliedMigration, error) {
	if err := EnsureSchemaMigrationsTable(ctx, db); err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, "SELECT version, name, checksum, applied_at FROM schema_migrations ORDER BY version ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to query schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]AppliedMigration)
	for rows.Next() {
		var (
			rec     AppliedMigration
			timeStr string
		)
		if err := rows.Scan(&rec.Version, &rec.Name, &rec.Checksum, &timeStr); err != nil {
			return nil, fmt.Errorf("failed to scan schema_migrations row: %w", err)
		}
		t, err := time.Parse(time.RFC3339Nano, timeStr)
		if err != nil {
			// Fallback standard parse
			t, _ = time.Parse(time.RFC3339, timeStr)
		}
		rec.AppliedAt = t
		applied[rec.Version] = rec
	}

	return applied, rows.Err()
}

// RunMigrations executes all unapplied migrations in ascending order.
// If backupFn is provided and there are unapplied migrations on an existing schema,
// it triggers a pre-migration backup.
func RunMigrations(ctx context.Context, db *sql.DB, backupFn func() error) error {
	if err := EnsureSchemaMigrationsTable(ctx, db); err != nil {
		return err
	}

	allMigrations, err := LoadMigrations()
	if err != nil {
		return err
	}

	applied, err := GetAppliedMigrations(ctx, db)
	if err != nil {
		return err
	}

	// Verify existing checksums
	var unapplied []Migration
	for _, m := range allMigrations {
		if rec, exists := applied[m.Version]; exists {
			if rec.Checksum != m.Checksum {
				return fmt.Errorf("migration %03d_%s checksum mismatch: database recorded %s, codebase has %s (possible tampered migration)",
					m.Version, m.Name, rec.Checksum, m.Checksum)
			}
		} else {
			unapplied = append(unapplied, m)
		}
	}

	if len(unapplied) == 0 {
		return nil
	}

	// If there are already applied migrations (upgrading an existing database),
	// create a backup before applying new migrations
	if len(applied) > 0 && backupFn != nil {
		if err := backupFn(); err != nil {
			return fmt.Errorf("failed to create backup before applying migrations: %w", err)
		}
	}

	// Apply pending migrations inside separate transactions
	for _, m := range unapplied {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %03d_%s: %w", m.Version, m.Name, err)
		}

		if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute migration %03d_%s: %w", m.Version, m.Name, err)
		}

		nowStr := time.Now().UTC().Format(time.RFC3339Nano)
		_, err = tx.ExecContext(ctx,
			"INSERT INTO schema_migrations (version, name, checksum, applied_at) VALUES (?, ?, ?, ?)",
			m.Version, m.Name, m.Checksum, nowStr)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration %03d_%s: %w", m.Version, m.Name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %03d_%s: %w", m.Version, m.Name, err)
		}
	}

	return nil
}
