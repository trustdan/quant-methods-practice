package mathengine

import (
	"math"
	"testing"
)

func TestPerformMeanZTest(t *testing.T) {
	// H0: mu = 50 vs Ha: mu != 50 (two-tailed)
	// xBar = 53, sigma = 10, n = 25, alpha = 0.05
	// SE = 10 / 5 = 2.0
	// Z = (53 - 50) / 2.0 = 1.5
	// p-value = 2 * (1 - Phi(1.5)) = 2 * (1 - 0.93319) = 0.1336
	// At alpha = 0.05, p-value > alpha -> Fail to reject H0
	res, err := PerformMeanZTest(50.0, 53.0, 10.0, 25, TailTwoSided, 0.05)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(res.TestStatistic-1.5) > 1e-6 {
		t.Errorf("expected Z = 1.5, got %v", res.TestStatistic)
	}
	if math.Abs(res.PValue-0.1336) > 1e-3 {
		t.Errorf("expected p-value ~ 0.1336, got %v", res.PValue)
	}
	if res.RejectNull {
		t.Errorf("expected fail to reject H0")
	}
	if res.Decision != "Fail to reject H0" {
		t.Errorf("expected Decision 'Fail to reject H0', got %q", res.Decision)
	}
}

func TestPerformMeanTTest(t *testing.T) {
	// H0: mu = 100 vs Ha: mu > 100 (right-tailed)
	// xBar = 106, s = 8, n = 16, alpha = 0.05
	// df = 15
	// SE = 8 / 4 = 2.0
	// t = (106 - 100) / 2.0 = 3.0
	// t_{crit, 0.05, 15} = 1.753
	// t = 3.0 > 1.753 -> Reject H0
	res, err := PerformMeanTTest(100.0, 106.0, 8.0, 16, TailRight, 0.05)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(res.TestStatistic-3.0) > 1e-6 {
		t.Errorf("expected t = 3.0, got %v", res.TestStatistic)
	}
	if *res.DegreesOfFree != 15.0 {
		t.Errorf("expected df 15, got %v", *res.DegreesOfFree)
	}
	if !res.RejectNull {
		t.Errorf("expected reject H0")
	}
	if res.Decision != "Reject H0" {
		t.Errorf("expected Decision 'Reject H0', got %q", res.Decision)
	}
	if res.PValue >= 0.05 {
		t.Errorf("expected p-value < 0.05, got %v", res.PValue)
	}
}

func TestPerformProportionTest(t *testing.T) {
	// H0: p = 0.50 vs Ha: p < 0.50 (left-tailed)
	// successes = 40, n = 100 (pHat = 0.40), alpha = 0.05
	// SE = sqrt(0.5 * 0.5 / 100) = 0.05
	// Z = (0.40 - 0.50) / 0.05 = -2.0
	// p-value = Phi(-2.0) = 0.02275
	// p-value < 0.05 -> Reject H0
	res, err := PerformProportionTest(0.50, 40, 100, TailLeft, 0.05)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(res.TestStatistic - -2.0) > 1e-6 {
		t.Errorf("expected Z = -2.0, got %v", res.TestStatistic)
	}
	if math.Abs(res.PValue-0.0228) > 1e-3 {
		t.Errorf("expected p-value ~ 0.0228, got %v", res.PValue)
	}
	if !res.RejectNull {
		t.Errorf("expected reject H0")
	}
}
