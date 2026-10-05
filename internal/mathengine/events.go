package mathengine

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// EventRelation defines the canonical relational operator for a discrete random variable event.
type EventRelation string

const (
	RelationEqual            EventRelation = "equal"             // X = k ("exactly k")
	RelationAtMost           EventRelation = "at_most"           // X <= k ("at most k", "no more than k")
	RelationLessThan         EventRelation = "less_than"         // X < k ("fewer than k", "less than k") -> X <= k-1
	RelationAtLeast          EventRelation = "at_least"          // X >= k ("at least k", "no fewer than k")
	RelationGreaterThan      EventRelation = "greater_than"      // X > k ("more than k", "greater than k") -> X >= k+1
	RelationBetweenInclusive EventRelation = "between_inclusive" // k1 <= X <= k2
)

// DiscreteEvent represents a canonical query event over an integer-supported random variable.
type DiscreteEvent struct {
	Relation EventRelation `json:"relation"`
	K        int           `json:"k"`
	K2       int           `json:"k2,omitempty"`
}

// Exactly constructs an exact count event X = k.
func Exactly(k int) DiscreteEvent {
	return DiscreteEvent{Relation: RelationEqual, K: k}
}

// AtMost constructs an upper-bounded event X <= k.
func AtMost(k int) DiscreteEvent {
	return DiscreteEvent{Relation: RelationAtMost, K: k}
}

// LessThan constructs a strict upper-bounded event X < k.
func LessThan(k int) DiscreteEvent {
	return DiscreteEvent{Relation: RelationLessThan, K: k}
}

// AtLeast constructs a lower-bounded event X >= k.
func AtLeast(k int) DiscreteEvent {
	return DiscreteEvent{Relation: RelationAtLeast, K: k}
}

// GreaterThan constructs a strict lower-bounded event X > k.
func GreaterThan(k int) DiscreteEvent {
	return DiscreteEvent{Relation: RelationGreaterThan, K: k}
}

// BetweenInclusive constructs a two-sided inclusive interval event k1 <= X <= k2.
func BetweenInclusive(k1, k2 int) DiscreteEvent {
	if k1 > k2 {
		k1, k2 = k2, k1
	}
	return DiscreteEvent{Relation: RelationBetweenInclusive, K: k1, K2: k2}
}

// Canonicalize converts strict inequalities on integers into equivalent weak inequalities:
// X < k becomes X <= k - 1
// X > k becomes X >= k + 1
func (e DiscreteEvent) Canonicalize() DiscreteEvent {
	switch e.Relation {
	case RelationLessThan:
		return DiscreteEvent{Relation: RelationAtMost, K: e.K - 1}
	case RelationGreaterThan:
		return DiscreteEvent{Relation: RelationAtLeast, K: e.K + 1}
	default:
		return e
	}
}

// LaTeX formats the event into a standard LaTeX mathematical expression.
func (e DiscreteEvent) LaTeX(varName string) string {
	if varName == "" {
		varName = "X"
	}
	switch e.Relation {
	case RelationEqual:
		return fmt.Sprintf("%s = %d", varName, e.K)
	case RelationAtMost:
		return fmt.Sprintf("%s \\le %d", varName, e.K)
	case RelationLessThan:
		return fmt.Sprintf("%s < %d", varName, e.K)
	case RelationAtLeast:
		return fmt.Sprintf("%s \\ge %d", varName, e.K)
	case RelationGreaterThan:
		return fmt.Sprintf("%s > %d", varName, e.K)
	case RelationBetweenInclusive:
		return fmt.Sprintf("%d \\le %s \\le %d", e.K, varName, e.K2)
	default:
		return fmt.Sprintf("%s = %d", varName, e.K)
	}
}

// CanonicalLaTeX formats the canonicalized form into LaTeX.
func (e DiscreteEvent) CanonicalLaTeX(varName string) string {
	return e.Canonicalize().LaTeX(varName)
}

// Description returns a clear human-readable description of the event.
func (e DiscreteEvent) Description() string {
	switch e.Relation {
	case RelationEqual:
		return fmt.Sprintf("exactly %d", e.K)
	case RelationAtMost:
		return fmt.Sprintf("at most %d", e.K)
	case RelationLessThan:
		return fmt.Sprintf("fewer than %d", e.K)
	case RelationAtLeast:
		return fmt.Sprintf("at least %d", e.K)
	case RelationGreaterThan:
		return fmt.Sprintf("more than %d", e.K)
	case RelationBetweenInclusive:
		return fmt.Sprintf("between %d and %d inclusive", e.K, e.K2)
	default:
		return fmt.Sprintf("count %d", e.K)
	}
}

// ParseEventRelation converts common textual expressions or mathematical operators to EventRelation.
func ParseEventRelation(s string) (EventRelation, error) {
	norm := strings.TrimSpace(strings.ToLower(s))
	norm = strings.ReplaceAll(norm, "-", "_")
	norm = strings.ReplaceAll(norm, " ", "_")

	switch norm {
	case "equal", "exactly", "=", "==":
		return RelationEqual, nil
	case "at_most", "no_more_than", "<=":
		return RelationAtMost, nil
	case "less_than", "fewer_than", "strictly_less_than", "<":
		return RelationLessThan, nil
	case "at_least", "no_fewer_than", ">=":
		return RelationAtLeast, nil
	case "greater_than", "more_than", "strictly_greater_than", ">":
		return RelationGreaterThan, nil
	case "between", "between_inclusive":
		return RelationBetweenInclusive, nil
	default:
		return "", fmt.Errorf("unknown event relation %q; valid values include equal, at_most, less_than, at_least, greater_than, between_inclusive", s)
	}
}

// EvaluateBinomial computes the exact probability of the event under a Binomial distribution.
func (e DiscreteEvent) EvaluateBinomial(dist *BinomialDistribution) float64 {
	if dist == nil {
		return 0.0
	}
	canon := e.Canonicalize()
	switch canon.Relation {
	case RelationEqual:
		return dist.PMF(canon.K)
	case RelationAtMost:
		return dist.CDF(canon.K)
	case RelationAtLeast:
		return dist.Survival(canon.K)
	case RelationBetweenInclusive:
		if canon.K > canon.K2 {
			return 0.0
		}
		// P(k1 <= X <= k2) = CDF(k2) - CDF(k1 - 1)
		return dist.CDF(canon.K2) - dist.CDF(canon.K-1)
	default:
		return 0.0
	}
}

// EvaluateBinomialRat computes the probability of the event as an exact *big.Rat under a Binomial distribution.
func (e DiscreteEvent) EvaluateBinomialRat(dist *BinomialDistribution, pRat *big.Rat) (*big.Rat, error) {
	if dist == nil {
		return nil, errors.New("dist cannot be nil")
	}
	canon := e.Canonicalize()
	switch canon.Relation {
	case RelationEqual:
		return dist.ExactPMFRat(canon.K, pRat)
	case RelationAtMost:
		res := big.NewRat(0, 1)
		for i := 0; i <= canon.K && i <= dist.n; i++ {
			p, err := dist.ExactPMFRat(i, pRat)
			if err != nil {
				return nil, err
			}
			res.Add(res, p)
		}
		return res, nil
	case RelationAtLeast:
		res := big.NewRat(0, 1)
		start := canon.K
		if start < 0 {
			start = 0
		}
		for i := start; i <= dist.n; i++ {
			p, err := dist.ExactPMFRat(i, pRat)
			if err != nil {
				return nil, err
			}
			res.Add(res, p)
		}
		return res, nil
	case RelationBetweenInclusive:
		if canon.K > canon.K2 {
			return big.NewRat(0, 1), nil
		}
		res := big.NewRat(0, 1)
		start := canon.K
		if start < 0 {
			start = 0
		}
		end := canon.K2
		if end > dist.n {
			end = dist.n
		}
		for i := start; i <= end; i++ {
			p, err := dist.ExactPMFRat(i, pRat)
			if err != nil {
				return nil, err
			}
			res.Add(res, p)
		}
		return res, nil
	default:
		return nil, fmt.Errorf("unsupported event relation for rational evaluation: %q", canon.Relation)
	}
}
