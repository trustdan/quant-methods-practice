package mathengine

import (
	"math"
	"testing"
)

func TestComputeTwoSampleTTestPooled(t *testing.T) {
	// Sample 1: Mean 78.4, Var 24.5, N 15
	// Sample 2: Mean 72.1, Var 20.2, N 15
	// df = 15 + 15 - 2 = 28
	// pooled var = (14*24.5 + 14*20.2) / 28 = (343 + 282.8) / 28 = 625.8 / 28 = 22.35
	// SE = sqrt(22.35 * (1/15 + 1/15)) = sqrt(22.35 * 2 / 15) = sqrt(2.98) = 1.7262676...
	// t = (78.4 - 72.1) / 1.7262676 = 6.3 / 1.7262676 = 3.64949...
	res, err := ComputeTwoSampleTTestPooled(78.4, 24.5, 15, 72.1, 20.2, 15, 0.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.DF != 28 {
		t.Errorf("expected df 28, got %d", res.DF)
	}
	if math.Abs(res.PooledVariance-22.35) > 1e-6 {
		t.Errorf("expected pooled variance 22.35, got %v", res.PooledVariance)
	}
	if math.Abs(res.SE-math.Sqrt(2.98)) > 1e-4 {
		t.Errorf("expected SE %v, got %v", math.Sqrt(2.98), res.SE)
	}
	if res.TStat < 3.64 || res.TStat > 3.66 {
		t.Errorf("expected tStat around 3.65, got %v", res.TStat)
	}
	if res.PValueTwoTail >= 0.01 {
		t.Errorf("expected p-value two-tail < 0.01, got %v", res.PValueTwoTail)
	}
}

func TestComputePairedTTest(t *testing.T) {
	// 10 subjects, mean diff = 4.0, sd diff = 3.0
	// df = 9, SE = 3.0 / sqrt(10) = 0.94868
	// t = 4.0 / 0.94868 = 4.21637
	res, err := ComputePairedTTest(4.0, 3.0, 10, 0.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.DF != 9 {
		t.Errorf("expected df 9, got %d", res.DF)
	}
	if math.Abs(res.SEDiff-3.0/math.Sqrt(10)) > 1e-6 {
		t.Errorf("expected SE %v, got %v", 3.0/math.Sqrt(10), res.SEDiff)
	}
	if math.Abs(res.TStat-4.0/(3.0/math.Sqrt(10))) > 1e-4 {
		t.Errorf("expected t %v, got %v", 4.0/(3.0/math.Sqrt(10)), res.TStat)
	}
	if res.PValueTwo >= 0.01 {
		t.Errorf("expected p-value two tail < 0.01, got %v", res.PValueTwo)
	}
}

func TestComputeTwoSampleProportionTest(t *testing.T) {
	// Sample 1: 48 / 600 = 0.08
	// Sample 2: 72 / 600 = 0.12
	// pooled = (48 + 72) / 1200 = 120 / 1200 = 0.10
	// SE = sqrt(0.10 * 0.90 * (1/600 + 1/600)) = sqrt(0.09 * (2/600)) = sqrt(0.0003) = 0.0173205
	// z = (0.08 - 0.12) / 0.0173205 = -0.04 / 0.0173205 = -2.3094
	res, err := ComputeTwoSampleProportionTest(48, 600, 72, 600)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(res.PooledP-0.10) > 1e-6 {
		t.Errorf("expected pooled p 0.10, got %v", res.PooledP)
	}
	if math.Abs(res.ZStat-(-2.3094)) > 1e-3 {
		t.Errorf("expected z around -2.3094, got %v", res.ZStat)
	}
}
