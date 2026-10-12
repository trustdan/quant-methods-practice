package mathengine

import (
	"fmt"
	"math"
	"sort"
)

// DescriptiveSummary holds common sample descriptive statistics for a dataset.
type DescriptiveSummary struct {
	Count              int       `json:"count"`
	Mean               float64   `json:"mean"`
	Median             float64   `json:"median"`
	Modes              []float64 `json:"modes,omitempty"`
	Min                float64   `json:"min"`
	Max                float64   `json:"max"`
	Range              float64   `json:"range"`
	SampleVariance     float64   `json:"sample_variance"`
	SampleStdDev       float64   `json:"sample_std_dev"`
	PopulationVariance float64   `json:"population_variance"`
	PopulationStdDev   float64   `json:"population_std_dev"`
	CVPercent          float64   `json:"cv_percent"` // (s / mean) * 100
}

// ComputeDescriptiveSummary calculates comprehensive summary statistics for a dataset.
func ComputeDescriptiveSummary(data []float64) (*DescriptiveSummary, error) {
	n := len(data)
	if n < 2 {
		return nil, fmt.Errorf("descriptive summary requires at least 2 observations, got %d", n)
	}

	mean, err := Mean(data)
	if err != nil {
		return nil, err
	}
	median, err := Median(data)
	if err != nil {
		return nil, err
	}
	sVar, err := SampleVariance(data)
	if err != nil {
		return nil, err
	}
	sSD, err := SampleStdDev(data)
	if err != nil {
		return nil, err
	}
	pVar, err := PopulationVariance(data)
	if err != nil {
		return nil, err
	}
	pSD, err := PopulationStdDev(data)
	if err != nil {
		return nil, err
	}

	sorted := make([]float64, n)
	copy(sorted, data)
	sort.Float64s(sorted)
	minVal := sorted[0]
	maxVal := sorted[n-1]
	rangeVal := maxVal - minVal

	cv := 0.0
	if math.Abs(mean) > 1e-9 {
		cv = (sSD / mean) * 100.0
	}

	modes := Mode(sorted)

	return &DescriptiveSummary{
		Count:              n,
		Mean:               mean,
		Median:             median,
		Modes:              modes,
		Min:                minVal,
		Max:                maxVal,
		Range:              rangeVal,
		SampleVariance:     sVar,
		SampleStdDev:       sSD,
		PopulationVariance: pVar,
		PopulationStdDev:   pSD,
		CVPercent:          cv,
	}, nil
}

// Mean computes the arithmetic mean of a slice of floats.
func Mean(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("mean of empty slice is undefined")
	}
	sum := 0.0
	for _, x := range data {
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return 0, fmt.Errorf("invalid float value %v in dataset", x)
		}
		sum += x
	}
	return sum / float64(len(data)), nil
}

// Median computes the sample median.
func Median(data []float64) (float64, error) {
	n := len(data)
	if n == 0 {
		return 0, fmt.Errorf("median of empty slice is undefined")
	}
	sorted := make([]float64, n)
	copy(sorted, data)
	sort.Float64s(sorted)

	if n%2 == 1 {
		return sorted[n/2], nil
	}
	mid := n / 2
	return (sorted[mid-1] + sorted[mid]) / 2.0, nil
}

// Mode returns the modal values in sorted data.
func Mode(data []float64) []float64 {
	if len(data) == 0 {
		return nil
	}
	counts := make(map[float64]int)
	maxCount := 0
	for _, x := range data {
		counts[x]++
		if counts[x] > maxCount {
			maxCount = counts[x]
		}
	}
	if maxCount <= 1 {
		return nil // no mode
	}
	var modes []float64
	for val, count := range counts {
		if count == maxCount {
			modes = append(modes, val)
		}
	}
	sort.Float64s(modes)
	return modes
}

// SampleVariance computes the sample variance s^2 with Bessel's correction (n - 1 divisor).
func SampleVariance(data []float64) (float64, error) {
	n := len(data)
	if n < 2 {
		return 0, fmt.Errorf("sample variance requires at least 2 observations, got %d", n)
	}
	mean, err := Mean(data)
	if err != nil {
		return 0, err
	}
	sumSq := 0.0
	for _, x := range data {
		diff := x - mean
		sumSq += diff * diff
	}
	return sumSq / float64(n-1), nil
}

// SampleStdDev computes the sample standard deviation s = sqrt(s^2).
func SampleStdDev(data []float64) (float64, error) {
	v, err := SampleVariance(data)
	if err != nil {
		return 0, err
	}
	return math.Sqrt(v), nil
}

// PopulationVariance computes the population variance sigma^2 with n divisor.
func PopulationVariance(data []float64) (float64, error) {
	n := len(data)
	if n < 1 {
		return 0, fmt.Errorf("population variance requires at least 1 observation, got %d", n)
	}
	mean, err := Mean(data)
	if err != nil {
		return 0, err
	}
	sumSq := 0.0
	for _, x := range data {
		diff := x - mean
		sumSq += diff * diff
	}
	return sumSq / float64(n), nil
}

// PopulationStdDev computes the population standard deviation sigma = sqrt(sigma^2).
func PopulationStdDev(data []float64) (float64, error) {
	v, err := PopulationVariance(data)
	if err != nil {
		return 0, err
	}
	return math.Sqrt(v), nil
}

// CoefficientOfVariation computes CV = (sd / mean) * 100.
func CoefficientOfVariation(mean, sd float64) (float64, error) {
	if math.Abs(mean) < 1e-9 {
		return 0, fmt.Errorf("coefficient of variation undefined for zero mean")
	}
	if sd < 0 {
		return 0, fmt.Errorf("standard deviation must be non-negative, got %v", sd)
	}
	return (sd / mean) * 100.0, nil
}

// ZScore computes standardized z = (x - mean) / sd.
func ZScore(x, mean, sd float64) (float64, error) {
	if sd <= 0 {
		return 0, fmt.Errorf("standard deviation must be strictly positive for z-score, got %v", sd)
	}
	return (x - mean) / sd, nil
}

// WeightedMean computes sum(w_i * x_i) / sum(w_i).
func WeightedMean(values, weights []float64) (float64, error) {
	if len(values) == 0 || len(values) != len(weights) {
		return 0, fmt.Errorf("values and weights must be non-empty and of equal length")
	}
	sumWeight := 0.0
	sumProduct := 0.0
	for i := range values {
		if weights[i] < 0 {
			return 0, fmt.Errorf("weight at index %d is negative: %v", i, weights[i])
		}
		sumWeight += weights[i]
		sumProduct += values[i] * weights[i]
	}
	if sumWeight <= 0 {
		return 0, fmt.Errorf("total weight must be strictly positive, got %v", sumWeight)
	}
	return sumProduct / sumWeight, nil
}
