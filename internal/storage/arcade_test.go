package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestArcadeScoreStorage(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_arcade.db")

	ctx := context.Background()
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()

	if err = RunMigrations(ctx, db, nil); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	store := NewStore(db, nil)

	// Initial high score should be 0
	highScore, err := store.GetHighScore(ctx)
	if err != nil {
		t.Fatalf("unexpected error getting initial high score: %v", err)
	}
	if highScore != 0 {
		t.Errorf("expected 0 initial high score, got %d", highScore)
	}

	// Demo scores should be rejected
	demoRec := ArcadeScoreRecord{
		GameVersion:     "0.1.0",
		Seed:            123,
		IsDemo:          true,
		DurationSeconds: 45.2,
		SectorsCleared:  1,
		Score:           1500,
		Initials:        "BOT",
	}
	if err := store.RecordArcadeScore(ctx, demoRec); err == nil {
		t.Errorf("expected error recording demo score, got nil")
	}

	// Record manual scores
	rec1 := ArcadeScoreRecord{
		GameVersion:     "0.1.0",
		Seed:            42,
		IsDemo:          false,
		DurationSeconds: 120.5,
		SectorsCleared:  3,
		Score:           4500,
		Initials:        "ACE",
		CreatedAt:       time.Now().UTC(),
	}
	if err := store.RecordArcadeScore(ctx, rec1); err != nil {
		t.Fatalf("failed to record score 1: %v", err)
	}

	rec2 := ArcadeScoreRecord{
		GameVersion:     "0.1.0",
		Seed:            84,
		IsDemo:          false,
		DurationSeconds: 210.0,
		SectorsCleared:  4,
		Score:           12800,
		Initials:        "TOP",
		CreatedAt:       time.Now().UTC(),
	}
	if err := store.RecordArcadeScore(ctx, rec2); err != nil {
		t.Fatalf("failed to record score 2: %v", err)
	}

	// Verify high score is 12800
	highScore, err = store.GetHighScore(ctx)
	if err != nil {
		t.Fatalf("failed to get high score: %v", err)
	}
	if highScore != 12800 {
		t.Errorf("expected high score 12800, got %d", highScore)
	}

	// Verify top scores ordering
	top, err := store.GetTopArcadeScores(ctx, 10)
	if err != nil {
		t.Fatalf("failed to get top scores: %v", err)
	}
	if len(top) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(top))
	}
	if top[0].Initials != "TOP" || top[0].Score != 12800 {
		t.Errorf("expected rank 1 to be TOP (12800), got %s (%d)", top[0].Initials, top[0].Score)
	}
	if top[1].Initials != "ACE" || top[1].Score != 4500 {
		t.Errorf("expected rank 2 to be ACE (4500), got %s (%d)", top[1].Initials, top[1].Score)
	}
}
