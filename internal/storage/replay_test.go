package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/drill"
)

// TestCloseReopenRestoresDrill satisfies Stage 05 Exit Evidence:
// "Close/reopen restores drill"
func TestCloseReopenRestoresDrill(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "practice.db")

	// Phase 1: Initialize database, store, and create drill
	db1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	ctx := context.Background()
	if err := RunMigrations(ctx, db1, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	store1 := NewStore(db1, nil)
	sm1 := drill.NewSessionManagerWithStore(store1, nil)
	tmpl := createTestTemplate()

	sess1, err := sm1.CreateSession(tmpl, 12345)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Learner completes stage 1
	submitCmd := drill.SessionCommand{
		CommandID:        "cmd_submit_st1",
		ExpectedRevision: 1,
		Type:             drill.CmdSubmitAnswer,
		StageID:          "stage_1",
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "opt_a",
		},
	}
	res1, err := sess1.ExecuteCommand(submitCmd, nil)
	if err != nil || !res1.Success {
		t.Fatalf("ExecuteCommand stage 1 failed: %v", err)
	}

	// Learner requests a hint on stage 2
	hintCmd := drill.SessionCommand{
		CommandID:        "cmd_hint_st2",
		ExpectedRevision: 2,
		Type:             drill.CmdRequestHint,
		StageID:          "stage_2",
	}
	res2, err := sess1.ExecuteCommand(hintCmd, nil)
	if err != nil || !res2.Success {
		t.Fatalf("ExecuteCommand hint stage 2 failed: %v", err)
	}

	// Phase 2: Simulate complete shutdown / close
	_ = db1.Close()

	// Phase 3: Simulate application reopen
	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Reopen database failed: %v", err)
	}
	defer db2.Close()

	store2 := NewStore(db2, nil)
	sm2 := drill.NewSessionManagerWithStore(store2, nil)

	// Restore active session
	restored, ok := sm2.GetActiveSession()
	if !ok || restored == nil {
		t.Fatalf("expected GetActiveSession to restore session on reopen")
	}

	if restored.ID != sess1.ID {
		t.Errorf("restored ID %q != original ID %q", restored.ID, sess1.ID)
	}
	if restored.Revision != 3 {
		t.Errorf("expected restored revision 3, got %d", restored.Revision)
	}
	if restored.CurrentStageIndex != 1 {
		t.Errorf("expected restored current stage 1, got %d", restored.CurrentStageIndex)
	}

	// Verify Stage 1 is completed with attempt history intact
	st1 := restored.Stages[0]
	if st1.Status != drill.StageStatusCompleted {
		t.Errorf("expected Stage 1 status 'completed', got %v", st1.Status)
	}
	if !st1.FirstTryCorrect {
		t.Errorf("expected Stage 1 FirstTryCorrect true")
	}
	if len(st1.Attempts) != 1 {
		t.Fatalf("expected 1 attempt on Stage 1, got %d", len(st1.Attempts))
	}
	if st1.Attempts[0].SubmittedAnswer.OptionID != "opt_a" {
		t.Errorf("expected attempt answer 'opt_a', got %q", st1.Attempts[0].SubmittedAnswer.OptionID)
	}

	// Verify Stage 2 is active with hint assistance preserved
	st2 := restored.Stages[1]
	if st2.Status != drill.StageStatusActive {
		t.Errorf("expected Stage 2 status 'active', got %v", st2.Status)
	}
	if st2.ActiveHint == "" {
		t.Errorf("expected Stage 2 ActiveHint to be preserved")
	}
	hasHint := false
	for _, a := range st2.Assistance {
		if a == domain.AssistanceHint {
			hasHint = true
			break
		}
	}
	if !hasHint {
		t.Errorf("expected Stage 2 Assistance to contain 'hint', got %v", st2.Assistance)
	}

	// Learner can continue and complete stage 2 on the restored session
	submitNumCmd := drill.SessionCommand{
		CommandID:        "cmd_submit_st2",
		ExpectedRevision: 3,
		Type:             drill.CmdSubmitAnswer,
		StageID:          "stage_2",
		Answer: &domain.SubmittedAnswer{
			Kind:       domain.StageKindNumeric,
			NumericRaw: "0.375",
		},
	}
	res3, err := restored.ExecuteCommand(submitNumCmd, nil)
	if err != nil || !res3.Success {
		t.Fatalf("failed to complete restored stage 2: %v", err)
	}
	if !restored.Completed {
		t.Errorf("expected overall drill to be completed after solving both stages")
	}
}

// TestDuplicateCommandCannotInflateEvidence satisfies Stage 05 Exit Evidence:
// "Duplicate command cannot inflate evidence"
func TestDuplicateCommandCannotInflateEvidence(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	sm := drill.NewSessionManagerWithStore(store, nil)
	tmpl := createTestTemplate()

	sess, err := sm.CreateSession(tmpl, 999)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	const fixedCmdID = "cmd_idempotent_submit_001"
	cmd := drill.SessionCommand{
		CommandID:        fixedCmdID,
		ExpectedRevision: 1,
		Type:             drill.CmdSubmitAnswer,
		StageID:          "stage_1",
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "opt_a",
		},
	}

	// 1. Initial submission
	res1, err := sess.ExecuteCommand(cmd, nil)
	if err != nil || !res1.Success {
		t.Fatalf("initial ExecuteCommand failed: %v", err)
	}
	if res1.SessionState.Revision != 2 {
		t.Errorf("expected revision 2, got %d", res1.SessionState.Revision)
	}

	// Verify database record count
	var attemptCount int
	err = store.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM attempts WHERE session_id = ?", sess.ID).Scan(&attemptCount)
	if err != nil || attemptCount != 1 {
		t.Fatalf("expected exactly 1 attempt in DB, got %d (err: %v)", attemptCount, err)
	}

	// 2. Duplicate submission with identical command_id (network retry / double click)
	res2, err := sess.ExecuteCommand(cmd, nil)
	if err != nil {
		t.Fatalf("duplicate ExecuteCommand failed: %v", err)
	}
	if !res2.Success {
		t.Errorf("expected cached success response")
	}

	// Database attempt count MUST NOT INFLATE
	err = store.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM attempts WHERE session_id = ?", sess.ID).Scan(&attemptCount)
	if err != nil || attemptCount != 1 {
		t.Fatalf("evidence inflated: expected exactly 1 attempt in DB after duplicate command, got %d", attemptCount)
	}

	// Session revision in DB MUST NOT increment
	var currentRev int64
	err = store.DB().QueryRowContext(ctx, "SELECT revision FROM sessions WHERE id = ?", sess.ID).Scan(&currentRev)
	if err != nil || currentRev != 2 {
		t.Fatalf("expected revision to remain 2, got %d", currentRev)
	}

	// 3. Test after a simulated restart (clearing memory cache)
	smRestart := drill.NewSessionManagerWithStore(store, nil)
	reloadedSess, ok := smRestart.GetSession(sess.ID)
	if !ok || reloadedSess == nil {
		t.Fatalf("failed to reload session from store")
	}

	res3, err := reloadedSess.ExecuteCommand(cmd, nil)
	if err != nil || !res3.Success {
		t.Fatalf("ExecuteCommand after restart returned error: %v", err)
	}

	// Still exactly 1 attempt in DB
	err = store.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM attempts WHERE session_id = ?", sess.ID).Scan(&attemptCount)
	if err != nil || attemptCount != 1 {
		t.Fatalf("evidence inflated after restart: expected 1 attempt, got %d", attemptCount)
	}
}

// TestChangedBankCannotChangeHistoricalContentOrGrades satisfies Stage 05 Exit Evidence:
// "Changed bank cannot change historical content/grades"
func TestChangedBankCannotChangeHistoricalContentOrGrades(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	sm := drill.NewSessionManagerWithStore(store, nil)

	// Create mutable original template: n=4, p=0.5, k=2, key = 0.375
	tmpl := createTestTemplate()
	origTitle := tmpl.Title
	origScenario := tmpl.ScenarioMarkdown
	origPromptSt1 := tmpl.Stages[0].PromptMarkdown

	sess, err := sm.CreateSession(tmpl, 100)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Complete stage 1
	_, err = sess.ExecuteCommand(drill.SessionCommand{
		CommandID:        "cmd_1",
		ExpectedRevision: 1,
		Type:             drill.CmdSubmitAnswer,
		StageID:          "stage_1",
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "opt_a",
		},
	}, nil)
	if err != nil {
		t.Fatalf("failed to complete stage 1: %v", err)
	}

	// SIMULATE ACTIVE BANK MUTATION:
	// A curriculum author modifies the template in the bank file:
	// They change title, change scenario, change parameters to n=10, p=0.9, k=8,
	// change expected answers to 0.193, change options, and change explanations!
	tmpl.Title = "MUTATED BANK TITLE - DO NOT USE FOR HISTORICAL SESSIONS"
	tmpl.ScenarioMarkdown = "MUTATED SCENARIO n=10, p=0.9, k=8"
	tmpl.Parameters = map[string]interface{}{"n": 10, "p": 0.9, "k": 8}
	tmpl.Stages[0].PromptMarkdown = "MUTATED STAGE 1 PROMPT"
	tmpl.Stages[1].ExpectedAnswer = domain.ExpectedAnswer{
		Kind:  domain.AnswerKindNumeric,
		Value: floatPtr(0.193),
		Units: "probability",
	}

	// Now re-open / resume the session from storage
	smRestart := drill.NewSessionManagerWithStore(store, nil)
	restored, ok := smRestart.GetSession(sess.ID)
	if !ok || restored == nil {
		t.Fatalf("failed to reload session")
	}

	// 1. Verify historical content is 100% UNCHANGED
	if restored.QuestionInstance.Title != origTitle {
		t.Errorf("historical title was corrupted by bank mutation: got %q, expected %q",
			restored.QuestionInstance.Title, origTitle)
	}
	if restored.QuestionInstance.ScenarioMarkdown != origScenario {
		t.Errorf("historical scenario was corrupted by bank mutation: got %q, expected %q",
			restored.QuestionInstance.ScenarioMarkdown, origScenario)
	}
	if restored.Stages[0].Instance.PromptMarkdown != origPromptSt1 {
		t.Errorf("historical stage prompt was corrupted by bank mutation: got %q, expected %q",
			restored.Stages[0].Instance.PromptMarkdown, origPromptSt1)
	}
	if restored.Stages[1].Instance.ExpectedAnswer.Value == nil || *restored.Stages[1].Instance.ExpectedAnswer.Value != 0.375 {
		t.Errorf("historical expected answer was corrupted: got %v, expected 0.375",
			restored.Stages[1].Instance.ExpectedAnswer.Value)
	}

	// 2. Verify grading on unresolved stage 2 STILL grades against original snapshot (0.375),
	// NOT the mutated bank value (0.193)!
	gradeRes, err := restored.ExecuteCommand(drill.SessionCommand{
		CommandID:        "cmd_grade_st2",
		ExpectedRevision: 2,
		Type:             drill.CmdSubmitAnswer,
		StageID:          "stage_2",
		Answer: &domain.SubmittedAnswer{
			Kind:       domain.StageKindNumeric,
			NumericRaw: "0.375",
		},
	}, nil)
	if err != nil || !gradeRes.Success {
		t.Fatalf("grading against original snapshot failed: %v", err)
	}

	if !restored.Stages[1].FirstTryCorrect {
		t.Errorf("expected answer 0.375 to be graded correct under historical snapshot policy")
	}
	if !restored.Completed {
		t.Errorf("expected session to complete")
	}

	// Verify public view recap generated from stored snapshot
	pub := restored.ToPublicView()
	if pub.Recap == nil {
		t.Fatalf("expected recap to be present on completed session")
	}
	if pub.Title != origTitle {
		t.Errorf("recap title mismatch: got %q, expected %q", pub.Title, origTitle)
	}
}

func TestMultiTabRevisionConflictRejection(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	sm := drill.NewSessionManagerWithStore(store, nil)
	tmpl := createTestTemplate()

	sess, err := sm.CreateSession(tmpl, 42)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Tab 1 submits answer at revision 1 -> advances revision to 2
	cmdTab1 := drill.SessionCommand{
		CommandID:        "cmd_tab1",
		ExpectedRevision: 1,
		Type:             drill.CmdSubmitAnswer,
		StageID:          "stage_1",
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "opt_a",
		},
	}
	res1, err := sess.ExecuteCommand(cmdTab1, nil)
	if err != nil || !res1.Success {
		t.Fatalf("Tab 1 command failed: %v", err)
	}

	// Tab 2 (stale tab) tries to submit with ExpectedRevision = 1
	cmdTab2 := drill.SessionCommand{
		CommandID:        "cmd_tab2",
		ExpectedRevision: 1,
		Type:             drill.CmdSubmitAnswer,
		StageID:          "stage_1",
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "opt_b",
		},
	}
	res2, err := sess.ExecuteCommand(cmdTab2, nil)
	if err != drill.ErrRevisionConflict {
		t.Errorf("expected ErrRevisionConflict for stale revision, got err: %v", err)
	}
	if res2.Success {
		t.Errorf("expected success to be false on conflict")
	}
	if res2.SessionState.Revision != 2 {
		t.Errorf("expected conflict response to return current session state at revision 2, got %d", res2.SessionState.Revision)
	}
}

// TestNavigationAndDraftPersistence satisfies Stage 06 requirements:
// "prior/future stage visits don't inflate evidence; navigation preserves unsent answers"
func TestNavigationAndDraftPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "navigation_draft.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	ctx := context.Background()
	if err := RunMigrations(ctx, db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	store := NewStore(db, nil)
	sm := drill.NewSessionManagerWithStore(store, nil)
	tmpl := createTestTemplate()

	sess, err := sm.CreateSession(tmpl, 54321)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// 1. Solve stage 1
	submitCmd1 := drill.SessionCommand{
		CommandID:        "cmd_submit_st1",
		ExpectedRevision: 1,
		Type:             drill.CmdSubmitAnswer,
		StageID:          "stage_1",
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "opt_a",
		},
	}
	if _, err := sess.ExecuteCommand(submitCmd1, nil); err != nil {
		t.Fatalf("submit stage 1 failed: %v", err)
	}

	// Now session is at stage 2. Verify attempts count in DB = 1
	var attCountBefore int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM attempts WHERE session_id = ?", sess.ID).Scan(&attCountBefore); err != nil {
		t.Fatalf("query attempts failed: %v", err)
	}
	if attCountBefore != 1 {
		t.Fatalf("expected 1 attempt, got %d", attCountBefore)
	}

	// 2. Navigate back to Stage 1, then forward to Stage 2 multiple times
	target0 := 0
	navCmdBack := drill.SessionCommand{
		CommandID:        "cmd_nav_back",
		ExpectedRevision: sess.Revision,
		Type:             drill.CmdNavigateStage,
		TargetStageIndex: &target0,
	}
	if _, err := sess.ExecuteCommand(navCmdBack, nil); err != nil {
		t.Fatalf("nav back failed: %v", err)
	}

	target1 := 1
	navCmdFwd := drill.SessionCommand{
		CommandID:        "cmd_nav_fwd",
		ExpectedRevision: sess.Revision,
		Type:             drill.CmdNavigateStage,
		TargetStageIndex: &target1,
	}
	if _, err := sess.ExecuteCommand(navCmdFwd, nil); err != nil {
		t.Fatalf("nav fwd failed: %v", err)
	}

	// Invariant: Navigation cannot inflate evidence
	var attCountAfter int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM attempts WHERE session_id = ?", sess.ID).Scan(&attCountAfter); err != nil {
		t.Fatalf("query attempts failed: %v", err)
	}
	if attCountAfter != 1 {
		t.Errorf("navigation inflated attempts! expected 1, got %d", attCountAfter)
	}

	// 3. User types unsent draft on Stage 2
	draftAns := &domain.SubmittedAnswer{
		Kind:       domain.StageKindNumeric,
		NumericRaw: "0.375",
	}
	saveDraftCmd := drill.SessionCommand{
		CommandID:        "cmd_draft_st2",
		ExpectedRevision: sess.Revision,
		Type:             drill.CmdSaveDraft,
		StageID:          "stage_2",
		Answer:           draftAns,
	}
	resDraft, err := sess.ExecuteCommand(saveDraftCmd, nil)
	if err != nil || !resDraft.Success {
		t.Fatalf("save draft failed: %v", err)
	}

	// Verify draft in public projection
	view := sess.ToPublicView()
	if view.Stages[1].DraftAnswer == nil || view.Stages[1].DraftAnswer.NumericRaw != "0.375" {
		t.Fatalf("expected stage 2 to have draft answer 0.375, got %+v", view.Stages[1].DraftAnswer)
	}

	// 4. Simulate process restart: close DB and reopen
	if err := db.Close(); err != nil {
		t.Fatalf("db close failed: %v", err)
	}

	dbReopen, err := Open(dbPath)
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}
	defer dbReopen.Close()

	storeReopen := NewStore(dbReopen, nil)
	sessRestored, err := storeReopen.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("failed to restore session: %v", err)
	}

	restoredView := sessRestored.ToPublicView()
	if restoredView.Stages[1].DraftAnswer == nil || restoredView.Stages[1].DraftAnswer.NumericRaw != "0.375" {
		t.Fatalf("reopened session lost draft answer! got %+v", restoredView.Stages[1].DraftAnswer)
	}

	// 5. Submit answer on stage 2: draft should be cleared from database
	submitCmd2 := drill.SessionCommand{
		CommandID:        "cmd_submit_st2",
		ExpectedRevision: sessRestored.Revision,
		Type:             drill.CmdSubmitAnswer,
		StageID:          "stage_2",
		Answer:           draftAns,
	}
	if _, err := sessRestored.ExecuteCommand(submitCmd2, nil); err != nil {
		t.Fatalf("submit stage 2 failed: %v", err)
	}

	var draftCount int
	if err := dbReopen.QueryRowContext(ctx, "SELECT COUNT(*) FROM session_drafts WHERE session_id = ? AND stage_id = 'stage_2'", sess.ID).Scan(&draftCount); err != nil {
		t.Fatalf("query session_drafts failed: %v", err)
	}
	if draftCount != 0 {
		t.Errorf("expected session_drafts to be 0 after submit, got %d", draftCount)
	}
}

// TestMultiQuestionSessionPersistenceAndReplay verifies Stage 07 exit criteria:
// 10-question session persistence, restart, question jumping, and state preservation.
func TestMultiQuestionSessionPersistenceAndReplay(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "ten_questions.db")

	db1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	ctx := context.Background()
	if err := RunMigrations(ctx, db1, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	store1 := NewStore(db1, nil)
	sm1 := drill.NewSessionManagerWithStore(store1, nil)

	// Create 10 templates
	templates := make([]*domain.QuestionTemplate, 10)
	for i := 0; i < 10; i++ {
		tCopy := *createTestTemplate()
		tCopy.ID = fmt.Sprintf("tmpl_q_%d", i)
		tCopy.Title = fmt.Sprintf("Question %d Title", i+1)
		templates[i] = &tCopy
	}

	settings := domain.SessionSettings{
		QuestionCount: 10,
		Intensity:     "standard",
	}

	sess1, err := sm1.CreateMultiQuestionSession(templates, settings, 9999)
	if err != nil {
		t.Fatalf("CreateMultiQuestionSession failed: %v", err)
	}

	if len(sess1.Questions) != 10 {
		t.Fatalf("expected 10 questions in session, got %d", len(sess1.Questions))
	}

	// 1. Submit answer on Q0 Stage 0
	subCmd := drill.SessionCommand{
		CommandID:        "cmd_q0_sub",
		ExpectedRevision: sess1.Revision,
		Type:             drill.CmdSubmitAnswer,
		StageID:          "stage_1",
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "opt_a",
		},
	}
	res, err := sess1.ExecuteCommand(subCmd, nil)
	if err != nil || !res.Success {
		t.Fatalf("ExecuteCommand on Q0 failed: %v", err)
	}

	// 2. Navigate to Question 1
	targetQ1 := 1
	navQCmd := drill.SessionCommand{
		CommandID:           "cmd_nav_q1",
		ExpectedRevision:    sess1.Revision,
		Type:                drill.CmdNavigateQuestion,
		TargetQuestionIndex: &targetQ1,
	}
	resNav, err := sess1.ExecuteCommand(navQCmd, nil)
	if err != nil || !resNav.Success {
		t.Fatalf("ExecuteCommand navigate to Q1 failed: %v", err)
	}
	if sess1.CurrentQuestionIndex != 1 {
		t.Fatalf("expected CurrentQuestionIndex=1, got %d", sess1.CurrentQuestionIndex)
	}

	// 3. Save draft on Question 1 Stage 1
	draftAns := &domain.SubmittedAnswer{
		Kind:       domain.StageKindNumeric,
		NumericRaw: "0.45",
	}
	draftCmd := drill.SessionCommand{
		CommandID:        "cmd_draft_q1",
		ExpectedRevision: sess1.Revision,
		Type:             drill.CmdSaveDraft,
		StageID:          "stage_2",
		Answer:           draftAns,
	}
	resDraft, err := sess1.ExecuteCommand(draftCmd, nil)
	if err != nil || !resDraft.Success {
		t.Fatalf("SaveDraft on Q1 failed: %v", err)
	}

	// 4. Simulate complete shutdown
	if err := db1.Close(); err != nil {
		t.Fatalf("db1.Close failed: %v", err)
	}

	// 5. Reopen database and recover session
	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("reopen db2 failed: %v", err)
	}
	defer db2.Close()

	store2 := NewStore(db2, nil)
	restoredSess, err := store2.GetLatestActiveSession(ctx)
	if err != nil {
		t.Fatalf("GetLatestActiveSession failed: %v", err)
	}

	if restoredSess.ID != sess1.ID {
		t.Fatalf("expected restored session ID %q, got %q", sess1.ID, restoredSess.ID)
	}
	if len(restoredSess.Questions) != 10 {
		t.Fatalf("expected 10 questions in restored session, got %d", len(restoredSess.Questions))
	}
	if restoredSess.CurrentQuestionIndex != 1 {
		t.Fatalf("expected restored CurrentQuestionIndex=1, got %d", restoredSess.CurrentQuestionIndex)
	}

	// Verify Q0 stage 0 is completed
	q0Stages := restoredSess.Questions[0].Stages
	if len(q0Stages) == 0 || q0Stages[0].Status != drill.StageStatusCompleted {
		t.Fatalf("expected Q0 stage 0 to be completed, got status %v", q0Stages[0].Status)
	}

	// Verify Q1 draft is preserved
	q1Stages := restoredSess.Questions[1].Stages
	if len(q1Stages) < 2 || q1Stages[1].DraftAnswer == nil || q1Stages[1].DraftAnswer.NumericRaw != "0.45" {
		t.Fatalf("expected Q1 stage 1 draft to be '0.45', got %+v", q1Stages[1].DraftAnswer)
	}

	// 6. Navigate to Question 9 (last question) on restored session
	targetQ9 := 9
	navQ9Cmd := drill.SessionCommand{
		CommandID:           "cmd_nav_q9",
		ExpectedRevision:    restoredSess.Revision,
		Type:                drill.CmdNavigateQuestion,
		TargetQuestionIndex: &targetQ9,
	}
	resNav9, err := restoredSess.ExecuteCommand(navQ9Cmd, nil)
	if err != nil || !resNav9.Success {
		t.Fatalf("Navigate to Q9 on restored session failed: %v", err)
	}
	if restoredSess.CurrentQuestionIndex != 9 {
		t.Fatalf("expected CurrentQuestionIndex=9, got %d", restoredSess.CurrentQuestionIndex)
	}
	if restoredSess.TemplateID != "tmpl_q_9" {
		t.Fatalf("expected TemplateID='tmpl_q_9', got %q", restoredSess.TemplateID)
	}
}

