package mathengine

import (
	"fmt"
	"math"
)

// LinearCombMean computes E[aX + bY + c] = a*E[X] + b*E[Y] + c.
// Linearity of expectation holds universally, regardless of whether X and Y are independent.
func LinearCombMean(a, meanX, b, meanY, c float64) float64 {
	return a*meanX + b*meanY + c
}

// LinearCombVariance computes Var(aX + bY + c) = a^2*Var(X) + b^2*Var(Y) + 2*a*b*Cov(X, Y).
// Invariant: The additive constant c contributes zero variance.
// Requires Var(X) >= 0, Var(Y) >= 0, and Cov(X, Y) within Cauchy-Schwarz bounds.
func LinearCombVariance(a, varX, b, varY, covXY float64) (float64, error) {
	if math.IsNaN(varX) || varX < 0.0 {
		return 0, fmt.Errorf("Var(X) must be non-negative, got %v", varX)
	}
	if math.IsNaN(varY) || varY < 0.0 {
		return 0, fmt.Errorf("Var(Y) must be non-negative, got %v", varY)
	}
	maxCov := math.Sqrt(varX * varY)
	if math.IsNaN(covXY) || math.Abs(covXY) > maxCov+1e-9 {
		return 0, fmt.Errorf("covariance Cov(X, Y)=%v violates Cauchy-Schwarz bound [-%v, %v]", covXY, maxCov, maxCov)
	}

	v := a*a*varX + b*b*varY + 2.0*a*b*covXY
	if v < 0.0 && v > -1e-9 {
		v = 0.0
	}
	return v, nil
}

// LinearCombVarianceIndependent computes Var(aX + bY + c) under independence (Cov(X, Y) = 0).
// Invariant: Independence is sufficient for zero covariance; zero covariance alone does not establish independence.
func LinearCombVarianceIndependent(a, varX, b, varY float64) (float64, error) {
	return LinearCombVariance(a, varX, b, varY, 0.0)
}

// LinearCombStdDev computes SD(aX + bY + c) = sqrt(Var(aX + bY + c)).
func LinearCombStdDev(a, varX, b, varY, covXY float64) (float64, error) {
	v, err := LinearCombVariance(a, varX, b, varY, covXY)
	if err != nil {
		return 0, err
	}
	return math.Sqrt(v), nil
}

// CovarianceFromCorrelation computes Cov(X, Y) = rho * SD(X) * SD(Y) where rho in [-1, 1].
func CovarianceFromCorrelation(rho, varX, varY float64) (float64, error) {
	if math.IsNaN(rho) || rho < -1.0-1e-9 || rho > 1.0+1e-9 {
		return 0, fmt.Errorf("correlation coefficient rho must be in [-1, 1], got %v", rho)
	}
	if varX < 0.0 || varY < 0.0 {
		return 0, fmt.Errorf("variances must be non-negative, got Var(X)=%v, Var(Y)=%v", varX, varY)
	}
	return rho * math.Sqrt(varX*varY), nil
}
