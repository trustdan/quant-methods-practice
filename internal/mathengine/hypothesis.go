package mathengine

import (
	"fmt"
	"math"
)

// TailDirection represents the alternative hypothesis direction.
type TailDirection string

const (
	TailTwoSided TailDirection = "two_tailed"
	TailLeft     TailDirection = "left_tailed"
	TailRight    TailDirection = "right_tailed"
)

// HypothesisTestResult holds the mathematical results, decision, and educational derivation.
type HypothesisTestResult struct {
	TestType        string        `json:"test_type"` // "mean_z", "mean_t", "proportion_z"
	Tail            TailDirection `json:"tail"`
	NullHypothesis  string        `json:"null_hypothesis"`
	AltHypothesis   string        `json:"alt_hypothesis"`
	NullValue       float64       `json:"null_value"`
	SampleStat      float64       `json:"sample_stat"`
	SampleSize      int           `json:"sample_size"`
	StandardError   float64       `json:"standard_error"`
	TestStatistic   float64       `json:"test_statistic"`
	DegreesOfFree   *float64      `json:"degrees_of_freedom,omitempty"`
	PValue          float64       `json:"p_value"`
	Alpha           float64       `json:"alpha"`
	CriticalValue   float64       `json:"critical_value"`
	CriticalValue2  *float64      `json:"critical_value_2,omitempty"` // For two-tailed test: -crit and +crit
	RejectNull      bool          `json:"reject_null"`
	Decision        string        `json:"decision"`
	RejectionRegion string        `json:"rejection_region"`
	ExcelFormula    string        `json:"excel_formula"`
	Explanation     string        `json:"explanation"`
}

// PerformMeanZTest performs a one-sample Z-test for the population mean (known sigma).
func PerformMeanZTest(nullMu, xBar, sigma float64, n int, tail TailDirection, alpha float64) (*HypothesisTestResult, error) {
	if n <= 0 {
		return nil, fmt.Errorf("sample size n must be positive, got %d", n)
	}
	if sigma <= 0 {
		return nil, fmt.Errorf("sigma must be positive, got %v", sigma)
	}
	if alpha <= 0 || alpha >= 1 {
		return nil, fmt.Errorf("alpha must be in (0, 1), got %v", alpha)
	}

	se := sigma / math.Sqrt(float64(n))
	zStat := (xBar - nullMu) / se
	norm := StandardNormal()

	var pVal float64
	var critVal float64
	var critVal2 *float64
	var rejRegion string
	var hA string

	switch tail {
	case TailLeft:
		hA = fmt.Sprintf("mu < %.4g", nullMu)
		pVal = norm.CDF(zStat)
		crit, _ := norm.Quantile(alpha)
		critVal = crit
		rejRegion = fmt.Sprintf("Z <= %.3f", crit)
	case TailRight:
		hA = fmt.Sprintf("mu > %.4g", nullMu)
		pVal = norm.Survival(zStat)
		crit, _ := norm.Quantile(1.0 - alpha)
		critVal = crit
		rejRegion = fmt.Sprintf("Z >= %.3f", crit)
	default: // Two-tailed
		tail = TailTwoSided
		hA = fmt.Sprintf("mu != %.4g", nullMu)
		pVal = 2.0 * norm.Survival(math.Abs(zStat))
		crit, _ := norm.Quantile(1.0 - 0.5*alpha)
		critVal = crit
		negCrit := -crit
		critVal2 = &negCrit
		rejRegion = fmt.Sprintf("|Z| >= %.3f", crit)
	}

	reject := pVal < alpha
	decision := "Fail to reject H0"
	if reject {
		decision = "Reject H0"
	}

	expl := fmt.Sprintf(
		"Test statistic Z = (%.4g - %.4g) / (%.4g / sqrt(%d)) = %.4g. P-value = %.4g. At alpha=%.4g, we %s.",
		xBar, nullMu, sigma, n, zStat, pVal, alpha, decision,
	)
	if !reject {
		expl += " Notice: failing to reject H0 does not prove H0 is true; there is simply insufficient evidence to reject it."
	}

	excelFormula := fmt.Sprintf("=Z.TEST(data_range, %.4g, %.4g)", nullMu, sigma)

	return &HypothesisTestResult{
		TestType:        "mean_z",
		Tail:            tail,
		NullHypothesis:  fmt.Sprintf("mu = %.4g", nullMu),
		AltHypothesis:   hA,
		NullValue:       nullMu,
		SampleStat:      xBar,
		SampleSize:      n,
		StandardError:   se,
		TestStatistic:   zStat,
		PValue:          pVal,
		Alpha:           alpha,
		CriticalValue:   critVal,
		CriticalValue2:  critVal2,
		RejectNull:      reject,
		Decision:        decision,
		RejectionRegion: rejRegion,
		ExcelFormula:    excelFormula,
		Explanation:     expl,
	}, nil
}

// PerformMeanTTest performs a one-sample t-test for the population mean (unknown sigma, using sample s).
func PerformMeanTTest(nullMu, xBar, s float64, n int, tail TailDirection, alpha float64) (*HypothesisTestResult, error) {
	if n <= 1 {
		return nil, fmt.Errorf("sample size n must be >= 2 for t-test, got %d", n)
	}
	if s <= 0 {
		return nil, fmt.Errorf("sample standard deviation s must be positive, got %v", s)
	}
	if alpha <= 0 || alpha >= 1 {
		return nil, fmt.Errorf("alpha must be in (0, 1), got %v", alpha)
	}

	df := float64(n - 1)
	tDist, err := NewTDistribution(df)
	if err != nil {
		return nil, err
	}

	se := s / math.Sqrt(float64(n))
	tStat := (xBar - nullMu) / se

	var pVal float64
	var critVal float64
	var critVal2 *float64
	var rejRegion string
	var hA string

	switch tail {
	case TailLeft:
		hA = fmt.Sprintf("mu < %.4g", nullMu)
		pVal = tDist.CDF(tStat)
		crit, _ := tDist.Quantile(alpha)
		critVal = crit
		rejRegion = fmt.Sprintf("t <= %.3f", crit)
	case TailRight:
		hA = fmt.Sprintf("mu > %.4g", nullMu)
		pVal = tDist.Survival(tStat)
		crit, _ := tDist.Quantile(1.0 - alpha)
		critVal = crit
		rejRegion = fmt.Sprintf("t >= %.3f", crit)
	default: // Two-tailed
		tail = TailTwoSided
		hA = fmt.Sprintf("mu != %.4g", nullMu)
		pVal = 2.0 * tDist.Survival(math.Abs(tStat))
		crit, _ := tDist.Quantile(1.0 - 0.5*alpha)
		critVal = crit
		negCrit := -crit
		critVal2 = &negCrit
		rejRegion = fmt.Sprintf("|t| >= %.3f", crit)
	}

	reject := pVal < alpha
	decision := "Fail to reject H0"
	if reject {
		decision = "Reject H0"
	}

	expl := fmt.Sprintf(
		"Test statistic t = (%.4g - %.4g) / (%.4g / sqrt(%d)) = %.4g with df = %d. P-value = %.4g. At alpha=%.4g, we %s.",
		xBar, nullMu, s, n, tStat, n-1, pVal, alpha, decision,
	)
	if !reject {
		expl += " Notice: failing to reject H0 does not prove H0 is true; there is simply insufficient evidence to reject it."
	}

	excelFormula := fmt.Sprintf("=T.DIST.2T(ABS(%.4g), %d)", tStat, n-1)
	if tail == TailRight {
		excelFormula = fmt.Sprintf("=T.DIST.RT(%.4g, %d)", tStat, n-1)
	} else if tail == TailLeft {
		excelFormula = fmt.Sprintf("=T.DIST(%.4g, %d, TRUE)", tStat, n-1)
	}

	return &HypothesisTestResult{
		TestType:        "mean_t",
		Tail:            tail,
		NullHypothesis:  fmt.Sprintf("mu = %.4g", nullMu),
		AltHypothesis:   hA,
		NullValue:       nullMu,
		SampleStat:      xBar,
		SampleSize:      n,
		StandardError:   se,
		TestStatistic:   tStat,
		DegreesOfFree:   &df,
		PValue:          pVal,
		Alpha:           alpha,
		CriticalValue:   critVal,
		CriticalValue2:  critVal2,
		RejectNull:      reject,
		Decision:        decision,
		RejectionRegion: rejRegion,
		ExcelFormula:    excelFormula,
		Explanation:     expl,
	}, nil
}

// PerformProportionTest performs a one-sample Z-test for population proportion p.
func PerformProportionTest(nullP float64, successes int, n int, tail TailDirection, alpha float64) (*HypothesisTestResult, error) {
	if nullP <= 0 || nullP >= 1 {
		return nil, fmt.Errorf("null proportion p0 must be in (0, 1), got %v", nullP)
	}
	if n <= 0 {
		return nil, fmt.Errorf("sample size n must be positive, got %d", n)
	}
	if successes < 0 || successes > n {
		return nil, fmt.Errorf("successes x (%d) must be in [0, %d]", successes, n)
	}
	if alpha <= 0 || alpha >= 1 {
		return nil, fmt.Errorf("alpha must be in (0, 1), got %v", alpha)
	}

	pHat := float64(successes) / float64(n)
	se := math.Sqrt(nullP * (1.0 - nullP) / float64(n))
	zStat := (pHat - nullP) / se
	norm := StandardNormal()

	var pVal float64
	var critVal float64
	var critVal2 *float64
	var rejRegion string
	var hA string

	switch tail {
	case TailLeft:
		hA = fmt.Sprintf("p < %.4g", nullP)
		pVal = norm.CDF(zStat)
		crit, _ := norm.Quantile(alpha)
		critVal = crit
		rejRegion = fmt.Sprintf("Z <= %.3f", crit)
	case TailRight:
		hA = fmt.Sprintf("p > %.4g", nullP)
		pVal = norm.Survival(zStat)
		crit, _ := norm.Quantile(1.0 - alpha)
		critVal = crit
		rejRegion = fmt.Sprintf("Z >= %.3f", crit)
	default:
		tail = TailTwoSided
		hA = fmt.Sprintf("p != %.4g", nullP)
		pVal = 2.0 * norm.Survival(math.Abs(zStat))
		crit, _ := norm.Quantile(1.0 - 0.5*alpha)
		critVal = crit
		negCrit := -crit
		critVal2 = &negCrit
		rejRegion = fmt.Sprintf("|Z| >= %.3f", crit)
	}

	reject := pVal < alpha
	decision := "Fail to reject H0"
	if reject {
		decision = "Reject H0"
	}

	expl := fmt.Sprintf(
		"Test statistic Z = (%.4g - %.4g) / sqrt(%.4g * %.4g / %d) = %.4g. P-value = %.4g. At alpha=%.4g, we %s.",
		pHat, nullP, nullP, 1.0-nullP, n, zStat, pVal, alpha, decision,
	)

	return &HypothesisTestResult{
		TestType:        "proportion_z",
		Tail:            tail,
		NullHypothesis:  fmt.Sprintf("p = %.4g", nullP),
		AltHypothesis:   hA,
		NullValue:       nullP,
		SampleStat:      pHat,
		SampleSize:      n,
		StandardError:   se,
		TestStatistic:   zStat,
		PValue:          pVal,
		Alpha:           alpha,
		CriticalValue:   critVal,
		CriticalValue2:  critVal2,
		RejectNull:      reject,
		Decision:        decision,
		RejectionRegion: rejRegion,
		ExcelFormula:    fmt.Sprintf("=NORM.S.DIST(%.4g, TRUE)", zStat),
		Explanation:     expl,
	}, nil
}
