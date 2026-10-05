package mathengine_test

import (
	"math"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

func TestLinearCombinationExpectation(t *testing.T) {
	// E[2X + 3Y + 5] = 2*E[X] + 3*E[Y] + 5
	meanX := 4.0
	meanY := 7.0
	got := mathengine.LinearCombMean(2.0, meanX, 3.0, meanY, 5.0)
	want := 2.0*4.0 + 3.0*7.0 + 5.0 // 8 + 21 + 5 = 34
	if math.Abs(got-want) > 1e-12 {
		t.Errorf("LinearCombMean = %v, want %v", got, want)
	}
}

func TestLinearCombinationVariance(t *testing.T) {
	// Var(aX + bY + c) = a^2*Var(X) + b^2*Var(Y) + 2*a*b*Cov(X, Y)
	varX := 4.0
	varY := 9.0
	covXY := 3.0 // valid, since sqrt(4*9) = 6 >= 3

	v, err := mathengine.LinearCombVariance(2.0, varX, -1.0, varY, covXY)
	if err != nil {
		t.Fatalf("LinearCombVariance failed: %v", err)
	}
	// 4*(4) + 1*(9) + 2*(2)*(-1)*(3) = 16 + 9 - 12 = 13
	wantV := 13.0
	if math.Abs(v-wantV) > 1e-12 {
		t.Errorf("LinearCombVariance = %v, want %v", v, wantV)
	}

	// Constant c adds zero variance: Var(X + c) = Var(X)
	vInd, err := mathengine.LinearCombVarianceIndependent(1.0, varX, 0.0, 0.0)
	if err != nil {
		t.Fatalf("LinearCombVarianceIndependent failed: %v", err)
	}
	if math.Abs(vInd-varX) > 1e-12 {
		t.Errorf("Var(X + c) = %v, want %v", vInd, varX)
	}
}

func TestVarianceSumDistinguishesDependence(t *testing.T) {
	// Invariant from NUMERICS.md:
	// "Variance-sum cases distinguish equal means but different dependence."
	meanX, varX := 10.0, 4.0
	meanY, varY := 10.0, 4.0

	// Equal means: E[X + Y] = 20 for all three cases
	meanSum := mathengine.LinearCombMean(1.0, meanX, 1.0, meanY, 0.0)
	if meanSum != 20.0 {
		t.Errorf("Mean sum = %v, want 20.0", meanSum)
	}

	// Case 1: Independent (Cov = 0)
	varInd, err := mathengine.LinearCombVarianceIndependent(1.0, varX, 1.0, varY)
	if err != nil {
		t.Fatalf("Independent variance failed: %v", err)
	}
	wantInd := 8.0 // 4 + 4
	if varInd != wantInd {
		t.Errorf("Independent Var(X+Y) = %v, want %v", varInd, wantInd)
	}

	// Case 2: Positively correlated (Cov = 2)
	varPos, err := mathengine.LinearCombVariance(1.0, varX, 1.0, varY, 2.0)
	if err != nil {
		t.Fatalf("Positively correlated variance failed: %v", err)
	}
	wantPos := 12.0 // 4 + 4 + 2(2) = 12
	if varPos != wantPos {
		t.Errorf("Positively correlated Var(X+Y) = %v, want %v", varPos, wantPos)
	}

	// Case 3: Negatively correlated (Cov = -2)
	varNeg, err := mathengine.LinearCombVariance(1.0, varX, 1.0, varY, -2.0)
	if err != nil {
		t.Fatalf("Negatively correlated variance failed: %v", err)
	}
	wantNeg := 4.0 // 4 + 4 + 2(-2) = 4
	if varNeg != wantNeg {
		t.Errorf("Negatively correlated Var(X+Y) = %v, want %v", varNeg, wantNeg)
	}

	if varInd == varPos || varInd == varNeg || varPos == varNeg {
		t.Errorf("Variances must distinguish dependence: ind=%v, pos=%v, neg=%v", varInd, varPos, varNeg)
	}
}

func TestCauchySchwarzCovarianceViolation(t *testing.T) {
	varX := 4.0
	varY := 9.0
	// maxCov = sqrt(4*9) = 6.0. Cov=7.0 violates Cauchy-Schwarz.
	_, err := mathengine.LinearCombVariance(1.0, varX, 1.0, varY, 7.0)
	if err == nil {
		t.Errorf("Expected Cauchy-Schwarz violation error, got nil")
	}
}

func TestCovarianceFromCorrelation(t *testing.T) {
	cov, err := mathengine.CovarianceFromCorrelation(0.5, 4.0, 9.0)
	if err != nil {
		t.Fatalf("CovarianceFromCorrelation failed: %v", err)
	}
	// 0.5 * 2 * 3 = 3.0
	if math.Abs(cov-3.0) > 1e-12 {
		t.Errorf("CovarianceFromCorrelation = %v, want 3.0", cov)
	}

	// Invalid correlation rho > 1
	if _, err := mathengine.CovarianceFromCorrelation(1.1, 4.0, 9.0); err == nil {
		t.Errorf("Expected error for rho > 1, got nil")
	}
}
