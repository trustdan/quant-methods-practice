package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// TutorThread represents a conversational interaction thread attached to a problem/stage.
type TutorThread struct {
	ID         string         `json:"id"`
	SessionID  string         `json:"session_id,omitempty"`
	InstanceID string         `json:"instance_id"`
	StageID    string         `json:"stage_id"`
	Provider   string         `json:"provider"`
	Model      string         `json:"model"`
	Title      string         `json:"title"`
	Messages   []TutorMessage `json:"messages,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// TutorMessage represents an individual turn (user or assistant) within a TutorThread.
type TutorMessage struct {
	ID            string    `json:"id"`
	ThreadID      string    `json:"thread_id"`
	Role          string    `json:"role"` // "user", "assistant", "system"
	Content       string    `json:"content"`
	FallbackLabel string    `json:"fallback_label,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreateTutorThread persists a new conversation thread in SQLite.
func (s *Store) CreateTutorThread(ctx context.Context, thread *TutorThread) error {
	if thread == nil {
		return errors.New("thread cannot be nil")
	}
	if strings.TrimSpace(thread.ID) == "" {
		return errors.New("thread ID cannot be empty")
	}
	if strings.TrimSpace(thread.InstanceID) == "" {
		return errors.New("instance ID cannot be empty")
	}
	if strings.TrimSpace(thread.StageID) == "" {
		return errors.New("stage ID cannot be empty")
	}

	now := s.clock().UTC()
	if thread.CreatedAt.IsZero() {
		thread.CreatedAt = now
	}
	thread.UpdatedAt = now

	const query = `
INSERT INTO tutor_threads (
    id, session_id, instance_id, stage_id, provider, model, title, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    title = excluded.title,
    provider = excluded.provider,
    model = excluded.model,
    updated_at = excluded.updated_at;`

	_, err := s.db.ExecContext(ctx, query,
		thread.ID,
		thread.SessionID,
		thread.InstanceID,
		thread.StageID,
		thread.Provider,
		thread.Model,
		thread.Title,
		thread.CreatedAt.Format(time.RFC3339Nano),
		thread.UpdatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("failed to create tutor thread %q: %w", thread.ID, err)
	}

	return nil
}

// GetTutorThread retrieves a conversation thread and all its messages in chronological order.
func (s *Store) GetTutorThread(ctx context.Context, id string) (*TutorThread, error) {
	const query = `
SELECT id, session_id, instance_id, stage_id, provider, model, title, created_at, updated_at
FROM tutor_threads
WHERE id = ?;`

	row := s.db.QueryRowContext(ctx, query, id)
	var (
		thread    TutorThread
		sessID    sql.NullString
		createdAt string
		updatedAt string
	)

	err := row.Scan(
		&thread.ID,
		&sessID,
		&thread.InstanceID,
		&thread.StageID,
		&thread.Provider,
		&thread.Model,
		&thread.Title,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get tutor thread %q: %w", id, err)
	}

	if sessID.Valid {
		thread.SessionID = sessID.String
	}
	thread.CreatedAt = parseDBTime(createdAt)
	thread.UpdatedAt = parseDBTime(updatedAt)

	// Fetch messages for this thread
	messages, err := s.GetTutorMessages(ctx, id)
	if err != nil {
		return nil, err
	}
	thread.Messages = make([]TutorMessage, len(messages))
	for i, m := range messages {
		thread.Messages[i] = *m
	}

	return &thread, nil
}

// ListTutorThreads retrieves threads matching optional instance and stage filters, ordered by most recently updated.
func (s *Store) ListTutorThreads(ctx context.Context, instanceID, stageID string) ([]*TutorThread, error) {
	var (
		query strings.Builder
		args  []any
	)

	query.WriteString(`
SELECT id, session_id, instance_id, stage_id, provider, model, title, created_at, updated_at
FROM tutor_threads`)

	var whereClauses []string
	if instanceID != "" {
		whereClauses = append(whereClauses, "instance_id = ?")
		args = append(args, instanceID)
	}
	if stageID != "" {
		whereClauses = append(whereClauses, "stage_id = ?")
		args = append(args, stageID)
	}

	if len(whereClauses) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(whereClauses, " AND "))
	}

	query.WriteString(" ORDER BY updated_at DESC;")

	rows, err := s.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tutor threads: %w", err)
	}
	defer rows.Close()

	var threads []*TutorThread
	for rows.Next() {
		var (
			th        TutorThread
			sessID    sql.NullString
			createdAt string
			updatedAt string
		)
		if err := rows.Scan(
			&th.ID,
			&sessID,
			&th.InstanceID,
			&th.StageID,
			&th.Provider,
			&th.Model,
			&th.Title,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan tutor thread: %w", err)
		}
		if sessID.Valid {
			th.SessionID = sessID.String
		}
		th.CreatedAt = parseDBTime(createdAt)
		th.UpdatedAt = parseDBTime(updatedAt)
		threads = append(threads, &th)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading tutor threads: %w", err)
	}

	return threads, nil
}

// DeleteTutorThread removes a thread and all associated messages.
func (s *Store) DeleteTutorThread(ctx context.Context, id string) error {
	// Delete messages first
	_, _ = s.db.ExecContext(ctx, "DELETE FROM tutor_messages WHERE thread_id = ?", id)

	res, err := s.db.ExecContext(ctx, "DELETE FROM tutor_threads WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete tutor thread %q: %w", id, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read affected rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// AppendTutorMessage appends a new message turn to a thread and touches the thread's updated_at timestamp.
func (s *Store) AppendTutorMessage(ctx context.Context, msg *TutorMessage) error {
	if msg == nil {
		return errors.New("message cannot be nil")
	}
	if strings.TrimSpace(msg.ID) == "" {
		return errors.New("message ID cannot be empty")
	}
	if strings.TrimSpace(msg.ThreadID) == "" {
		return errors.New("thread ID cannot be empty")
	}

	now := s.clock().UTC()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	const insertMsg = `
INSERT INTO tutor_messages (
    id, thread_id, role, content, fallback_label, created_at
) VALUES (?, ?, ?, ?, ?, ?);`

	_, err = tx.ExecContext(ctx, insertMsg,
		msg.ID,
		msg.ThreadID,
		msg.Role,
		msg.Content,
		msg.FallbackLabel,
		msg.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("failed to append tutor message %q: %w", msg.ID, err)
	}

	// Update parent thread's updated_at
	const touchThread = `UPDATE tutor_threads SET updated_at = ? WHERE id = ?;`
	_, err = tx.ExecContext(ctx, touchThread, now.Format(time.RFC3339Nano), msg.ThreadID)
	if err != nil {
		return fmt.Errorf("failed to touch tutor thread: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tutor message transaction: %w", err)
	}

	return nil
}

// GetTutorMessages retrieves all messages for a thread ordered chronologically.
func (s *Store) GetTutorMessages(ctx context.Context, threadID string) ([]*TutorMessage, error) {
	const query = `
SELECT id, thread_id, role, content, fallback_label, created_at
FROM tutor_messages
WHERE thread_id = ?
ORDER BY created_at ASC;`

	rows, err := s.db.QueryContext(ctx, query, threadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages for thread %q: %w", threadID, err)
	}
	defer rows.Close()

	var messages []*TutorMessage
	for rows.Next() {
		var (
			msg           TutorMessage
			fallbackLabel sql.NullString
			createdAt     string
		)
		if err := rows.Scan(
			&msg.ID,
			&msg.ThreadID,
			&msg.Role,
			&msg.Content,
			&fallbackLabel,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan tutor message: %w", err)
		}
		if fallbackLabel.Valid {
			msg.FallbackLabel = fallbackLabel.String
		}
		msg.CreatedAt = parseDBTime(createdAt)
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading tutor messages: %w", err)
	}

	return messages, nil
}

func parseDBTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return time.Time{}
}
