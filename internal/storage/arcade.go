package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ArcadeScoreRecord represents an immutable persistent arcade game score record.
type ArcadeScoreRecord struct {
	ID              string    `json:"id"`
	GameVersion     string    `json:"game_version"`
	Seed            int64     `json:"seed"`
	IsDemo          bool      `json:"is_demo"`
	DurationSeconds float64   `json:"duration_seconds"`
	SectorsCleared  int       `json:"sectors_cleared"`
	Score           int       `json:"score"`
	Initials        string    `json:"initials"`
	CreatedAt       time.Time `json:"created_at"`
}

// RecordArcadeScore records a completed manual arcade game score in SQLite.
// Demo scores are rejected from persistent records.
func (s *Store) RecordArcadeScore(ctx context.Context, rec ArcadeScoreRecord) error {
	if rec.IsDemo {
		return fmt.Errorf("demo scores cannot be recorded in persistent leaderboard")
	}
	if rec.Score < 0 {
		return fmt.Errorf("score cannot be negative")
	}
	if rec.ID == "" {
		rec.ID = uuid.NewString()
	}
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = time.Now().UTC()
	}
	if rec.Initials == "" {
		rec.Initials = "PIL"
	}
	if len(rec.Initials) > 3 {
		rec.Initials = rec.Initials[:3]
	}

	query := `
		INSERT INTO arcade_scores (
			id, game_version, seed, is_demo, duration_seconds, sectors_cleared, score, initials, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	isDemoInt := 0
	if rec.IsDemo {
		isDemoInt = 1
	}

	_, err := s.db.ExecContext(
		ctx,
		query,
		rec.ID,
		rec.GameVersion,
		rec.Seed,
		isDemoInt,
		rec.DurationSeconds,
		rec.SectorsCleared,
		rec.Score,
		rec.Initials,
		rec.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to insert arcade score: %w", err)
	}
	return nil
}

// GetTopArcadeScores retrieves the top scores from the persistent leaderboard.
func (s *Store) GetTopArcadeScores(ctx context.Context, limit int) ([]ArcadeScoreRecord, error) {
	if limit <= 0 {
		limit = 10
	}
	query := `
		SELECT id, game_version, seed, is_demo, duration_seconds, sectors_cleared, score, initials, created_at
		FROM arcade_scores
		WHERE is_demo = 0
		ORDER BY score DESC, duration_seconds ASC, created_at DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top arcade scores: %w", err)
	}
	defer rows.Close()

	var records []ArcadeScoreRecord
	for rows.Next() {
		var rec ArcadeScoreRecord
		var isDemoInt int
		var createdAtStr string

		if err := rows.Scan(
			&rec.ID,
			&rec.GameVersion,
			&rec.Seed,
			&isDemoInt,
			&rec.DurationSeconds,
			&rec.SectorsCleared,
			&rec.Score,
			&rec.Initials,
			&createdAtStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan arcade score row: %w", err)
		}

		rec.IsDemo = (isDemoInt == 1)
		t, err := time.Parse(time.RFC3339, createdAtStr)
		if err == nil {
			rec.CreatedAt = t
		}
		records = append(records, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// GetHighScore retrieves the all-time highest manual arcade score.
func (s *Store) GetHighScore(ctx context.Context) (int, error) {
	query := `SELECT COALESCE(MAX(score), 0) FROM arcade_scores WHERE is_demo = 0`
	var highScore int
	err := s.db.QueryRowContext(ctx, query).Scan(&highScore)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query high score: %w", err)
	}
	return highScore, nil
}
