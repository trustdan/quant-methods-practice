package mathengine

import (
	"math"
	"testing"
)

func TestComputeSampleSizeMean(t *testing.T) {
	// sigma = 25, E = 5, 95% confidence
	// z ~= 1.95996
	// exactN = (1.95996 * 25 / 5)^2 = (9.7998)^2 = 96.036
	// requiredN = 97
	res, err := ComputeSampleSizeMean(25.0, 5.0, 0.95)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequiredN != 97 {
		t.Errorf("expected RequiredN = 97, got %d (exact: %f)", res.RequiredN, res.ExactN)
	}
	if math.Abs(res.CriticalZ-1.95996) > 1e-4 {
		t.Errorf("expected z ~= 1.96, got %f", res.CriticalZ)
	}

	// Halving margin of error from 5 to 2.5 should quadruple sample size (96.036 * 4 = 384.15 => 385)
	resHalf, err := ComputeSampleSizeMean(25.0, 2.5, 0.95)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resHalf.RequiredN != 385 {
		t.Errorf("expected RequiredN = 385, got %d (exact: %f)", resHalf.RequiredN, resHalf.ExactN)
	}

	// Invalid inputs
	if _, err := ComputeSampleSizeMean(-10, 5, 0.95); err == nil {
		t.Error("expected error on negative sigma")
	}
	if _, err := ComputeSampleSizeMean(10, 0, 0.95); err == nil {
		t.Error("expected error on zero margin of error")
	}
	if _, err := ComputeSampleSizeMean(10, 5, 1.0); err == nil {
		t.Error("expected error on invalid confidence level 1.0")
	}
}

func TestComputeSampleSizeProportion(t *testing.T) {
	// Pilot p = 0.20, E = 0.03, 95% confidence
	// exactN = (1.95996^2 * 0.20 * 0.80) / 0.03^2 = 3.84146 * 0.16 / 0.0009 = 682.93 => 683
	res, err := ComputeSampleSizeProportion(0.20, 0.03, 0.95)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequiredN != 683 {
		t.Errorf("expected RequiredN = 683, got %d (exact: %f)", res.RequiredN, res.ExactN)
	}
	if res.Conservative {
		t.Errorf("expected Conservative = false, got true")
	}

	// Conservative planning (p = 0 or negative): uses p = 0.50
	// E = 0.04, 90% confidence (z ~= 1.64485)
	// exactN = 1.64485^2 * 0.25 / 0.0016 = 2.7055 * 0.25 / 0.0016 = 422.74 => 423
	resCons, err := ComputeSampleSizeProportion(0.0, 0.04, 0.90)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resCons.Conservative {
		t.Errorf("expected Conservative = true")
	}
	if resCons.RequiredN != 423 {
		t.Errorf("expected RequiredN = 423, got %d (exact: %f)", resCons.RequiredN, resCons.ExactN)
	}
}

func TestComputeSampleSizeTargetSE(t *testing.T) {
	// sigma = 15, targetSE = 2 => exact = (15/2)^2 = 7.5^2 = 56.25 => required = 57
	res, err := ComputeSampleSizeTargetSE(15.0, 2.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequiredN != 57 {
		t.Errorf("expected RequiredN = 57, got %d (exact: %f)", res.RequiredN, res.ExactN)
	}
	if math.Abs(res.ExactN-56.25) > 1e-6 {
		t.Errorf("expected exactN = 56.25, got %f", res.ExactN)
	}
}

func TestStandardErrorMean(t *testing.T) {
	// sigma = 20, n = 25 => SE = 20 / 5 = 4.0
	se1, err := StandardErrorMean(20.0, 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(se1-4.0) > 1e-6 {
		t.Errorf("expected SE = 4.0, got %f", se1)
	}

	// Quadruple n: n = 100 => SE = 20 / 10 = 2.0 (halves SE)
	se2, err := StandardErrorMean(20.0, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(se2-2.0) > 1e-6 {
		t.Errorf("expected SE = 2.0, got %f", se2)
	}
}

func TestStandardErrorProportion(t *testing.T) {
	// n = 400, p = 0.50 => SE = sqrt(0.25 / 400) = sqrt(0.000625) = 0.025
	se1, err := StandardErrorProportion(0.50, 400)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(se1-0.025) > 1e-6 {
		t.Errorf("expected SE = 0.025, got %f", se1)
	}

	// n = 400, p = 0.10 => SE = sqrt(0.09 / 400) = sqrt(0.000225) = 0.015
	se2, err := StandardErrorProportion(0.10, 400)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(se2-0.015) > 1e-6 {
		t.Errorf("expected SE = 0.015, got %f", se2)
	}
}
