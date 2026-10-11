package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/trustdan/quant-methods-practice/internal/exam"
)

// HistoryItem represents an immutable record in the paginated activity history.
type HistoryItem struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"` // "drill" or "exam"
	Mode         string    `json:"mode"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	ScorePercent *float64  `json:"score_percent,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Completed    bool      `json:"completed"`
}

// CreateExam persists an initialized assessment, its frozen question snapshots, and deadline.
func (s *Store) CreateExam(ctx context.Context, rec *exam.ExamRecord) error {
	if rec == nil {
		return errors.New("exam record cannot be nil")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	dataJSON, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshaling exam data: %w", err)
	}

	var deadlineStr *string
	if rec.DeadlineAt != nil {
		ds := rec.DeadlineAt.UTC().Format(time.RFC3339Nano)
		deadlineStr = &ds
	}

	completedInt := 0
	if rec.Status != exam.StatusActive {
		completedInt = 1
	}

	firstTmplID := "exam_multi"
	if len(rec.Questions) > 0 {
		firstTmplID = rec.Questions[0].TemplateID
	}

	const insertSession = `
INSERT INTO sessions (
    id, type, mode, template_id, template_version, seed, revision,
    current_question_index, current_stage_index, completed,
    settings_json, created_at, updated_at, exam_deadline_at
) VALUES (?, 'exam', 'assessment', ?, 1, ?, ?, ?, 0, ?, ?, ?, ?, ?);`

	createdAtStr := rec.CreatedAt.UTC().Format(time.RFC3339Nano)
	updatedAtStr := rec.UpdatedAt.UTC().Format(time.RFC3339Nano)

	_, err = tx.ExecContext(ctx, insertSession,
		rec.ID,
		firstTmplID,
		rec.Seed,
		rec.Revision,
		rec.CurrentQuestionIndex,
		completedInt,
		string(dataJSON),
		createdAtStr,
		updatedAtStr,
		deadlineStr,
	)
	if err != nil {
		return fmt.Errorf("inserting exam session: %w", err)
	}

	const insertInstance = `
INSERT OR IGNORE INTO question_instances (
    id, session_id, template_id, template_version, seed,
    parameters_json, title, scenario_markdown, stages_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	for _, q := range rec.Questions {
		paramsJSON, _ := json.Marshal(q.Parameters)
		stagesJSON, _ := json.Marshal(q.Stages)

		_, err = tx.ExecContext(ctx, insertInstance,
			q.ID,
			rec.ID,
			q.TemplateID,
			q.TemplateVersion,
			q.Seed,
			string(paramsJSON),
			q.Title,
			q.ScenarioMarkdown,
			string(stagesJSON),
			createdAtStr,
		)
		if err != nil {
			return fmt.Errorf("inserting exam question instance snapshot: %w", err)
		}
	}

	return tx.Commit()
}

// GetExam retrieves an authoritative exam record by ID.
// If an active timed exam has exceeded its deadline, it is dynamically marked expired.
func (s *Store) GetExam(ctx context.Context, id string) (*exam.ExamRecord, error) {
	const query = `
SELECT settings_json, exam_deadline_at
FROM sessions
WHERE id = ? AND type = 'exam';`

	var settingsJSON string
	var deadlineStr sql.NullString
	err := s.db.QueryRowContext(ctx, query, id).Scan(&settingsJSON, &deadlineStr)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var rec exam.ExamRecord
	if err := json.Unmarshal([]byte(settingsJSON), &rec); err != nil {
		return nil, fmt.Errorf("unmarshaling exam record: %w", err)
	}

	// Verify questions snapshot presence
	if len(rec.Questions) == 0 {
		return nil, fmt.Errorf("exam %q has no question snapshots", id)
	}

	// Check deadline against storage clock
	now := s.clock().UTC()
	if rec.Status == exam.StatusActive && rec.DeadlineAt != nil && now.After(*rec.DeadlineAt) {
		rec.Status = exam.StatusExpired
		rec.CompletedAt = rec.DeadlineAt
		rec.Report = exam.EvaluateExam(&rec, rec.DeadlineAt)
	}

	return &rec, nil
}

// ListExams returns summaries of all saved assessments sorted newest first.
func (s *Store) ListExams(ctx context.Context) ([]exam.ExamSummary, error) {
	const query = `
SELECT settings_json
FROM sessions
WHERE type = 'exam'
ORDER BY created_at DESC;`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]exam.ExamSummary, 0)
	now := s.clock().UTC()

	for rows.Next() {
		var settingsJSON string
		if err := rows.Scan(&settingsJSON); err != nil {
			return nil, err
		}

		var rec exam.ExamRecord
		if err := json.Unmarshal([]byte(settingsJSON), &rec); err != nil {
			continue
		}

		if rec.Status == exam.StatusActive && rec.DeadlineAt != nil && now.After(*rec.DeadlineAt) {
			rec.Status = exam.StatusExpired
			rec.Report = exam.EvaluateExam(&rec, rec.DeadlineAt)
		}

		result = append(result, exam.Summary(&rec))
	}

	return result, rows.Err()
}

// CommandExam executes an assessment command within an atomic transaction.
// Repeated requests with the same command ID return the idempotent result.
func (s *Store) CommandExam(ctx context.Context, id string, cmd exam.ExamCommand) (*exam.ExamRecord, error) {
	reqBytes, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(reqBytes)
	hash := hex.EncodeToString(sum[:])

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Check idempotency cache
	var previousResult string
	err = tx.QueryRowContext(ctx, `
SELECT result_json FROM command_idempotency
WHERE session_id = ? AND command_id = ?;`, id, cmd.CommandID).Scan(&previousResult)
	if err == nil {
		var cached exam.ExamRecord
		if err := json.Unmarshal([]byte(previousResult), &cached); err == nil {
			return &cached, nil
		}
	} else if err != sql.ErrNoRows {
		return nil, err
	}

	// 2. Fetch current record
	var currentJSON string
	err = tx.QueryRowContext(ctx, `
SELECT settings_json FROM sessions
WHERE id = ? AND type = 'exam';`, id).Scan(&currentJSON)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var old exam.ExamRecord
	if err := json.Unmarshal([]byte(currentJSON), &old); err != nil {
		return nil, err
	}

	// 3. Apply command
	next, err := exam.Apply(&old, cmd, s.clock())
	if err != nil {
		return nil, err
	}

	nextBytes, err := json.Marshal(next)
	if err != nil {
		return nil, err
	}

	completedInt := 0
	if next.Status != exam.StatusActive {
		completedInt = 1
	}

	// 4. Update session row with optimistic revision concurrency check
	res, err := tx.ExecContext(ctx, `
UPDATE sessions SET
    revision = ?,
    current_question_index = ?,
    completed = ?,
    settings_json = ?,
    updated_at = ?
WHERE id = ? AND revision = ?;`,
		next.Revision,
		next.CurrentQuestionIndex,
		completedInt,
		string(nextBytes),
		next.UpdatedAt.UTC().Format(time.RFC3339Nano),
		id,
		cmd.ExpectedRevision,
	)
	if err != nil {
		return nil, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected != 1 {
		return nil, exam.ErrConflict
	}

	// 5. Persist submitted response in exam_responses table if applicable
	if cmd.Type == "save_answer" && cmd.Answer != nil && len(next.Questions) > 0 {
		qIdx := next.CurrentQuestionIndex
		if cmd.QuestionIndex != nil {
			qIdx = *cmd.QuestionIndex
		}
		if qIdx >= 0 && qIdx < len(next.Questions) {
			qInst := next.Questions[qIdx]
			ansJSON, _ := json.Marshal(cmd.Answer)
			respID := fmt.Sprintf("resp_%s_%s_%s", id, qInst.ID, cmd.StageID)

			const upsertResp = `
INSERT INTO exam_responses (
    id, session_id, instance_id, stage_id, order_index, response_json, status, submitted_at
) VALUES (?, ?, ?, ?, ?, ?, 'saved', ?)
ON CONFLICT(id) DO UPDATE SET
    response_json = excluded.response_json,
    submitted_at = excluded.submitted_at;`

			_, _ = tx.ExecContext(ctx, upsertResp,
				respID,
				id,
				qInst.ID,
				cmd.StageID,
				qIdx,
				string(ansJSON),
				next.UpdatedAt.UTC().Format(time.RFC3339Nano),
			)
		}
	}

	// 6. Record in command_idempotency table
	_, err = tx.ExecContext(ctx, `
INSERT INTO command_idempotency (
    command_id, session_id, command_type, result_json, created_at
) VALUES (?, ?, ?, ?, ?);`,
		cmd.CommandID,
		id,
		cmd.Type+"#"+hash[:8],
		string(nextBytes),
		next.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return nil, fmt.Errorf("persisting command idempotency: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return next, nil
}

// GetExamReport returns the full evaluation report for an exam.
// If the exam is currently active and unexpired, an error is returned.
func (s *Store) GetExamReport(ctx context.Context, id string) (*exam.ExamReport, error) {
	rec, err := s.GetExam(ctx, id)
	if err != nil {
		return nil, err
	}

	if rec.Status == exam.StatusActive {
		return nil, errors.New("assessment is currently active; results and answers are withheld")
	}

	if rec.Report != nil {
		return rec.Report, nil
	}

	return exam.EvaluateExam(rec, rec.CompletedAt), nil
}

// GetHistory returns a paginated list of all practice and exam sessions.
func (s *Store) GetHistory(ctx context.Context, limit, offset int) ([]HistoryItem, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	const query = `
SELECT id, type, mode, completed, settings_json, created_at, updated_at
FROM sessions
ORDER BY updated_at DESC
LIMIT ? OFFSET ?;`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []HistoryItem
	for rows.Next() {
		var (
			id           string
			sessType     string
			mode         string
			completedInt int
			settingsJSON string
			createdStr   string
			updatedStr   string
		)

		if err := rows.Scan(&id, &sessType, &mode, &completedInt, &settingsJSON, &createdStr, &updatedStr); err != nil {
			return nil, err
		}

		cTime, _ := time.Parse(time.RFC3339Nano, createdStr)
		uTime, _ := time.Parse(time.RFC3339Nano, updatedStr)

		title := "Practice Session"
		var scorePct *float64
		status := "in_progress"
		if completedInt == 1 {
			status = "completed"
		}

		if sessType == "exam" {
			var rec exam.ExamRecord
			if err := json.Unmarshal([]byte(settingsJSON), &rec); err == nil {
				title = rec.Title
				status = string(rec.Status)
				if rec.Report != nil {
					p := rec.Report.ScorePercent
					scorePct = &p
				}
			}
		}

		items = append(items, HistoryItem{
			ID:           id,
			Type:         sessType,
			Mode:         mode,
			Title:        title,
			Status:       status,
			ScorePercent: scorePct,
			CreatedAt:    cTime,
			UpdatedAt:    uTime,
			Completed:    completedInt == 1,
		})
	}

	return items, rows.Err()
}

// GenerateExamID returns a unique identifier for an assessment session.
func GenerateExamID() string {
	return "exam_" + uuid.NewString()[:8]
}
