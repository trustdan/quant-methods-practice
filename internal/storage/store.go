package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/drill"
)

var ErrNotFound = errors.New("record not found")

// Store provides transactional SQLite persistence for drill sessions, snapshots,
// attempts, assistance, drafts, and command idempotency.
type Store struct {
	db    *sql.DB
	clock func() time.Time
}

// NewStore constructs a Store backed by the provided SQLite database connection.
func NewStore(db *sql.DB, clock func() time.Time) *Store {
	if clock == nil {
		clock = time.Now
	}
	return &Store{
		db:    db,
		clock: clock,
	}
}

// DB returns the underlying database handle.
func (s *Store) DB() *sql.DB {
	return s.db
}

type sessionMetadata struct {
	Settings       domain.SessionSettings `json:"settings"`
	QuestionStates []drill.QuestionState  `json:"question_states,omitempty"`
}

func scopedStageKey(instanceID, stageID string, multi bool) string {
	if multi {
		return fmt.Sprintf("%s:%s", instanceID, stageID)
	}
	return stageID
}

// SaveSession atomically persists a DrillSession, its immutable QuestionInstance snapshot(s),
// current stage states, attempts, and assistance records.
func (s *Store) SaveSession(ctx context.Context, session *drill.DrillSession) error {
	if session == nil {
		return errors.New("session cannot be nil")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 1. Prepare questions to persist
	var questionsToSave []drill.QuestionState
	isMulti := len(session.Questions) > 1
	if len(session.Questions) > 0 {
		session.SyncCurrentQuestion()
		questionsToSave = session.Questions
	} else {
		questionsToSave = []drill.QuestionState{
			{
				Index:             0,
				TemplateID:        session.TemplateID,
				TemplateVersion:   session.TemplateVersion,
				QuestionInstance:  session.QuestionInstance,
				Stages:            session.Stages,
				CurrentStageIndex: session.CurrentStageIndex,
				Completed:         session.Completed,
			},
		}
	}

	meta := sessionMetadata{
		Settings: session.Settings,
	}
	if len(session.Questions) > 0 {
		meta.QuestionStates = session.Questions
	}
	settingsJSONBytes, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal settings json: %w", err)
	}
	settingsJSON := string(settingsJSONBytes)

	// 2. Upsert session row
	completedInt := 0
	if session.Completed {
		completedInt = 1
	}

	const upsertSession = `
INSERT INTO sessions (
    id, type, mode, template_id, template_version, seed, revision,
    current_question_index, current_stage_index, completed,
    settings_json, created_at, updated_at
) VALUES (?, 'drill', 'practice', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    revision = excluded.revision,
    current_question_index = excluded.current_question_index,
    current_stage_index = excluded.current_stage_index,
    completed = excluded.completed,
    settings_json = excluded.settings_json,
    updated_at = excluded.updated_at;`

	createdAtStr := session.CreatedAt.UTC().Format(time.RFC3339Nano)
	updatedAtStr := session.UpdatedAt.UTC().Format(time.RFC3339Nano)

	_, err = tx.ExecContext(ctx, upsertSession,
		session.ID,
		session.TemplateID,
		session.TemplateVersion,
		session.Seed,
		session.Revision,
		session.CurrentQuestionIndex,
		session.CurrentStageIndex,
		completedInt,
		settingsJSON,
		createdAtStr,
		updatedAtStr,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert session %q: %w", session.ID, err)
	}

	const insertInstance = `
INSERT OR IGNORE INTO question_instances (
    id, session_id, template_id, template_version, seed,
    parameters_json, title, scenario_markdown, stages_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	const upsertStageState = `
INSERT INTO drill_stage_states (
    session_id, stage_index, stage_id, status,
    first_try_correct, solved_on_retry, revealed,
    active_hint, last_feedback, misconception_id, invalid_input_notice
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(session_id, stage_id) DO UPDATE SET
    stage_index = excluded.stage_index,
    status = excluded.status,
    first_try_correct = excluded.first_try_correct,
    solved_on_retry = excluded.solved_on_retry,
    revealed = excluded.revealed,
    active_hint = excluded.active_hint,
    last_feedback = excluded.last_feedback,
    misconception_id = excluded.misconception_id,
    invalid_input_notice = excluded.invalid_input_notice;`

	const insertAttempt = `
INSERT OR IGNORE INTO attempts (
    id, session_id, instance_id, stage_id, attempt_number,
    command_id, submitted_answer_json, assistance_json,
    is_correct, feedback_markdown, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	const insertAssistance = `
INSERT OR IGNORE INTO assistance_events (
    id, session_id, instance_id, stage_id, event_type, scope, created_at
) VALUES (?, ?, ?, ?, ?, 'stage', ?);`

	// 3. Persist question instances, stage states, attempts, and assistance
	for _, q := range questionsToSave {
		paramsJSON, err := json.Marshal(q.QuestionInstance.Parameters)
		if err != nil {
			return fmt.Errorf("failed to marshal instance parameters: %w", err)
		}

		stagesJSON, err := json.Marshal(q.QuestionInstance.Stages)
		if err != nil {
			return fmt.Errorf("failed to marshal instance stages: %w", err)
		}

		_, err = tx.ExecContext(ctx, insertInstance,
			q.QuestionInstance.ID,
			session.ID,
			q.QuestionInstance.TemplateID,
			q.QuestionInstance.TemplateVersion,
			q.QuestionInstance.Seed,
			string(paramsJSON),
			q.QuestionInstance.Title,
			q.QuestionInstance.ScenarioMarkdown,
			string(stagesJSON),
			createdAtStr,
		)
		if err != nil {
			return fmt.Errorf("failed to insert question instance snapshot %q: %w", q.QuestionInstance.ID, err)
		}

		for i, st := range q.Stages {
			firstTryInt := 0
			if st.FirstTryCorrect {
				firstTryInt = 1
			}
			solvedOnRetryInt := 0
			if st.SolvedOnRetry {
				solvedOnRetryInt = 1
			}
			revealedInt := 0
			if st.Revealed {
				revealedInt = 1
			}

			stKey := scopedStageKey(q.QuestionInstance.ID, st.Instance.ID, isMulti)

			_, err = tx.ExecContext(ctx, upsertStageState,
				session.ID,
				i,
				stKey,
				string(st.Status),
				firstTryInt,
				solvedOnRetryInt,
				revealedInt,
				st.ActiveHint,
				st.LastFeedback,
				st.MisconceptionID,
				st.InvalidInputNotice,
			)
			if err != nil {
				return fmt.Errorf("failed to upsert stage state %q: %w", stKey, err)
			}

			// Insert any attempts
			for _, att := range st.Attempts {
				ansJSON, err := json.Marshal(att.SubmittedAnswer)
				if err != nil {
					return fmt.Errorf("failed to marshal attempt answer: %w", err)
				}
				asstJSON, err := json.Marshal(att.Assistance)
				if err != nil {
					return fmt.Errorf("failed to marshal attempt assistance: %w", err)
				}

				attCorrectInt := 0
				if att.IsCorrect {
					attCorrectInt = 1
				}

				attCreatedStr := att.CreatedAt.UTC().Format(time.RFC3339Nano)
				_, err = tx.ExecContext(ctx, insertAttempt,
					att.ID,
					session.ID,
					q.QuestionInstance.ID,
					st.Instance.ID,
					att.AttemptNumber,
					nil,
					string(ansJSON),
					string(asstJSON),
					attCorrectInt,
					att.FeedbackMarkdown,
					attCreatedStr,
				)
				if err != nil {
					return fmt.Errorf("failed to insert attempt %q: %w", att.ID, err)
				}
			}

			// Insert assistance events
			for _, asst := range st.Assistance {
				asstID := fmt.Sprintf("asst_%s_%s_%s", session.ID, stKey, string(asst))
				_, err = tx.ExecContext(ctx, insertAssistance,
					asstID,
					session.ID,
					q.QuestionInstance.ID,
					st.Instance.ID,
					string(asst),
					updatedAtStr,
				)
				if err != nil {
					return fmt.Errorf("failed to insert assistance event: %w", err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit session transaction: %w", err)
	}

	return nil
}

// GetSession reconstructs an authoritative DrillSession from the stored immutable snapshot and attempts.
func (s *Store) GetSession(ctx context.Context, id string) (*drill.DrillSession, error) {
	if id == "" {
		return nil, ErrNotFound
	}

	// 1. Fetch session row
	const querySession = `
SELECT id, template_id, template_version, seed, revision,
       current_question_index, current_stage_index, completed,
       settings_json, created_at, updated_at
FROM sessions
WHERE id = ?;`

	var (
		sessID            string
		tmplID            string
		tmplVer           int
		seed              int64
		revision          int64
		currentQIndex     int
		currentStageIndex int
		completedInt      int
		settingsJSON      string
		createdAtStr      string
		updatedAtStr      string
	)

	err := s.db.QueryRowContext(ctx, querySession, id).Scan(
		&sessID, &tmplID, &tmplVer, &seed, &revision,
		&currentQIndex, &currentStageIndex, &completedInt,
		&settingsJSON, &createdAtStr, &updatedAtStr,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query session %q: %w", id, err)
	}

	createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)
	updatedAt, _ := time.Parse(time.RFC3339Nano, updatedAtStr)

	var meta sessionMetadata
	_ = json.Unmarshal([]byte(settingsJSON), &meta)

	// 2. Fetch all QuestionInstance snapshots for session
	const queryInstances = `
SELECT id, template_id, template_version, seed, parameters_json,
       title, scenario_markdown, stages_json
FROM question_instances
WHERE session_id = ?
ORDER BY rowid ASC;`

	instRows, err := s.db.QueryContext(ctx, queryInstances, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query question instances: %w", err)
	}
	defer instRows.Close()

	var rawInstances []domain.QuestionInstance
	for instRows.Next() {
		var (
			instID      string
			instTmplID  string
			instTmplVer int
			instSeed    int64
			paramsJSON  string
			title       string
			scenarioMD  string
			stagesJSON  string
		)
		if err := instRows.Scan(
			&instID, &instTmplID, &instTmplVer, &instSeed,
			&paramsJSON, &title, &scenarioMD, &stagesJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan question instance: %w", err)
		}

		var params map[string]interface{}
		_ = json.Unmarshal([]byte(paramsJSON), &params)

		var stageInstances []domain.StageInstance
		_ = json.Unmarshal([]byte(stagesJSON), &stageInstances)

		rawInstances = append(rawInstances, domain.QuestionInstance{
			ID:               instID,
			TemplateID:       instTmplID,
			TemplateVersion:  instTmplVer,
			Seed:             instSeed,
			Parameters:       params,
			Title:            title,
			ScenarioMarkdown: scenarioMD,
			Stages:           stageInstances,
		})
	}
	if err := instRows.Err(); err != nil {
		return nil, err
	}
	if len(rawInstances) == 0 {
		return nil, fmt.Errorf("question instance snapshot missing for session %q", id)
	}

	// 3. Fetch drill_stage_states
	const queryStageStates = `
SELECT stage_index, stage_id, status, first_try_correct, solved_on_retry,
       revealed, active_hint, last_feedback, misconception_id, invalid_input_notice
FROM drill_stage_states
WHERE session_id = ?
ORDER BY stage_index ASC;`

	rows, err := s.db.QueryContext(ctx, queryStageStates, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query drill_stage_states: %w", err)
	}
	defer rows.Close()

	type rawStageState struct {
		stageIndex         int
		stageID            string
		status             drill.StageProgressStatus
		firstTryCorrect    bool
		solvedOnRetry      bool
		revealed           bool
		activeHint         string
		lastFeedback       string
		misconceptionID    *string
		invalidInputNotice string
	}

	stageStatesByKey := make(map[string]rawStageState)
	for rows.Next() {
		var (
			stIndex       int
			stID          string
			stStatus      string
			firstTryInt   int
			solvedRetry   int
			revealedInt   int
			activeHint    string
			lastFeedback  string
			miscID        *string
			invalidNotice string
		)
		if err := rows.Scan(
			&stIndex, &stID, &stStatus, &firstTryInt, &solvedRetry,
			&revealedInt, &activeHint, &lastFeedback, &miscID, &invalidNotice,
		); err != nil {
			return nil, fmt.Errorf("failed to scan stage state: %w", err)
		}

		stageStatesByKey[stID] = rawStageState{
			stageIndex:         stIndex,
			stageID:            stID,
			status:             drill.StageProgressStatus(stStatus),
			firstTryCorrect:    firstTryInt == 1,
			solvedOnRetry:      solvedRetry == 1,
			revealed:           revealedInt == 1,
			activeHint:         activeHint,
			lastFeedback:       lastFeedback,
			misconceptionID:    miscID,
			invalidInputNotice: invalidNotice,
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 4. Fetch attempts
	const queryAttempts = `
SELECT id, instance_id, stage_id, attempt_number,
       submitted_answer_json, assistance_json, is_correct,
       feedback_markdown, created_at
FROM attempts
WHERE session_id = ?
ORDER BY stage_id, attempt_number ASC;`

	attRows, err := s.db.QueryContext(ctx, queryAttempts, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query attempts: %w", err)
	}
	defer attRows.Close()

	attemptsByInstStage := make(map[string][]domain.StageAttempt)
	attemptsByStage := make(map[string][]domain.StageAttempt)
	for attRows.Next() {
		var (
			attID      string
			instIDVal  string
			stageIDVal string
			attNum     int
			ansJSON    string
			asstJSON   string
			isCorrInt  int
			feedbackMD string
			attCreated string
		)
		if err := attRows.Scan(
			&attID, &instIDVal, &stageIDVal, &attNum,
			&ansJSON, &asstJSON, &isCorrInt, &feedbackMD, &attCreated,
		); err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
		}

		var ans domain.SubmittedAnswer
		_ = json.Unmarshal([]byte(ansJSON), &ans)

		var asstList []domain.AssistanceType
		_ = json.Unmarshal([]byte(asstJSON), &asstList)

		t, _ := time.Parse(time.RFC3339Nano, attCreated)

		att := domain.StageAttempt{
			ID:               attID,
			SessionID:        id,
			InstanceID:       instIDVal,
			StageID:          stageIDVal,
			AttemptNumber:    attNum,
			SubmittedAnswer:  ans,
			Assistance:       asstList,
			IsCorrect:        isCorrInt == 1,
			FeedbackMarkdown: feedbackMD,
			CreatedAt:        t,
		}
		scopedKey := fmt.Sprintf("%s:%s", instIDVal, stageIDVal)
		attemptsByInstStage[scopedKey] = append(attemptsByInstStage[scopedKey], att)
		attemptsByStage[stageIDVal] = append(attemptsByStage[stageIDVal], att)
	}
	if err := attRows.Err(); err != nil {
		return nil, err
	}

	// 5. Fetch assistance events
	const queryAssistance = `
SELECT instance_id, stage_id, event_type
FROM assistance_events
WHERE session_id = ?
ORDER BY created_at ASC;`

	asstRows, err := s.db.QueryContext(ctx, queryAssistance, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query assistance events: %w", err)
	}
	defer asstRows.Close()

	assistanceByInstStage := make(map[string][]domain.AssistanceType)
	assistanceByStage := make(map[string][]domain.AssistanceType)
	for asstRows.Next() {
		var (
			instID string
			stID   string
			evtTyp string
		)
		if err := asstRows.Scan(&instID, &stID, &evtTyp); err != nil {
			return nil, fmt.Errorf("failed to scan assistance event: %w", err)
		}
		scopedKey := fmt.Sprintf("%s:%s", instID, stID)
		assistanceByInstStage[scopedKey] = append(assistanceByInstStage[scopedKey], domain.AssistanceType(evtTyp))
		assistanceByStage[stID] = append(assistanceByStage[stID], domain.AssistanceType(evtTyp))
	}

	// 6. Fetch session drafts
	draftsByKey := make(map[string]domain.SubmittedAnswer)
	const queryDrafts = `SELECT stage_id, draft_answer_json FROM session_drafts WHERE session_id = ?;`
	if draftRows, err := s.db.QueryContext(ctx, queryDrafts, id); err == nil {
		defer draftRows.Close()
		for draftRows.Next() {
			var stID, dJSON string
			if err := draftRows.Scan(&stID, &dJSON); err == nil {
				var draftAns domain.SubmittedAnswer
				if err := json.Unmarshal([]byte(dJSON), &draftAns); err == nil {
					draftsByKey[stID] = draftAns
				}
			}
		}
	}

	// 7. Reconstruct single or multi question session
	isMulti := len(rawInstances) > 1 || len(meta.QuestionStates) > 0
	var sess *drill.DrillSession

	if isMulti {
		questions := make([]drill.QuestionState, len(rawInstances))
		for k, qInst := range rawInstances {
			qStages := make([]drill.StageState, len(qInst.Stages))
			for i, stInst := range qInst.Stages {
				scopedKey := fmt.Sprintf("%s:%s", qInst.ID, stInst.ID)
				st := drill.StageState{
					Instance:   stInst,
					Status:     drill.StageStatusUnvisited,
					Attempts:   attemptsByInstStage[scopedKey],
					Assistance: assistanceByInstStage[scopedKey],
				}
				if len(st.Attempts) == 0 {
					st.Attempts = attemptsByStage[stInst.ID]
				}
				if len(st.Assistance) == 0 {
					st.Assistance = assistanceByStage[stInst.ID]
				}
				if st.Attempts == nil {
					st.Attempts = make([]domain.StageAttempt, 0)
				}
				if st.Assistance == nil {
					st.Assistance = make([]domain.AssistanceType, 0)
				}

				saved, hasSaved := stageStatesByKey[scopedKey]
				if !hasSaved {
					saved, hasSaved = stageStatesByKey[stInst.ID]
				}
				if hasSaved {
					st.Status = saved.status
					st.FirstTryCorrect = saved.firstTryCorrect
					st.SolvedOnRetry = saved.solvedOnRetry
					st.Revealed = saved.revealed
					st.ActiveHint = saved.activeHint
					st.LastFeedback = saved.lastFeedback
					st.MisconceptionID = saved.misconceptionID
					st.InvalidInputNotice = saved.invalidInputNotice
				} else if i == 0 {
					st.Status = drill.StageStatusActive
				}

				if d, ok := draftsByKey[scopedKey]; ok {
					dCopy := d
					st.DraftAnswer = &dCopy
				} else if d, ok := draftsByKey[stInst.ID]; ok {
					dCopy := d
					st.DraftAnswer = &dCopy
				}

				qStages[i] = st
			}

			qCurrentStage := 0
			qCompleted := false
			if k < len(meta.QuestionStates) {
				qCurrentStage = meta.QuestionStates[k].CurrentStageIndex
				qCompleted = meta.QuestionStates[k].Completed
			} else {
				allDone := true
				for _, st := range qStages {
					if st.Status != drill.StageStatusCompleted {
						allDone = false
						break
					}
				}
				qCompleted = allDone
			}

			questions[k] = drill.QuestionState{
				Index:             k,
				TemplateID:        qInst.TemplateID,
				TemplateVersion:   qInst.TemplateVersion,
				QuestionInstance:  qInst,
				Stages:            qStages,
				CurrentStageIndex: qCurrentStage,
				Completed:         qCompleted,
			}
		}

		if currentQIndex < 0 || currentQIndex >= len(questions) {
			currentQIndex = 0
		}

		activeQ := questions[currentQIndex]
		sess = &drill.DrillSession{
			ID:                   sessID,
			TemplateID:           activeQ.TemplateID,
			TemplateVersion:      activeQ.TemplateVersion,
			Seed:                 seed,
			Revision:             revision,
			QuestionInstance:     activeQ.QuestionInstance,
			Stages:               activeQ.Stages,
			CurrentStageIndex:    activeQ.CurrentStageIndex,
			Completed:            completedInt == 1,
			CreatedAt:            createdAt,
			UpdatedAt:            updatedAt,
			CurrentQuestionIndex: currentQIndex,
			Questions:            questions,
			Settings:             meta.Settings,
		}
	} else {
		// Single question backward compatible
		qInst := rawInstances[0]
		stages := make([]drill.StageState, len(qInst.Stages))
		for i, inst := range qInst.Stages {
			st := drill.StageState{
				Instance:   inst,
				Status:     drill.StageStatusUnvisited,
				Attempts:   attemptsByStage[inst.ID],
				Assistance: assistanceByStage[inst.ID],
			}
			if st.Attempts == nil {
				st.Attempts = make([]domain.StageAttempt, 0)
			}
			if st.Assistance == nil {
				st.Assistance = make([]domain.AssistanceType, 0)
			}

			if saved, ok := stageStatesByKey[inst.ID]; ok {
				st.Status = saved.status
				st.FirstTryCorrect = saved.firstTryCorrect
				st.SolvedOnRetry = saved.solvedOnRetry
				st.Revealed = saved.revealed
				st.ActiveHint = saved.activeHint
				st.LastFeedback = saved.lastFeedback
				st.MisconceptionID = saved.misconceptionID
				st.InvalidInputNotice = saved.invalidInputNotice
			} else if i == 0 {
				st.Status = drill.StageStatusActive
			}

			if ans, ok := draftsByKey[inst.ID]; ok {
				ansCopy := ans
				st.DraftAnswer = &ansCopy
			}

			stages[i] = st
		}

		sess = &drill.DrillSession{
			ID:                sessID,
			TemplateID:        tmplID,
			TemplateVersion:   tmplVer,
			Seed:              seed,
			Revision:          revision,
			QuestionInstance:  qInst,
			Stages:            stages,
			CurrentStageIndex: currentStageIndex,
			Completed:         completedInt == 1,
			CreatedAt:         createdAt,
			UpdatedAt:         updatedAt,
			Settings:          meta.Settings,
		}
	}

	sess.SetStore(s)

	// 7. Load cached idempotency results
	const queryIdempotency = `
SELECT command_id, result_json
FROM command_idempotency
WHERE session_id = ?;`

	idempRows, err := s.db.QueryContext(ctx, queryIdempotency, id)
	if err == nil {
		defer idempRows.Close()
		for idempRows.Next() {
			var (
				cmdID   string
				resJSON string
			)
			if err := idempRows.Scan(&cmdID, &resJSON); err == nil {
				var res drill.CommandResult
				if err := json.Unmarshal([]byte(resJSON), &res); err == nil {
					sess.SetCachedResult(cmdID, &res)
				}
			}
		}
	}

	return sess, nil
}

// GetLatestActiveSession retrieves the most recently updated active (incomplete) session,
// or falls back to the most recently updated completed session.
func (s *Store) GetLatestActiveSession(ctx context.Context) (*drill.DrillSession, error) {
	// First check incomplete session
	var sessionID string
	err := s.db.QueryRowContext(ctx, `
SELECT id FROM sessions
WHERE completed = 0
ORDER BY updated_at DESC
LIMIT 1;`).Scan(&sessionID)

	if err == nil && sessionID != "" {
		return s.GetSession(ctx, sessionID)
	}

	// Fallback to most recent completed session
	err = s.db.QueryRowContext(ctx, `
SELECT id FROM sessions
WHERE completed = 1
ORDER BY updated_at DESC
LIMIT 1;`).Scan(&sessionID)

	if err == nil && sessionID != "" {
		return s.GetSession(ctx, sessionID)
	}

	return nil, ErrNotFound
}

// SaveCommandResult records the evaluated command result for idempotency caching.
func (s *Store) SaveCommandResult(ctx context.Context, sessionID, cmdID string, cmdType drill.CommandType, res *drill.CommandResult) error {
	if cmdID == "" || res == nil {
		return nil
	}

	resJSON, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshal command result: %w", err)
	}

	nowStr := s.clock().UTC().Format(time.RFC3339Nano)
	const query = `
INSERT OR REPLACE INTO command_idempotency (
    command_id, session_id, command_type, result_json, created_at
) VALUES (?, ?, ?, ?, ?);`

	_, err = s.db.ExecContext(ctx, query, cmdID, sessionID, string(cmdType), string(resJSON), nowStr)
	if err != nil {
		return fmt.Errorf("failed to save command idempotency result: %w", err)
	}

	return nil
}

// GetCommandResult retrieves a previously evaluated command result if present.
func (s *Store) GetCommandResult(ctx context.Context, cmdID string) (*drill.CommandResult, error) {
	if cmdID == "" {
		return nil, nil
	}

	const query = `SELECT result_json FROM command_idempotency WHERE command_id = ?;`
	var resJSON string
	err := s.db.QueryRowContext(ctx, query, cmdID).Scan(&resJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query command idempotency: %w", err)
	}

	var res drill.CommandResult
	if err := json.Unmarshal([]byte(resJSON), &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal command result: %w", err)
	}

	return &res, nil
}

// SaveDraftAnswer stores an unsubmitted stage input draft, fulfilling drill.SessionStore.
func (s *Store) SaveDraftAnswer(ctx context.Context, sessionID, stageID string, ans domain.SubmittedAnswer) error {
	draft := SessionDraft{
		SessionID:      sessionID,
		StageID:        stageID,
		DraftAnswer:    ans,
		ActivePosition: 0,
		Revision:       0,
		UpdatedAt:      s.clock(),
	}
	return s.SaveSessionDraft(ctx, draft)
}

// SaveSessionDraft persists an unsubmitted stage input draft.
func (s *Store) SaveSessionDraft(ctx context.Context, draft SessionDraft) error {
	ansJSON, err := json.Marshal(draft.DraftAnswer)
	if err != nil {
		return fmt.Errorf("failed to marshal draft answer: %w", err)
	}

	nowStr := draft.UpdatedAt.UTC().Format(time.RFC3339Nano)
	const query = `
INSERT INTO session_drafts (
    session_id, stage_id, draft_answer_json, active_position, revision, updated_at
) VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(session_id, stage_id) DO UPDATE SET
    draft_answer_json = excluded.draft_answer_json,
    active_position = excluded.active_position,
    revision = excluded.revision,
    updated_at = excluded.updated_at;`

	_, err = s.db.ExecContext(ctx, query,
		draft.SessionID,
		draft.StageID,
		string(ansJSON),
		draft.ActivePosition,
		draft.Revision,
		nowStr,
	)
	return err
}

// GetSessionDraft retrieves an unsubmitted stage input draft if one exists.
func (s *Store) GetSessionDraft(ctx context.Context, sessionID, stageID string) (*SessionDraft, error) {
	const query = `
SELECT draft_answer_json, active_position, revision, updated_at
FROM session_drafts
WHERE session_id = ? AND stage_id = ?;`

	var (
		ansJSON   string
		activePos int
		rev       int64
		timeStr   string
	)

	err := s.db.QueryRowContext(ctx, query, sessionID, stageID).Scan(&ansJSON, &activePos, &rev, &timeStr)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var ans domain.SubmittedAnswer
	_ = json.Unmarshal([]byte(ansJSON), &ans)
	t, _ := time.Parse(time.RFC3339Nano, timeStr)

	return &SessionDraft{
		SessionID:      sessionID,
		StageID:        stageID,
		DraftAnswer:    ans,
		ActivePosition: activePos,
		Revision:       rev,
		UpdatedAt:      t,
	}, nil
}

// ClearSessionDraft removes an unsubmitted draft once submitted or discarded.
func (s *Store) ClearSessionDraft(ctx context.Context, sessionID, stageID string) error {
	const query = `DELETE FROM session_drafts WHERE session_id = ? AND stage_id = ?;`
	_, err := s.db.ExecContext(ctx, query, sessionID, stageID)
	return err
}

// SaveSettings stores a key-value application preference.
func (s *Store) SaveSettings(ctx context.Context, key, valueJSON string) error {
	nowStr := s.clock().UTC().Format(time.RFC3339Nano)
	const query = `
INSERT INTO settings (key, value_json, updated_at)
VALUES (?, ?, ?)
ON CONFLICT(key) DO UPDATE SET
    value_json = excluded.value_json,
    updated_at = excluded.updated_at;`
	_, err := s.db.ExecContext(ctx, query, key, valueJSON, nowStr)
	return err
}

// GetSettings retrieves a key-value application preference.
func (s *Store) GetSettings(ctx context.Context, key string) (string, error) {
	const query = `SELECT value_json FROM settings WHERE key = ?;`
	var val string
	err := s.db.QueryRowContext(ctx, query, key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	return val, err
}
