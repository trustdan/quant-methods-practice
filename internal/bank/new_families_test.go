package bank

import (
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
)

func TestConceptCheckFamily(t *testing.T) {
	reg := DefaultRegistry()
	fam, ok := reg.Get("concept_check")
	if !ok {
		t.Fatal("concept_check family not found in DefaultRegistry")
	}

	if fam.ID() != "concept_check" {
		t.Errorf("expected ID concept_check, got %q", fam.ID())
	}

	// Parameter validation
	if err := fam.ValidateParameters(2, nil); err == nil {
		t.Error("expected error for unsupported rule_version 2")
	}

	if err := fam.ValidateParameters(1, map[string]interface{}{"unknown": 123}); err == nil {
		t.Error("expected error for unknown parameter")
	}

	if err := fam.ValidateParameters(1, map[string]interface{}{"topic": "descriptive_vs_inferential"}); err != nil {
		t.Errorf("unexpected error for valid topic: %v", err)
	}

	if err := fam.ValidateParameters(1, map[string]interface{}{}); err != nil {
		t.Errorf("unexpected error for empty parameters: %v", err)
	}

	// Template validation
	tv, ok := fam.(TemplateValidator)
	if !ok {
		t.Fatal("concept_check family does not implement TemplateValidator")
	}

	strPtr := func(s string) *string { return &s }

	validTmpl := &domain.QuestionTemplate{
		SchemaVersion:    1,
		TemplateKind:     domain.TemplateKindFixedReference,
		ID:               "concept_test_sample",
		Version:          1,
		Status:           domain.StatusActive,
		FamilyID:         "concept_check",
		RuleVersion:      1,
		ModuleID:         "module_1",
		Title:            "Sample Concept Question",
		ScenarioMarkdown: "A sample conceptual scenario.",
		Assumptions:      []string{"Valid assumptions"},
		SourceRefs:       []string{"Course Module 1"},
		SettingGroup:     "test_group",
		ConceptIDs:       []string{"concept_1"},
		Approval: &domain.ApprovalRecord{
			Reviewer:   "test_reviewer",
			ReviewedAt: time.Now(),
			Notes:      "Test notes",
		},
		Stages: []domain.StageTemplate{
			{
				ID:             "stage_1",
				Kind:           domain.StageKindChoice,
				PromptMarkdown: "Which statement is correct?",
				Options: []domain.Option{
					{ID: "opt_correct", TextMarkdown: "Correct answer", MisconceptionID: nil, HintMarkdown: nil},
					{ID: "opt_wrong", TextMarkdown: "Wrong answer", MisconceptionID: strPtr("misc_1"), HintMarkdown: strPtr("Hint for wrong")},
				},
				ExpectedAnswer:      domain.ExpectedAnswer{Kind: domain.AnswerKindChoice, OptionID: "opt_correct"},
				EvidenceConceptIDs:  []string{"concept_1"},
				ExplanationMarkdown: "Full explanation of why this is correct.",
			},
		},
	}

	if err := tv.ValidateTemplate(validTmpl); err != nil {
		t.Fatalf("unexpected template validation error: %v", err)
	}

	// Also test full ValidateTemplate on valid template
	if err := ValidateTemplate(validTmpl, reg); err != nil {
		t.Fatalf("ValidateTemplate failed on valid concept template: %v", err)
	}

	// Rejection: numeric stage
	tmplNumeric := *validTmpl
	tmplNumeric.Stages = []domain.StageTemplate{
		{
			ID:             "stage_num",
			Kind:           domain.StageKindNumeric,
			PromptMarkdown: "Enter a number",
			ExpectedAnswer: domain.ExpectedAnswer{Kind: domain.AnswerKindNumeric},
		},
	}
	if err := tv.ValidateTemplate(&tmplNumeric); err == nil {
		t.Error("expected error for numeric stage in concept_check")
	}

	// Rejection: distractor missing misconception
	tmplMissingMisc := *validTmpl
	tmplMissingMisc.Stages = []domain.StageTemplate{
		{
			ID:             "stage_1",
			Kind:           domain.StageKindChoice,
			PromptMarkdown: "Which statement is correct?",
			Options: []domain.Option{
				{ID: "opt_correct", TextMarkdown: "Correct answer", MisconceptionID: nil, HintMarkdown: nil},
				{ID: "opt_wrong", TextMarkdown: "Wrong answer", MisconceptionID: nil, HintMarkdown: strPtr("Hint")},
			},
			ExpectedAnswer:      domain.ExpectedAnswer{Kind: domain.AnswerKindChoice, OptionID: "opt_correct"},
			EvidenceConceptIDs:  []string{"concept_1"},
			ExplanationMarkdown: "Explanation",
		},
	}
	if err := tv.ValidateTemplate(&tmplMissingMisc); err == nil {
		t.Error("expected error for distractor without misconception_id")
	}

	// Rejection: distractor missing hint
	tmplMissingHint := *validTmpl
	tmplMissingHint.Stages = []domain.StageTemplate{
		{
			ID:             "stage_1",
			Kind:           domain.StageKindChoice,
			PromptMarkdown: "Which statement is correct?",
			Options: []domain.Option{
				{ID: "opt_correct", TextMarkdown: "Correct answer", MisconceptionID: nil, HintMarkdown: nil},
				{ID: "opt_wrong", TextMarkdown: "Wrong answer", MisconceptionID: strPtr("misc_1"), HintMarkdown: nil},
			},
			ExpectedAnswer:      domain.ExpectedAnswer{Kind: domain.AnswerKindChoice, OptionID: "opt_correct"},
			EvidenceConceptIDs:  []string{"concept_1"},
			ExplanationMarkdown: "Explanation",
		},
	}
	if err := tv.ValidateTemplate(&tmplMissingHint); err == nil {
		t.Error("expected error for distractor without hint_markdown")
	}

	// Rejection: correct option with non-nil misconception
	tmplCorrectHasMisc := *validTmpl
	tmplCorrectHasMisc.Stages = []domain.StageTemplate{
		{
			ID:             "stage_1",
			Kind:           domain.StageKindChoice,
			PromptMarkdown: "Which statement is correct?",
			Options: []domain.Option{
				{ID: "opt_correct", TextMarkdown: "Correct answer", MisconceptionID: strPtr("misc_err"), HintMarkdown: nil},
				{ID: "opt_wrong", TextMarkdown: "Wrong answer", MisconceptionID: strPtr("misc_1"), HintMarkdown: strPtr("Hint")},
			},
			ExpectedAnswer:      domain.ExpectedAnswer{Kind: domain.AnswerKindChoice, OptionID: "opt_correct"},
			EvidenceConceptIDs:  []string{"concept_1"},
			ExplanationMarkdown: "Explanation",
		},
	}
	if err := tv.ValidateTemplate(&tmplCorrectHasMisc); err == nil {
		t.Error("expected error for correct option with misconception_id")
	}
}

func TestRegressionOutputFamily(t *testing.T) {
	reg := DefaultRegistry()
	fam, ok := reg.Get("regression_output")
	if !ok {
		t.Fatal("regression_output family not found in DefaultRegistry")
	}

	if fam.ID() != "regression_output" {
		t.Errorf("expected ID regression_output, got %q", fam.ID())
	}

	// Unsupported rule version
	if err := fam.ValidateParameters(2, nil); err == nil {
		t.Error("expected error for rule_version 2")
	}

	// Unknown parameter
	if err := fam.ValidateParameters(1, map[string]interface{}{"unknown_key": 1}); err == nil {
		t.Error("expected error for unknown parameter")
	}

	// Valid simple regression parameters:
	// n = 25, k = 1, SSR = 144, SSE = 368, SST = 512
	// MSR = 144, MSE = 16, F = 9.0
	// b1 = 2.4, se_b1 = 0.8 -> t = 3.0, t^2 = 9.0 == F
	validSimpleParams := map[string]interface{}{
		"n":                25,
		"k":                1,
		"ssr":              144.0,
		"sse":              368.0,
		"sst":              512.0,
		"b0":               10.0,
		"se_b0":            2.5,
		"b1":               2.4,
		"se_b1":            0.8,
		"x1_name":          "Advertising",
		"confidence_level": 0.95,
	}

	if err := fam.ValidateParameters(1, validSimpleParams); err != nil {
		t.Fatalf("unexpected error validating valid simple regression parameters: %v", err)
	}

	// Valid multiple regression parameters (k = 2)
	validMultiParams := map[string]interface{}{
		"n":                30,
		"k":                2,
		"ssr":              300.0,
		"sse":              200.0,
		"b0":               5.0,
		"se_b0":            1.0,
		"b1":               1.5,
		"se_b1":            0.3,
		"b2":               -0.8,
		"se_b2":            0.2,
		"confidence_level": 0.95,
	}

	if err := fam.ValidateParameters(1, validMultiParams); err != nil {
		t.Fatalf("unexpected error validating valid multiple regression parameters: %v", err)
	}

	// Validation failure: missing b2 for k = 2
	missingPredParams := map[string]interface{}{
		"n":     30,
		"k":     2,
		"ssr":   300.0,
		"sse":   200.0,
		"b0":    5.0,
		"se_b0": 1.0,
		"b1":    1.5,
		"se_b1": 0.3,
	}
	if err := fam.ValidateParameters(1, missingPredParams); err == nil {
		t.Error("expected error when predictor b2 is missing for k=2")
	}

	// Validation failure: k = 1 consistency failure (t^2 != F)
	inconsistentTParams := map[string]interface{}{
		"n":     25,
		"k":     1,
		"ssr":   144.0,
		"sse":   368.0,
		"b0":    10.0,
		"se_b0": 2.5,
		"b1":    10.0, // t = 12.5 -> t^2 = 156.25 != 9.0
		"se_b1": 0.8,
	}
	if err := fam.ValidateParameters(1, inconsistentTParams); err == nil {
		t.Error("expected error for k=1 when t^2 != F")
	}

	// Validation failure: k out of range [1, 4]
	kZeroParams := map[string]interface{}{
		"n": 25, "k": 0, "ssr": 100.0, "sse": 100.0, "b0": 1.0, "se_b0": 1.0,
	}
	if err := fam.ValidateParameters(1, kZeroParams); err == nil {
		t.Error("expected error for k=0")
	}
}
