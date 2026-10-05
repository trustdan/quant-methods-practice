package mathengine_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

func TestChooseBigInt(t *testing.T) {
	tests := []struct {
		n, k     int
		expected string
	}{
		{4, 2, "6"},
		{5, 3, "10"},
		{10, 5, "252"},
		{52, 5, "2598960"},
		{0, 0, "1"},
		{5, 0, "1"},
		{5, 5, "1"},
		{5, 6, "0"},
		{5, -1, "0"},
		{-1, 0, "0"},
	}

	for _, tt := range tests {
		actual := mathengine.ChooseBigInt(tt.n, tt.k)
		expected, _ := new(big.Int).SetString(tt.expected, 10)
		if actual.Cmp(expected) != 0 {
			t.Errorf("ChooseBigInt(%d, %d) = %s, want %s", tt.n, tt.k, actual.String(), tt.expected)
		}
	}
}

func TestChooseSymmetry(t *testing.T) {
	for n := 0; n <= 20; n++ {
		for k := 0; k <= n; k++ {
			c1 := mathengine.ChooseBigInt(n, k)
			c2 := mathengine.ChooseBigInt(n, n-k)
			if c1.Cmp(c2) != 0 {
				t.Fatalf("Symmetry broken for C(%d, %d): %s != %s", n, k, c1.String(), c2.String())
			}
		}
	}
}

func TestChooseAndLogChoose(t *testing.T) {
	tests := []struct {
		n, k int
		want float64
	}{
		{4, 2, 6.0},
		{10, 3, 120.0},
		{20, 10, 184756.0},
		{5, 0, 1.0},
		{5, 5, 1.0},
		{5, 6, 0.0},
		{5, -1, 0.0},
	}

	for _, tt := range tests {
		got := mathengine.Choose(tt.n, tt.k)
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("Choose(%d, %d) = %v, want %v", tt.n, tt.k, got, tt.want)
		}

		logVal := mathengine.LogChoose(tt.n, tt.k)
		if tt.want > 0 {
			expectedLog := math.Log(tt.want)
			if math.Abs(logVal-expectedLog) > 1e-9 {
				t.Errorf("LogChoose(%d, %d) = %v, want %v", tt.n, tt.k, logVal, expectedLog)
			}
		} else {
			if !math.IsInf(logVal, -1) {
				t.Errorf("LogChoose(%d, %d) = %v, want -Inf", tt.n, tt.k, logVal)
			}
		}
	}
}

func TestFactorials(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "1"},
		{1, "1"},
		{5, "120"},
		{10, "3628800"},
		{-1, "0"},
	}

	for _, tt := range tests {
		actual := mathengine.FactorialBigInt(tt.n)
		expected, _ := new(big.Int).SetString(tt.want, 10)
		if actual.Cmp(expected) != 0 {
			t.Errorf("FactorialBigInt(%d) = %s, want %s", tt.n, actual.String(), tt.want)
		}
	}
}

func TestFormatCombinationLaTeX(t *testing.T) {
	got := mathengine.FormatCombinationLaTeX(4, 2)
	want := "\\binom{4}{2}"
	if got != want {
		t.Errorf("FormatCombinationLaTeX(4, 2) = %q, want %q", got, want)
	}
}
