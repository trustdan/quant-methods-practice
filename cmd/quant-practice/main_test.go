package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/storage"
)

func TestLoadDefaultBank(t *testing.T) {
	b, source, err := loadDefaultBank()
	if err != nil {
		t.Fatalf("loadDefaultBank failed: %v", err)
	}

	if b.Count() != 137 {
		t.Errorf("expected 137 approved templates, got %d", b.Count())
	}

	if source == "" {
		t.Errorf("expected non-empty source description")
	}

	// Verify required templates across modules
	for _, id := range []string{
		"binomial_fair_coin_exactly_two",
		"normal_between_package_weight",
		"regression_sales_advertising",
		"twosample_choose_design",
	} {
		if _, ok := b.Get(id); !ok {
			t.Errorf("bank missing expected template: %s", id)
		}
	}
}

func TestExportData(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_export.db")
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := storage.RunMigrations(ctx, db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	store := storage.NewStore(db, nil)
	exportPath := filepath.Join(tempDir, "export.json")

	runExportData(store, exportPath)

	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("reading export file failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshaling export json failed: %v", err)
	}

	if parsed["application"] != "Quant Methods Practice" {
		t.Errorf("expected application 'Quant Methods Practice', got %v", parsed["application"])
	}
	if parsed["version"] != version {
		t.Errorf("expected version %q, got %v", version, parsed["version"])
	}
	if _, ok := parsed["exported_at"]; !ok {
		t.Errorf("export missing 'exported_at' field")
	}
}

func TestRunHighScoresEmpty(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_hs.db")
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := storage.RunMigrations(ctx, db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	store := storage.NewStore(db, nil)
	// Must run without panicking or erroring
	runHighScores(store)
}
