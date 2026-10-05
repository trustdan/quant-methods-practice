package mathengine_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

func TestDeriveBinomialProblemReferenceFixture(t *testing.T) {
	// n=4, p=0.5, k=2, Event: Exactly(2)
	event := mathengine.Exactly(2)
	deriv, err := mathengine.DeriveBinomialProblem(4, 0.5, 2, event)
	if err != nil {
		t.Fatalf("DeriveBinomialProblem failed: %v", err)
	}

	// 1. Canonical probability and rational
	if math.Abs(deriv.CanonicalProbability-0.375) > 1e-12 {
		t.Errorf("CanonicalProbability = %v, want 0.375", deriv.CanonicalProbability)
	}
	expectedRat := big.NewRat(3, 8)
	if deriv.CanonicalRational == nil || deriv.CanonicalRational.Cmp(expectedRat) != 0 {
		t.Errorf("CanonicalRational = %v, want %s", deriv.CanonicalRational, expectedRat.RatString())
	}

	// 2. Moments
	if deriv.Mean != 2.0 {
		t.Errorf("Mean = %v, want 2.0", deriv.Mean)
	}
	if deriv.Variance != 1.0 {
		t.Errorf("Variance = %v, want 1.0", deriv.Variance)
	}
	if deriv.StdDev != 1.0 {
		t.Errorf("StdDev = %v, want 1.0", deriv.StdDev)
	}

	// 3. LaTeX formatting
	if deriv.EventTeX != "P(X = 2)" {
		t.Errorf("EventTeX = %q, want \"P(X = 2)\"", deriv.EventTeX)
	}

	// 4. Misconception distractors
	distractorMap := make(map[string]float64)
	for _, d := range deriv.Distractors {
		if d.Value != nil {
			distractorMap[d.MisconceptionID] = *d.Value
		}
	}

	// Missing combination: 0.5^2 * 0.5^2 = 0.0625 (1/16)
	if val, ok := distractorMap["missing_combination"]; !ok || math.Abs(val-0.0625) > 1e-12 {
		t.Errorf("missing_combination = %v, want 0.0625", val)
	}

	// Missing failure factor: C(4,2) * 0.5^2 = 6 * 0.25 = 1.5
	if val, ok := distractorMap["missing_failure_factor"]; !ok || math.Abs(val-1.5) > 1e-12 {
		t.Errorf("missing_failure_factor = %v, want 1.5", val)
	}

	// Exactly as at most: P(X <= 2) = 11/16 = 0.6875
	if val, ok := distractorMap["exactly_as_at_most"]; !ok || math.Abs(val-0.6875) > 1e-12 {
		t.Errorf("exactly_as_at_most = %v, want 0.6875", val)
	}

	// Exactly as at least: P(X >= 2) = 11/16 = 0.6875
	if val, ok := distractorMap["exactly_as_at_least"]; !ok || math.Abs(val-0.6875) > 1e-12 {
		t.Errorf("exactly_as_at_least = %v, want 0.6875", val)
	}

	// Complement event: 1 - P(X=2) = 1 - 0.375 = 0.625 (5/8)
	if val, ok := distractorMap["complement_event"]; !ok || math.Abs(val-0.625) > 1e-12 {
		t.Errorf("complement_event = %v, want 0.625", val)
	}

	// Single trial instead of experiment: 0.5
	if val, ok := distractorMap["experiment_as_single_trial"]; !ok || math.Abs(val-0.5) > 1e-12 {
		t.Errorf("experiment_as_single_trial = %v, want 0.5", val)
	}

	// 5. Verify no distractor collisions with canonical answer for n=4, p=0.5, k=2
	policy := domain.NumericPolicy{
		Version:           1,
		AbsoluteTolerance: 0.0000005,
		RelativeTolerance: 0.000001,
		AllowedForms: []domain.NumericForm{
			domain.NumericFormDecimal,
			domain.NumericFormPercent,
			domain.NumericFormFraction,
		},
	}
	collisions := deriv.CheckDistractorCollisions(policy)
	if len(collisions) > 0 {
		t.Errorf("Expected 0 collisions for reference fixture, got: %v", collisions)
	}
}

func TestDeriveBinomialCollisionDetection(t *testing.T) {
	// For n=2, p=0.5, k=2:
	// C(2,2) = 1, so missing_combination (p^2 * q^0 = 0.25) equals C(2,2)*p^2*q^0 = 0.25.
	// CheckDistractorCollisions should detect this ambiguous distractor.
	deriv, err := mathengine.DeriveBinomialProblem(2, 0.5, 2, mathengine.Exactly(2))
	if err != nil {
		t.Fatalf("DeriveBinomialProblem failed: %v", err)
	}

	policy := domain.NumericPolicy{
		Version:           1,
		AbsoluteTolerance: 0.0000005,
		RelativeTolerance: 0.000001,
	}
	collisions := deriv.CheckDistractorCollisions(policy)
	if len(collisions) == 0 {
		t.Errorf("Expected collision for n=2, k=2 missing_combination, got 0 collisions")
	}
}
