package mathengine

import (
	"math"
	"testing"
)

func TestFDistribution_Validation(t *testing.T) {
	if _, err := NewFDistribution(0, 10); err == nil {
		t.Error("expected error for d1 <= 0")
	}
	if _, err := NewFDistribution(2, 0); err == nil {
		t.Error("expected error for d2 <= 0")
	}
	if _, err := NewFDistribution(math.NaN(), 10); err == nil {
		t.Error("expected error for NaN d1")
	}
	f, err := NewFDistribution(2, 10)
	if err != nil || f == nil {
		t.Fatalf("unexpected error creating F-distribution: %v", err)
	}
}

func TestFDistribution_EquivalenceWithStudentT(t *testing.T) {
	// For d1 = 1, F = T^2 has the exact property:
	// P(F >= t^2) = 2 * P(T >= |t|) where T ~ t(d2)
	dfs := []float64{5, 10, 20, 30}
	tVals := []float64{0.5, 1.0, 1.96, 2.086, 3.0}

	for _, df := range dfs {
		fDist, err := NewFDistribution(1, df)
		if err != nil {
			t.Fatalf("failed to create F(1, %v): %v", df, err)
		}
		tDist, err := NewTDistribution(df)
		if err != nil {
			t.Fatalf("failed to create T(%v): %v", df, err)
		}

		for _, tVal := range tVals {
			fVal := tVal * tVal
			fSurv := fDist.Survival(fVal)
			tSurv := 2.0 * tDist.Survival(tVal)

			if math.Abs(fSurv-tSurv) > 1e-10 {
				t.Errorf("df=%v, t=%v: F.Survival(%.6g)=%.10g != 2*T.Survival(%.6g)=%.10g (diff: %g)",
					df, tVal, fVal, fSurv, tVal, tSurv, math.Abs(fSurv-tSurv))
			}
		}
	}
}

func TestFDistribution_KnownCriticalValues(t *testing.T) {
	// Critical value F(3, 20) at alpha = 0.05 is approximately 3.0984
	fDist, err := NewFDistribution(3, 20)
	if err != nil {
		t.Fatal(err)
	}

	pVal := fDist.Survival(3.09839)
	if math.Abs(pVal-0.05) > 1e-4 {
		t.Errorf("expected Survival(3.09839) ≈ 0.05, got %v", pVal)
	}

	// CDF + Survival == 1
	cdfVal := fDist.CDF(3.09839)
	if math.Abs(cdfVal+pVal-1.0) > 1e-12 {
		t.Errorf("CDF + Survival != 1, got %v + %v = %v", cdfVal, pVal, cdfVal+pVal)
	}

	// Boundary behavior
	if fDist.CDF(-1) != 0.0 {
		t.Errorf("expected CDF(-1) == 0, got %v", fDist.CDF(-1))
	}
	if fDist.Survival(-1) != 1.0 {
		t.Errorf("expected Survival(-1) == 1, got %v", fDist.Survival(-1))
	}
}

func TestFDistribution_Moments(t *testing.T) {
	fDist, err := NewFDistribution(4, 10)
	if err != nil {
		t.Fatal(err)
	}

	// Mean = 10 / (10 - 2) = 1.25
	if math.Abs(fDist.Mean()-1.25) > 1e-12 {
		t.Errorf("expected Mean == 1.25, got %v", fDist.Mean())
	}

	// Variance = 2 * 10^2 * (4 + 10 - 2) / (4 * (10 - 2)^2 * (10 - 4))
	// = 200 * 12 / (4 * 64 * 6) = 2400 / 1536 = 1.5625
	if math.Abs(fDist.Variance()-1.5625) > 1e-12 {
		t.Errorf("expected Variance == 1.5625, got %v", fDist.Variance())
	}
}
