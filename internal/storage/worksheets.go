package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/worksheets"
)

func (s *Store) CreateWorksheet(ctx context.Context, r *worksheets.Record) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO worksheets(id,revision,data_json,updated_at) VALUES(?,?,?,?)`, r.ID, r.Revision, string(b), r.UpdatedAt.Format(time.RFC3339Nano))
	return err
}
func (s *Store) GetWorksheet(ctx context.Context, id string) (*worksheets.Record, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT data_json FROM worksheets WHERE id=?`, id).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var r worksheets.Record
	err = json.Unmarshal([]byte(raw), &r)
	return &r, err
}
func (s *Store) ListWorksheets(ctx context.Context) ([]worksheets.View, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT data_json FROM worksheets ORDER BY updated_at DESC,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []worksheets.View{}
	for rows.Next() {
		var raw string
		if err = rows.Scan(&raw); err != nil {
			return nil, err
		}
		var r worksheets.Record
		if err = json.Unmarshal([]byte(raw), &r); err != nil {
			return nil, err
		}
		result = append(result, worksheets.Public(&r))
	}
	return result, rows.Err()
}

// CommandWorksheet commits the form, all attempts, and command result in one
// transaction. Repeating the same request returns the original durable result.
func (s *Store) CommandWorksheet(ctx context.Context, id string, cmd worksheets.Command) (*worksheets.Record, error) {
	request, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(request)
	hash := hex.EncodeToString(sum[:])
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var result, previousHash string
	err = tx.QueryRowContext(ctx, `SELECT request_hash,result_json FROM worksheet_commands WHERE worksheet_id=? AND command_id=?`, id, cmd.ID).Scan(&previousHash, &result)
	if err == nil {
		if hash != previousHash {
			return nil, worksheets.ErrConflict
		}
		var r worksheets.Record
		err = json.Unmarshal([]byte(result), &r)
		return &r, err
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	var raw string
	err = tx.QueryRowContext(ctx, `SELECT data_json FROM worksheets WHERE id=?`, id).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var old worksheets.Record
	if err = json.Unmarshal([]byte(raw), &old); err != nil {
		return nil, err
	}
	next, err := worksheets.Apply(&old, cmd, s.clock())
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(next)
	if err != nil {
		return nil, err
	}
	res, err := tx.ExecContext(ctx, `UPDATE worksheets SET revision=?,data_json=?,updated_at=? WHERE id=? AND revision=?`, next.Revision, string(b), next.UpdatedAt.Format(time.RFC3339Nano), id, cmd.Revision)
	if err != nil {
		return nil, err
	}
	changed, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if changed != 1 {
		return nil, worksheets.ErrConflict
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO worksheet_commands(worksheet_id,command_id,request_hash,result_json) VALUES(?,?,?,?)`, id, cmd.ID, hash, string(b))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return next, nil
}
