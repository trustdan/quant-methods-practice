package mathengine_test

import (
	"math"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

func TestPoissonReferenceAndMoments(t *testing.T) {
	p, err := mathengine.NewPoisson(2.0)
	if err != nil {
		t.Fatalf("NewPoisson(2.0) failed: %v", err)
	}

	if p.Mean() != 2.0 || p.Variance() != 2.0 {
		t.Errorf("Poisson moments incorrect: Mean=%v, Var=%v (want 2.0)", p.Mean(), p.Variance())
	}
	if math.Abs(p.StdDev()-math.Sqrt(2.0)) > 1e-12 {
		t.Errorf("Poisson StdDev=%v, want %v", p.StdDev(), math.Sqrt(2.0))
	}

	// P(X=1) = lambda * exp(-lambda) = 2 * exp(-2)
	wantP1 := 2.0 * math.Exp(-2.0)
	gotP1 := p.PMF(1)
	if math.Abs(gotP1-wantP1) > 1e-12 {
		t.Errorf("PMF(1) = %v, want %v", gotP1, wantP1)
	}

	// P(X=2) = (lambda^2 / 2!) * exp(-lambda) = 2 * exp(-2) = P(X=1)
	gotP2 := p.PMF(2)
	if math.Abs(gotP2-wantP1) > 1e-12 {
		t.Errorf("PMF(2) = %v, want %v", gotP2, wantP1)
	}
}

func TestPoissonZeroRateBoundary(t *testing.T) {
	pZero, err := mathengine.NewPoisson(0.0)
	if err != nil {
		t.Fatalf("NewPoisson(0.0) failed: %v", err)
	}

	if pZero.PMF(0) != 1.0 {
		t.Errorf("PMF(0) = %v, want 1.0", pZero.PMF(0))
	}
	if pZero.PMF(1) != 0.0 {
		t.Errorf("PMF(1) = %v, want 0.0", pZero.PMF(1))
	}
	if pZero.PMF(-1) != 0.0 {
		t.Errorf("PMF(-1) = %v, want 0.0", pZero.PMF(-1))
	}
	if pZero.CDF(0) != 1.0 {
		t.Errorf("CDF(0) = %v, want 1.0", pZero.CDF(0))
	}
	if pZero.Survival(0) != 1.0 {
		t.Errorf("Survival(0) = %v, want 1.0", pZero.Survival(0))
	}
	if pZero.Survival(1) != 0.0 {
		t.Errorf("Survival(1) = %v, want 0.0", pZero.Survival(1))
	}
}

func TestPoissonNormalizationAndComplement(t *testing.T) {
	p, _ := mathengine.NewPoisson(4.0)

	sum := 0.0
	for k := 0; k <= 30; k++ {
		sum += p.PMF(k)
	}
	if math.Abs(sum-1.0) > 1e-10 {
		t.Errorf("Poisson(4.0) sum of PMFs = %v, want 1.0", sum)
	}

	// Complement identity
	for k := 0; k <= 10; k++ {
		cdf := p.CDF(k)
		tail := p.Tail(k)
		if math.Abs((cdf+tail)-1.0) > 1e-12 {
			t.Errorf("Poisson complement failed at k=%d: CDF=%v, Tail=%v, sum=%v",
				k, cdf, tail, cdf+tail)
		}
	}
}

func TestPoissonEventEvaluation(t *testing.T) {
	p, _ := mathengine.NewPoisson(2.0)

	pmf1 := p.PMF(1)
	eventProb := p.EvaluateEvent(mathengine.Exactly(1))
	if math.Abs(eventProb-pmf1) > 1e-12 {
		t.Errorf("EvaluateEvent(Exactly(1)) = %v, want %v", eventProb, pmf1)
	}

	cdf2 := p.CDF(2)
	eventAtMost := p.EvaluateEvent(mathengine.AtMost(2))
	if math.Abs(eventAtMost-cdf2) > 1e-12 {
		t.Errorf("EvaluateEvent(AtMost(2)) = %v, want %v", eventAtMost, cdf2)
	}
}

func TestSumIndependentPoissons(t *testing.T) {
	p1, _ := mathengine.NewPoisson(1.5)
	p2, _ := mathengine.NewPoisson(2.5)

	sumP, err := mathengine.SumIndependentPoissons(p1, p2)
	if err != nil {
		t.Fatalf("SumIndependentPoissons failed: %v", err)
	}
	if sumP.Lambda() != 4.0 {
		t.Errorf("sumP.Lambda() = %v, want 4.0", sumP.Lambda())
	}
}
