package bank_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/curriculum"
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

func TestRejectInvalidContinuousDistributionParameters(t *testing.T) {
	reg := bank.DefaultRegistry()

	// Normal distribution
	normFam, ok := reg.Get("normal_dist")
	if !ok {
		t.Fatal("normal_dist not registered")
	}
	normCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid standard normal", map[string]interface{}{"mu": 0.0, "sigma": 1.0}, false},
		{"valid with cutoff x", map[string]interface{}{"mu": 100.0, "sigma": 15.0, "x": 115.0, "tail": "upper"}, false},
		{"valid interval bounds", map[string]interface{}{"mu": 0.0, "sigma": 1.0, "x1": -1.96, "x2": 1.96}, false},
		{"valid with quantile p", map[string]interface{}{"mu": 80000.0, "sigma": 12000.0, "p": 0.95, "tail": "upper"}, false},
		{"valid with empirical k", map[string]interface{}{"mu": 100.0, "sigma": 15.0, "k": 2.0}, false},
		{"valid with comparison params", map[string]interface{}{"mu": 500000.0, "sigma": 50000.0, "x": 580000.0, "mu_b": 300000.0, "sigma_b": 40000.0, "x_b": 370000.0}, false},
		{"invalid p zero", map[string]interface{}{"mu": 0.0, "sigma": 1.0, "p": 0.0}, true},
		{"invalid p one", map[string]interface{}{"mu": 0.0, "sigma": 1.0, "p": 1.0}, true},
		{"invalid negative k", map[string]interface{}{"mu": 100.0, "sigma": 15.0, "k": -1.0}, true},
		{"invalid negative sigma_b", map[string]interface{}{"mu": 500.0, "sigma": 50.0, "mu_b": 300.0, "sigma_b": -10.0}, true},
		{"inverted interval bounds", map[string]interface{}{"mu": 0.0, "sigma": 1.0, "x1": 2.0, "x2": -1.0}, true},
		{"missing mu", map[string]interface{}{"sigma": 1.0}, true},
		{"missing sigma", map[string]interface{}{"mu": 0.0}, true},
		{"negative sigma", map[string]interface{}{"mu": 0.0, "sigma": -1.0}, true},
		{"zero sigma", map[string]interface{}{"mu": 0.0, "sigma": 0.0}, true},
		{"unknown param", map[string]interface{}{"mu": 0.0, "sigma": 1.0, "extra": 123}, true},
	}
	for _, tc := range normCases {
		t.Run("normal_"+tc.name, func(t *testing.T) {
			err := normFam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("normal ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}

	// Uniform distribution
	uniFam, ok := reg.Get("uniform_dist")
	if !ok {
		t.Fatal("uniform_dist not registered")
	}
	uniCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid bounds", map[string]interface{}{"a": 0.0, "b": 10.0}, false},
		{"a equals b", map[string]interface{}{"a": 5.0, "b": 5.0}, true},
		{"a greater than b", map[string]interface{}{"a": 10.0, "b": 2.0}, true},
		{"missing b", map[string]interface{}{"a": 0.0}, true},
		{"unknown param", map[string]interface{}{"a": 0.0, "b": 1.0, "c": 2.0}, true},
	}
	for _, tc := range uniCases {
		t.Run("uniform_"+tc.name, func(t *testing.T) {
			err := uniFam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("uniform ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}

	// Exponential distribution
	expFam, ok := reg.Get("exponential_dist")
	if !ok {
		t.Fatal("exponential_dist not registered")
	}
	expCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid rate", map[string]interface{}{"lambda": 0.5}, false},
		{"valid full params", map[string]interface{}{"lambda": 0.25, "mean": 4.0, "x": 2.0, "tail": "lower"}, false},
		{"valid quantile p", map[string]interface{}{"lambda": 0.40, "p": 0.90}, false},
		{"valid memoryless", map[string]interface{}{"lambda": 0.1667, "s": 5.0, "t": 3.0}, false},
		{"valid interval", map[string]interface{}{"lambda": 0.001, "x1": 500.0, "x2": 1500.0}, false},
		{"missing lambda", map[string]interface{}{}, true},
		{"negative lambda", map[string]interface{}{"lambda": -2.0}, true},
		{"zero lambda", map[string]interface{}{"lambda": 0.0}, true},
		{"invalid quantile p zero", map[string]interface{}{"lambda": 0.5, "p": 0.0}, true},
		{"invalid quantile p one", map[string]interface{}{"lambda": 0.5, "p": 1.0}, true},
		{"invalid interval bounds", map[string]interface{}{"lambda": 0.5, "x1": 10.0, "x2": 5.0}, true},
		{"missing interval bound", map[string]interface{}{"lambda": 0.5, "x1": 5.0}, true},
		{"negative mean", map[string]interface{}{"lambda": 0.5, "mean": -2.0}, true},
		{"unknown param", map[string]interface{}{"lambda": 1.0, "rate": 1.0}, true},
	}
	for _, tc := range expCases {
		t.Run("exponential_"+tc.name, func(t *testing.T) {
			err := expFam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("exponential ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}

	// Sampling distribution
	sampFam, ok := reg.Get("sampling_distribution")
	if !ok {
		t.Fatal("sampling_distribution not registered")
	}
	sampCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid mean", map[string]interface{}{"stat": "mean", "mu": 50.0, "sigma": 10.0, "n": 25}, false},
		{"valid proportion", map[string]interface{}{"stat": "proportion", "p": 0.4, "n": 100}, false},
		{"valid mean with x_bar and pop_size", map[string]interface{}{"stat": "mean", "mu": 50000.0, "sigma": 8000.0, "n": 64, "pop_size": 400, "x_bar": 51835.33}, false},
		{"valid proportion with p_hat", map[string]interface{}{"stat": "proportion", "p": 0.30, "n": 200, "p_hat": 0.35}, false},
		{"valid skewed pop_shape", map[string]interface{}{"stat": "mean", "mu": 6.0, "sigma": 6.0, "n": 64, "pop_shape": "exponential", "pop_is_normal": false, "x_bar": 4.8}, false},
		{"missing n", map[string]interface{}{"mu": 50.0, "sigma": 10.0}, true},
		{"negative n", map[string]interface{}{"mu": 50.0, "sigma": 10.0, "n": -5}, true},
		{"missing sigma for mean", map[string]interface{}{"stat": "mean", "mu": 50.0, "n": 25}, true},
		{"n exceeds pop_size", map[string]interface{}{"mu": 50.0, "sigma": 10.0, "n": 100, "pop_size": 50}, true},
		{"invalid p_hat", map[string]interface{}{"stat": "proportion", "p": 0.4, "n": 100, "p_hat": 1.5}, true},
		{"invalid pop_is_normal type", map[string]interface{}{"mu": 50.0, "sigma": 10.0, "n": 25, "pop_is_normal": "yes"}, true},
	}
	for _, tc := range sampCases {
		t.Run("sampling_"+tc.name, func(t *testing.T) {
			err := sampFam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("sampling ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}

	// Confidence interval
	ciFam, ok := reg.Get("confidence_interval")
	if !ok {
		t.Fatal("confidence_interval not registered")
	}
	ciCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid mean z", map[string]interface{}{"parameter": "mean_z", "n": 36, "confidence_level": 0.95, "sigma": 12.0}, false},
		{"missing confidence_level", map[string]interface{}{"parameter": "mean_z", "n": 36}, true},
		{"invalid confidence_level", map[string]interface{}{"n": 36, "confidence_level": 1.5}, true},
	}
	for _, tc := range ciCases {
		t.Run("ci_"+tc.name, func(t *testing.T) {
			err := ciFam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("ci ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}

	// Hypothesis test
	hypFam, ok := reg.Get("hypothesis_test")
	if !ok {
		t.Fatal("hypothesis_test not registered")
	}
	hypCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid test", map[string]interface{}{"null_value": 50.0, "n": 25, "alpha": 0.05, "tail": "two_tailed"}, false},
		{"missing alpha", map[string]interface{}{"null_value": 50.0, "n": 25}, true},
		{"missing null_value", map[string]interface{}{"n": 25, "alpha": 0.05}, true},
	}
	for _, tc := range hypCases {
		t.Run("hyp_"+tc.name, func(t *testing.T) {
			err := hypFam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("hyp ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}

	// Linear regression
	regFam, ok := reg.Get("linear_regression")
	if !ok {
		t.Fatal("linear_regression not registered")
	}
	regCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid regression", map[string]interface{}{"slope": 1.5, "intercept": 2.0, "n": 10}, false},
		{"missing slope", map[string]interface{}{"intercept": 2.0, "n": 10}, true},
		{"missing intercept", map[string]interface{}{"slope": 1.5, "n": 10}, true},
	}
	for _, tc := range regCases {
		t.Run("regression_"+tc.name, func(t *testing.T) {
			err := regFam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("regression ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}

	// Sample size
	ssFam, ok := reg.Get("sample_size")
	if !ok {
		t.Fatal("sample_size not registered")
	}
	ssCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid mean", map[string]interface{}{"target": "mean", "sigma": 25.0, "margin_of_error": 5.0, "confidence_level": 0.95}, false},
		{"missing sigma for mean", map[string]interface{}{"target": "mean", "margin_of_error": 5.0, "confidence_level": 0.95}, true},
		{"valid proportion", map[string]interface{}{"target": "proportion", "p": 0.20, "margin_of_error": 0.03, "confidence_level": 0.95}, false},
		{"valid conservative proportion", map[string]interface{}{"target": "proportion", "margin_of_error": 0.04, "confidence_level": 0.90}, false},
		{"valid target_se", map[string]interface{}{"target": "target_se", "sigma": 15.0, "target_se": 2.0}, false},
		{"missing target_se value", map[string]interface{}{"target": "target_se", "sigma": 15.0}, true},
		{"unsupported target", map[string]interface{}{"target": "unknown"}, true},
	}
	for _, tc := range ssCases {
		t.Run("ss_"+tc.name, func(t *testing.T) {
			err := ssFam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("sample_size ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}

	// Set probability
	setFam, ok := reg.Get("set_probability")
	if !ok {
		t.Fatal("set_probability not registered")
	}
	setCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid union", map[string]interface{}{"operation": "union", "p_a": 0.4, "p_b": 0.5, "p_intersection": 0.2}, false},
		{"valid complement", map[string]interface{}{"operation": "complement", "p_a": 0.3}, false},
		{"valid conditional", map[string]interface{}{"operation": "conditional", "p_a": 0.4, "p_b": 0.6, "p_intersection": 0.24}, false},
		{"valid multiplication", map[string]interface{}{"operation": "multiplication", "p_a": 0.90, "p_b": 0.85, "p_intersection": 0.765}, false},
		{"valid independence_check", map[string]interface{}{"operation": "independence_check", "p_a": 0.40, "p_b": 0.50, "p_intersection": 0.20}, false},
		{"independence_check missing p_intersection", map[string]interface{}{"operation": "independence_check", "p_a": 0.40, "p_b": 0.50}, true},
		{"invalid operation", map[string]interface{}{"operation": "xor", "p_a": 0.5}, true},
		{"missing p_b for union", map[string]interface{}{"operation": "union", "p_a": 0.4}, true},
	}
	for _, tc := range setCases {
		t.Run("set_"+tc.name, func(t *testing.T) {
			err := setFam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("set_probability ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}

	// Binomial mean_sd operation
	binomFam, ok := reg.Get("binomial_pmf")
	if !ok {
		t.Fatal("binomial_pmf not registered")
	}
	binomCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid mean_sd", map[string]interface{}{"operation": "mean_sd", "n": 200, "p": 0.02, "mean": 4.0, "sd": 1.979898987322333}, false},
		{"mismatched mean", map[string]interface{}{"operation": "mean_sd", "n": 200, "p": 0.02, "mean": 5.0}, true},
		{"mismatched sd", map[string]interface{}{"operation": "mean_sd", "n": 200, "p": 0.02, "sd": 3.0}, true},
	}
	for _, tc := range binomCases {
		t.Run("binom_"+tc.name, func(t *testing.T) {
			err := binomFam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("binomial_pmf ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}

	// Descriptive stats

	descFam, ok := reg.Get("descriptive_stats")
	if !ok {
		t.Fatal("descriptive_stats not registered")
	}
	descCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid summary", map[string]interface{}{"operation": "summary", "data": []interface{}{12.0, 16.0, 18.0, 20.0, 22.0, 32.0}}, false},
		{"missing operation", map[string]interface{}{"data": []interface{}{10.0, 20.0}}, true},
		{"unknown param", map[string]interface{}{"operation": "summary", "invalid_param": 123}, true},
	}
	for _, tc := range descCases {
		t.Run("desc_"+tc.name, func(t *testing.T) {
			err := descFam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("descriptive_stats ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}

	// Discrete RV
	drvFam, ok := reg.Get("discrete_rv")
	if !ok {
		t.Fatal("discrete_rv not registered")
	}
	drvCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid expected_value", map[string]interface{}{"operation": "expected_value", "values": []interface{}{0.0, 1.0, 2.0}, "probabilities": []interface{}{0.2, 0.5, 0.3}}, false},
		{"missing operation", map[string]interface{}{"values": []interface{}{0.0, 1.0}}, true},
		{"unknown param", map[string]interface{}{"operation": "variance", "invalid": 456}, true},
	}
	for _, tc := range drvCases {
		t.Run("drv_"+tc.name, func(t *testing.T) {
			err := drvFam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("discrete_rv ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}

	// Two sample test
	twoFam, ok := reg.Get("two_sample_test")
	if !ok {
		t.Fatal("two_sample_test not registered")
	}
	twoCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{"valid paired_t", map[string]interface{}{"operation": "paired_t", "mean_diff": 4.0, "sd_diff": 3.0, "n": 10}, false},
		{"missing operation", map[string]interface{}{"n": 10}, true},
		{"unknown param", map[string]interface{}{"operation": "paired_t", "foo": 999}, true},
	}
	for _, tc := range twoCases {
		t.Run("two_"+tc.name, func(t *testing.T) {
			err := twoFam.ValidateParameters(1, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("two_sample_test ValidateParameters() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestLoadActiveBankFS(t *testing.T) {
	b, err := bank.LoadActiveBankFS(curriculum.ApprovedFS, "approved", nil)
	if err != nil {
		t.Fatalf("LoadActiveBankFS failed on embedded curriculum: %v", err)
	}

	if b.Count() != 137 {
		t.Fatalf("expected 137 approved templates from embedded FS, got %d", b.Count())
	}

	// Verify key templates are present and active
	checkIDs := []string{
		"binomial_fair_coin_exactly_two",
		"normal_between_package_weight",
		"regression_sales_advertising",
		"twosample_choose_design",
	}
	for _, id := range checkIDs {
		tmpl, ok := b.Get(id)
		if !ok {
			t.Errorf("expected embedded bank to contain %s", id)
			continue
		}
		if tmpl.Status != domain.StatusActive {
			t.Errorf("template %s has status %s, expected active", id, tmpl.Status)
		}
	}
}
