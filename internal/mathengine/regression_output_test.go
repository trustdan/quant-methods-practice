package mathengine

import (
	"math"
	"testing"
)

func TestEvaluateRegressionOutput_SimpleRegression(t *testing.T) {
	// Simple linear regression: n = 25, k = 1
	// SSR = 144, SSE = 368, SST = 512
	// dfReg = 1, dfRes = 23, dfTotal = 24
	// MSR = 144, MSE = 368 / 23 = 16.0
	// F = 144 / 16.0 = 9.0
	// Predictor: b1 = 2.4, SE = 0.8 -> t = 3.0. t^2 = 9.0 == F!
	sst := 512.0
	params := RegressionOutputParams{
		N:               25,
		K:               1,
		SSR:             144.0,
		SSE:             368.0,
		SST:             &sst,
		ConfidenceLevel: 0.95,
		Intercept: OutputPredictorInput{
			Name:   "Intercept",
			Coeff:  10.0,
			StdErr: 2.5,
		},
		Predictors: []OutputPredictorInput{
			{
				Name:   "Advertising",
				Coeff:  2.4,
				StdErr: 0.8,
			},
		},
	}

	table, err := EvaluateRegressionOutput(params)
	if err != nil {
		t.Fatalf("unexpected error evaluating regression output: %v", err)
	}

	if table.DFReg != 1 || table.DFRes != 23 || table.DFTotal != 24 {
		t.Errorf("degrees of freedom mismatch: reg=%d, res=%d, total=%d", table.DFReg, table.DFRes, table.DFTotal)
	}

	if math.Abs(table.MSE-16.0) > 1e-9 {
		t.Errorf("expected MSE == 16.0, got %v", table.MSE)
	}
	if math.Abs(table.StdErrEst-4.0) > 1e-9 {
		t.Errorf("expected s_e == 4.0, got %v", table.StdErrEst)
	}
	if math.Abs(table.FStat-9.0) > 1e-9 {
		t.Errorf("expected F == 9.0, got %v", table.FStat)
	}

	expectedR2 := 144.0 / 512.0 // 0.28125
	if math.Abs(table.RSquared-expectedR2) > 1e-9 {
		t.Errorf("expected R^2 == %v, got %v", expectedR2, table.RSquared)
	}

	// Verify t-stat and F equivalence
	pred := table.Predictors[0]
	if math.Abs(pred.TStat-3.0) > 1e-9 {
		t.Errorf("expected slope t == 3.0, got %v", pred.TStat)
	}
	if math.Abs(pred.PValue-table.SignificanceF) > 1e-9 {
		t.Errorf("expected slope p-value (%v) to equal Significance F (%v)", pred.PValue, table.SignificanceF)
	}

	// Slope CI: b1 ± t* * SE
	// df = 23, 95% two-tailed t* ≈ 2.068658
	expectedHalfWidth := table.TCritical * 0.8
	if math.Abs((pred.UpperCI-pred.Coeff)-expectedHalfWidth) > 1e-6 {
		t.Errorf("expected CI half-width %v, got %v", expectedHalfWidth, pred.UpperCI-pred.Coeff)
	}

	// Prediction at x = 5.0: yHat = 10.0 + 2.4 * 5.0 = 22.0
	yHat, err := table.Predict([]float64{5.0})
	if err != nil {
		t.Fatalf("prediction error: %v", err)
	}
	if math.Abs(yHat-22.0) > 1e-9 {
		t.Errorf("expected yHat == 22.0, got %v", yHat)
	}
}

func TestEvaluateRegressionOutput_MultipleRegression(t *testing.T) {
	// Multiple regression: n = 30, k = 3
	// dfReg = 3, dfRes = 26, dfTotal = 29
	params := RegressionOutputParams{
		N:               30,
		K:               3,
		SSR:             450.0,
		SSE:             150.0,
		ConfidenceLevel: 0.95,
		Intercept: OutputPredictorInput{
			Name:   "Intercept",
			Coeff:  50.0,
			StdErr: 10.0,
		},
		Predictors: []OutputPredictorInput{
			{Name: "AdSpend", Coeff: 3.5, StdErr: 0.5},
			{Name: "StoreSize", Coeff: 1.2, StdErr: 0.4},
			{Name: "Competitors", Coeff: -2.0, StdErr: 1.0},
		},
	}

	table, err := EvaluateRegressionOutput(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if table.DFReg != 3 || table.DFRes != 26 || table.DFTotal != 29 {
		t.Errorf("degrees of freedom mismatch: reg=%d, res=%d, total=%d", table.DFReg, table.DFRes, table.DFTotal)
	}

	// MSR = 450 / 3 = 150
	// MSE = 150 / 26 ≈ 5.76923
	// F = 150 / (150/26) = 26.0
	if math.Abs(table.FStat-26.0) > 1e-9 {
		t.Errorf("expected F == 26.0, got %v", table.FStat)
	}

	// R^2 = 450 / 600 = 0.75
	if math.Abs(table.RSquared-0.75) > 1e-9 {
		t.Errorf("expected R^2 == 0.75, got %v", table.RSquared)
	}

	// Prediction at [10, 20, 2]:
	// 50 + 3.5*10 + 1.2*20 + (-2.0)*2 = 50 + 35 + 24 - 4 = 105
	yHat, err := table.Predict([]float64{10.0, 20.0, 2.0})
	if err != nil {
		t.Fatalf("prediction error: %v", err)
	}
	if math.Abs(yHat-105.0) > 1e-9 {
		t.Errorf("expected yHat == 105.0, got %v", yHat)
	}
}

func TestEvaluateRegressionOutput_ValidationFailures(t *testing.T) {
	// k = 0
	_, err := EvaluateRegressionOutput(RegressionOutputParams{K: 0})
	if err == nil {
		t.Error("expected error for k=0")
	}

	// k = 5 (exceeds max 4)
	_, err = EvaluateRegressionOutput(RegressionOutputParams{K: 5})
	if err == nil {
		t.Error("expected error for k=5")
	}

	// n too small (n < k + 3)
	_, err = EvaluateRegressionOutput(RegressionOutputParams{
		N: 3, K: 1, SSR: 10, SSE: 10,
		Predictors: []OutputPredictorInput{{Name: "X1", Coeff: 1, StdErr: 1}},
	})
	if err == nil {
		t.Error("expected error for n < k + 3")
	}

	// Inconsistent SST
	sstBad := 500.0
	_, err = EvaluateRegressionOutput(RegressionOutputParams{
		N: 20, K: 1, SSR: 100, SSE: 100, SST: &sstBad, // 100 + 100 != 500
		Predictors: []OutputPredictorInput{{Name: "X1", Coeff: 1, StdErr: 1}},
	})
	if err == nil {
		t.Error("expected error for inconsistent SST")
	}

	// k = 1 consistency failure (t^2 != F)
	_, err = EvaluateRegressionOutput(RegressionOutputParams{
		N: 25, K: 1, SSR: 144, SSE: 368,
		Predictors: []OutputPredictorInput{{Name: "X1", Coeff: 10.0, StdErr: 0.8}}, // t = 12.5 -> t^2 = 156 != 9.0
	})
	if err == nil {
		t.Error("expected error for k=1 when t^2 != F")
	}
}
