package mathengine

import (
	"fmt"
	"math"
)

// ValidateSetProbabilities checks that P(A), P(B), and P(A ∩ B) are valid probabilities in [0, 1]
// and satisfy Fréchet inequality bounds:
// max(0, P(A) + P(B) - 1) <= P(A ∩ B) <= min(P(A), P(B)).
func ValidateSetProbabilities(pA, pB, pAandB float64) error {
	const tol = 1e-9
	if math.IsNaN(pA) || pA < -tol || pA > 1.0+tol {
		return fmt.Errorf("P(A) must be in [0, 1], got %v", pA)
	}
	if math.IsNaN(pB) || pB < -tol || pB > 1.0+tol {
		return fmt.Errorf("P(B) must be in [0, 1], got %v", pB)
	}
	if math.IsNaN(pAandB) || pAandB < -tol || pAandB > 1.0+tol {
		return fmt.Errorf("P(A ∩ B) must be in [0, 1], got %v", pAandB)
	}

	minOverlap := math.Max(0.0, pA+pB-1.0)
	maxOverlap := math.Min(pA, pB)

	if pAandB < minOverlap-tol {
		return fmt.Errorf("P(A ∩ B)=%v violates lower Fréchet bound: cannot be less than max(0, P(A)+P(B)-1)=%v", pAandB, minOverlap)
	}
	if pAandB > maxOverlap+tol {
		return fmt.Errorf("P(A ∩ B)=%v violates upper Fréchet bound: cannot exceed min(P(A), P(B))=%v", pAandB, maxOverlap)
	}

	return nil
}

// Union computes P(A ∪ B) = P(A) + P(B) - P(A ∩ B).
func Union(pA, pB, pAandB float64) (float64, error) {
	if err := ValidateSetProbabilities(pA, pB, pAandB); err != nil {
		return 0, err
	}
	u := pA + pB - pAandB
	return math.Max(0.0, math.Min(1.0, u)), nil
}

// Complement computes P(A^c) = 1 - P(A).
func Complement(pA float64) (float64, error) {
	if math.IsNaN(pA) || pA < 0.0 || pA > 1.0 {
		return 0, fmt.Errorf("probability must be in [0, 1], got %v", pA)
	}
	return 1.0 - pA, nil
}

// Conditional computes P(A | B) = P(A ∩ B) / P(B).
func Conditional(pAandB, pB float64) (float64, error) {
	if math.IsNaN(pB) || pB <= 0.0 || pB > 1.0 {
		return 0, fmt.Errorf("conditioning probability P(B) must be in (0, 1], got %v", pB)
	}
	if math.IsNaN(pAandB) || pAandB < 0.0 || pAandB > pB+1e-9 {
		return 0, fmt.Errorf("joint probability P(A ∩ B)=%v cannot exceed conditioning event P(B)=%v", pAandB, pB)
	}
	res := pAandB / pB
	return math.Max(0.0, math.Min(1.0, res)), nil
}

// AreIndependent returns true if |P(A ∩ B) - P(A)*P(B)| <= tol.
func AreIndependent(pA, pB, pAandB, tol float64) bool {
	expected := pA * pB
	return math.Abs(pAandB-expected) <= tol
}

// AreDisjoint returns true if P(A ∩ B) <= tol (i.e. mutually exclusive events).
// Note invariant: Disjointness does NOT imply independence!
// For events with P(A) > 0 and P(B) > 0, disjoint events are never independent since P(A)*P(B) > 0 != 0.
func AreDisjoint(pAandB, tol float64) bool {
	return math.Abs(pAandB) <= tol
}
