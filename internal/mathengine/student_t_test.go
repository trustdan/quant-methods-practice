package mathengine

import (
	"math"
	"testing"
)

func TestStudentTDistributionBasic(t *testing.T) {
	// nu must be strictly positive
	if _, err := NewTDistribution(0); err == nil {
		t.Errorf("expected error for nu = 0, got nil")
	}
	if _, err := NewTDistribution(-5); err == nil {
		t.Errorf("expected error for nu = -5, got nil")
	}

	dist, err := NewTDistribution(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Symmetry: CDF(0) = 0.5, PDF(x) = PDF(-x)
	if math.Abs(dist.CDF(0)-0.5) > 1e-12 {
		t.Errorf("expected CDF(0) = 0.5, got %v", dist.CDF(0))
	}
	if math.Abs(dist.Survival(0)-0.5) > 1e-12 {
		t.Errorf("expected Survival(0) = 0.5, got %v", dist.Survival(0))
	}
	if math.Abs(dist.PDF(1.5)-dist.PDF(-1.5)) > 1e-12 {
		t.Errorf("PDF should be symmetric: f(1.5)=%v, f(-1.5)=%v", dist.PDF(1.5), dist.PDF(-1.5))
	}

	// Moments
	if dist.Mean() != 0.0 {
		t.Errorf("expected mean 0, got %v", dist.Mean())
	}
	// Var for nu=10 is 10 / (10 - 2) = 1.25
	expectedVar := 10.0 / 8.0
	if math.Abs(dist.Variance()-expectedVar) > 1e-12 {
		t.Errorf("expected variance %v, got %v", expectedVar, dist.Variance())
	}
}

func TestStudentTDistributionReferenceQuantiles(t *testing.T) {
	// Standard statistical textbook reference critical values:
	// nu = 9: t_{0.975} = 2.262157 (two-tailed 0.05 critical value)
	// nu = 19: t_{0.975} = 2.093024
	// nu = 29: t_{0.975} = 2.045230
	// nu = 20: t_{0.95}  = 1.724718 (one-tailed 0.05 upper tail)

	cases := []struct {
		nu       float64
		p        float64
		expected float64
		tol      float64
	}{
		{nu: 9, p: 0.975, expected: 2.262157, tol: 1e-4},
		{nu: 9, p: 0.025, expected: -2.262157, tol: 1e-4},
		{nu: 19, p: 0.975, expected: 2.093024, tol: 1e-4},
		{nu: 29, p: 0.975, expected: 2.045230, tol: 1e-4},
		{nu: 20, p: 0.95, expected: 1.724718, tol: 1e-4},
		{nu: 10, p: 0.90, expected: 1.372184, tol: 1e-4},
	}

	for _, tc := range cases {
		dist, err := NewTDistribution(tc.nu)
		if err != nil {
			t.Fatalf("nu=%.1f error: %v", tc.nu, err)
		}

		q, err := dist.Quantile(tc.p)
		if err != nil {
			t.Fatalf("Quantile(%.4f) error: %v", tc.p, err)
		}

		if math.Abs(q-tc.expected) > tc.tol {
			t.Errorf("nu=%.1f, p=%.4f: expected quantile %v, got %v (diff %v)", tc.nu, tc.p, tc.expected, q, math.Abs(q-tc.expected))
		}

		// Reversible: CDF(q) should equal p
		cdf := dist.CDF(q)
		if math.Abs(cdf-tc.p) > 1e-6 {
			t.Errorf("nu=%.1f, p=%.4f: CDF(quantile)=%v, diff=%v", tc.nu, tc.p, cdf, math.Abs(cdf-tc.p))
		}
	}
}

func TestStudentTConvergenceToNormal(t *testing.T) {
	// As nu -> infinity, t distribution converges to standard normal N(0, 1)
	dist, err := NewTDistribution(5000)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	norm := StandardNormal()

	// Compare PDF and CDF at several points
	testPoints := []float64{-2.0, -1.0, 0.0, 1.0, 1.96, 2.58}
	for _, x := range testPoints {
		if math.Abs(dist.PDF(x)-norm.PDF(x)) > 0.005 {
			t.Errorf("x=%v: t PDF %v vs norm PDF %v", x, dist.PDF(x), norm.PDF(x))
		}
		if math.Abs(dist.CDF(x)-norm.CDF(x)) > 0.005 {
			t.Errorf("x=%v: t CDF %v vs norm CDF %v", x, dist.CDF(x), norm.CDF(x))
		}
	}
}
