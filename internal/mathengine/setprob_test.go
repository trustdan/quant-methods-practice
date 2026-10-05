package mathengine_test

import (
	"math"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

func TestValidateSetProbabilitiesFrechetBounds(t *testing.T) {
	// Valid cases
	validCases := []struct {
		pA, pB, pAandB float64
	}{
		{0.6, 0.7, 0.4}, // min overlap: 0.3, max overlap: 0.6
		{0.3, 0.4, 0.0}, // min overlap: 0.0, max overlap: 0.3 (disjoint)
		{0.5, 0.5, 0.5}, // min overlap: 0.0, max overlap: 0.5 (coincident)
		{1.0, 1.0, 1.0}, // sure events
		{0.0, 0.5, 0.0}, // impossible event
	}

	for _, vc := range validCases {
		if err := mathengine.ValidateSetProbabilities(vc.pA, vc.pB, vc.pAandB); err != nil {
			t.Errorf("ValidateSetProbabilities(%v, %v, %v) unexpected error: %v",
				vc.pA, vc.pB, vc.pAandB, err)
		}
	}

	// Lower Fréchet bound violation: P(A)=0.7, P(B)=0.8 -> min overlap is 0.5
	if err := mathengine.ValidateSetProbabilities(0.7, 0.8, 0.4); err == nil {
		t.Errorf("Expected error for lower Fréchet violation (overlap 0.4 < 0.5), got nil")
	}

	// Upper Fréchet bound violation: P(A)=0.3, P(B)=0.6 -> max overlap is 0.3
	if err := mathengine.ValidateSetProbabilities(0.3, 0.6, 0.4); err == nil {
		t.Errorf("Expected error for upper Fréchet violation (overlap 0.4 > 0.3), got nil")
	}

	// Range violations
	if err := mathengine.ValidateSetProbabilities(-0.1, 0.5, 0.0); err == nil {
		t.Errorf("Expected error for negative P(A), got nil")
	}
	if err := mathengine.ValidateSetProbabilities(0.5, 1.2, 0.5); err == nil {
		t.Errorf("Expected error for P(B) > 1, got nil")
	}
}

func TestSetUnionAndComplement(t *testing.T) {
	// P(A u B) = P(A) + P(B) - P(A n B)
	u, err := mathengine.Union(0.4, 0.3, 0.1)
	if err != nil {
		t.Fatalf("Union failed: %v", err)
	}
	wantU := 0.6
	if math.Abs(u-wantU) > 1e-12 {
		t.Errorf("Union(0.4, 0.3, 0.1) = %v, want %v", u, wantU)
	}

	// Complement P(A^c) = 1 - P(A)
	c, err := mathengine.Complement(0.35)
	if err != nil {
		t.Fatalf("Complement failed: %v", err)
	}
	wantC := 0.65
	if math.Abs(c-wantC) > 1e-12 {
		t.Errorf("Complement(0.35) = %v, want %v", c, wantC)
	}
}

func TestConditionalProbability(t *testing.T) {
	// P(A|B) = P(A n B) / P(B)
	cond, err := mathengine.Conditional(0.15, 0.5)
	if err != nil {
		t.Fatalf("Conditional failed: %v", err)
	}
	wantCond := 0.3
	if math.Abs(cond-wantCond) > 1e-12 {
		t.Errorf("Conditional(0.15, 0.5) = %v, want %v", cond, wantCond)
	}

	// Conditioning on zero probability event must fail
	if _, err := mathengine.Conditional(0.0, 0.0); err == nil {
		t.Errorf("Conditional with P(B)=0 expected error, got nil")
	}
}

func TestDisjointnessDoesNotImplyIndependence(t *testing.T) {
	// Invariant from NUMERICS.md: "Disjointness does not imply independence."
	// Case: Disjoint events with non-zero probability
	pA := 0.4
	pB := 0.5
	pAandB := 0.0 // mutually exclusive

	isDisjoint := mathengine.AreDisjoint(pAandB, 1e-9)
	isIndependent := mathengine.AreIndependent(pA, pB, pAandB, 1e-9)

	if !isDisjoint {
		t.Errorf("Expected AreDisjoint to be true")
	}
	if isIndependent {
		t.Errorf("Invariant broken: Disjoint events with positive probability cannot be independent (P(A)*P(B)=%v != 0)", pA*pB)
	}

	// Case: Truly independent events
	pAandBInd := pA * pB // 0.2
	if !mathengine.AreIndependent(pA, pB, pAandBInd, 1e-9) {
		t.Errorf("Expected AreIndependent to be true for product overlap")
	}
	if mathengine.AreDisjoint(pAandBInd, 1e-9) {
		t.Errorf("Independent events with positive probability are not disjoint")
	}
}
