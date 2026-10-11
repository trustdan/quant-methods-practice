package mathengine

import (
	"fmt"
	"math"
	"math/rand"
)

// SamplingDistributionMean evaluates the sampling distribution of the sample mean X_bar.
type SamplingDistributionMean struct {
	Mu          float64 `json:"mu"`
	Sigma       float64 `json:"sigma"`
	SampleSize  int     `json:"sample_size"`
	PopSize     int     `json:"pop_size,omitempty"`
	StandardErr float64 `json:"standard_error"`
	FPCApplied  bool    `json:"fpc_applied"`
	FPCFactor   float64 `json:"fpc_factor"`
	IsNormal    bool    `json:"is_normal"` // True if population is normal or n >= 30 (CLT)
}

// ComputeSamplingDistributionMean computes the sampling distribution moments and standard error.
func ComputeSamplingDistributionMean(mu, sigma float64, n int, popSize int, popIsNormal bool) (*SamplingDistributionMean, error) {
	if math.IsNaN(mu) || math.IsInf(mu, 0) {
		return nil, fmt.Errorf("mean mu must be finite, got %v", mu)
	}
	if math.IsNaN(sigma) || math.IsInf(sigma, 0) || sigma <= 0 {
		return nil, fmt.Errorf("standard deviation sigma must be strictly positive, got %v", sigma)
	}
	if n <= 0 {
		return nil, fmt.Errorf("sample size n must be strictly positive, got %d", n)
	}
	if popSize > 0 && n > popSize {
		return nil, fmt.Errorf("sample size n (%d) cannot exceed population size N (%d)", n, popSize)
	}

	se := sigma / math.Sqrt(float64(n))
	fpcApplied := false
	fpc := 1.0

	// Finite population correction if sampling without replacement and n / N > 0.05
	if popSize > 0 {
		frac := float64(n) / float64(popSize)
		if frac > 0.05 && popSize > 1 {
			fpcApplied = true
			fpc = math.Sqrt(float64(popSize-n) / float64(popSize-1))
			se *= fpc
		}
	}

	isNormal := popIsNormal || n >= 30

	return &SamplingDistributionMean{
		Mu:          mu,
		Sigma:       sigma,
		SampleSize:  n,
		PopSize:     popSize,
		StandardErr: se,
		FPCApplied:  fpcApplied,
		FPCFactor:   fpc,
		IsNormal:    isNormal,
	}, nil
}

// ConfidenceIntervalResult encapsulates the computed interval, critical value, margin of error, and derivation.
type ConfidenceIntervalResult struct {
	Parameter       string   `json:"parameter"` // "mean_z", "mean_t", "proportion"
	ConfidenceLevel float64  `json:"confidence_level"`
	Alpha           float64  `json:"alpha"`
	PointEstimate   float64  `json:"point_estimate"`
	StandardError   float64  `json:"standard_error"`
	CriticalValue   float64  `json:"critical_value"`
	MarginOfError   float64  `json:"margin_of_error"`
	LowerBound      float64  `json:"lower_bound"`
	UpperBound      float64  `json:"upper_bound"`
	DegreesOfFree   *float64 `json:"degrees_of_freedom,omitempty"`
	ExcelFormula    string   `json:"excel_formula"`
	Interpretation  string   `json:"interpretation"`
}

// ConfidenceIntervalMeanZ computes a confidence interval for the population mean when sigma is known (Z-interval).
func ConfidenceIntervalMeanZ(xBar, sigma float64, n int, confLevel float64) (*ConfidenceIntervalResult, error) {
	if confLevel <= 0 || confLevel >= 1 {
		return nil, fmt.Errorf("confidence level must be in open interval (0, 1), got %v", confLevel)
	}
	if n <= 0 {
		return nil, fmt.Errorf("sample size n must be positive, got %d", n)
	}
	if sigma <= 0 {
		return nil, fmt.Errorf("sigma must be positive, got %v", sigma)
	}

	alpha := 1.0 - confLevel
	norm := StandardNormal()
	zCrit, err := norm.Quantile(1.0 - 0.5*alpha)
	if err != nil {
		return nil, err
	}

	se := sigma / math.Sqrt(float64(n))
	moe := zCrit * se
	lower := xBar - moe
	upper := xBar + moe

	excelFormula := fmt.Sprintf("=CONFIDENCE.NORM(%.4g, %.4g, %d)", alpha, sigma, n)
	interp := fmt.Sprintf("We are %.1f%% confident that the true population mean mu falls between %.4g and %.4g.", confLevel*100, lower, upper)

	return &ConfidenceIntervalResult{
		Parameter:       "mean_z",
		ConfidenceLevel: confLevel,
		Alpha:           alpha,
		PointEstimate:   xBar,
		StandardError:   se,
		CriticalValue:   zCrit,
		MarginOfError:   moe,
		LowerBound:      lower,
		UpperBound:      upper,
		ExcelFormula:    excelFormula,
		Interpretation:  interp,
	}, nil
}

// ConfidenceIntervalMeanT computes a confidence interval for the population mean when sigma is unknown (t-interval).
func ConfidenceIntervalMeanT(xBar, s float64, n int, confLevel float64) (*ConfidenceIntervalResult, error) {
	if confLevel <= 0 || confLevel >= 1 {
		return nil, fmt.Errorf("confidence level must be in open interval (0, 1), got %v", confLevel)
	}
	if n <= 1 {
		return nil, fmt.Errorf("sample size n must be at least 2 for t-interval, got %d", n)
	}
	if s <= 0 {
		return nil, fmt.Errorf("sample standard deviation s must be positive, got %v", s)
	}

	alpha := 1.0 - confLevel
	df := float64(n - 1)
	tDist, err := NewTDistribution(df)
	if err != nil {
		return nil, err
	}

	tCrit, err := tDist.Quantile(1.0 - 0.5*alpha)
	if err != nil {
		return nil, err
	}

	se := s / math.Sqrt(float64(n))
	moe := tCrit * se
	lower := xBar - moe
	upper := xBar + moe

	excelFormula := fmt.Sprintf("=CONFIDENCE.T(%.4g, %.4g, %d)", alpha, s, n)
	interp := fmt.Sprintf("With df=%d, we are %.1f%% confident that the true population mean mu falls between %.4g and %.4g.", n-1, confLevel*100, lower, upper)

	return &ConfidenceIntervalResult{
		Parameter:       "mean_t",
		ConfidenceLevel: confLevel,
		Alpha:           alpha,
		PointEstimate:   xBar,
		StandardError:   se,
		CriticalValue:   tCrit,
		MarginOfError:   moe,
		LowerBound:      lower,
		UpperBound:      upper,
		DegreesOfFree:   &df,
		ExcelFormula:    excelFormula,
		Interpretation:  interp,
	}, nil
}

// ConfidenceIntervalProportion computes a confidence interval for a population proportion p.
func ConfidenceIntervalProportion(successes, n int, confLevel float64) (*ConfidenceIntervalResult, error) {
	if confLevel <= 0 || confLevel >= 1 {
		return nil, fmt.Errorf("confidence level must be in open interval (0, 1), got %v", confLevel)
	}
	if n <= 0 {
		return nil, fmt.Errorf("sample size n must be positive, got %d", n)
	}
	if successes < 0 || successes > n {
		return nil, fmt.Errorf("successes x (%d) must be in [0, %d]", successes, n)
	}

	pHat := float64(successes) / float64(n)
	alpha := 1.0 - confLevel
	norm := StandardNormal()
	zCrit, err := norm.Quantile(1.0 - 0.5*alpha)
	if err != nil {
		return nil, err
	}

	se := math.Sqrt(pHat * (1.0 - pHat) / float64(n))
	moe := zCrit * se
	lower := math.Max(0.0, pHat-moe)
	upper := math.Min(1.0, pHat+moe)

	excelFormula := fmt.Sprintf("=%.4g +/- NORM.S.INV(%.4g) * SQRT(%.4g * (1 - %.4g) / %d)", pHat, 1.0-0.5*alpha, pHat, pHat, n)
	interp := fmt.Sprintf("We are %.1f%% confident that the true population proportion p falls between %.4g and %.4g.", confLevel*100, lower, upper)

	return &ConfidenceIntervalResult{
		Parameter:       "proportion",
		ConfidenceLevel: confLevel,
		Alpha:           alpha,
		PointEstimate:   pHat,
		StandardError:   se,
		CriticalValue:   zCrit,
		MarginOfError:   moe,
		LowerBound:      lower,
		UpperBound:      upper,
		ExcelFormula:    excelFormula,
		Interpretation:  interp,
	}, nil
}

// SimulatedInterval represents one simulated confidence interval for coverage analysis.
type SimulatedInterval struct {
	Index         int     `json:"index"`
	SampleMean    float64 `json:"sample_mean"`
	LowerBound    float64 `json:"lower_bound"`
	UpperBound    float64 `json:"upper_bound"`
	CapturesTruth bool    `json:"captures_truth"`
}

// SimulateConfidenceIntervalCoverage runs a Monte Carlo simulation of confidence interval coverage.
func SimulateConfidenceIntervalCoverage(mu, sigma float64, n int, confLevel float64, numTrials int, seed int64) ([]SimulatedInterval, float64, error) {
	if numTrials <= 0 || numTrials > 500 {
		return nil, 0, fmt.Errorf("number of trials must be in [1, 500], got %d", numTrials)
	}
	if n <= 1 {
		return nil, 0, fmt.Errorf("sample size n must be >= 2, got %d", n)
	}

	rng := rand.New(rand.NewSource(seed))
	intervals := make([]SimulatedInterval, numTrials)
	capturedCount := 0

	for i := 0; i < numTrials; i++ {
		// Draw sample of size n from N(mu, sigma^2)
		sum := 0.0
		sample := make([]float64, n)
		for j := 0; j < n; j++ {
			val := mu + sigma*rng.NormFloat64()
			sample[j] = val
			sum += val
		}
		xBar := sum / float64(n)

		// Sample standard deviation s
		sumSqDiff := 0.0
		for _, val := range sample {
			diff := val - xBar
			sumSqDiff += diff * diff
		}
		s := math.Sqrt(sumSqDiff / float64(n-1))
		if s == 0 {
			s = sigma
		}

		ci, err := ConfidenceIntervalMeanT(xBar, s, n, confLevel)
		if err != nil {
			return nil, 0, err
		}

		covers := ci.LowerBound <= mu && mu <= ci.UpperBound
		if covers {
			capturedCount++
		}

		intervals[i] = SimulatedInterval{
			Index:         i + 1,
			SampleMean:    math.Round(xBar*1000) / 1000,
			LowerBound:    math.Round(ci.LowerBound*1000) / 1000,
			UpperBound:    math.Round(ci.UpperBound*1000) / 1000,
			CapturesTruth: covers,
		}
	}

	empiricalCoverage := float64(capturedCount) / float64(numTrials)
	return intervals, empiricalCoverage, nil
}
