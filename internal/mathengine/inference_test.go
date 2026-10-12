package mathengine

import (
	"math"
	"testing"
)

func TestSamplingDistributionMean(t *testing.T) {
	// Infinite population
	sd, err := ComputeSamplingDistributionMean(100.0, 15.0, 25, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedSE := 15.0 / math.Sqrt(25.0) // 3.0
	if math.Abs(sd.StandardErr-expectedSE) > 1e-9 {
		t.Errorf("expected SE %v, got %v", expectedSE, sd.StandardErr)
	}
	if sd.FPCApplied {
		t.Errorf("FPC should not be applied for infinite population")
	}

	// Finite population with n/N > 0.05
	// N = 100, n = 20 (n/N = 0.20 > 0.05)
	// FPC = sqrt((100 - 20) / (100 - 1)) = sqrt(80 / 99) = 0.898933
	sdFPC, err := ComputeSamplingDistributionMean(50.0, 10.0, 20, 100, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sdFPC.FPCApplied {
		t.Errorf("expected FPC to be applied")
	}
	expectedFPC := math.Sqrt(80.0 / 99.0)
	if math.Abs(sdFPC.FPCFactor-expectedFPC) > 1e-4 {
		t.Errorf("expected FPC factor %v, got %v", expectedFPC, sdFPC.FPCFactor)
	}
}

func TestConfidenceIntervalMeanZ(t *testing.T) {
	// Sample mean xBar = 100, sigma = 15, n = 36, 95% confidence
	// z_{0.975} = 1.95996
	// SE = 15 / sqrt(36) = 2.5
	// MOE = 1.95996 * 2.5 = 4.8999
	// Bounds: [95.1001, 104.8999]
	ci, err := ConfidenceIntervalMeanZ(100.0, 15.0, 36, 0.95)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(ci.StandardError-2.5) > 1e-9 {
		t.Errorf("expected SE 2.5, got %v", ci.StandardError)
	}
	if math.Abs(ci.CriticalValue-1.95996) > 1e-3 {
		t.Errorf("expected z* ~ 1.96, got %v", ci.CriticalValue)
	}
	if math.Abs(ci.LowerBound-95.10) > 0.02 {
		t.Errorf("expected lower bound ~ 95.10, got %v", ci.LowerBound)
	}
	if math.Abs(ci.UpperBound-104.90) > 0.02 {
		t.Errorf("expected upper bound ~ 104.90, got %v", ci.UpperBound)
	}
}

func TestConfidenceIntervalMeanT(t *testing.T) {
	// Sample mean xBar = 50, s = 10, n = 10, 95% confidence
	// df = 9, t_{0.975, 9} = 2.2622
	// SE = 10 / sqrt(10) = 3.162277
	// MOE = 2.2622 * 3.162277 = 7.1537
	// Bounds: [42.846, 57.154]
	ci, err := ConfidenceIntervalMeanT(50.0, 10.0, 10, 0.95)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *ci.DegreesOfFree != 9.0 {
		t.Errorf("expected df 9, got %v", *ci.DegreesOfFree)
	}
	if math.Abs(ci.CriticalValue-2.2622) > 1e-3 {
		t.Errorf("expected t* ~ 2.262, got %v", ci.CriticalValue)
	}
	if math.Abs(ci.LowerBound-42.846) > 0.02 {
		t.Errorf("expected lower bound ~ 42.85, got %v", ci.LowerBound)
	}
	if math.Abs(ci.UpperBound-57.154) > 0.02 {
		t.Errorf("expected upper bound ~ 57.15, got %v", ci.UpperBound)
	}
}

func TestConfidenceIntervalProportion(t *testing.T) {
	// Sample proportion: successes = 60, n = 100 -> pHat = 0.60
	// 95% CI: z* = 1.96
	// SE = sqrt(0.60 * 0.40 / 100) = sqrt(0.0024) = 0.0489898
	// MOE = 1.95996 * 0.0489898 = 0.096018
	// Bounds: [0.50398, 0.69602]
	ci, err := ConfidenceIntervalProportion(60, 100, 0.95)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ci.PointEstimate != 0.60 {
		t.Errorf("expected pHat 0.60, got %v", ci.PointEstimate)
	}
	if math.Abs(ci.MarginOfError-0.0960) > 1e-3 {
		t.Errorf("expected MOE ~ 0.0960, got %v", ci.MarginOfError)
	}
	if math.Abs(ci.LowerBound-0.5040) > 1e-3 {
		t.Errorf("expected lower bound ~ 0.5040, got %v", ci.LowerBound)
	}
}

func TestSimulateConfidenceIntervalCoverage(t *testing.T) {
	// 100 simulated 95% intervals with seed
	intervals, coverage, err := SimulateConfidenceIntervalCoverage(50.0, 10.0, 20, 0.95, 100, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(intervals) != 100 {
		t.Errorf("expected 100 intervals, got %d", len(intervals))
	}
	// Coverage should be close to 0.95 (e.g. 0.90 to 0.98)
	if coverage < 0.88 || coverage > 1.0 {
		t.Errorf("coverage out of expected range for 95%% CI: %v", coverage)
	}
}

func TestSamplingDistributionProportion(t *testing.T) {
	// p = 0.30, n = 200
	// SE = sqrt(0.30 * 0.70 / 200) = sqrt(0.00105) ~ 0.0324037
	// np = 60 >= 10, n(1-p) = 140 >= 10 -> isNormal = true
	sd, err := ComputeSamplingDistributionProportion(0.30, 200, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedSE := math.Sqrt(0.30 * 0.70 / 200.0)
	if math.Abs(sd.StandardErr-expectedSE) > 1e-9 {
		t.Errorf("expected SE %v, got %v", expectedSE, sd.StandardErr)
	}
	if !sd.IsNormal {
		t.Errorf("expected isNormal=true for np=60, n(1-p)=140")
	}
	if sd.FPCApplied {
		t.Errorf("FPC should not be applied when popSize = 0")
	}

	// z-score for pHat = 0.35
	z := sd.ZScore(0.35)
	expectedZ := (0.35 - 0.30) / expectedSE
	if math.Abs(z-expectedZ) > 1e-9 {
		t.Errorf("expected Z %v, got %v", expectedZ, z)
	}

	// ProbabilityAbove(0.35)
	pAbove := sd.ProbabilityAbove(0.35)
	if math.Abs(pAbove-0.0614) > 0.002 {
		t.Errorf("expected P(pHat > 0.35) ~ 0.0614, got %v", pAbove)
	}

	// Finite population with FPC: N = 500, n = 100, p = 0.40
	// n/N = 0.20 > 0.05
	sdFPC, err := ComputeSamplingDistributionProportion(0.40, 100, 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sdFPC.FPCApplied {
		t.Errorf("expected FPC applied")
	}
	expectedFPC := math.Sqrt(400.0 / 499.0)
	if math.Abs(sdFPC.FPCFactor-expectedFPC) > 1e-4 {
		t.Errorf("expected FPC factor %v, got %v", expectedFPC, sdFPC.FPCFactor)
	}
}

func TestSamplingDistributionMeanHelpers(t *testing.T) {
	// mu = 80, sigma = 24, n = 36 -> SE = 4.0
	sd, err := ComputeSamplingDistributionMean(80.0, 24.0, 36, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	z := sd.ZScore(86.0)
	if math.Abs(z-1.50) > 1e-9 {
		t.Errorf("expected Z = 1.50, got %v", z)
	}
	pAbove := sd.ProbabilityAbove(86.0)
	if math.Abs(pAbove-0.0668) > 0.001 {
		t.Errorf("expected P(Xbar > 86) ~ 0.0668, got %v", pAbove)
	}
	pBelow := sd.ProbabilityBelow(86.0)
	if math.Abs(pBelow-0.9332) > 0.001 {
		t.Errorf("expected P(Xbar <= 86) ~ 0.9332, got %v", pBelow)
	}
}

