package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/drill"
	"github.com/trustdan/quant-methods-practice/internal/mastery"
)

func setupStorageTestDB(t *testing.T) (*Store, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "practice.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	ctx := context.Background()
	if err := RunMigrations(ctx, db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}
	store := NewStore(db, nil)
	return store, func() { _ = db.Close() }
}

func TestMasteryProjectionSaveAndRetrieve(t *testing.T) {
	store, cleanup := setupStorageTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)

	proj := &mastery.ConceptMastery{
		ConceptID:               "concept_calc",
		PolicyVersion:           1,
		IndependentSuccesses:    3,
		IndependentErrors:       1,
		AssistedCount:           2,
		TotalEvidenceCount:      4,
		LastTestedAt:            &now,
		BaseScore:               0.67,
		DecayedScore:            0.65,
		RetentionFactor:         0.97,
		HalfLifeDays:            3.0,
		Status:                  mastery.StatusLearning,
		ScaffoldLevel:           mastery.ScaffoldFull,
		SettingGroupsSeen:       []string{"group_a"},
		DelayedTransferAchieved: false,
		RecentError:             false,
		PriorityScore:           1.8,
	}

	if err := store.SaveMasteryProjection(ctx, proj); err != nil {
		t.Fatalf("SaveMasteryProjection failed: %v", err)
	}

	all, err := store.GetAllMasteryProjections(ctx)
	if err != nil {
		t.Fatalf("GetAllMasteryProjections failed: %v", err)
	}

	retrieved, ok := all["concept_calc"]
	if !ok {
		t.Fatalf("expected concept_calc in projections")
	}

	if retrieved.IndependentSuccesses != 3 || retrieved.IndependentErrors != 1 {
		t.Errorf("expected 3 successes, 1 error, got %d and %d",
			retrieved.IndependentSuccesses, retrieved.IndependentErrors)
	}
	if retrieved.TotalEvidenceCount != 4 {
		t.Errorf("expected TotalEvidenceCount 4, got %d", retrieved.TotalEvidenceCount)
	}
}

func TestRebuildMasteryFromHistoricalSessions(t *testing.T) {
	store, cleanup := setupStorageTestDB(t)
	defer cleanup()

	ctx := context.Background()
	clockTime := time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)
	clock := func() time.Time { return clockTime }

	mgr := drill.NewSessionManagerWithStore(store, clock)

	// Create a template with designated evidence concepts
	tmpl := &domain.QuestionTemplate{
		ID:           "tmpl_test_1",
		Version:      1,
		Title:        "Test Question 1",
		SettingGroup: "group_alpha",
		ConceptIDs:   []string{"concept_1", "concept_2"},
		Stages: []domain.StageTemplate{
			{
				ID:                 "st_1",
				Kind:               domain.StageKindChoice,
				EvidenceConceptIDs: []string{"concept_1"},
				Options: []domain.Option{
					{ID: "opt_correct", TextMarkdown: "Correct"},
					{ID: "opt_wrong", TextMarkdown: "Wrong"},
				},
				ExpectedAnswer: domain.ExpectedAnswer{Kind: domain.AnswerKindChoice, OptionID: "opt_correct"},
			},
			{
				ID:                 "st_2",
				Kind:               domain.StageKindChoice,
				EvidenceConceptIDs: []string{"concept_2"},
				Options: []domain.Option{
					{ID: "opt_correct", TextMarkdown: "Correct"},
					{ID: "opt_wrong", TextMarkdown: "Wrong"},
				},
				ExpectedAnswer: domain.ExpectedAnswer{Kind: domain.AnswerKindChoice, OptionID: "opt_correct"},
			},
		},
	}

	session, err := mgr.CreateMultiQuestionSession([]*domain.QuestionTemplate{tmpl}, domain.SessionSettings{QuestionCount: 1}, 10)
	if err != nil {
		t.Fatalf("CreateMultiQuestionSession failed: %v", err)
	}

	// Stage 1: Independent success on concept_1
	subCorrect := domain.SubmittedAnswer{Kind: domain.StageKindChoice, OptionID: "opt_correct"}
	_, err = session.ExecuteCommand(drill.SessionCommand{
		CommandID: "c1",
		Type:      drill.CmdSubmitAnswer,
		Answer:    &subCorrect,
	}, clock)
	if err != nil {
		t.Fatalf("execute c1 failed: %v", err)
	}

	// Stage 2: Attempt 1 is WRONG on concept_2
	subWrong := domain.SubmittedAnswer{Kind: domain.StageKindChoice, OptionID: "opt_wrong"}
	_, err = session.ExecuteCommand(drill.SessionCommand{
		CommandID: "c2_wrong",
		Type:      drill.CmdSubmitAnswer,
		Answer:    &subWrong,
	}, clock)
	if err != nil {
		t.Fatalf("execute c2_wrong failed: %v", err)
	}

	// Stage 2: Attempt 2 on retry is CORRECT
	clockTime = clockTime.Add(1 * time.Minute)
	_, err = session.ExecuteCommand(drill.SessionCommand{
		CommandID: "c2_retry",
		Type:      drill.CmdSubmitAnswer,
		Answer:    &subCorrect,
	}, clock)
	if err != nil {
		t.Fatalf("execute c2_retry failed: %v", err)
	}

	// Now RebuildMastery from SQLite
	ledger := mastery.NewEvidenceLedger(3.0)
	rebuilt, err := store.RebuildMastery(ctx, ledger, clockTime.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("RebuildMastery failed: %v", err)
	}

	// concept_1: 1 independent success, 0 errors, 1 total evidence
	c1, ok := rebuilt["concept_1"]
	if !ok {
		t.Fatalf("expected concept_1 in rebuilt projections")
	}
	if c1.IndependentSuccesses != 1 || c1.IndependentErrors != 0 {
		t.Errorf("concept_1 expected 1 success, 0 errors, got %d and %d",
			c1.IndependentSuccesses, c1.IndependentErrors)
	}
	if c1.TotalEvidenceCount != 1 {
		t.Errorf("concept_1 expected TotalEvidenceCount 1, got %d", c1.TotalEvidenceCount)
	}

	// concept_2: Hinted retry never erases first error!
	// 0 independent successes, 1 independent error, 1 total evidence
	c2, ok := rebuilt["concept_2"]
	if !ok {
		t.Fatalf("expected concept_2 in rebuilt projections")
	}
	if c2.IndependentErrors != 1 || c2.IndependentSuccesses != 0 {
		t.Errorf("concept_2 expected 1 error, 0 successes (retry does not erase first error), got %d and %d",
			c2.IndependentErrors, c2.IndependentSuccesses)
	}
	if c2.TotalEvidenceCount != 1 {
		t.Errorf("concept_2 expected TotalEvidenceCount 1, got %d", c2.TotalEvidenceCount)
	}

	// Test GetMasterySummary
	summary, err := store.GetMasterySummary(ctx, ledger, clockTime.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("GetMasterySummary failed: %v", err)
	}
	if len(summary.Concepts) != 2 {
		t.Errorf("expected 2 concepts in summary, got %d", len(summary.Concepts))
	}
}
