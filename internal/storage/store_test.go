package storage

import (
	"context"
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/drill"
)

func floatPtr(v float64) *float64 { return &v }
func intPtr(v int) *int           { return &v }

func createTestTemplate() *domain.QuestionTemplate {
	return &domain.QuestionTemplate{
		ID:               "tmpl_test_binomial",
		FamilyID:         "binomial_pmf",
		Version:          1,
		ModuleID:         "discrete_distributions",
		ConceptIDs:       []string{"binomial_formula"},
		Title:            "Test Binomial Question",
		ScenarioMarkdown: "A test scenario with $n=4, p=0.5, k=2$.",
		Parameters: map[string]interface{}{
			"n": 4,
			"p": 0.5,
			"k": 2,
		},
		Stages: []domain.StageTemplate{
			{
				ID:             "stage_1",
				Kind:           domain.StageKindChoice,
				PromptMarkdown: "Stage 1 prompt",
				Options: []domain.Option{
					{ID: "opt_a", TextMarkdown: "Option A"},
					{ID: "opt_b", TextMarkdown: "Option B"},
				},
				ExpectedAnswer: domain.ExpectedAnswer{
					Kind:     domain.AnswerKindChoice,
					OptionID: "opt_a",
				},
				EvidenceConceptIDs:  []string{"binomial_formula"},
				ExplanationMarkdown: "Stage 1 explanation",
			},
			{
				ID:             "stage_2",
				Kind:           domain.StageKindNumeric,
				PromptMarkdown: "Stage 2 numeric prompt",
				Options:        nil,
				ExpectedAnswer: domain.ExpectedAnswer{
					Kind:  domain.AnswerKindNumeric,
					Value: floatPtr(0.375),
					Units: "probability",
				},
				EvidenceConceptIDs:  []string{"binomial_formula"},
				ExplanationMarkdown: "Stage 2 explanation",
				NumericPolicy: &domain.NumericPolicy{
					Version:           1,
					AbsoluteTolerance: 0.001,
					RelativeTolerance: 0.01,
					AllowedForms:      []domain.NumericForm{domain.NumericFormDecimal, domain.NumericFormPercent, domain.NumericFormFraction},
					DisplayDecimals:   intPtr(3),
				},
			},
		},
	}
}

func setupTestStore(t *testing.T) (*Store, func()) {
	db, err := OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory failed: %v", err)
	}
	ctx := context.Background()
	if err := RunMigrations(ctx, db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}
	store := NewStore(db, nil)
	return store, func() { _ = db.Close() }
}

func TestStoreSessionRoundTrip(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	tmpl := createTestTemplate()

	// Create a session via drill SessionManager
	sm := drill.NewSessionManagerWithStore(store, nil)
	sess, err := sm.CreateSession(tmpl, 42)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Add an attempt to Stage 1
	sess.Stages[0].Attempts = append(sess.Stages[0].Attempts, domain.StageAttempt{
		ID:            "att_1",
		SessionID:     sess.ID,
		InstanceID:    sess.QuestionInstance.ID,
		StageID:       "stage_1",
		AttemptNumber: 1,
		SubmittedAnswer: domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "opt_a",
		},
		Assistance:       []domain.AssistanceType{domain.AssistanceNone},
		IsCorrect:        true,
		FeedbackMarkdown: "Correct!",
		CreatedAt:        time.Now().UTC(),
	})
	sess.Stages[0].Status = drill.StageStatusCompleted
	sess.Stages[0].FirstTryCorrect = true
	sess.CurrentStageIndex = 1
	sess.Revision = 2
	sess.Stages[1].Status = drill.StageStatusActive

	// Persist
	if err := store.SaveSession(ctx, sess); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	// Load from database
	loaded, err := store.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if loaded.ID != sess.ID {
		t.Errorf("expected session ID %q, got %q", sess.ID, loaded.ID)
	}
	if loaded.Revision != 2 {
		t.Errorf("expected revision 2, got %d", loaded.Revision)
	}
	if loaded.CurrentStageIndex != 1 {
		t.Errorf("expected current stage index 1, got %d", loaded.CurrentStageIndex)
	}
	if len(loaded.Stages) != 2 {
		t.Fatalf("expected 2 stages, got %d", len(loaded.Stages))
	}
	if loaded.Stages[0].Status != drill.StageStatusCompleted {
		t.Errorf("expected stage 0 completed, got %v", loaded.Stages[0].Status)
	}
	if !loaded.Stages[0].FirstTryCorrect {
		t.Errorf("expected stage 0 FirstTryCorrect true")
	}
	if len(loaded.Stages[0].Attempts) != 1 {
		t.Fatalf("expected 1 attempt on stage 0, got %d", len(loaded.Stages[0].Attempts))
	}
	att := loaded.Stages[0].Attempts[0]
	if att.ID != "att_1" || !att.IsCorrect || att.SubmittedAnswer.OptionID != "opt_a" {
		t.Errorf("attempt mismatch: %+v", att)
	}
	if loaded.Stages[1].Status != drill.StageStatusActive {
		t.Errorf("expected stage 1 active, got %v", loaded.Stages[1].Status)
	}

	// Verify immutable QuestionInstance snapshot is preserved
	if loaded.QuestionInstance.ID != sess.QuestionInstance.ID {
		t.Errorf("expected instance ID %q, got %q", sess.QuestionInstance.ID, loaded.QuestionInstance.ID)
	}
	if loaded.QuestionInstance.Title != tmpl.Title {
		t.Errorf("expected title %q, got %q", tmpl.Title, loaded.QuestionInstance.Title)
	}
	if len(loaded.QuestionInstance.Stages) != 2 {
		t.Errorf("expected 2 stage instances, got %d", len(loaded.QuestionInstance.Stages))
	}
}

func TestStoreGetLatestActiveSession(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	tmpl := createTestTemplate()

	// When no sessions exist, should return ErrNotFound
	_, err := store.GetLatestActiveSession(ctx)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound for empty db, got %v", err)
	}

	// Create first session (will be completed)
	sm := drill.NewSessionManagerWithStore(store, nil)
	sess1, err := sm.CreateSession(tmpl, 1)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sess1.Completed = true
	sess1.UpdatedAt = time.Now().UTC().Add(-10 * time.Minute)
	_ = store.SaveSession(ctx, sess1)

	// Create second session (active, incomplete)
	sess2, err := sm.CreateSession(tmpl, 2)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sess2.Completed = false
	sess2.UpdatedAt = time.Now().UTC().Add(-5 * time.Minute)
	_ = store.SaveSession(ctx, sess2)

	// GetLatestActiveSession should prefer incomplete session (sess2)
	latest, err := store.GetLatestActiveSession(ctx)
	if err != nil {
		t.Fatalf("GetLatestActiveSession failed: %v", err)
	}
	if latest.ID != sess2.ID {
		t.Errorf("expected active session %q, got %q", sess2.ID, latest.ID)
	}

	// If sess2 completes, and no other session is active, it should return sess2 as latest completed
	sess2.Completed = true
	sess2.UpdatedAt = time.Now().UTC()
	_ = store.SaveSession(ctx, sess2)

	latest, err = store.GetLatestActiveSession(ctx)
	if err != nil {
		t.Fatalf("GetLatestActiveSession failed: %v", err)
	}
	if latest.ID != sess2.ID {
		t.Errorf("expected latest completed session %q, got %q", sess2.ID, latest.ID)
	}
}

func TestStoreDraftsAndSettings(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Settings
	if err := store.SaveSettings(ctx, "theme", `{"dark":true}`); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}
	val, err := store.GetSettings(ctx, "theme")
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}
	if val != `{"dark":true}` {
		t.Errorf("expected setting %q, got %q", `{"dark":true}`, val)
	}

	// Drafts
	draft := SessionDraft{
		SessionID: "sess_100",
		StageID:   "stage_3",
		DraftAnswer: domain.SubmittedAnswer{
			Kind:       domain.StageKindNumeric,
			NumericRaw: "0.375",
		},
		ActivePosition: 2,
		Revision:       5,
		UpdatedAt:      time.Now().UTC(),
	}

	// Save session first to satisfy foreign key
	sess := &drill.DrillSession{
		ID:               "sess_100",
		TemplateID:       "tmpl_1",
		QuestionInstance: domain.QuestionInstance{ID: "inst_100"},
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	if err := store.SaveSession(ctx, sess); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	if err := store.SaveSessionDraft(ctx, draft); err != nil {
		t.Fatalf("SaveSessionDraft failed: %v", err)
	}

	retrieved, err := store.GetSessionDraft(ctx, "sess_100", "stage_3")
	if err != nil {
		t.Fatalf("GetSessionDraft failed: %v", err)
	}
	if retrieved.DraftAnswer.NumericRaw != "0.375" || retrieved.Revision != 5 {
		t.Errorf("draft mismatch: %+v", retrieved)
	}

	if err := store.ClearSessionDraft(ctx, "sess_100", "stage_3"); err != nil {
		t.Fatalf("ClearSessionDraft failed: %v", err)
	}

	_, err = store.GetSessionDraft(ctx, "sess_100", "stage_3")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after clear, got %v", err)
	}
}
