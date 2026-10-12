package mathengine

import (
	"math"
	"testing"
)

func TestDescriptiveSummary(t *testing.T) {
	// Sample data: [12, 16, 18, 20, 22, 32]
	// n = 6
	// sum = 120
	// mean = 20.0
	// median = (18 + 20) / 2 = 19.0
	// deviations from 20: -8, -4, -2, 0, 2, 12
	// squared deviations: 64, 16, 4, 0, 4, 144 -> sum = 232
	// sample variance = 232 / 5 = 46.4
	// sample sd = sqrt(46.4) = 6.81175454637
	// population variance = 232 / 6 = 38.6666666667
	// population sd = sqrt(232/6) = 6.218252702
	// CV = (6.8117545 / 20.0) * 100 = 34.05877%
	data := []float64{12, 16, 18, 20, 22, 32}

	summary, err := ComputeDescriptiveSummary(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.Count != 6 {
		t.Errorf("expected count 6, got %d", summary.Count)
	}
	if math.Abs(summary.Mean-20.0) > 1e-9 {
		t.Errorf("expected mean 20.0, got %v", summary.Mean)
	}
	if math.Abs(summary.Median-19.0) > 1e-9 {
		t.Errorf("expected median 19.0, got %v", summary.Median)
	}
	if math.Abs(summary.SampleVariance-46.4) > 1e-9 {
		t.Errorf("expected sample variance 46.4, got %v", summary.SampleVariance)
	}
	if math.Abs(summary.SampleStdDev-math.Sqrt(46.4)) > 1e-6 {
		t.Errorf("expected sample sd %v, got %v", math.Sqrt(46.4), summary.SampleStdDev)
	}
	if math.Abs(summary.PopulationVariance-232.0/6.0) > 1e-6 {
		t.Errorf("expected pop variance %v, got %v", 232.0/6.0, summary.PopulationVariance)
	}
	if math.Abs(summary.Range-20.0) > 1e-9 {
		t.Errorf("expected range 20.0, got %v", summary.Range)
	}
}

func TestMeanMedianOutlier(t *testing.T) {
	// Store monthly sales: [18, 20, 22, 24, 25, 26, 120]
	// n = 7
	// sum = 255
	// mean = 255 / 7 = 36.42857...
	// median = 24.0 (sorted 4th element)
	data := []float64{18, 20, 22, 24, 25, 26, 120}

	mean, err := Mean(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(mean-255.0/7.0) > 1e-6 {
		t.Errorf("expected mean %v, got %v", 255.0/7.0, mean)
	}

	median, err := Median(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(median-24.0) > 1e-9 {
		t.Errorf("expected median 24.0, got %v", median)
	}
}

func TestCoefficientOfVariation(t *testing.T) {
	// Stock A: Mean 12%, SD 4% -> CV = (4/12)*100 = 33.333%
	// Stock B: Mean 4%, SD 2%  -> CV = (2/4)*100 = 50.0%
	cvA, err := CoefficientOfVariation(12.0, 4.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(cvA-100.0/3.0) > 1e-4 {
		t.Errorf("expected CV A 33.333%%, got %v", cvA)
	}

	cvB, err := CoefficientOfVariation(4.0, 2.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(cvB-50.0) > 1e-4 {
		t.Errorf("expected CV B 50.0%%, got %v", cvB)
	}
}

func TestWeightedMeanAndZScore(t *testing.T) {
	vals := []float64{10, 20, 30}
	weights := []float64{1, 2, 3}
	// sum(w*x) = 10 + 40 + 90 = 140; sum(w) = 6; weighted mean = 140 / 6 = 23.3333
	wm, err := WeightedMean(vals, weights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(wm-140.0/6.0) > 1e-4 {
		t.Errorf("expected weighted mean %v, got %v", 140.0/6.0, wm)
	}

	z, err := ZScore(25.0, 20.0, 2.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(z-2.0) > 1e-9 {
		t.Errorf("expected z 2.0, got %v", z)
	}
}
