package mathengine

import (
	"math"
	"testing"
)

func TestNormalDistributionStandard(t *testing.T) {
	norm := StandardNormal()
	if norm.Mu != 0 || norm.Sigma != 1 {
		t.Fatalf("expected standard normal N(0, 1), got mu=%v, sigma=%v", norm.Mu, norm.Sigma)
	}

	// Symmetry: f(-x) == f(x)
	for _, x := range []float64{0.5, 1.0, 1.96, 2.58, 3.0} {
		if math.Abs(norm.PDF(-x)-norm.PDF(x)) > 1e-15 {
			t.Errorf("PDF symmetry failed at x=%v", x)
		}
	}

	// Peak density at x=0 is 1/sqrt(2*pi) approx 0.3989422804014327
	expectedPeak := 1.0 / math.Sqrt(2*math.Pi)
	if math.Abs(norm.PDF(0)-expectedPeak) > 1e-14 {
		t.Errorf("expected peak density %v, got %v", expectedPeak, norm.PDF(0))
	}

	// Standard Normal CDF reference values
	// Phi(0) = 0.5
	if math.Abs(norm.CDF(0)-0.5) > 1e-15 {
		t.Errorf("expected Phi(0) = 0.5, got %v", norm.CDF(0))
	}

	// Phi(1) approx 0.8413447460685429
	if math.Abs(norm.CDF(1.0)-0.841344746) > 1e-7 {
		t.Errorf("expected Phi(1) approx 0.8413447, got %v", norm.CDF(1.0))
	}

	// Empirical Rule: 68-95-99.7
	// P(-1 <= Z <= 1) approx 0.682689492
	p1, _ := norm.Interval(-1.0, 1.0)
	if math.Abs(p1-0.682689492) > 1e-6 {
		t.Errorf("expected 1-sigma interval approx 0.6827, got %v", p1)
	}

	// P(-2 <= Z <= 2) approx 0.954499736
	p2, _ := norm.Interval(-2.0, 2.0)
	if math.Abs(p2-0.954499736) > 1e-6 {
		t.Errorf("expected 2-sigma interval approx 0.9545, got %v", p2)
	}

	// P(-3 <= Z <= 3) approx 0.997300204
	p3, _ := norm.Interval(-3.0, 3.0)
	if math.Abs(p3-0.997300204) > 1e-6 {
		t.Errorf("expected 3-sigma interval approx 0.9973, got %v", p3)
	}

	// Survival tail: P(Z >= 1.96) approx 0.024997895 (alpha/2 = 0.025)
	if math.Abs(norm.Survival(1.95996398454)-0.025) > 1e-6 {
		t.Errorf("expected upper tail at 1.96 to be 0.025, got %v", norm.Survival(1.95996398454))
	}

	// Extreme tail stability: Survival should not underflow to zero prematurely
	s := norm.Survival(8.0)
	if s <= 0 || math.IsNaN(s) {
		t.Errorf("extreme survival tail underflowed or returned NaN: %v", s)
	}

	// Quantiles
	q50, err := norm.Quantile(0.5)
	if err != nil || math.Abs(q50) > 1e-12 {
		t.Errorf("expected quantile(0.5) = 0, got %v, err=%v", q50, err)
	}

	q975, err := norm.Quantile(0.975)
	if err != nil || math.Abs(q975-1.95996398454) > 1e-6 {
		t.Errorf("expected quantile(0.975) approx 1.96, got %v, err=%v", q975, err)
	}

	q95, err := norm.Quantile(0.95)
	if err != nil || math.Abs(q95-1.64485362695) > 1e-6 {
		t.Errorf("expected quantile(0.95) approx 1.645, got %v, err=%v", q95, err)
	}
}

func TestNormalDistributionNonStandard(t *testing.T) {
	// N(100, 15^2) (e.g. IQ distribution)
	norm, err := NewNormalDistribution(100, 15)
	if err != nil {
		t.Fatalf("failed to create N(100, 15): %v", err)
	}

	// P(X <= 100) = 0.5
	if math.Abs(norm.CDF(100)-0.5) > 1e-12 {
		t.Errorf("expected CDF(100) = 0.5, got %v", norm.CDF(100))
	}

	// P(X <= 130) = Phi(2) approx 0.977249868
	if math.Abs(norm.CDF(130)-0.977249868) > 1e-6 {
		t.Errorf("expected CDF(130) approx 0.97725, got %v", norm.CDF(130))
	}

	// ZScore
	if norm.ZScore(130) != 2.0 {
		t.Errorf("expected z-score 2.0 for x=130, got %v", norm.ZScore(130))
	}

	// Quantile at 0.5 should be 100
	q, err := norm.Quantile(0.5)
	if err != nil || math.Abs(q-100) > 1e-10 {
		t.Errorf("expected quantile 100, got %v, err=%v", q, err)
	}

	// Invalid parameters
	if _, err := NewNormalDistribution(0, -1); err == nil {
		t.Errorf("expected error for negative sigma")
	}
	if _, err := NewNormalDistribution(0, 0); err == nil {
		t.Errorf("expected error for zero sigma")
	}
}

func TestUniformDistribution(t *testing.T) {
	u, err := NewUniformDistribution(10, 20)
	if err != nil {
		t.Fatalf("failed to create U(10, 20): %v", err)
	}

	if u.Mean() != 15.0 {
		t.Errorf("expected mean 15.0, got %v", u.Mean())
	}
	expectedVar := 100.0 / 12.0
	if math.Abs(u.Variance()-expectedVar) > 1e-12 {
		t.Errorf("expected variance %v, got %v", expectedVar, u.Variance())
	}

	// PDF
	if u.PDF(5) != 0 || u.PDF(25) != 0 {
		t.Errorf("PDF outside support must be 0")
	}
	if u.PDF(15) != 0.1 {
		t.Errorf("expected PDF inside support to be 0.1, got %v", u.PDF(15))
	}

	// CDF
	if u.CDF(5) != 0 || u.CDF(25) != 1.0 {
		t.Errorf("CDF boundary checks failed")
	}
	if u.CDF(15) != 0.5 {
		t.Errorf("expected CDF(15) = 0.5, got %v", u.CDF(15))
	}

	// Interval P(12 <= X <= 17) = (17 - 12)/10 = 0.5
	p, err := u.Interval(12, 17)
	if err != nil || math.Abs(p-0.5) > 1e-12 {
		t.Errorf("expected interval probability 0.5, got %v, err=%v", p, err)
	}

	// Quantile
	q, err := u.Quantile(0.7)
	if err != nil || math.Abs(q-17.0) > 1e-12 {
		t.Errorf("expected quantile(0.7) = 17, got %v", q)
	}

	// Invalid
	if _, err := NewUniformDistribution(20, 10); err == nil {
		t.Errorf("expected error when a >= b")
	}
}

func TestExponentialDistribution(t *testing.T) {
	exp, err := NewExponentialDistribution(2.0)
	if err != nil {
		t.Fatalf("failed to create Exp(lambda=2): %v", err)
	}

	if exp.Mean() != 0.5 {
		t.Errorf("expected mean 0.5, got %v", exp.Mean())
	}
	if exp.Variance() != 0.25 {
		t.Errorf("expected variance 0.25, got %v", exp.Variance())
	}

	// PDF: f(0) = lambda = 2.0
	if exp.PDF(0) != 2.0 {
		t.Errorf("expected f(0) = 2.0, got %v", exp.PDF(0))
	}

	// CDF: P(X <= 0) = 0
	if exp.CDF(0) != 0 {
		t.Errorf("expected CDF(0) = 0, got %v", exp.CDF(0))
	}

	// Survival: P(X >= 1) = exp(-2) approx 0.135335283
	if math.Abs(exp.Survival(1.0)-math.Exp(-2)) > 1e-12 {
		t.Errorf("expected Survival(1) = exp(-2), got %v", exp.Survival(1.0))
	}

	// Memoryless property: P(X > s + t | X > s) == P(X > t)
	tVal := 1.5
	sVal := 3.0
	condProb := exp.Survival(sVal+tVal) / exp.Survival(sVal)
	directProb := exp.Survival(tVal)
	if math.Abs(condProb-directProb) > 1e-12 {
		t.Errorf("memoryless property failed: cond=%v, direct=%v", condProb, directProb)
	}

	// Quantile
	// Median of exponential is ln(2)/lambda = ln(2)/2 approx 0.34657359
	median, err := exp.Quantile(0.5)
	if err != nil || math.Abs(median-math.Ln2/2.0) > 1e-10 {
		t.Errorf("expected median %v, got %v", math.Ln2/2.0, median)
	}

	// Interval: P(0.5 <= X <= 1.5)
	intervalProb, err := exp.Interval(0.5, 1.5)
	if err != nil {
		t.Fatalf("unexpected interval error: %v", err)
	}
	expectedInterval := math.Exp(-2.0*0.5) - math.Exp(-2.0*1.5)
	if math.Abs(intervalProb-expectedInterval) > 1e-12 {
		t.Errorf("expected interval %v, got %v", expectedInterval, intervalProb)
	}

	// ExcelQuantile
	excelQ := exp.ExcelQuantile(0.90)
	if excelQ != "=-LN(1 - 0.9) / 2" {
		t.Errorf("unexpected ExcelQuantile: %s", excelQ)
	}

	// Invalid lambda
	if _, err := NewExponentialDistribution(0); err == nil {
		t.Errorf("expected error for lambda=0")
	}
	if _, err := NewExponentialDistribution(-0.5); err == nil {
		t.Errorf("expected error for negative lambda")
	}
}

func TestContinuousEvents(t *testing.T) {
	norm := StandardNormal()

	// P(X = 0) must be 0 with IsSinglePointZero = true
	resEq, err := EvaluateContinuousEvent(norm, ContinuousEventEqual, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resEq.Probability != 0.0 || !resEq.IsSinglePointZero {
		t.Errorf("expected P(X=0) = 0 with IsSinglePointZero=true, got %v, flag=%v", resEq.Probability, resEq.IsSinglePointZero)
	}
	if resEq.DensityAtCutoff <= 0 {
		t.Errorf("expected non-zero density at cutoff, got %v", resEq.DensityAtCutoff)
	}

	// P(X <= 0) = 0.5
	resLe, err := EvaluateContinuousEvent(norm, ContinuousEventAtMost, 0, 0)
	if err != nil || math.Abs(resLe.Probability-0.5) > 1e-12 {
		t.Errorf("expected P(X <= 0) = 0.5, got %v, err=%v", resLe.Probability, err)
	}

	// P(X >= 0) = 0.5
	resGe, err := EvaluateContinuousEvent(norm, ContinuousEventAtLeast, 0, 0)
	if err != nil || math.Abs(resGe.Probability-0.5) > 1e-12 {
		t.Errorf("expected P(X >= 0) = 0.5, got %v, err=%v", resGe.Probability, err)
	}

	// Between -1.96 and 1.96 approx 0.95
	resBet, err := EvaluateContinuousEvent(norm, ContinuousEventBetween, -1.95996398454, 1.95996398454)
	if err != nil || math.Abs(resBet.Probability-0.95) > 1e-6 {
		t.Errorf("expected P(-1.96 <= Z <= 1.96) approx 0.95, got %v, err=%v", resBet.Probability, err)
	}
}
