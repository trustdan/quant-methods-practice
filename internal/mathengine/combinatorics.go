package mathengine

import (
	"fmt"
	"math"
	"math/big"
)

// ChooseBigInt computes the exact binomial coefficient C(n, k) as an arbitrary-precision integer.
// Returns 0 if n < 0, k < 0, or k > n.
func ChooseBigInt(n, k int) *big.Int {
	if n < 0 || k < 0 || k > n {
		return big.NewInt(0)
	}
	if k == 0 || k == n {
		return big.NewInt(1)
	}
	if k > n-k {
		k = n - k
	}

	result := big.NewInt(1)
	temp := new(big.Int)
	for i := 1; i <= k; i++ {
		result.Mul(result, temp.SetInt64(int64(n-k+i)))
		result.Div(result, temp.SetInt64(int64(i)))
	}
	return result
}

// LogFactorial computes ln(n!) using math.Lgamma for numerical stability.
// Returns math.Inf(-1) if n < 0.
func LogFactorial(n int) float64 {
	if n < 0 {
		return math.Inf(-1)
	}
	if n == 0 || n == 1 {
		return 0.0
	}
	lg, _ := math.Lgamma(float64(n + 1))
	return lg
}

// FactorialBigInt computes exact n! as an arbitrary-precision integer.
// Returns 0 if n < 0.
func FactorialBigInt(n int) *big.Int {
	if n < 0 {
		return big.NewInt(0)
	}
	res := big.NewInt(1)
	if n == 0 || n == 1 {
		return res
	}
	temp := new(big.Int)
	for i := 2; i <= n; i++ {
		res.Mul(res, temp.SetInt64(int64(i)))
	}
	return res
}

// LogChoose computes ln(C(n, k)) using log-gamma to prevent numerical overflow for large n.
// Returns math.Inf(-1) if n < 0, k < 0, or k > n.
func LogChoose(n, k int) float64 {
	if n < 0 || k < 0 || k > n {
		return math.Inf(-1)
	}
	if k == 0 || k == n {
		return 0.0
	}
	if k > n-k {
		k = n - k
	}

	// For small n (<= 60), compute exact ChooseBigInt and take natural logarithm.
	if n <= 60 {
		comb := ChooseBigInt(n, k)
		bf := new(big.Float).SetInt(comb)
		f, _ := bf.Float64()
		return math.Log(f)
	}

	lgN, _ := math.Lgamma(float64(n + 1))
	lgK, _ := math.Lgamma(float64(k + 1))
	lgNK, _ := math.Lgamma(float64(n - k + 1))

	return lgN - lgK - lgNK
}

// Choose computes C(n, k) as float64.
// For large values where C(n, k) overflows float64, it returns math.Inf(1).
func Choose(n, k int) float64 {
	if n < 0 || k < 0 || k > n {
		return 0.0
	}
	if k == 0 || k == n {
		return 1.0
	}
	if k > n-k {
		k = n - k
	}

	if n <= 60 {
		comb := ChooseBigInt(n, k)
		bf := new(big.Float).SetInt(comb)
		f, _ := bf.Float64()
		return f
	}

	logVal := LogChoose(n, k)
	if logVal > 709.782712893384 { // math.MaxFloat64 boundary
		return math.Inf(1)
	}
	return math.Exp(logVal)
}

// FormatCombinationLaTeX returns a LaTeX representation of C(n, k), e.g. \binom{4}{2}.
func FormatCombinationLaTeX(n, k int) string {
	return fmt.Sprintf("\\binom{%d}{%d}", n, k)
}
