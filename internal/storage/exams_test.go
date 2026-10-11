package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/exam"
)

func makeTestExamTemplate() *domain.QuestionTemplate {
	val := 0.375
	return &domain.QuestionTemplate{
		ID:               "tmpl_storage_exam",
		Version:          1,
		Status:           domain.StatusActive,
		Title:            "Coin Tossing Assessment",
		ModuleID:         "module_01",
		FamilyID:         "binomial_pmf",
		ScenarioMarkdown: "Coin is tossed four times.",
		Approval: &domain.ApprovalRecord{
			Reviewer:   "Test Reviewer",
			ReviewedAt: time.Now().UTC(),
			Notes:      "Verified for test",
		},
		Stages: []domain.StageTemplate{
			{
				ID:             "stage_1",
				Kind:           domain.StageKindChoice,
				PromptMarkdown: "Choose model",
				Options: []domain.Option{
					{ID: "opt_binom", TextMarkdown: "Binomial"},
					{ID: "opt_norm", TextMarkdown: "Normal"},
				},
				ExpectedAnswer: domain.ExpectedAnswer{
					Kind:     domain.AnswerKindChoice,
					OptionID: "opt_binom",
				},
				ExplanationMarkdown: "Binary independent trials.",
			},
			{
				ID:             "stage_2",
				Kind:           domain.StageKindNumeric,
				PromptMarkdown: "Calculate P(X=2)",
				ExpectedAnswer: domain.ExpectedAnswer{
					Kind:  domain.AnswerKindNumeric,
					Value: &val,
					Units: "probability",
				},
				NumericPolicy: &domain.NumericPolicy{
					Version:           1,
					AbsoluteTolerance: 1e-6,
					RelativeTolerance: 1e-6,
					AllowedForms:      []domain.NumericForm{domain.NumericFormDecimal, domain.NumericFormPercent, domain.NumericFormFraction},
				},
				ExplanationMarkdown: "6 * (0.5)^4 = 0.375",
			},
		},
	}
}

func TestExamPersistenceRestartAndIdempotency(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "exams.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err = RunMigrations(ctx, db, nil); err != nil {
		t.Fatal(err)
	}

	testClock := time.Date(2026, 10, 10, 10, 0, 0, 0, time.UTC)
	store := NewStore(db, func() time.Time { return testClock })

	tmpl := makeTestExamTemplate()
	rec, err := exam.New("exam_persist_01", exam.ExamConfig{
		Title:           "Final Exam",
		QuestionCount:   1,
		DurationMinutes: 30,
		Seed:            101,
	}, []*domain.QuestionTemplate{tmpl}, testClock)
	if err != nil {
		t.Fatalf("failed to initialize exam: %v", err)
	}

	if err := store.CreateExam(ctx, rec); err != nil {
		t.Fatalf("failed to store exam: %v", err)
	}

	// Active report cannot be fetched
	_, err = store.GetExamReport(ctx, rec.ID)
	if err == nil {
		t.Errorf("expected error fetching report for active exam")
	}

	// 1. Submit answer via CommandExam
	cmd := exam.ExamCommand{
		CommandID:        "cmd_ans_choice",
		ExpectedRevision: 1,
		Type:             "save_answer",
		StageID:          "stage_1",
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "opt_binom",
		},
	}
	next, err := store.CommandExam(ctx, rec.ID, cmd)
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}
	if next.Revision != 2 {
		t.Errorf("expected revision 2, got %d", next.Revision)
	}

	// Idempotent retry returns identical result without error
	nextReplay, err := store.CommandExam(ctx, rec.ID, cmd)
	if err != nil {
		t.Fatalf("idempotent replay failed: %v", err)
	}
	if nextReplay.Revision != 2 {
		t.Errorf("expected replayed revision 2, got %d", nextReplay.Revision)
	}

	// Restart database process
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	db, err = Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store = NewStore(db, func() time.Time { return testClock.Add(5 * time.Minute) })

	// Verify restored state after reopen
	restored, err := store.GetExam(ctx, rec.ID)
	if err != nil {
		t.Fatalf("failed to retrieve exam after restart: %v", err)
	}
	if restored.ID != rec.ID || restored.Revision != 2 || restored.Status != exam.StatusActive {
		t.Errorf("unexpected restored exam state: %+v", restored)
	}
	if restored.DeadlineAt == nil || !restored.DeadlineAt.Equal(testClock.Add(30*time.Minute)) {
		t.Errorf("deadline was not accurately restored across restart")
	}

	// 2. Submit numeric answer & finalize exam
	cmdNumeric := exam.ExamCommand{
		CommandID:        "cmd_ans_num",
		ExpectedRevision: 2,
		Type:             "save_answer",
		StageID:          "stage_2",
		Answer: &domain.SubmittedAnswer{
			Kind:       domain.StageKindNumeric,
			NumericRaw: "37.5%",
		},
	}
	next, err = store.CommandExam(ctx, rec.ID, cmdNumeric)
	if err != nil {
		t.Fatalf("save numeric answer failed: %v", err)
	}

	cmdSubmit := exam.ExamCommand{
		CommandID:        "cmd_finish",
		ExpectedRevision: 3,
		Type:             "submit_exam",
	}
	finalRec, err := store.CommandExam(ctx, rec.ID, cmdSubmit)
	if err != nil {
		t.Fatalf("submit exam failed: %v", err)
	}
	if finalRec.Status != exam.StatusCompleted || finalRec.Report == nil {
		t.Fatalf("exam not finalized properly: %+v", finalRec)
	}
	if finalRec.Report.ScorePercent != 100.0 {
		t.Errorf("expected 100%% score, got %.1f", finalRec.Report.ScorePercent)
	}

	// Report is now accessible
	report, err := store.GetExamReport(ctx, rec.ID)
	if err != nil {
		t.Fatalf("failed to get report: %v", err)
	}
	if report.CorrectStages != 2 {
		t.Errorf("expected 2 correct stages in report, got %d", report.CorrectStages)
	}

	// Check History listing
	history, err := store.GetHistory(ctx, 10, 0)
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}
	if len(history) == 0 || history[0].ID != rec.ID || history[0].Status != "completed" {
		t.Errorf("unexpected history entry: %+v", history)
	}
}

func TestDeadlineExpirationAcrossRestart(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "exams_expire.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err = RunMigrations(ctx, db, nil); err != nil {
		t.Fatal(err)
	}

	startTime := time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)
	store := NewStore(db, func() time.Time { return startTime })

	tmpl := makeTestExamTemplate()
	rec, err := exam.New("exam_expire_01", exam.ExamConfig{
		Title:           "Quick 10-Minute Quiz",
		QuestionCount:   1,
		DurationMinutes: 10,
	}, []*domain.QuestionTemplate{tmpl}, startTime)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateExam(ctx, rec); err != nil {
		t.Fatal(err)
	}

	// Close database
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	// Reopen after deadline (15 minutes later)
	db, err = Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	afterDeadline := startTime.Add(15 * time.Minute)
	store = NewStore(db, func() time.Time { return afterDeadline })

	// GetExam must recognize that deadline elapsed
	expired, err := store.GetExam(ctx, rec.ID)
	if err != nil {
		t.Fatalf("failed to fetch exam: %v", err)
	}
	if expired.Status != exam.StatusExpired {
		t.Errorf("expected status to be expired, got %s", expired.Status)
	}
	if expired.Report == nil {
		t.Errorf("expected report to be generated for expired exam")
	}

	// Attempting command on expired exam must be rejected
	cmdLate := exam.ExamCommand{
		CommandID:        "cmd_late",
		ExpectedRevision: 1,
		Type:             "save_answer",
		StageID:          "stage_1",
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "opt_binom",
		},
	}
	next, err := store.CommandExam(ctx, rec.ID, cmdLate)
	if err != nil {
		t.Fatalf("unexpected error on late command: %v", err)
	}
	if next.Status != exam.StatusExpired {
		t.Errorf("expected command outcome on expired exam to remain expired, got %s", next.Status)
	}
}
