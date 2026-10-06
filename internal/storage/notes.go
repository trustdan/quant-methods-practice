package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ProviderInfo captures metadata regarding the source tutor provider and problem mapping.
type ProviderInfo struct {
	Title          string   `json:"title"`
	Concepts       []string `json:"concepts"`
	Provider       string   `json:"provider"`
	Model          string   `json:"model"`
	Route          string   `json:"route"`
	ThreadID       string   `json:"thread_id,omitempty"`
	ParentNoteID   string   `json:"parent_note_id,omitempty"`
	FallbackLabel  string   `json:"fallback_label,omitempty"`
	AdvisoryStatus string   `json:"advisory_status"`
}

// SavedExplanation represents an authoritative, persistent student note or explanation.
type SavedExplanation struct {
	ID               string       `json:"id"`
	RawMarkdown      string       `json:"raw_markdown"`
	OriginInstanceID string       `json:"origin_instance_id"`
	OriginStageID    string       `json:"origin_stage_id"`
	Topic            string       `json:"topic"`
	ProviderInfo     ProviderInfo `json:"provider_info"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

// TutorDraft captures an unsaved in-progress recovery draft for the AI tutor panel.
type TutorDraft struct {
	ID           string    `json:"id"`
	ContextJSON  string    `json:"context_json"`
	RecoveryText string    `json:"recovery_text"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SaveExplanation persists a learner note to SQLite. It is idempotent by note ID.
func (s *Store) SaveExplanation(ctx context.Context, note *SavedExplanation) error {
	if note == nil {
		return errors.New("note cannot be nil")
	}
	if strings.TrimSpace(note.ID) == "" {
		return errors.New("note ID cannot be empty")
	}

	now := s.clock().UTC()
	if note.CreatedAt.IsZero() {
		note.CreatedAt = now
	}
	note.UpdatedAt = now

	if note.ProviderInfo.AdvisoryStatus == "" {
		note.ProviderInfo.AdvisoryStatus = "Advisory AI explanation: for self-study only, does not affect score or official grades"
	}

	infoBytes, err := json.Marshal(note.ProviderInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal provider info: %w", err)
	}

	const query = `
INSERT INTO saved_explanations (
    id, raw_markdown, origin_instance_id, origin_stage_id, topic, provider_info_json, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    raw_markdown = excluded.raw_markdown,
    topic = excluded.topic,
    provider_info_json = excluded.provider_info_json,
    updated_at = excluded.updated_at;`

	_, err = s.db.ExecContext(ctx, query,
		note.ID,
		note.RawMarkdown,
		note.OriginInstanceID,
		note.OriginStageID,
		note.Topic,
		string(infoBytes),
		note.CreatedAt.Format(time.RFC3339Nano),
		note.UpdatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("failed to save explanation %q: %w", note.ID, err)
	}

	return nil
}

// GetExplanation retrieves a single saved explanation by ID.
func (s *Store) GetExplanation(ctx context.Context, id string) (*SavedExplanation, error) {
	const query = `
SELECT id, raw_markdown, origin_instance_id, origin_stage_id, topic, provider_info_json, created_at, updated_at
FROM saved_explanations
WHERE id = ?;`

	row := s.db.QueryRowContext(ctx, query, id)
	var (
		rec       SavedExplanation
		infoJSON  string
		createdAt string
		updatedAt string
	)

	err := row.Scan(
		&rec.ID,
		&rec.RawMarkdown,
		&rec.OriginInstanceID,
		&rec.OriginStageID,
		&rec.Topic,
		&infoJSON,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get explanation %q: %w", id, err)
	}

	_ = json.Unmarshal([]byte(infoJSON), &rec.ProviderInfo)
	rec.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	rec.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)

	return &rec, nil
}

// ListExplanations returns saved explanations optionally filtered by query and topic.
func (s *Store) ListExplanations(ctx context.Context, queryFilter, topicFilter string) ([]*SavedExplanation, error) {
	const querySQL = `
SELECT id, raw_markdown, origin_instance_id, origin_stage_id, topic, provider_info_json, created_at, updated_at
FROM saved_explanations
ORDER BY updated_at DESC;`

	rows, err := s.db.QueryContext(ctx, querySQL)
	if err != nil {
		return nil, fmt.Errorf("failed to list explanations: %w", err)
	}
	defer rows.Close()

	var results []*SavedExplanation
	qLower := strings.ToLower(strings.TrimSpace(queryFilter))
	topicLower := strings.ToLower(strings.TrimSpace(topicFilter))

	for rows.Next() {
		var (
			rec       SavedExplanation
			infoJSON  string
			createdAt string
			updatedAt string
		)

		if err := rows.Scan(
			&rec.ID,
			&rec.RawMarkdown,
			&rec.OriginInstanceID,
			&rec.OriginStageID,
			&rec.Topic,
			&infoJSON,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan explanation row: %w", err)
		}

		_ = json.Unmarshal([]byte(infoJSON), &rec.ProviderInfo)
		rec.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		rec.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)

		// Filter by topic if specified
		if topicLower != "" && !strings.Contains(strings.ToLower(rec.Topic), topicLower) {
			continue
		}

		// Filter by search query if specified (matches title, topic, concepts, or raw markdown)
		if qLower != "" {
			matched := false
			if strings.Contains(strings.ToLower(rec.Topic), qLower) ||
				strings.Contains(strings.ToLower(rec.RawMarkdown), qLower) ||
				strings.Contains(strings.ToLower(rec.ProviderInfo.Title), qLower) {
				matched = true
			} else {
				for _, c := range rec.ProviderInfo.Concepts {
					if strings.Contains(strings.ToLower(c), qLower) {
						matched = true
						break
					}
				}
			}
			if !matched {
				continue
			}
		}

		results = append(results, &rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating explanations: %w", err)
	}

	return results, nil
}

// DeleteExplanation removes a saved explanation by ID.
func (s *Store) DeleteExplanation(ctx context.Context, id string) error {
	const query = `DELETE FROM saved_explanations WHERE id = ?;`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete explanation %q: %w", id, err)
	}
	return nil
}

// SaveTutorDraft stores an unsaved in-progress recovery draft.
func (s *Store) SaveTutorDraft(ctx context.Context, draft *TutorDraft) error {
	if draft == nil {
		return errors.New("draft cannot be nil")
	}
	if strings.TrimSpace(draft.ID) == "" {
		return errors.New("draft ID cannot be empty")
	}

	now := s.clock().UTC()
	draft.UpdatedAt = now

	const query = `
INSERT INTO tutor_drafts (id, context_json, recovery_text, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    context_json = excluded.context_json,
    recovery_text = excluded.recovery_text,
    updated_at = excluded.updated_at;`

	_, err := s.db.ExecContext(ctx, query,
		draft.ID,
		draft.ContextJSON,
		draft.RecoveryText,
		draft.UpdatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("failed to save tutor draft %q: %w", draft.ID, err)
	}
	return nil
}

// GetTutorDraft retrieves an unsaved recovery draft by ID.
func (s *Store) GetTutorDraft(ctx context.Context, id string) (*TutorDraft, error) {
	const query = `SELECT id, context_json, recovery_text, updated_at FROM tutor_drafts WHERE id = ?;`
	row := s.db.QueryRowContext(ctx, query, id)

	var (
		draft     TutorDraft
		updatedAt string
	)
	err := row.Scan(&draft.ID, &draft.ContextJSON, &draft.RecoveryText, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get tutor draft %q: %w", id, err)
	}

	draft.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return &draft, nil
}

// ClearTutorDraft removes an unsaved recovery draft once saved or discarded.
func (s *Store) ClearTutorDraft(ctx context.Context, id string) error {
	const query = `DELETE FROM tutor_drafts WHERE id = ?;`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to clear tutor draft %q: %w", id, err)
	}
	return nil
}
