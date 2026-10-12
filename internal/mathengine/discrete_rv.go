package mathengine

import (
	"fmt"
	"math"
)

// DiscreteRV represents a discrete random variable defined by a finite probability mass function.
type DiscreteRV struct {
	Values        []float64 `json:"values"`
	Probabilities []float64 `json:"probabilities"`
}

// NewDiscreteRV creates and validates a DiscreteRV ensuring non-negative probabilities summing to 1.
func NewDiscreteRV(values, probs []float64) (*DiscreteRV, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("discrete RV must have at least one outcome")
	}
	if len(values) != len(probs) {
		return nil, fmt.Errorf("length mismatch between values (%d) and probabilities (%d)", len(values), len(probs))
	}

	sumP := 0.0
	for i, p := range probs {
		if math.IsNaN(p) || p < -1e-9 || p > 1.0+1e-9 {
			return nil, fmt.Errorf("probability at index %d is invalid: %v (must be in [0, 1])", i, p)
		}
		if p < 0.0 {
			p = 0.0
		}
		sumP += p
	}

	if math.Abs(sumP-1.0) > 1e-4 {
		return nil, fmt.Errorf("probabilities must sum to 1.0, got sum = %v", sumP)
	}

	return &DiscreteRV{
		Values:        values,
		Probabilities: probs,
	}, nil
}

// SolveMissingProbability solves for a single missing probability in a PMF where sum must equal 1.
func SolveMissingProbability(probsWithNegativeOne []float64) (int, float64, error) {
	missingIdx := -1
	sumKnown := 0.0

	for i, p := range probsWithNegativeOne {
		if p < 0 || math.IsNaN(p) {
			if missingIdx != -1 {
				return -1, 0, fmt.Errorf("multiple missing probabilities found (at index %d and %d)", missingIdx, i)
			}
			missingIdx = i
		} else {
			if p > 1.0 {
				return -1, 0, fmt.Errorf("probability at index %d exceeds 1: %v", i, p)
			}
			sumKnown += p
		}
	}

	if missingIdx == -1 {
		return -1, 0, fmt.Errorf("no missing probability indicated (values must contain negative/NaN placeholder)")
	}

	if sumKnown > 1.0+1e-7 {
		return -1, 0, fmt.Errorf("sum of known probabilities (%v) exceeds 1.0", sumKnown)
	}

	solvedP := 1.0 - sumKnown
	if solvedP < 0.0 && solvedP > -1e-9 {
		solvedP = 0.0
	}

	return missingIdx, solvedP, nil
}

// ExpectedValue computes E[X] = sum(x_i * p_i).
func (rv *DiscreteRV) ExpectedValue() float64 {
	e := 0.0
	for i := range rv.Values {
		e += rv.Values[i] * rv.Probabilities[i]
	}
	return e
}

// ExpectedValueSquared computes E[X^2] = sum(x_i^2 * p_i).
func (rv *DiscreteRV) ExpectedValueSquared() float64 {
	e2 := 0.0
	for i := range rv.Values {
		x := rv.Values[i]
		e2 += x * x * rv.Probabilities[i]
	}
	return e2
}

// Variance computes Var(X) = E[X^2] - (E[X])^2.
func (rv *DiscreteRV) Variance() float64 {
	mu := rv.ExpectedValue()
	e2 := rv.ExpectedValueSquared()
	v := e2 - mu*mu
	if v < 0.0 && v > -1e-9 {
		v = 0.0
	}
	return v
}

// StdDev computes SD(X) = sqrt(Var(X)).
func (rv *DiscreteRV) StdDev() float64 {
	return math.Sqrt(rv.Variance())
}

// ProbabilityLeq computes P(X <= k) = sum_{x_i <= k} p_i.
func (rv *DiscreteRV) ProbabilityLeq(k float64) float64 {
	p := 0.0
	for i, x := range rv.Values {
		if x <= k+1e-9 {
			p += rv.Probabilities[i]
		}
	}
	return p
}

// ProbabilityGeq computes P(X >= k) = sum_{x_i >= k} p_i.
func (rv *DiscreteRV) ProbabilityGeq(k float64) float64 {
	p := 0.0
	for i, x := range rv.Values {
		if x >= k-1e-9 {
			p += rv.Probabilities[i]
		}
	}
	return p
}

// ProbabilityBetween computes P(a <= X <= b) = sum_{a <= x_i <= b} p_i.
func (rv *DiscreteRV) ProbabilityBetween(a, b float64) float64 {
	p := 0.0
	for i, x := range rv.Values {
		if x >= a-1e-9 && x <= b+1e-9 {
			p += rv.Probabilities[i]
		}
	}
	return p
}
