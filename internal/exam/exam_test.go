package exam_test

import (
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/exam"
)

func makeTestTemplate() *domain.QuestionTemplate {
	val := 0.375
	hint := "Remember combinations formula"
	return &domain.QuestionTemplate{
		ID:               "test_binomial",
		Version:          1,
		Status:           domain.StatusActive,
		Title:            "Test Coin Toss",
		ModuleID:         "module_01",
		FamilyID:         "binomial_pmf",
		ScenarioMarkdown: "A fair coin is tossed 4 times.",
		Approval: &domain.ApprovalRecord{
			Reviewer:   "Test Reviewer",
			ReviewedAt: time.Now().UTC(),
			Notes:      "Verified for tests",
		},
		Stages: []domain.StageTemplate{
			{
				ID:             "stage_choice",
				Kind:           domain.StageKindChoice,
				PromptMarkdown: "What distribution model applies?",
				Options: []domain.Option{
					{ID: "opt_binomial", TextMarkdown: "Binomial", HintMarkdown: &hint},
					{ID: "opt_poisson", TextMarkdown: "Poisson", HintMarkdown: &hint},
				},
				ExpectedAnswer: domain.ExpectedAnswer{
					Kind:     domain.AnswerKindChoice,
					OptionID: "opt_binomial",
				},
				ExplanationMarkdown: "Binary outcomes with constant probability give a Binomial model.",
			},
			{
				ID:             "stage_numeric",
				Kind:           domain.StageKindNumeric,
				PromptMarkdown: "What is P(X = 2)?",
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
				ExplanationMarkdown: "6 * (0.5)^4 = 0.375 (3/8).",
			},
		},
	}
}

func TestNewExamValidationAndSnapshotIsolation(t *testing.T) {
	tmpl := makeTestTemplate()
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)

	// Valid creation
	cfg := exam.ExamConfig{
		Title:           "Midterm Assessment",
		QuestionCount:   1,
		DurationMinutes: 30,
		Seed:            42,
	}
	rec, err := exam.New("exam_001", cfg, []*domain.QuestionTemplate{tmpl}, now)
	if err != nil {
		t.Fatalf("unexpected error creating exam: %v", err)
	}

	if rec.ID != "exam_001" || rec.Status != exam.StatusActive || rec.Revision != 1 {
		t.Errorf("unexpected initial exam record state: %+v", rec)
	}
	if rec.DeadlineAt == nil || !rec.DeadlineAt.Equal(now.Add(30*time.Minute)) {
		t.Errorf("unexpected deadline: %v", rec.DeadlineAt)
	}
	if len(rec.Questions) != 1 || len(rec.Questions[0].Stages) != 2 {
		t.Fatalf("unexpected question/stage structure in exam")
	}

	// Verify template mutation does not mutate stored exam snapshot
	tmpl.Title = "Mutated Bank Title"
	if rec.Questions[0].Title == "Mutated Bank Title" {
		t.Errorf("exam question snapshot was not isolated from bank template mutation")
	}

	// Rejection of unapproved template
	unapproved := makeTestTemplate()
	unapproved.Status = domain.StatusDraft
	unapproved.Approval = nil
	_, err = exam.New("exam_bad", cfg, []*domain.QuestionTemplate{unapproved}, now)
	if err == nil {
		t.Errorf("expected error when creating exam from unapproved template")
	}
}

func TestPublicViewWithholdingWhileActive(t *testing.T) {
	tmpl := makeTestTemplate()
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	rec, err := exam.New("exam_withhold", exam.ExamConfig{QuestionCount: 1, DurationMinutes: 20}, []*domain.QuestionTemplate{tmpl}, now)
	if err != nil {
		t.Fatalf("failed to create exam: %v", err)
	}

	view := exam.Public(rec, now)
	if view.Status != exam.StatusActive {
		t.Errorf("expected active status, got %s", view.Status)
	}
	if view.Report != nil {
		t.Errorf("expected report to be nil while exam is active")
	}

	q := view.Questions[0]
	for _, s := range q.Stages {
		// Verify options are sanitized PublicOptionView
		for _, opt := range s.Options {
			if opt.ID == "" || opt.TextMarkdown == "" {
				t.Errorf("unexpected public option: %+v", opt)
			}
		}
	}
}

func TestExamLifecycleAndGrading(t *testing.T) {
	tmpl := makeTestTemplate()
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	rec, err := exam.New("exam_flow", exam.ExamConfig{QuestionCount: 1, DurationMinutes: 0}, []*domain.QuestionTemplate{tmpl}, now)
	if err != nil {
		t.Fatalf("failed to create exam: %v", err)
	}

	// 1. Save answer for choice stage
	qID := rec.Questions[0].ID
	cmd1 := exam.ExamCommand{
		CommandID:        "cmd_ans1",
		ExpectedRevision: 1,
		Type:             "save_answer",
		StageID:          "stage_choice",
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "opt_binomial",
		},
	}
	rec, err = exam.Apply(rec, cmd1, now.Add(1*time.Minute))
	if err != nil {
		t.Fatalf("save_answer failed: %v", err)
	}
	if rec.Revision != 2 {
		t.Errorf("expected revision 2, got %d", rec.Revision)
	}
	if ans, ok := rec.Responses[qID+":stage_choice"]; !ok || ans.OptionID != "opt_binomial" {
		t.Errorf("response not recorded properly: %+v", rec.Responses)
	}

	// 2. Revision conflict check
	badCmd := exam.ExamCommand{
		CommandID:        "cmd_conflict",
		ExpectedRevision: 1, // stale
		Type:             "save_answer",
		StageID:          "stage_choice",
	}
	_, err = exam.Apply(rec, badCmd, now)
	if err != exam.ErrConflict {
		t.Errorf("expected ErrConflict, got %v", err)
	}

	// 3. Save answer for numeric stage (as fraction 3/8)
	cmd2 := exam.ExamCommand{
		CommandID:        "cmd_ans2",
		ExpectedRevision: 2,
		Type:             "save_answer",
		StageID:          "stage_numeric",
		Answer: &domain.SubmittedAnswer{
			Kind:       domain.StageKindNumeric,
			NumericRaw: "3/8",
		},
	}
	rec, err = exam.Apply(rec, cmd2, now.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("save_answer numeric failed: %v", err)
	}
	if rec.Revision != 3 {
		t.Errorf("expected revision 3, got %d", rec.Revision)
	}

	// 4. Submit exam
	cmdSubmit := exam.ExamCommand{
		CommandID:        "cmd_submit",
		ExpectedRevision: 3,
		Type:             "submit_exam",
	}
	rec, err = exam.Apply(rec, cmdSubmit, now.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("submit_exam failed: %v", err)
	}

	if rec.Status != exam.StatusCompleted {
		t.Errorf("expected completed status, got %s", rec.Status)
	}
	if rec.Report == nil {
		t.Fatalf("expected report to be populated after submission")
	}

	rep := rec.Report
	if rep.CorrectStages != 2 || rep.TotalStages != 2 || rep.ScorePercent != 100.0 {
		t.Errorf("unexpected score report: %+v", rep)
	}

	// Public view should now include full report and feedback
	pubView := exam.Public(rec, now.Add(4*time.Minute))
	if pubView.Status != exam.StatusCompleted || pubView.Report == nil {
		t.Fatalf("public view does not expose report for completed exam")
	}

	// Generating markdown report
	md := exam.FormatMarkdownReport(rec)
	if len(md) == 0 || !pubView.Report.QuestionReports[0].IsFullyCorrect {
		t.Errorf("markdown report empty or unexpected")
	}
}

func TestTimedExamDeadlineExpiration(t *testing.T) {
	tmpl := makeTestTemplate()
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)

	// Create 10-minute exam
	rec, err := exam.New("exam_timed", exam.ExamConfig{QuestionCount: 1, DurationMinutes: 10}, []*domain.QuestionTemplate{tmpl}, now)
	if err != nil {
		t.Fatalf("failed to create exam: %v", err)
	}

	// Time check before deadline: remaining time positive
	vBefore := exam.Public(rec, now.Add(5*time.Minute))
	if vBefore.Status != exam.StatusActive || vBefore.TimeRemainingSeconds == nil || *vBefore.TimeRemainingSeconds != 300 {
		t.Errorf("unexpected remaining time before deadline: %+v", vBefore.TimeRemainingSeconds)
	}

	// Time check after deadline: dynamically expires
	vAfter := exam.Public(rec, now.Add(11*time.Minute))
	if vAfter.Status != exam.StatusExpired || vAfter.Report == nil {
		t.Errorf("expected expired status and report after deadline, got %s", vAfter.Status)
	}

	// Attempting command after deadline transitions record to StatusExpired
	cmdAfter := exam.ExamCommand{
		CommandID:        "cmd_late",
		ExpectedRevision: 1,
		Type:             "save_answer",
		StageID:          "stage_choice",
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "opt_binomial",
		},
	}
	expiredRec, err := exam.Apply(rec, cmdAfter, now.Add(12*time.Minute))
	if err != nil {
		t.Fatalf("unexpected error on late command: %v", err)
	}
	if expiredRec.Status != exam.StatusExpired || expiredRec.Report == nil {
		t.Errorf("expected record to be expired with report, got %s", expiredRec.Status)
	}
}

func TestAbandonExam(t *testing.T) {
	tmpl := makeTestTemplate()
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	rec, err := exam.New("exam_abandon", exam.ExamConfig{QuestionCount: 1}, []*domain.QuestionTemplate{tmpl}, now)
	if err != nil {
		t.Fatalf("failed to create exam: %v", err)
	}

	cmdAbandon := exam.ExamCommand{
		CommandID:        "cmd_ab",
		ExpectedRevision: 1,
		Type:             "abandon_exam",
	}
	rec, err = exam.Apply(rec, cmdAbandon, now.Add(1*time.Minute))
	if err != nil {
		t.Fatalf("abandon_exam failed: %v", err)
	}
	if rec.Status != exam.StatusAbandoned || rec.Report == nil {
		t.Errorf("expected abandoned status with report, got %s", rec.Status)
	}
}
