package mathengine

import (
	"errors"
	"fmt"
	"math"
)

// PoissonDistribution represents a Poisson random variable X ~ Poisson(lambda),
// representing the count of events occurring in a fixed interval or exposure.
type PoissonDistribution struct {
	lambda float64
}

// NewPoisson creates and validates a PoissonDistribution.
// Requires lambda >= 0.
func NewPoisson(lambda float64) (*PoissonDistribution, error) {
	if math.IsNaN(lambda) || math.IsInf(lambda, 0) || lambda < 0.0 {
		return nil, fmt.Errorf("poisson parameter lambda must be a finite non-negative number, got %v", lambda)
	}
	return &PoissonDistribution{lambda: lambda}, nil
}

// Lambda returns the rate parameter.
func (p *PoissonDistribution) Lambda() float64 {
	return p.lambda
}

// Mean returns the expected value E[X] = lambda.
func (p *PoissonDistribution) Mean() float64 {
	return p.lambda
}

// Variance returns Var(X) = lambda.
func (p *PoissonDistribution) Variance() float64 {
	return p.lambda
}

// StdDev returns the standard deviation SD(X) = sqrt(lambda).
func (p *PoissonDistribution) StdDev() float64 {
	return math.Sqrt(p.lambda)
}

// PMF computes P(X = k) = exp(-lambda) * lambda^k / k!.
// Returns 0.0 for k < 0.
// Handles lambda = 0 explicitly: P(X=0)=1, P(X>0)=0.
func (p *PoissonDistribution) PMF(k int) float64 {
	if k < 0 {
		return 0.0
	}
	if p.lambda == 0.0 {
		if k == 0 {
			return 1.0
		}
		return 0.0
	}

	// For small k and moderate lambda, evaluate in log-space to prevent overflow of k! and underflow of exp(-lambda).
	logP := p.LogPMF(k)
	if math.IsInf(logP, -1) {
		return 0.0
	}
	return math.Max(0.0, math.Min(1.0, math.Exp(logP)))
}

// LogPMF computes ln(P(X = k)) = -lambda + k*ln(lambda) - ln(k!).
func (p *PoissonDistribution) LogPMF(k int) float64 {
	if k < 0 {
		return math.Inf(-1)
	}
	if p.lambda == 0.0 {
		if k == 0 {
			return 0.0
		}
		return math.Inf(-1)
	}

	lgK1, _ := math.Lgamma(float64(k + 1))
	return -p.lambda + float64(k)*math.Log(p.lambda) - lgK1
}

// CDF computes the cumulative distribution function P(X <= k).
func (p *PoissonDistribution) CDF(k int) float64 {
	if k < 0 {
		return 0.0
	}
	if p.lambda == 0.0 {
		return 1.0
	}

	sum := 0.0
	for i := 0; i <= k; i++ {
		sum += p.PMF(i)
	}
	return math.Max(0.0, math.Min(1.0, sum))
}

// Survival computes P(X >= k).
func (p *PoissonDistribution) Survival(k int) float64 {
	if k <= 0 {
		return 1.0
	}
	if p.lambda == 0.0 {
		return 0.0
	}

	// If k is well below the mean lambda, 1 - CDF(k-1) is stable.
	// If k is above the mean lambda, sum upper tail directly to avoid catastrophic cancellation.
	if float64(k) <= p.lambda {
		return math.Max(0.0, math.Min(1.0, 1.0-p.CDF(k-1)))
	}

	sum := 0.0
	i := k
	for {
		term := p.PMF(i)
		sum += term
		if term < 1e-18 || i > k+5000 {
			break
		}
		i++
	}
	return math.Max(0.0, math.Min(1.0, sum))
}

// Tail computes P(X > k) = Survival(k+1).
func (p *PoissonDistribution) Tail(k int) float64 {
	return p.Survival(k + 1)
}

// EvaluatePoisson computes the probability of a DiscreteEvent under a Poisson distribution.
func (p *PoissonDistribution) EvaluateEvent(e DiscreteEvent) float64 {
	canon := e.Canonicalize()
	switch canon.Relation {
	case RelationEqual:
		return p.PMF(canon.K)
	case RelationAtMost:
		return p.CDF(canon.K)
	case RelationAtLeast:
		return p.Survival(canon.K)
	case RelationBetweenInclusive:
		if canon.K > canon.K2 {
			return 0.0
		}
		return p.CDF(canon.K2) - p.CDF(canon.K-1)
	default:
		return 0.0
	}
}

// SumIndependentPoissons returns the distribution of the sum of two independent Poisson variables:
// X + Y ~ Poisson(lambda1 + lambda2).
func SumIndependentPoissons(p1, p2 *PoissonDistribution) (*PoissonDistribution, error) {
	if p1 == nil || p2 == nil {
		return nil, errors.New("cannot sum nil Poisson distributions")
	}
	return NewPoisson(p1.lambda + p2.lambda)
}
