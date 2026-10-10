package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/trustdan/quant-methods-practice/internal/candidates"
)

func (s *Store) CreateCandidate(ctx context.Context, rec *candidates.Record) error {
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO candidate_questions(id,source,status,data_json,created_at,updated_at) VALUES(?,?,?,?,?,?)`, rec.ID, rec.Source.Mode, rec.Status, string(data), rec.CreatedAt.Format(time.RFC3339Nano), rec.CreatedAt.Format(time.RFC3339Nano))
	return err
}
func (s *Store) ListCandidates(ctx context.Context) ([]*candidates.Record, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT data_json FROM candidate_questions ORDER BY created_at DESC,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []*candidates.Record{}
	for rows.Next() {
		var data string
		if err = rows.Scan(&data); err != nil {
			return nil, err
		}
		var rec candidates.Record
		if err = json.Unmarshal([]byte(data), &rec); err != nil {
			return nil, err
		}
		result = append(result, &rec)
	}
	return result, rows.Err()
}
func (s *Store) GetCandidate(ctx context.Context, id string) (*candidates.Record, error) {
	var data string
	err := s.db.QueryRowContext(ctx, `SELECT data_json FROM candidate_questions WHERE id=?`, id).Scan(&data)
	if err != nil {
		return nil, err
	}
	var rec candidates.Record
	if err = json.Unmarshal([]byte(data), &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// ReviewCandidate atomically appends the audit event and changes eligibility using
// the exact preview revision. No attempt, snapshot, or mastery table is touched.
func (s *Store) ReviewCandidate(ctx context.Context, rec *candidates.Record, expectedRevision int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var previous string
	if err = tx.QueryRowContext(ctx, `SELECT data_json FROM candidate_questions WHERE id=?`, rec.ID).Scan(&previous); err != nil {
		return err
	}
	var old candidates.Record
	if err = json.Unmarshal([]byte(previous), &old); err != nil {
		return err
	}
	if old.Revision != expectedRevision || rec.Revision != expectedRevision+1 {
		return candidates.ErrConflict
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	event := rec.Reviews[len(rec.Reviews)-1]
	result, err := tx.ExecContext(ctx, `UPDATE candidate_questions SET status=?,data_json=?,updated_at=? WHERE id=? AND data_json=?`, rec.Status, string(data), event.Timestamp.Format(time.RFC3339Nano), rec.ID, previous)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return candidates.ErrConflict
	}
	// Store complete review provenance in notes; existing schema remains compatible.
	audit, _ := json.Marshal(event)
	_, err = tx.ExecContext(ctx, `INSERT INTO content_approval_events(id,candidate_id,version,reviewer,action,notes,timestamp) VALUES(?,?,?,?,?,?,?)`, uuid.NewString(), rec.ID, rec.Template.Version, event.Reviewer, event.Action, string(audit), event.Timestamp.Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("saving approval audit: %w", err)
	}
	return tx.Commit()
}
