package mathengine

import (
	"errors"
	"fmt"
	"math"
	"math/big"
)

// BinomialDistribution represents a binomial random variable X ~ Bin(n, p)
// with n independent and identically distributed Bernoulli trials, each with success probability p.
type BinomialDistribution struct {
	n int
	p float64
}

// NewBinomial creates and validates a new BinomialDistribution.
// Requires n >= 0 and p in [0, 1].
func NewBinomial(n int, p float64) (*BinomialDistribution, error) {
	if n < 0 {
		return nil, fmt.Errorf("binomial parameter n (trials) must be non-negative integer, got %d", n)
	}
	if math.IsNaN(p) || math.IsInf(p, 0) || p < 0.0 || p > 1.0 {
		return nil, fmt.Errorf("binomial parameter p (probability) must be in [0, 1], got %v", p)
	}
	return &BinomialDistribution{n: n, p: p}, nil
}

// N returns the number of trials.
func (b *BinomialDistribution) N() int {
	return b.n
}

// P returns the single-trial success probability.
func (b *BinomialDistribution) P() float64 {
	return b.p
}

// Mean returns the expected value E[X] = n * p.
func (b *BinomialDistribution) Mean() float64 {
	return float64(b.n) * b.p
}

// Variance returns Var(X) = n * p * (1 - p).
func (b *BinomialDistribution) Variance() float64 {
	return float64(b.n) * b.p * (1.0 - b.p)
}

// StdDev returns the standard deviation SD(X) = sqrt(Var(X)).
func (b *BinomialDistribution) StdDev() float64 {
	return math.Sqrt(b.Variance())
}

// Support returns the minimum and maximum integer values of the support [0, n].
func (b *BinomialDistribution) Support() (int, int) {
	return 0, b.n
}

// PMF computes the probability mass function P(X = k) = C(n, k) * p^k * (1-p)^(n-k).
// Returns 0.0 for k outside support [0, n].
func (b *BinomialDistribution) PMF(k int) float64 {
	if k < 0 || k > b.n {
		return 0.0
	}
	if b.n == 0 {
		if k == 0 {
			return 1.0
		}
		return 0.0
	}
	if b.p == 0.0 {
		if k == 0 {
			return 1.0
		}
		return 0.0
	}
	if b.p == 1.0 {
		if k == b.n {
			return 1.0
		}
		return 0.0
	}

	// For small n (<= 30), direct evaluation preserves exact dyadic rationals (e.g. 3/8 -> 0.375).
	if b.n <= 30 {
		comb := Choose(b.n, k)
		prob := comb * math.Pow(b.p, float64(k)) * math.Pow(1.0-b.p, float64(b.n-k))
		return math.Max(0.0, math.Min(1.0, prob))
	}

	// For larger n, use stable log-space computation.
	logProb := b.LogPMF(k)
	return math.Max(0.0, math.Min(1.0, math.Exp(logProb)))
}

// LogPMF computes ln(P(X = k)) using log-gamma and log1p for numerical stability.
// Returns math.Inf(-1) for k outside support [0, n] or impossible events.
func (b *BinomialDistribution) LogPMF(k int) float64 {
	if k < 0 || k > b.n {
		return math.Inf(-1)
	}
	if b.n == 0 {
		if k == 0 {
			return 0.0
		}
		return math.Inf(-1)
	}
	if b.p == 0.0 {
		if k == 0 {
			return 0.0
		}
		return math.Inf(-1)
	}
	if b.p == 1.0 {
		if k == b.n {
			return 0.0
		}
		return math.Inf(-1)
	}

	logComb := LogChoose(b.n, k)
	logP := math.Log(b.p)
	log1MinusP := math.Log1p(-b.p)

	return logComb + float64(k)*logP + float64(b.n-k)*log1MinusP
}

// ExactPMFRat computes P(X = k) as an exact rational *big.Rat given an exact rational pRat.
// Useful for canonical answer verification (e.g. n=4, p=1/2, k=2 -> 3/8).
func (b *BinomialDistribution) ExactPMFRat(k int, pRat *big.Rat) (*big.Rat, error) {
	if pRat == nil {
		return nil, errors.New("pRat cannot be nil")
	}
	zero := big.NewRat(0, 1)
	one := big.NewRat(1, 1)
	if pRat.Cmp(zero) < 0 || pRat.Cmp(one) > 0 {
		return nil, fmt.Errorf("pRat must be in [0, 1], got %s", pRat.RatString())
	}
	if k < 0 || k > b.n {
		return big.NewRat(0, 1), nil
	}
	if b.n == 0 {
		if k == 0 {
			return big.NewRat(1, 1), nil
		}
		return big.NewRat(0, 1), nil
	}
	if pRat.Cmp(zero) == 0 {
		if k == 0 {
			return big.NewRat(1, 1), nil
		}
		return big.NewRat(0, 1), nil
	}
	if pRat.Cmp(one) == 0 {
		if k == b.n {
			return big.NewRat(1, 1), nil
		}
		return big.NewRat(0, 1), nil
	}

	comb := ChooseBigInt(b.n, k)
	combRat := new(big.Rat).SetInt(comb)

	pPow := ratPow(pRat, k)
	qRat := new(big.Rat).Sub(one, pRat)
	qPow := ratPow(qRat, b.n-k)

	res := new(big.Rat).Mul(combRat, pPow)
	res.Mul(res, qPow)
	return res, nil
}

// CDF computes the cumulative distribution function P(X <= k).
// Uses complement summation when k > n/2 to prevent loss of precision.
func (b *BinomialDistribution) CDF(k int) float64 {
	if k < 0 {
		return 0.0
	}
	if k >= b.n {
		return 1.0
	}
	if b.p == 0.0 {
		return 1.0
	}
	if b.p == 1.0 {
		return 0.0
	}

	if k <= b.n/2 {
		sum := 0.0
		for i := 0; i <= k; i++ {
			sum += b.PMF(i)
		}
		return math.Max(0.0, math.Min(1.0, sum))
	}

	// For upper half, sum upper tail directly and subtract from 1 to avoid adding many terms.
	upperTail := 0.0
	for i := k + 1; i <= b.n; i++ {
		upperTail += b.PMF(i)
	}
	return math.Max(0.0, math.Min(1.0, 1.0-upperTail))
}

// Survival computes the survival / upper-tail function P(X >= k).
// When k >= n/2, sums upper tail directly to avoid catastrophic cancellation from 1 - CDF(k-1).
func (b *BinomialDistribution) Survival(k int) float64 {
	if k <= 0 {
		return 1.0
	}
	if k > b.n {
		return 0.0
	}
	if b.p == 0.0 {
		return 0.0
	}
	if b.p == 1.0 {
		return 1.0
	}

	if k >= b.n/2 {
		sum := 0.0
		for i := k; i <= b.n; i++ {
			sum += b.PMF(i)
		}
		return math.Max(0.0, math.Min(1.0, sum))
	}

	// For small k, 1 - CDF(k-1)
	lowerTail := 0.0
	for i := 0; i < k; i++ {
		lowerTail += b.PMF(i)
	}
	return math.Max(0.0, math.Min(1.0, 1.0-lowerTail))
}

// Tail computes P(X > k) = P(X >= k + 1).
func (b *BinomialDistribution) Tail(k int) float64 {
	return b.Survival(k + 1)
}

// SumIndependentBinomials checks that two independent binomials share the same success probability p
// and returns their sum distribution X + Y ~ Bin(n1 + n2, p).
// Fails if success probabilities differ, adhering to NUMERICS.md invariant:
// "The sum of independent binomials is binomial only when their success probability is the same."
func SumIndependentBinomials(b1, b2 *BinomialDistribution) (*BinomialDistribution, error) {
	if b1 == nil || b2 == nil {
		return nil, errors.New("cannot sum nil binomial distributions")
	}
	const tol = 1e-9
	if math.Abs(b1.p-b2.p) > tol {
		return nil, fmt.Errorf("sum of independent binomials is binomial only when success probabilities are equal; got p1=%v, p2=%v", b1.p, b2.p)
	}
	return NewBinomial(b1.n+b2.n, b1.p)
}

func ratPow(r *big.Rat, exp int) *big.Rat {
	if exp < 0 {
		inv := new(big.Rat).Inv(r)
		return ratPow(inv, -exp)
	}
	res := big.NewRat(1, 1)
	if exp == 0 {
		return res
	}
	base := new(big.Rat).Set(r)
	for exp > 0 {
		if exp%2 == 1 {
			res.Mul(res, base)
		}
		exp /= 2
		if exp > 0 {
			base.Mul(base, base)
		}
	}
	return res
}
