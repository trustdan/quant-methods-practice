package mathengine

import (
	"fmt"
	"math"
)

// FDistribution represents Snedecor's F-distribution with d1 (numerator) and d2 (denominator) degrees of freedom.
type FDistribution struct {
	D1 float64 // Numerator degrees of freedom (d1 > 0)
	D2 float64 // Denominator degrees of freedom (d2 > 0)
}

// NewFDistribution constructs an F-distribution with d1 and d2 degrees of freedom.
func NewFDistribution(d1, d2 float64) (*FDistribution, error) {
	if math.IsNaN(d1) || math.IsInf(d1, 0) || d1 <= 0 {
		return nil, fmt.Errorf("numerator degrees of freedom d1 must be strictly positive, got %v", d1)
	}
	if math.IsNaN(d2) || math.IsInf(d2, 0) || d2 <= 0 {
		return nil, fmt.Errorf("denominator degrees of freedom d2 must be strictly positive, got %v", d2)
	}
	return &FDistribution{D1: d1, D2: d2}, nil
}

// CDF returns P(F <= x), utilizing regularized incomplete beta I_z(d1/2, d2/2).
func (f *FDistribution) CDF(x float64) float64 {
	if math.IsNaN(x) {
		return math.NaN()
	}
	if x <= 0 {
		return 0.0
	}
	if math.IsInf(x, 1) {
		return 1.0
	}

	z := (f.D1 * x) / (f.D1*x + f.D2)
	return RegularizedIncompleteBeta(z, 0.5*f.D1, 0.5*f.D2)
}

// Survival returns P(F >= x) = 1 - CDF(x), computed stably via beta symmetry I_{1-z}(d2/2, d1/2).
// This is the p-value / Significance F in ANOVA and regression.
func (f *FDistribution) Survival(x float64) float64 {
	if math.IsNaN(x) {
		return math.NaN()
	}
	if x <= 0 {
		return 1.0
	}
	if math.IsInf(x, 1) {
		return 0.0
	}

	oneMinusZ := f.D2 / (f.D1*x + f.D2)
	return RegularizedIncompleteBeta(oneMinusZ, 0.5*f.D2, 0.5*f.D1)
}

// Mean returns the expected value of F: d2 / (d2 - 2) for d2 > 2.
func (f *FDistribution) Mean() float64 {
	if f.D2 > 2.0 {
		return f.D2 / (f.D2 - 2.0)
	}
	return math.NaN()
}

// Variance returns the variance of F for d2 > 4.
func (f *FDistribution) Variance() float64 {
	if f.D2 > 4.0 {
		num := 2.0 * f.D2 * f.D2 * (f.D1 + f.D2 - 2.0)
		denom := f.D1 * (f.D2 - 2.0) * (f.D2 - 2.0) * (f.D2 - 4.0)
		return num / denom
	}
	if f.D2 > 2.0 {
		return math.Inf(1)
	}
	return math.NaN()
}

// ExcelSurvival returns the Excel formula for upper-tail F probability: =F.DIST.RT(x, d1, d2).
func (f *FDistribution) ExcelSurvival(x float64) string {
	return fmt.Sprintf("=F.DIST.RT(%.4g, %.4g, %.4g)", x, f.D1, f.D2)
}
