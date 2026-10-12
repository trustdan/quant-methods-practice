package mathengine

import (
	"math"
	"testing"
)

func TestDiscreteRVDealership(t *testing.T) {
	// Values: 0, 1, 2, 3, 4
	// Probs: 0.10, 0.25, 0.35, 0.20, 0.10
	// E[X] = 0(0.10) + 1(0.25) + 2(0.35) + 3(0.20) + 4(0.10)
	//      = 0 + 0.25 + 0.70 + 0.60 + 0.40 = 1.95
	// E[X^2] = 0 + 1(0.25) + 4(0.35) + 9(0.20) + 16(0.10)
	//        = 0 + 0.25 + 1.40 + 1.80 + 1.60 = 5.05
	// Var(X) = 5.05 - (1.95)^2 = 5.05 - 3.8025 = 1.2475
	// SD(X) = sqrt(1.2475) = 1.11691539518...
	values := []float64{0, 1, 2, 3, 4}
	probs := []float64{0.10, 0.25, 0.35, 0.20, 0.10}

	rv, err := NewDiscreteRV(values, probs)
	if err != nil {
		t.Fatalf("unexpected error creating DiscreteRV: %v", err)
	}

	mu := rv.ExpectedValue()
	if math.Abs(mu-1.95) > 1e-9 {
		t.Errorf("expected E[X] = 1.95, got %v", mu)
	}

	e2 := rv.ExpectedValueSquared()
	if math.Abs(e2-5.05) > 1e-9 {
		t.Errorf("expected E[X^2] = 5.05, got %v", e2)
	}

	variance := rv.Variance()
	if math.Abs(variance-1.2475) > 1e-9 {
		t.Errorf("expected Var(X) = 1.2475, got %v", variance)
	}

	sd := rv.StdDev()
	if math.Abs(sd-math.Sqrt(1.2475)) > 1e-6 {
		t.Errorf("expected SD(X) = %v, got %v", math.Sqrt(1.2475), sd)
	}
}

func TestSolveMissingProbability(t *testing.T) {
	// [0.10, 0.25, 0.35, -1, 0.10] -> missing idx 3, solved = 0.20
	probs := []float64{0.10, 0.25, 0.35, -1.0, 0.10}
	idx, p, err := SolveMissingProbability(probs)
	if err != nil {
		t.Fatalf("unexpected error solving missing prob: %v", err)
	}
	if idx != 3 {
		t.Errorf("expected missing index 3, got %d", idx)
	}
	if math.Abs(p-0.20) > 1e-9 {
		t.Errorf("expected solved prob 0.20, got %v", p)
	}
}
