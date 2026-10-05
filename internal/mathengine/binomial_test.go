package mathengine_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

func TestBinomialReferenceFixture(t *testing.T) {
	// Canonical draft reference fixture: n=4, p=0.5, k=2 -> 0.375 (3/8)
	dist, err := mathengine.NewBinomial(4, 0.5)
	if err != nil {
		t.Fatalf("unexpected error creating Binomial: %v", err)
	}

	got := dist.PMF(2)
	want := 0.375
	if math.Abs(got-want) > 1e-15 {
		t.Errorf("PMF(2) = %v, want exact %v", got, want)
	}

	// Exact rational verification: 3/8
	pRat := big.NewRat(1, 2)
	exactRat, err := dist.ExactPMFRat(2, pRat)
	if err != nil {
		t.Fatalf("unexpected error computing exact rational: %v", err)
	}
	expectedRat := big.NewRat(3, 8)
	if exactRat.Cmp(expectedRat) != 0 {
		t.Errorf("ExactPMFRat(2, 1/2) = %s, want %s", exactRat.RatString(), expectedRat.RatString())
	}
}

func TestBinomialMoments(t *testing.T) {
	dist, err := mathengine.NewBinomial(10, 0.4)
	if err != nil {
		t.Fatalf("NewBinomial failed: %v", err)
	}

	wantMean := 4.0
	wantVar := 2.4
	wantStdDev := math.Sqrt(2.4)

	if math.Abs(dist.Mean()-wantMean) > 1e-12 {
		t.Errorf("Mean() = %v, want %v", dist.Mean(), wantMean)
	}
	if math.Abs(dist.Variance()-wantVar) > 1e-12 {
		t.Errorf("Variance() = %v, want %v", dist.Variance(), wantVar)
	}
	if math.Abs(dist.StdDev()-wantStdDev) > 1e-12 {
		t.Errorf("StdDev() = %v, want %v", dist.StdDev(), wantStdDev)
	}
}

func TestBinomialNormalization(t *testing.T) {
	cases := []struct {
		n int
		p float64
	}{
		{0, 0.5},
		{1, 0.5},
		{4, 0.5},
		{10, 0.2},
		{20, 0.75},
		{50, 0.3},
	}

	for _, c := range cases {
		dist, err := mathengine.NewBinomial(c.n, c.p)
		if err != nil {
			t.Fatalf("NewBinomial(%d, %v) failed: %v", c.n, c.p, err)
		}

		sum := 0.0
		for k := 0; k <= c.n; k++ {
			p := dist.PMF(k)
			if p < 0.0 || p > 1.0 {
				t.Errorf("PMF(%d) out of range [0, 1]: %v", k, p)
			}
			sum += p
		}

		if math.Abs(sum-1.0) > 1e-12 {
			t.Errorf("Binomial(%d, %v) sum of PMFs = %v, want 1.0", c.n, c.p, sum)
		}
	}
}

func TestBinomialDegenerateBoundaries(t *testing.T) {
	// 1. p = 0
	bZero, err := mathengine.NewBinomial(5, 0.0)
	if err != nil {
		t.Fatalf("NewBinomial with p=0 failed: %v", err)
	}
	if bZero.PMF(0) != 1.0 {
		t.Errorf("bZero.PMF(0) = %v, want 1.0", bZero.PMF(0))
	}
	if bZero.PMF(1) != 0.0 {
		t.Errorf("bZero.PMF(1) = %v, want 0.0", bZero.PMF(1))
	}
	if bZero.Mean() != 0.0 || bZero.Variance() != 0.0 {
		t.Errorf("bZero moments incorrect: Mean=%v, Var=%v", bZero.Mean(), bZero.Variance())
	}

	// 2. p = 1
	bOne, err := mathengine.NewBinomial(5, 1.0)
	if err != nil {
		t.Fatalf("NewBinomial with p=1 failed: %v", err)
	}
	if bOne.PMF(5) != 1.0 {
		t.Errorf("bOne.PMF(5) = %v, want 1.0", bOne.PMF(5))
	}
	if bOne.PMF(4) != 0.0 {
		t.Errorf("bOne.PMF(4) = %v, want 0.0", bOne.PMF(4))
	}
	if bOne.Mean() != 5.0 || bOne.Variance() != 0.0 {
		t.Errorf("bOne moments incorrect: Mean=%v, Var=%v", bOne.Mean(), bOne.Variance())
	}

	// 3. n = 0
	bEmpty, err := mathengine.NewBinomial(0, 0.5)
	if err != nil {
		t.Fatalf("NewBinomial with n=0 failed: %v", err)
	}
	if bEmpty.PMF(0) != 1.0 {
		t.Errorf("bEmpty.PMF(0) = %v, want 1.0", bEmpty.PMF(0))
	}
	if bEmpty.PMF(1) != 0.0 {
		t.Errorf("bEmpty.PMF(1) = %v, want 0.0", bEmpty.PMF(1))
	}

	// 4. Outside support
	dist, _ := mathengine.NewBinomial(4, 0.5)
	if dist.PMF(-1) != 0.0 {
		t.Errorf("dist.PMF(-1) = %v, want 0.0", dist.PMF(-1))
	}
	if dist.PMF(5) != 0.0 {
		t.Errorf("dist.PMF(5) = %v, want 0.0", dist.PMF(5))
	}

	// 5. Invalid parameters fail
	if _, err := mathengine.NewBinomial(-1, 0.5); err == nil {
		t.Errorf("NewBinomial(-1, 0.5) expected error, got nil")
	}
	if _, err := mathengine.NewBinomial(4, -0.1); err == nil {
		t.Errorf("NewBinomial(4, -0.1) expected error, got nil")
	}
	if _, err := mathengine.NewBinomial(4, 1.1); err == nil {
		t.Errorf("NewBinomial(4, 1.1) expected error, got nil")
	}
}

func TestBinomialCDFAndSurvivalComplement(t *testing.T) {
	dist, err := mathengine.NewBinomial(10, 0.35)
	if err != nil {
		t.Fatalf("NewBinomial failed: %v", err)
	}

	// Boundaries
	if dist.CDF(-1) != 0.0 {
		t.Errorf("CDF(-1) = %v, want 0.0", dist.CDF(-1))
	}
	if dist.CDF(10) != 1.0 {
		t.Errorf("CDF(10) = %v, want 1.0", dist.CDF(10))
	}
	if dist.Survival(0) != 1.0 {
		t.Errorf("Survival(0) = %v, want 1.0", dist.Survival(0))
	}
	if dist.Survival(11) != 0.0 {
		t.Errorf("Survival(11) = %v, want 0.0", dist.Survival(11))
	}

	// Complement identity: P(X <= k) + P(X > k) = 1.0
	for k := 0; k <= 10; k++ {
		cdf := dist.CDF(k)
		tail := dist.Tail(k)
		if math.Abs((cdf+tail)-1.0) > 1e-12 {
			t.Errorf("Complement identity failed at k=%d: CDF=%v + Tail=%v = %v (want 1.0)",
				k, cdf, tail, cdf+tail)
		}

		// CDF step equals PMF: CDF(k) - CDF(k-1) = PMF(k)
		prevCDF := dist.CDF(k - 1)
		pmf := dist.PMF(k)
		if math.Abs((cdf-prevCDF)-pmf) > 1e-12 {
			t.Errorf("CDF step identity failed at k=%d: CDF(k)-CDF(k-1)=%v, PMF(k)=%v",
				k, cdf-prevCDF, pmf)
		}

		// Survival step equals PMF: Survival(k) - Survival(k+1) = PMF(k)
		surv := dist.Survival(k)
		nextSurv := dist.Survival(k + 1)
		if math.Abs((surv-nextSurv)-pmf) > 1e-12 {
			t.Errorf("Survival step identity failed at k=%d: Survival(k)-Survival(k+1)=%v, PMF(k)=%v",
				k, surv-nextSurv, pmf)
		}
	}
}

func TestSumIndependentBinomials(t *testing.T) {
	b1, _ := mathengine.NewBinomial(3, 0.4)
	b2, _ := mathengine.NewBinomial(5, 0.4)

	sumDist, err := mathengine.SumIndependentBinomials(b1, b2)
	if err != nil {
		t.Fatalf("SumIndependentBinomials failed: %v", err)
	}
	if sumDist.N() != 8 {
		t.Errorf("sumDist.N() = %d, want 8", sumDist.N())
	}
	if sumDist.P() != 0.4 {
		t.Errorf("sumDist.P() = %v, want 0.4", sumDist.P())
	}

	// Sum of independent binomials with different p must fail
	bDiff, _ := mathengine.NewBinomial(5, 0.5)
	if _, err := mathengine.SumIndependentBinomials(b1, bDiff); err == nil {
		t.Errorf("SumIndependentBinomials with different p expected error, got nil")
	}
}

func TestLargeBinomialNumerics(t *testing.T) {
	// n=100, p=0.3, k=30
	dist, err := mathengine.NewBinomial(100, 0.3)
	if err != nil {
		t.Fatalf("NewBinomial failed: %v", err)
	}

	pmf30 := dist.PMF(30)
	if pmf30 <= 0.0 || pmf30 > 1.0 {
		t.Errorf("PMF(30) out of reasonable range: %v", pmf30)
	}

	// Check that sum over all 101 support points sums to 1.0 without NaN/Inf
	sum := 0.0
	for k := 0; k <= 100; k++ {
		p := dist.PMF(k)
		if math.IsNaN(p) || math.IsInf(p, 0) {
			t.Fatalf("NaN or Inf encountered at k=%d", k)
		}
		sum += p
	}
	if math.Abs(sum-1.0) > 1e-10 {
		t.Errorf("Binomial(100, 0.3) sum of PMFs = %v, want 1.0", sum)
	}
}
