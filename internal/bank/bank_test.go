package bank_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/domain"
)

func TestValidateDraftBinomialFixture(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "curriculum", "examples", "binomial-seven-stage.draft.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read draft fixture: %v", err)
	}

	tmpl, err := bank.StrictDecodeTemplate(data)
	if err != nil {
		t.Fatalf("failed to strictly decode draft fixture: %v", err)
	}

	if err := bank.ValidateTemplate(tmpl, nil); err != nil {
		t.Fatalf("failed to validate draft fixture: %v", err)
	}

	if tmpl.ID != "binomial_fair_coin_exactly_two" {
		t.Errorf("unexpected template ID: %s", tmpl.ID)
	}
	if tmpl.Status != domain.StatusDraft {
		t.Errorf("draft fixture status must be draft, got %s", tmpl.Status)
	}
	if len(tmpl.Stages) != 7 {
		t.Errorf("draft fixture must have 7 stages, got %d", len(tmpl.Stages))
	}
}

func TestRejectUnknownFields(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "curriculum", "examples", "binomial-seven-stage.draft.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read draft fixture: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}

	// Add unknown field at top level
	raw["extra_field"] = "not_allowed"
	mutated, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	_, err = bank.StrictDecodeTemplate(mutated)
	if err == nil {
		t.Fatal("expected StrictDecodeTemplate to reject unknown top-level field, got nil")
	}
}

func TestRejectInvalidParameters(t *testing.T) {
	reg := bank.DefaultRegistry()
	fam, ok := reg.Get("binomial_pmf")
	if !ok {
		t.Fatal("binomial_pmf not registered")
	}

	cases := []struct {
		name   string
		params map[string]interface{}
	}{
		{"missing n", map[string]interface{}{"p": 0.5, "k": 2}},
		{"missing p", map[string]interface{}{"n": 4, "k": 2}},
		{"missing k", map[string]interface{}{"n": 4, "p": 0.5}},
		{"negative n", map[string]interface{}{"n": -1, "p": 0.5, "k": 0}},
		{"fractional n", map[string]interface{}{"n": 4.5, "p": 0.5, "k": 2}},
		{"p negative", map[string]interface{}{"n": 4, "p": -0.1, "k": 2}},
		{"p greater than 1", map[string]interface{}{"n": 4, "p": 1.1, "k": 2}},
		{"negative k", map[string]interface{}{"n": 4, "p": 0.5, "k": -1}},
		{"fractional k", map[string]interface{}{"n": 4, "p": 0.5, "k": 2.5}},
		{"k exceeds n", map[string]interface{}{"n": 4, "p": 0.5, "k": 5}},
		{"unknown param", map[string]interface{}{"n": 4, "p": 0.5, "k": 2, "extra": "invalid"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := fam.ValidateParameters(1, tc.params)
			if err == nil {
				t.Errorf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

func TestRejectUnsupportedFamily(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "curriculum", "examples", "binomial-seven-stage.draft.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read draft fixture: %v", err)
	}

	tmpl, err := bank.StrictDecodeTemplate(data)
	if err != nil {
		t.Fatalf("strict decode: %v", err)
	}

	tmpl.FamilyID = "nonexistent_distribution"
	err = bank.ValidateTemplate(tmpl, nil)
	if err == nil {
		t.Fatal("expected error for unsupported family, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported family_id") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestActiveStatusRequiresApproval(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "curriculum", "examples", "binomial-seven-stage.draft.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read draft fixture: %v", err)
	}

	tmpl, err := bank.StrictDecodeTemplate(data)
	if err != nil {
		t.Fatalf("strict decode: %v", err)
	}

	// Change to active without approval
	tmpl.Status = domain.StatusActive
	tmpl.Approval = nil

	err = bank.ValidateTemplate(tmpl, nil)
	if err == nil {
		t.Fatal("expected error for active template without approval, got nil")
	}

	// Add valid approval
	now := time.Now().UTC()
	tmpl.Approval = &domain.ApprovalRecord{
		Reviewer:   "reviewer_dan",
		ReviewedAt: now,
		Notes:      "Verified against canonical binomial PMF formula",
	}

	if err := bank.ValidateTemplate(tmpl, nil); err != nil {
		t.Fatalf("expected valid approved template to pass validation, got %v", err)
	}
}

func TestLoadActiveBankRejectsDraft(t *testing.T) {
	tmpDir := t.TempDir()

	// Copy draft fixture into tmpDir
	fixturePath := filepath.Join("..", "..", "curriculum", "examples", "binomial-seven-stage.draft.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read draft fixture: %v", err)
	}

	destFile := filepath.Join(tmpDir, "sample.json")
	if err := os.WriteFile(destFile, data, 0644); err != nil {
		t.Fatalf("failed to write tmp template: %v", err)
	}

	// Loading active bank from directory containing a draft template must fail
	_, err = bank.LoadActiveBank(tmpDir, nil)
	if err == nil {
		t.Fatal("expected LoadActiveBank to reject draft template, got nil")
	}
	if !strings.Contains(err.Error(), "active bank only admits approved active records") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadActiveBankAdmitsApproved(t *testing.T) {
	tmpDir := t.TempDir()

	fixturePath := filepath.Join("..", "..", "curriculum", "examples", "binomial-seven-stage.draft.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read draft fixture: %v", err)
	}

	tmpl, err := bank.StrictDecodeTemplate(data)
	if err != nil {
		t.Fatalf("strict decode: %v", err)
	}

	// Make it approved active
	tmpl.Status = domain.StatusActive
	tmpl.Approval = &domain.ApprovalRecord{
		Reviewer:   "dan",
		ReviewedAt: time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC),
		Notes:      "Canonical approved fixture",
	}

	approvedBytes, err := json.MarshalIndent(tmpl, "", "  ")
	if err != nil {
		t.Fatalf("marshal approved: %v", err)
	}

	destFile := filepath.Join(tmpDir, "approved_sample.json")
	if err := os.WriteFile(destFile, approvedBytes, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	b, err := bank.LoadActiveBank(tmpDir, nil)
	if err != nil {
		t.Fatalf("expected LoadActiveBank to succeed on approved record, got: %v", err)
	}

	if b.Count() != 1 {
		t.Errorf("expected 1 template in bank, got %d", b.Count())
	}
	loaded, ok := b.Get(tmpl.ID)
	if !ok || loaded == nil {
		t.Fatalf("could not retrieve template %s from bank", tmpl.ID)
	}
	if loaded.Title != tmpl.Title {
		t.Errorf("title mismatch: %s vs %s", loaded.Title, tmpl.Title)
	}

	byMod := b.ListByModule("module_2")
	if len(byMod) != 1 {
		t.Errorf("expected 1 template in module_2, got %d", len(byMod))
	}
	byFam := b.ListByFamily("binomial_pmf")
	if len(byFam) != 1 {
		t.Errorf("expected 1 template in binomial_pmf, got %d", len(byFam))
	}
}

func TestMismatchedStageKindValidation(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "curriculum", "examples", "binomial-seven-stage.draft.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read draft fixture: %v", err)
	}

	tmpl, err := bank.StrictDecodeTemplate(data)
	if err != nil {
		t.Fatalf("strict decode: %v", err)
	}

	// Change a choice stage to have numeric policy
	tmpl.Stages[0].NumericPolicy = &domain.NumericPolicy{
		Version:           1,
		AbsoluteTolerance: 0.001,
		RelativeTolerance: 0.001,
		AllowedForms:      []domain.NumericForm{domain.NumericFormDecimal},
	}
	if err := bank.ValidateTemplate(tmpl, nil); err == nil {
		t.Fatal("expected error when choice stage has numeric_policy, got nil")
	}

	// Reset
	tmpl.Stages[0].NumericPolicy = nil

	// Give numeric stage options
	tmpl.Stages[5].Options = []domain.Option{
		{ID: "bad_opt", TextMarkdown: "Bad", MisconceptionID: nil, HintMarkdown: nil},
	}
	if err := bank.ValidateTemplate(tmpl, nil); err == nil {
		t.Fatal("expected error when numeric stage has options, got nil")
	}
}

func TestEvidenceConceptsMapping(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "curriculum", "examples", "binomial-seven-stage.draft.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read draft fixture: %v", err)
	}

	tmpl, err := bank.StrictDecodeTemplate(data)
	if err != nil {
		t.Fatalf("strict decode: %v", err)
	}

	// Duplicate an evidence concept into another stage
	tmpl.Stages[1].EvidenceConceptIDs = append(tmpl.Stages[1].EvidenceConceptIDs, tmpl.Stages[0].EvidenceConceptIDs[0])
	if err := bank.ValidateTemplate(tmpl, nil); err == nil {
		t.Fatal("expected error when evidence concept is assigned to multiple stages, got nil")
	}
}

func TestRejectInvalidPoissonParameters(t *testing.T) {
	reg := bank.DefaultRegistry()
	fam, ok := reg.Get("poisson_pmf")
	if !ok {
		t.Fatal("poisson_pmf not registered")
	}

	cases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid", map[string]interface{}{"lambda": 3.0, "k": 2}, false},
		{"valid zero rate", map[string]interface{}{"lambda": 0.0, "k": 0}, false},
		{"missing lambda", map[string]interface{}{"k": 2}, true},
		{"missing k", map[string]interface{}{"lambda": 3.0}, true},
		{"negative lambda", map[string]interface{}{"lambda": -1.0, "k": 2}, true},
		{"fractional k", map[string]interface{}{"lambda": 3.0, "k": 1.5}, true},
		{"negative k", map[string]interface{}{"lambda": 3.0, "k": -1}, true},
		{"unknown param", map[string]interface{}{"lambda": 3.0, "k": 2, "extra": 1}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := fam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestRejectInvalidSetProbabilityParameters(t *testing.T) {
	reg := bank.DefaultRegistry()
	fam, ok := reg.Get("set_probability")
	if !ok {
		t.Fatal("set_probability not registered")
	}

	cases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid union", map[string]interface{}{"operation": "union", "p_a": 0.6, "p_b": 0.5, "p_intersection": 0.3}, false},
		{"valid complement", map[string]interface{}{"operation": "complement", "p_a": 0.35}, false},
		{"valid conditional", map[string]interface{}{"operation": "conditional", "p_a": 0.4, "p_b": 0.6, "p_intersection": 0.24}, false},
		{"missing operation", map[string]interface{}{"p_a": 0.5}, true},
		{"invalid operation", map[string]interface{}{"operation": "unsupported", "p_a": 0.5}, true},
		{"p_a out of range", map[string]interface{}{"operation": "complement", "p_a": 1.5}, true},
		{"p_b out of range", map[string]interface{}{"operation": "union", "p_a": 0.5, "p_b": -0.1, "p_intersection": 0.2}, true},
		{"frechet violation lower", map[string]interface{}{"operation": "union", "p_a": 0.7, "p_b": 0.6, "p_intersection": 0.1}, true},
		{"frechet violation upper", map[string]interface{}{"operation": "union", "p_a": 0.4, "p_b": 0.6, "p_intersection": 0.5}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := fam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestRejectInvalidLinearCombinationParameters(t *testing.T) {
	reg := bank.DefaultRegistry()
	fam, ok := reg.Get("linear_combination")
	if !ok {
		t.Fatal("linear_combination not registered")
	}

	cases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid independent", map[string]interface{}{"operation": "variance_independent", "a": 2.0, "b": -3.0, "var_x": 4.0, "var_y": 9.0}, false},
		{"valid correlated", map[string]interface{}{"operation": "variance_correlated", "a": 2.0, "b": 3.0, "var_x": 4.0, "var_y": 9.0, "cov_xy": 3.0}, false},
		{"negative variance", map[string]interface{}{"operation": "variance_independent", "a": 1.0, "b": 1.0, "var_x": -1.0, "var_y": 4.0}, true},
		{"cauchy-schwarz violation", map[string]interface{}{"operation": "variance_correlated", "a": 1.0, "b": 1.0, "var_x": 4.0, "var_y": 9.0, "cov_xy": 7.0}, true},
		{"missing operation", map[string]interface{}{"a": 1.0, "b": 1.0}, true},
		{"missing scalar a", map[string]interface{}{"operation": "expectation", "b": 1.0}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := fam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

