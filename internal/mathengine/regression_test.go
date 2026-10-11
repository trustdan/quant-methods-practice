package mathengine

import (
	"math"
	"testing"
)

func TestLinearRegressionKnownDataset(t *testing.T) {
	// Standard textbook dataset:
	// X: [1, 2, 3, 4, 5]
	// Y: [2, 3, 5, 4, 6]
	// meanX = 3, meanY = 4
	// dx = [-2, -1, 0, 1, 2] -> SSxx = 4 + 1 + 0 + 1 + 4 = 10
	// dy = [-2, -1, 1, 0, 2] -> SSyy = 4 + 1 + 1 + 0 + 4 = 10
	// dx*dy = [4, 1, 0, 0, 4] -> SSxy = 9
	// Slope b1 = SSxy / SSxx = 9 / 10 = 0.9
	// Intercept b0 = 4 - 0.9 * 3 = 1.3
	// r = 9 / sqrt(10 * 10) = 0.9
	// R^2 = 0.81
	// SSR = 0.9 * 9 = 8.1
	// SSE = 10 - 8.1 = 1.9
	// df = 3
	// seEst = sqrt(1.9 / 3) = 0.795822

	points := []DataPoint{
		{X: 1, Y: 2},
		{X: 2, Y: 3},
		{X: 3, Y: 5},
		{X: 4, Y: 4},
		{X: 5, Y: 6},
	}

	reg, err := ComputeLinearRegression(points)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(reg.MeanX-3.0) > 1e-9 {
		t.Errorf("expected meanX 3.0, got %v", reg.MeanX)
	}
	if math.Abs(reg.MeanY-4.0) > 1e-9 {
		t.Errorf("expected meanY 4.0, got %v", reg.MeanY)
	}
	if math.Abs(reg.Slope-0.9) > 1e-9 {
		t.Errorf("expected slope 0.9, got %v", reg.Slope)
	}
	if math.Abs(reg.Intercept-1.3) > 1e-9 {
		t.Errorf("expected intercept 1.3, got %v", reg.Intercept)
	}
	if math.Abs(reg.Correlation-0.9) > 1e-9 {
		t.Errorf("expected r 0.9, got %v", reg.Correlation)
	}
	if math.Abs(reg.RSquared-0.81) > 1e-9 {
		t.Errorf("expected R^2 0.81, got %v", reg.RSquared)
	}
	if math.Abs(reg.SSE-1.9) > 1e-9 {
		t.Errorf("expected SSE 1.9, got %v", reg.SSE)
	}

	// Residuals sum to approximately 0 in OLS with intercept
	sumResid := 0.0
	for _, res := range reg.Residuals {
		sumResid += res.Residual
	}
	if math.Abs(sumResid) > 1e-3 {
		t.Errorf("residuals should sum to ~0, got %v", sumResid)
	}

	// Prediction at X = 6: yHat = 1.3 + 0.9*6 = 6.7
	yHat, sePred := reg.Predict(6.0)
	if math.Abs(yHat-6.7) > 1e-9 {
		t.Errorf("expected prediction 6.7, got %v", yHat)
	}
	if sePred <= 0 {
		t.Errorf("expected positive prediction standard error, got %v", sePred)
	}
}

func TestLinearRegressionIdenticalXFails(t *testing.T) {
	points := []DataPoint{
		{X: 2, Y: 3},
		{X: 2, Y: 5},
		{X: 2, Y: 7},
	}
	_, err := ComputeLinearRegression(points)
	if err == nil {
		t.Errorf("expected error for zero variance in X, got nil")
	}
}
