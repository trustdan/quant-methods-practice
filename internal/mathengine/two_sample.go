package mathengine

import (
	"fmt"
	"math"
)

// PairedTTestResult holds the results of a paired samples t-test.
type PairedTTestResult struct {
	N            int     `json:"n"`
	MeanDiff     float64 `json:"mean_diff"`
	SDDiff       float64 `json:"sd_diff"`
	SEDiff       float64 `json:"se_diff"`
	HypMeanDiff  float64 `json:"hyp_mean_diff"`
	DF           int     `json:"df"`
	TStat        float64 `json:"t_stat"`
	PValueTwo    float64 `json:"p_value_two_tail"`
	PValueOne    float64 `json:"p_value_one_tail"`
	ExcelFormula string  `json:"excel_formula"`
}

// ComputePairedTTest performs a paired samples t-test on differences d = x1 - x2.
func ComputePairedTTest(meanDiff, sdDiff float64, n int, hypDiff float64) (*PairedTTestResult, error) {
	if n < 2 {
		return nil, fmt.Errorf("paired t-test requires n >= 2, got %d", n)
	}
	if sdDiff <= 0 {
		return nil, fmt.Errorf("sd of differences must be strictly positive, got %v", sdDiff)
	}
	df := n - 1
	se := sdDiff / math.Sqrt(float64(n))
	tStat := (meanDiff - hypDiff) / se

	tDist, err := NewTDistribution(float64(df))
	if err != nil {
		return nil, err
	}
	cdf := tDist.CDF(math.Abs(tStat))
	pOne := 1.0 - cdf
	pTwo := 2.0 * pOne

	return &PairedTTestResult{
		N:            n,
		MeanDiff:     meanDiff,
		SDDiff:       sdDiff,
		SEDiff:       se,
		HypMeanDiff:  hypDiff,
		DF:           df,
		TStat:        tStat,
		PValueTwo:    pTwo,
		PValueOne:    pOne,
		ExcelFormula: fmt.Sprintf("=T.TEST(range1, range2, %d, 1)", 2),
	}, nil

}

// TwoSampleTTestResult holds the results of an independent two-sample t-test (pooled variances).
type TwoSampleTTestResult struct {
	N1             int     `json:"n1"`
	Mean1          float64 `json:"mean1"`
	Var1           float64 `json:"var1"`
	N2             int     `json:"n2"`
	Mean2          float64 `json:"mean2"`
	Var2           float64 `json:"var2"`
	PooledVariance float64 `json:"pooled_variance"`
	SE             float64 `json:"se"`
	DF             int     `json:"df"`
	TStat          float64 `json:"t_stat"`
	PValueOneTail  float64 `json:"p_value_one_tail"`
	PValueTwoTail  float64 `json:"p_value_two_tail"`
}

// ComputeTwoSampleTTestPooled performs an independent two-sample t-test assuming equal population variances.
func ComputeTwoSampleTTestPooled(mean1, var1 float64, n1 int, mean2, var2 float64, n2 int, hypDiff float64) (*TwoSampleTTestResult, error) {
	if n1 < 2 || n2 < 2 {
		return nil, fmt.Errorf("both sample sizes must be >= 2, got n1=%d, n2=%d", n1, n2)
	}
	if var1 <= 0 || var2 <= 0 {
		return nil, fmt.Errorf("variances must be strictly positive, got var1=%v, var2=%v", var1, var2)
	}

	df := n1 + n2 - 2
	pooledVar := (float64(n1-1)*var1 + float64(n2-1)*var2) / float64(df)
	se := math.Sqrt(pooledVar * (1.0/float64(n1) + 1.0/float64(n2)))
	tStat := ((mean1 - mean2) - hypDiff) / se

	tDist, err := NewTDistribution(float64(df))
	if err != nil {
		return nil, err
	}
	cdf := tDist.CDF(math.Abs(tStat))
	pOne := 1.0 - cdf
	pTwo := 2.0 * pOne

	return &TwoSampleTTestResult{
		N1:             n1,
		Mean1:          mean1,
		Var1:           var1,
		N2:             n2,
		Mean2:          mean2,
		Var2:           var2,
		PooledVariance: pooledVar,
		SE:             se,
		DF:             df,
		TStat:          tStat,
		PValueOneTail:  pOne,
		PValueTwoTail:  pTwo,
	}, nil
}

// TwoSampleProportionResult holds the results of a two-sample z-test for difference in proportions.
type TwoSampleProportionResult struct {
	N1            int     `json:"n1"`
	X1            int     `json:"x1"`
	P1Hat         float64 `json:"p1_hat"`
	N2            int     `json:"n2"`
	X2            int     `json:"x2"`
	P2Hat         float64 `json:"p2_hat"`
	PooledP       float64 `json:"pooled_p"`
	SE            float64 `json:"se"`
	ZStat         float64 `json:"z_stat"`
	PValueOneTail float64 `json:"p_value_one_tail"`
	PValueTwoTail float64 `json:"p_value_two_tail"`
}

// ComputeTwoSampleProportionTest performs a two-sample pooled z-test for proportions.
func ComputeTwoSampleProportionTest(x1, n1, x2, n2 int) (*TwoSampleProportionResult, error) {
	if n1 < 10 || n2 < 10 {
		return nil, fmt.Errorf("sample sizes must be >= 10, got n1=%d, n2=%d", n1, n2)
	}
	if x1 < 0 || x1 > n1 || x2 < 0 || x2 > n2 {
		return nil, fmt.Errorf("success counts must be in [0, n], got x1=%d/n1=%d, x2=%d/n2=%d", x1, n1, x2, n2)
	}

	p1 := float64(x1) / float64(n1)
	p2 := float64(x2) / float64(n2)
	pooledP := float64(x1+x2) / float64(n1+n2)

	se := math.Sqrt(pooledP * (1.0 - pooledP) * (1.0/float64(n1) + 1.0/float64(n2)))
	if se <= 0 {
		return nil, fmt.Errorf("standard error cannot be zero")
	}
	zStat := (p1 - p2) / se

	norm := StandardNormal()
	cdf := norm.CDF(math.Abs(zStat))
	pOne := 1.0 - cdf
	pTwo := 2.0 * pOne

	return &TwoSampleProportionResult{
		N1:            n1,
		X1:            x1,
		P1Hat:         p1,
		N2:            n2,
		X2:            x2,
		P2Hat:         p2,
		PooledP:       pooledP,
		SE:            se,
		ZStat:         zStat,
		PValueOneTail: pOne,
		PValueTwoTail: pTwo,
	}, nil
}
