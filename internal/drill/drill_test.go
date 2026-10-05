package drill

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/domain"
)

func loadApprovedBinomialTemplate(t *testing.T) *domain.QuestionTemplate {
	t.Helper()
	path := filepath.Join("..", "..", "curriculum", "approved", "binomial-fair-coin-exactly-two.json")
	tmpl, err := bank.ValidateTemplateFile(path, nil)
	if err != nil {
		t.Fatalf("failed to load approved binomial template from %s: %v", path, err)
	}
	return tmpl
}

func TestDrillSessionCreation(t *testing.T) {
	tmpl := loadApprovedBinomialTemplate(t)
	mgr := NewSessionManager(nil)

	session, err := mgr.CreateSession(tmpl, 42)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if session.ID == "" {
		t.Errorf("expected non-empty session ID")
	}
	if len(session.Stages) != 7 {
		t.Fatalf("expected 7 stages, got %d", len(session.Stages))
	}
	if session.Stages[0].Status != StageStatusActive {
		t.Errorf("expected stage 0 to be active, got %s", session.Stages[0].Status)
	}
	for i := 1; i < 7; i++ {
		if session.Stages[i].Status != StageStatusUnvisited {
			t.Errorf("expected stage %d to be unvisited, got %s", i, session.Stages[i].Status)
		}
	}

	pub := session.ToPublicView()
	if pub.Completed {
		t.Errorf("new session should not be completed")
	}
	if pub.Stages[0].ExpectedAnswer != nil {
		t.Errorf("active unresolved stage must withhold expected answer")
	}
	if pub.Stages[0].ExplanationMarkdown != "" {
		t.Errorf("active unresolved stage must withhold explanation")
	}
}

func TestSevenStageDrillHappyPath(t *testing.T) {
	tmpl := loadApprovedBinomialTemplate(t)
	mgr := NewSessionManager(func() time.Time { return time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC) })
	session, err := mgr.CreateSession(tmpl, 0)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// 7 stages happy path submissions
	submissions := []domain.SubmittedAnswer{
		{Kind: domain.StageKindChoice, OptionID: "count_heads"},
		{Kind: domain.StageKindChoice, OptionID: "binomial"},
		{Kind: domain.StageKindChoice, OptionID: "n4_p_half"},
		{Kind: domain.StageKindChoice, OptionID: "equal_two"},
		{Kind: domain.StageKindChoice, OptionID: "with_combination"},
		{Kind: domain.StageKindNumeric, NumericRaw: "0.375"},
		{Kind: domain.StageKindChoice, OptionID: "whole_experiment"},
	}

	for i, sub := range submissions {
		cmdID := "cmd_" + string(rune('a'+i))
		res, err := session.ExecuteCommand(SessionCommand{
			CommandID:        cmdID,
			ExpectedRevision: session.Revision,
			Type:             CmdSubmitAnswer,
			Answer:           &sub,
		}, nil)

		if err != nil {
			t.Fatalf("stage %d command failed: %v", i+1, err)
		}
		if !res.Success {
			t.Fatalf("stage %d expected success, got error %s", i+1, res.ErrorMessage)
		}
	}

	if !session.Completed {
		t.Fatalf("expected drill to be completed after 7 successful stages")
	}

	pub := session.ToPublicView()
	if pub.Recap == nil {
		t.Fatalf("expected completed session to have recap")
	}
	recap := pub.Recap
	if recap.TotalStages != 7 {
		t.Errorf("expected 7 total stages in recap, got %d", recap.TotalStages)
	}
	if recap.FirstTryCount != 7 {
		t.Errorf("expected 7 first try correct stages, got %d", recap.FirstTryCount)
	}
	if recap.CanonicalDerivation == nil {
		t.Fatalf("expected canonical derivation in recap")
	}
	deriv := recap.CanonicalDerivation
	if deriv.CanonicalProbability != 0.375 {
		t.Errorf("expected canonical probability 0.375, got %v", deriv.CanonicalProbability)
	}
	if deriv.CanonicalRational == nil || deriv.CanonicalRational.RatString() != "3/8" {
		t.Errorf("expected canonical rational 3/8, got %v", deriv.CanonicalRational)
	}
	if deriv.Mean != 2.0 || deriv.Variance != 1.0 || deriv.StdDev != 1.0 {
		t.Errorf("expected moments mean=2, var=1, sd=1, got mean=%v, var=%v, sd=%v", deriv.Mean, deriv.Variance, deriv.StdDev)
	}
}

func TestNumericGradingVariants(t *testing.T) {
	tmpl := loadApprovedBinomialTemplate(t)

	tests := []struct {
		name       string
		raw        string
		wantValid  bool
		wantCorr   bool
		diagnostic string
	}{
		{name: "canonical_decimal", raw: "0.375", wantValid: true, wantCorr: true},
		{name: "trailing_zero_decimal", raw: "0.375000", wantValid: true, wantCorr: true},
		{name: "explicit_percent", raw: "37.5%", wantValid: true, wantCorr: true},
		{name: "fraction", raw: "3/8", wantValid: true, wantCorr: true},
		{name: "unsimplified_fraction", raw: "6/16", wantValid: true, wantCorr: true},
		{name: "outside_tolerance", raw: "0.38", wantValid: true, wantCorr: false},
		{name: "bare_percentage_semantics", raw: "37.5", wantValid: true, wantCorr: false, diagnostic: "Entered 37.5. Did you mean 37.5%?"},
		{name: "ambiguous_comma", raw: "0,375", wantValid: false, wantCorr: false},
		{name: "invalid_syntax", raw: "abc", wantValid: false, wantCorr: false},
		{name: "empty", raw: "", wantValid: false, wantCorr: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mgr := NewSessionManager(nil)
			session, err := mgr.CreateSession(tmpl, 0)
			if err != nil {
				t.Fatalf("CreateSession failed: %v", err)
			}

			// Advance to stage 5 (index 5 = Stage 6 Calculate)
			session.CurrentStageIndex = 5
			session.Stages[5].Status = StageStatusActive

			res, err := session.ExecuteCommand(SessionCommand{
				CommandID: "test_num_" + tc.name,
				Type:      CmdSubmitAnswer,
				Answer: &domain.SubmittedAnswer{
					Kind:       domain.StageKindNumeric,
					NumericRaw: tc.raw,
				},
			}, nil)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !tc.wantValid {
				if !res.InvalidInput {
					t.Errorf("expected InvalidInput = true for raw %q", tc.raw)
				}
				if len(session.Stages[5].Attempts) != 0 {
					t.Errorf("invalid input must not generate an attempt, got %d", len(session.Stages[5].Attempts))
				}
				return
			}

			if res.InvalidInput {
				t.Fatalf("expected valid input, got invalid: %s", res.ErrorMessage)
			}

			if len(session.Stages[5].Attempts) != 1 {
				t.Fatalf("expected 1 attempt, got %d", len(session.Stages[5].Attempts))
			}

			att := session.Stages[5].Attempts[0]
			if att.IsCorrect != tc.wantCorr {
				t.Errorf("raw %q: expected IsCorrect=%v, got %v", tc.raw, tc.wantCorr, att.IsCorrect)
			}
		})
	}
}

func TestChoiceDistractorTargetingAndRetry(t *testing.T) {
	tmpl := loadApprovedBinomialTemplate(t)
	mgr := NewSessionManager(nil)
	session, err := mgr.CreateSession(tmpl, 0)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Stage 1: choose distractor "chance_one_head" (misconception: parameter_as_random_variable)
	res1, err := session.ExecuteCommand(SessionCommand{
		CommandID: "cmd_err1",
		Type:      CmdSubmitAnswer,
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "chance_one_head",
		},
	}, nil)

	if err != nil {
		t.Fatalf("unexpected error on err1: %v", err)
	}
	if !res1.Success {
		t.Fatalf("expected success response with incorrect feedback")
	}

	st := session.Stages[0]
	if st.Status != StageStatusRetry {
		t.Errorf("expected stage to transition to retry, got %s", st.Status)
	}
	if st.MisconceptionID == nil || *st.MisconceptionID != "parameter_as_random_variable" {
		t.Errorf("expected misconception parameter_as_random_variable, got %v", st.MisconceptionID)
	}
	if st.ActiveHint == "" {
		t.Errorf("expected non-empty active hint on first error")
	}

	// Retry: choose correct answer "count_heads"
	res2, err := session.ExecuteCommand(SessionCommand{
		CommandID: "cmd_retry",
		Type:      CmdSubmitAnswer,
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "count_heads",
		},
	}, nil)

	if err != nil {
		t.Fatalf("unexpected error on retry: %v", err)
	}
	if !res2.Success {
		t.Fatalf("expected success on retry")
	}

	stAfter := session.Stages[0]
	if stAfter.Status != StageStatusCompleted {
		t.Errorf("expected stage to be completed, got %s", stAfter.Status)
	}
	if !stAfter.SolvedOnRetry {
		t.Errorf("expected SolvedOnRetry = true")
	}
	if stAfter.FirstTryCorrect {
		t.Errorf("FirstTryCorrect should be false")
	}
}

func TestTwoErrorsRevealAnswer(t *testing.T) {
	tmpl := loadApprovedBinomialTemplate(t)
	mgr := NewSessionManager(nil)
	session, err := mgr.CreateSession(tmpl, 0)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Advance to stage 3 (translate_event)
	session.CurrentStageIndex = 3
	session.Stages[3].Status = StageStatusActive

	// Error 1
	_, err = session.ExecuteCommand(SessionCommand{
		CommandID: "cmd_err1",
		Type:      CmdSubmitAnswer,
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "at_most_two",
		},
	}, nil)
	if err != nil {
		t.Fatalf("err1 failed: %v", err)
	}

	// Error 2
	_, err = session.ExecuteCommand(SessionCommand{
		CommandID: "cmd_err2",
		Type:      CmdSubmitAnswer,
		Answer: &domain.SubmittedAnswer{
			Kind:     domain.StageKindChoice,
			OptionID: "at_least_two",
		},
	}, nil)
	if err != nil {
		t.Fatalf("err2 failed: %v", err)
	}

	st := session.Stages[3]
	if st.Status != StageStatusCompleted {
		t.Errorf("expected stage to be completed after 2 errors, got %s", st.Status)
	}
	if !st.Revealed {
		t.Errorf("expected Revealed = true after 2 errors")
	}

	pub := session.ToPublicView()
	if pub.Stages[3].ExpectedAnswer == nil {
		t.Errorf("expected answer to be revealed on completed stage")
	}
	if pub.Stages[3].ExpectedAnswer.OptionID != "equal_two" {
		t.Errorf("expected revealed option equal_two, got %v", pub.Stages[3].ExpectedAnswer.OptionID)
	}
}

func TestOfflineHintRequest(t *testing.T) {
	tmpl := loadApprovedBinomialTemplate(t)
	mgr := NewSessionManager(nil)
	session, err := mgr.CreateSession(tmpl, 0)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	res, err := session.ExecuteCommand(SessionCommand{
		CommandID: "cmd_hint",
		Type:      CmdRequestHint,
	}, nil)
	if err != nil {
		t.Fatalf("request hint failed: %v", err)
	}
	if !res.Success {
		t.Errorf("expected hint command to succeed")
	}

	st := session.Stages[0]
	if len(st.Assistance) == 0 || st.Assistance[0] != domain.AssistanceHint {
		t.Errorf("expected AssistanceHint to be recorded")
	}
	if st.ActiveHint == "" {
		t.Errorf("expected active hint text to be populated")
	}
}

func TestNavigationEligibility(t *testing.T) {
	tmpl := loadApprovedBinomialTemplate(t)
	mgr := NewSessionManager(nil)
	session, err := mgr.CreateSession(tmpl, 0)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Try jumping to stage 4 (unvisited)
	target := 4
	res, err := session.ExecuteCommand(SessionCommand{
		CommandID:        "cmd_jump",
		Type:             CmdNavigateStage,
		TargetStageIndex: &target,
	}, nil)

	if err == nil && res.Success {
		t.Errorf("expected jump to unvisited stage to fail")
	}

	// Complete stage 0
	_, _ = session.ExecuteCommand(SessionCommand{
		CommandID: "cmd_s0",
		Type:      CmdSubmitAnswer,
		Answer:    &domain.SubmittedAnswer{Kind: domain.StageKindChoice, OptionID: "count_heads"},
	}, nil)

	// Now stage 1 is active; learner should be able to navigate back to stage 0
	target0 := 0
	resNav, errNav := session.ExecuteCommand(SessionCommand{
		CommandID:        "cmd_nav_back",
		Type:             CmdNavigateStage,
		TargetStageIndex: &target0,
	}, nil)

	if errNav != nil || !resNav.Success {
		t.Errorf("expected navigating back to stage 0 to succeed, got %v", errNav)
	}
	if session.CurrentStageIndex != 0 {
		t.Errorf("expected current stage index to be 0, got %d", session.CurrentStageIndex)
	}
}

func TestCommandIdempotencyAndRevision(t *testing.T) {
	tmpl := loadApprovedBinomialTemplate(t)
	mgr := NewSessionManager(nil)
	session, err := mgr.CreateSession(tmpl, 0)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	cmd := SessionCommand{
		CommandID:        "cmd_idem_1",
		ExpectedRevision: 1,
		Type:             CmdSubmitAnswer,
		Answer:           &domain.SubmittedAnswer{Kind: domain.StageKindChoice, OptionID: "count_heads"},
	}

	res1, err1 := session.ExecuteCommand(cmd, nil)
	if err1 != nil || !res1.Success {
		t.Fatalf("first command execution failed: %v", err1)
	}

	revAfter := session.Revision

	// Resend exact duplicate command
	res2, err2 := session.ExecuteCommand(cmd, nil)
	if err2 != nil || !res2.Success {
		t.Fatalf("duplicate command execution failed: %v", err2)
	}
	if session.Revision != revAfter {
		t.Errorf("duplicate command must not advance revision, was %d now %d", revAfter, session.Revision)
	}

	// Send command with stale revision
	staleCmd := SessionCommand{
		CommandID:        "cmd_stale",
		ExpectedRevision: 1, // Stale! Current is revAfter
		Type:             CmdRequestHint,
	}
	_, errStale := session.ExecuteCommand(staleCmd, nil)
	if errStale != ErrRevisionConflict {
		t.Errorf("expected ErrRevisionConflict, got %v", errStale)
	}
}

func TestMultiQuestionSessionCreationAndNavigation(t *testing.T) {
	tmpl1 := loadApprovedBinomialTemplate(t)
	// Create a cloned template with another ID
	tmpl2 := *tmpl1
	tmpl2.ID = "binomial_second"
	tmpl2.Title = "Second problem"

	mgr := NewSessionManager(nil)
	settings := domain.SessionSettings{
		QuestionCount: 2,
		Intensity:     "standard",
	}

	session, err := mgr.CreateMultiQuestionSession([]*domain.QuestionTemplate{tmpl1, &tmpl2}, settings, 42)
	if err != nil {
		t.Fatalf("CreateMultiQuestionSession failed: %v", err)
	}

	if len(session.Questions) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(session.Questions))
	}
	if session.CurrentQuestionIndex != 0 {
		t.Errorf("expected starting question index 0, got %d", session.CurrentQuestionIndex)
	}

	pv := session.ToPublicView()
	if pv.TotalQuestions != 2 {
		t.Errorf("expected TotalQuestions = 2, got %d", pv.TotalQuestions)
	}
	if len(pv.Questions) != 2 {
		t.Fatalf("expected 2 questions in public view, got %d", len(pv.Questions))
	}
	if pv.Questions[0].Status != "in_progress" {
		t.Errorf("expected question 0 status in_progress, got %s", pv.Questions[0].Status)
	}
	if pv.Questions[1].Status != "pending" {
		t.Errorf("expected question 1 status pending, got %s", pv.Questions[1].Status)
	}

	// Answer question 0 stage 0
	_, err = session.ExecuteCommand(SessionCommand{
		CommandID: "cmd_q0_s0",
		Type:      CmdSubmitAnswer,
		Answer:    &domain.SubmittedAnswer{Kind: domain.StageKindChoice, OptionID: "count_heads"},
	}, nil)
	if err != nil {
		t.Fatalf("submit q0_s0 failed: %v", err)
	}

	// Navigate to question 1
	targetQ := 1
	navRes, errNav := session.ExecuteCommand(SessionCommand{
		CommandID:           "cmd_nav_q1",
		Type:                CmdNavigateQuestion,
		TargetQuestionIndex: &targetQ,
	}, nil)
	if errNav != nil || !navRes.Success {
		t.Fatalf("navigate to question 1 failed: %v", errNav)
	}
	if session.CurrentQuestionIndex != 1 {
		t.Fatalf("expected CurrentQuestionIndex = 1, got %d", session.CurrentQuestionIndex)
	}
	if session.QuestionInstance.ID == session.Questions[0].QuestionInstance.ID {
		t.Errorf("expected active QuestionInstance to switch to question 1")
	}

	// Verify public view reflects current question
	pvAfter := session.ToPublicView()
	if pvAfter.CurrentQuestionIndex != 1 {
		t.Errorf("expected pv CurrentQuestionIndex = 1, got %d", pvAfter.CurrentQuestionIndex)
	}
	if pvAfter.Questions[0].Status != "in_progress" {
		t.Errorf("expected question 0 status in_progress, got %s", pvAfter.Questions[0].Status)
	}
	if pvAfter.Questions[1].Status != "in_progress" {
		t.Errorf("expected question 1 status in_progress, got %s", pvAfter.Questions[1].Status)
	}

	// Navigate back to question 0
	targetQ0 := 0
	navRes0, errNav0 := session.ExecuteCommand(SessionCommand{
		CommandID:           "cmd_nav_q0",
		Type:                CmdNavigateQuestion,
		TargetQuestionIndex: &targetQ0,
	}, nil)
	if errNav0 != nil || !navRes0.Success {
		t.Fatalf("navigate back to question 0 failed: %v", errNav0)
	}
	if session.CurrentQuestionIndex != 0 {
		t.Fatalf("expected CurrentQuestionIndex = 0, got %d", session.CurrentQuestionIndex)
	}
	// Question 0 Stage 0 should still be completed!
	if session.Stages[0].Status != StageStatusCompleted {
		t.Errorf("expected question 0 stage 0 to remain completed, got %s", session.Stages[0].Status)
	}
}

