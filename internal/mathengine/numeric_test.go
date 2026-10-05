package mathengine_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

func TestParseNumericInputValidForms(t *testing.T) {
	allForms := []domain.NumericForm{
		domain.NumericFormDecimal,
		domain.NumericFormPercent,
		domain.NumericFormFraction,
	}

	tests := []struct {
		input    string
		wantForm domain.NumericForm
		wantVal  float64
		wantRat  string
	}{
		{"0.375", domain.NumericFormDecimal, 0.375, "3/8"},
		{"0", domain.NumericFormDecimal, 0.0, "0/1"},
		{"1.0", domain.NumericFormDecimal, 1.0, "1/1"},
		{"-0.5", domain.NumericFormDecimal, -0.5, "-1/2"},
		{"1.5e-3", domain.NumericFormDecimal, 0.0015, "3/2000"},
		{"37.5%", domain.NumericFormPercent, 0.375, "3/8"},
		{"37.5 %", domain.NumericFormPercent, 0.375, "3/8"},
		{"100%", domain.NumericFormPercent, 1.0, "1/1"},
		{"0%", domain.NumericFormPercent, 0.0, "0/1"},
		{"3/8", domain.NumericFormFraction, 0.375, "3/8"},
		{"6/16", domain.NumericFormFraction, 0.375, "3/8"},
		{"1/2", domain.NumericFormFraction, 0.5, "1/2"},
		{"-1/4", domain.NumericFormFraction, -0.25, "-1/4"},
	}

	for _, tt := range tests {
		parsed, err := mathengine.ParseNumericInput(tt.input, allForms)
		if err != nil {
			t.Fatalf("ParseNumericInput(%q) error: %v", tt.input, err)
		}
		if parsed.Form != tt.wantForm {
			t.Errorf("ParseNumericInput(%q) Form = %q, want %q", tt.input, parsed.Form, tt.wantForm)
		}
		if math.Abs(parsed.Value-tt.wantVal) > 1e-12 {
			t.Errorf("ParseNumericInput(%q) Value = %v, want %v", tt.input, parsed.Value, tt.wantVal)
		}
		if tt.wantRat != "" && parsed.ExactRat != nil {
			expRat, _ := new(big.Rat).SetString(tt.wantRat)
			if parsed.ExactRat.Cmp(expRat) != 0 {
				t.Errorf("ParseNumericInput(%q) ExactRat = %s, want %s", tt.input, parsed.ExactRat.RatString(), tt.wantRat)
			}
		}
	}
}

func TestParseNumericInputValidationRejections(t *testing.T) {
	allForms := []domain.NumericForm{
		domain.NumericFormDecimal,
		domain.NumericFormPercent,
		domain.NumericFormFraction,
	}

	rejections := []struct {
		input       string
		errContains string
	}{
		{"", "empty"},
		{"   ", "empty"},
		{"0,375", "ambiguous comma separator"},
		{"3/0", "denominator cannot be zero"},
		{"3/1.5", "must be an integer"},
		{"3/4/5", "invalid fraction format"},
		{"1+2", "expressions and LaTeX formulas are not evaluated"},
		{"3*4", "expressions and LaTeX formulas are not evaluated"},
		{"\\frac{3}{8}", "expressions and LaTeX formulas are not evaluated"},
		{"xyz", "expressions and LaTeX formulas are not evaluated"},
	}

	for _, r := range rejections {
		_, err := mathengine.ParseNumericInput(r.input, allForms)
		if err == nil {
			t.Errorf("ParseNumericInput(%q) expected error containing %q, got nil", r.input, r.errContains)
		}
	}
}

func TestParseNumericInputFormConstraints(t *testing.T) {
	// Only decimal allowed
	decimalsOnly := []domain.NumericForm{domain.NumericFormDecimal}
	if _, err := mathengine.ParseNumericInput("37.5%", decimalsOnly); err == nil {
		t.Errorf("Expected error when percent input used under decimal-only policy")
	}
	if _, err := mathengine.ParseNumericInput("3/8", decimalsOnly); err == nil {
		t.Errorf("Expected error when fraction input used under decimal-only policy")
	}
	if _, err := mathengine.ParseNumericInput("0.375", decimalsOnly); err != nil {
		t.Errorf("Decimal should succeed under decimal-only policy: %v", err)
	}
}

func TestBarePercentNotSilentlyInterpreted(t *testing.T) {
	// Invariant from NUMERICS.md:
	// "A bare 37.5 is not silently interpreted as 37.5%."
	allForms := []domain.NumericForm{domain.NumericFormDecimal, domain.NumericFormPercent, domain.NumericFormFraction}
	parsed, err := mathengine.ParseNumericInput("37.5", allForms)
	if err != nil {
		t.Fatalf("ParseNumericInput(37.5) failed: %v", err)
	}
	if parsed.Value != 37.5 {
		t.Errorf("parsed.Value = %v, want 37.5 (not silently 0.375)", parsed.Value)
	}

	policy := domain.NumericPolicy{
		Version:           1,
		AbsoluteTolerance: 0.0000005,
		RelativeTolerance: 0.000001,
		AllowedForms:      allForms,
	}

	res := mathengine.GradeNumericAnswer(parsed, 0.375, policy)
	if res.Correct {
		t.Errorf("GradeNumericAnswer with bare 37.5 against 0.375 should be INCORRECT")
	}
	if !containsSubstr(res.Reason, "bare percentage") {
		t.Errorf("Expected diagnostic hint about bare percentage in Reason, got: %q", res.Reason)
	}
}

func TestGradeNumericAnswerDraftFixture(t *testing.T) {
	// Draft binomial fixture policy:
	// Expected: 0.375, absTol: 0.0000005, relTol: 0.000001, allowed_forms: [decimal, percent, fraction]
	policy := domain.NumericPolicy{
		Version:           1,
		AbsoluteTolerance: 0.0000005,
		RelativeTolerance: 0.000001,
		AllowedForms: []domain.NumericForm{
			domain.NumericFormDecimal,
			domain.NumericFormPercent,
			domain.NumericFormFraction,
		},
		DisplayDecimals: nil,
	}

	correctInputs := []string{"0.375", "37.5%", "3/8", "6/16", "0.3750004"}
	for _, in := range correctInputs {
		parsed, err := mathengine.ParseNumericInput(in, policy.AllowedForms)
		if err != nil {
			t.Fatalf("Parse error for %q: %v", in, err)
		}
		grade := mathengine.GradeNumericAnswer(parsed, 0.375, policy)
		if !grade.Correct {
			t.Errorf("Input %q should be graded CORRECT, got false (reason: %s)", in, grade.Reason)
		}
	}

	incorrectInputs := []string{"0.376", "38%", "1/2", "0.0625", "0.6875"}
	for _, in := range incorrectInputs {
		parsed, err := mathengine.ParseNumericInput(in, policy.AllowedForms)
		if err != nil {
			t.Fatalf("Parse error for %q: %v", in, err)
		}
		grade := mathengine.GradeNumericAnswer(parsed, 0.375, policy)
		if grade.Correct {
			t.Errorf("Input %q should be graded INCORRECT, got true", in)
		}
	}
}

func TestGradeNumericAnswerWithStatedRounding(t *testing.T) {
	// Task requiring 3 decimal places: expected = 0.3754
	decimals := 3
	policy := domain.NumericPolicy{
		Version:           1,
		AbsoluteTolerance: 0.0000005,
		RelativeTolerance: 0.000001,
		AllowedForms:      []domain.NumericForm{domain.NumericFormDecimal},
		DisplayDecimals:   &decimals,
	}

	// 0.3754 rounded to 3 decimal places is 0.375
	p1, _ := mathengine.ParseNumericInput("0.375", policy.AllowedForms)
	g1 := mathengine.GradeNumericAnswer(p1, 0.3754, policy)
	if !g1.Correct {
		t.Errorf("Submitted 0.375 should match expected 0.3754 rounded to 3 decimals: %s", g1.Reason)
	}

	// Full-precision answer 0.3754 within half-ulp interval [0.3749, 0.3759]
	p2, _ := mathengine.ParseNumericInput("0.3754", policy.AllowedForms)
	g2 := mathengine.GradeNumericAnswer(p2, 0.3754, policy)
	if !g2.Correct {
		t.Errorf("Full precision answer within half-ulp interval should be accepted: %s", g2.Reason)
	}

	// Outside rounding interval: 0.374
	p3, _ := mathengine.ParseNumericInput("0.374", policy.AllowedForms)
	g3 := mathengine.GradeNumericAnswer(p3, 0.3754, policy)
	if g3.Correct {
		t.Errorf("Submitted 0.374 should be INCORRECT for expected 0.3754")
	}
}

func containsSubstr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && math.Max(0, 0) == 0 && stringContains(s, substr)))
}

func stringContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
