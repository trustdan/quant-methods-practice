package mathengine

import (
	"fmt"
	"math"
	"math/big"

	"github.com/trustdan/quant-methods-practice/internal/domain"
)

// DistractorValue represents a derived incorrect candidate answer corresponding to a known misconception.
type DistractorValue struct {
	MisconceptionID string   `json:"misconception_id"`
	Value           *float64 `json:"value,omitempty"`
	ExpressionTeX   string   `json:"expression_tex"`
	Description     string   `json:"description"`
}

// BinomialDerivation contains the canonical answers and pedagogically reviewed distractors
// derived deterministically by the mathematical engine for a binomial scenario.
type BinomialDerivation struct {
	N                    int               `json:"n"`
	P                    float64           `json:"p"`
	K                    int               `json:"k"`
	Event                DiscreteEvent     `json:"event"`
	CanonicalProbability float64           `json:"canonical_probability"`
	CanonicalRational    *big.Rat          `json:"canonical_rational,omitempty"`
	Mean                 float64           `json:"mean"`
	Variance             float64           `json:"variance"`
	StdDev               float64           `json:"std_dev"`
	EventTeX             string            `json:"event_tex"`
	ExpressionTeX        string            `json:"expression_tex"`
	CalculationTeX       string            `json:"calculation_tex"`
	Distractors          []DistractorValue `json:"distractors"`
}

// DeriveBinomialProblem deterministically derives the canonical answers and pedagogical distractors
// for a binomial scenario with parameters (n, p, k) and a target event.
func DeriveBinomialProblem(n int, p float64, k int, event DiscreteEvent) (*BinomialDerivation, error) {
	dist, err := NewBinomial(n, p)
	if err != nil {
		return nil, err
	}

	canonProb := event.EvaluateBinomial(dist)

	// Format expressions
	q := 1.0 - p
	eventTeX := fmt.Sprintf("P(%s)", event.CanonicalLaTeX("X"))
	combTeX := FormatCombinationLaTeX(n, k)
	exprTeX := fmt.Sprintf("%s(%.4g)^{%d}(%.4g)^{%d}", combTeX, p, k, q, n-k)

	// Build exact rational if p can be parsed cleanly as rational (e.g. 0.5 -> 1/2)
	pRat, _ := new(big.Rat).SetString(fmt.Sprintf("%v", p))
	var canonRat *big.Rat
	if pRat != nil {
		canonRat, _ = event.EvaluateBinomialRat(dist, pRat)
	}

	comb := Choose(n, k)
	calcTeX := fmt.Sprintf("%s = %.4g \\times %.4g \\times %.4g = %.6g",
		exprTeX, comb, math.Pow(p, float64(k)), math.Pow(q, float64(n-k)), canonProb)
	if canonRat != nil {
		calcTeX = fmt.Sprintf("%s = %s = %.6g", exprTeX, canonRat.RatString(), canonProb)
	}

	// 1. Missing combination factor: p^k * (1-p)^(n-k)
	missingCombVal := math.Pow(p, float64(k)) * math.Pow(q, float64(n-k))
	missingCombExpr := fmt.Sprintf("(%.4g)^{%d}(%.4g)^{%d}", p, k, q, n-k)

	// 2. Missing failure factor: C(n, k) * p^k
	missingFailureVal := comb * math.Pow(p, float64(k))
	missingFailureExpr := fmt.Sprintf("%s(%.4g)^{%d}", combTeX, p, k)

	// 3. Exactly as at-most: P(X <= k)
	atMostVal := dist.CDF(k)
	atMostExpr := fmt.Sprintf("P(X \\le %d)", k)

	// 4. Exactly as at-least: P(X >= k)
	atLeastVal := dist.Survival(k)
	atLeastExpr := fmt.Sprintf("P(X \\ge %d)", k)

	// 5. Complement event: 1 - P(X = k)
	compVal := 1.0 - dist.PMF(k)
	compExpr := fmt.Sprintf("1 - P(X = %d)", k)

	// 6. Single trial instead of entire experiment: p
	singleTrialVal := p
	singleTrialExpr := fmt.Sprintf("p = %.4g", p)

	// 7. Target as trial count: Binomial with n=k, k=k
	targetTrialsDist, _ := NewBinomial(k, p)
	var targetTrialsVal float64
	if targetTrialsDist != nil {
		targetTrialsVal = targetTrialsDist.PMF(k)
	}
	targetTrialsExpr := fmt.Sprintf("n=%d, p=%.4g", k, p)

	// 8. Count as probability: k (or k/n)
	countAsProbVal := float64(k)
	countAsProbExpr := fmt.Sprintf("k = %d", k)

	distractors := []DistractorValue{
		{
			MisconceptionID: "missing_combination",
			Value:           &missingCombVal,
			ExpressionTeX:   missingCombExpr,
			Description:     "Calculates probability of a single ordered sequence without the combination factor",
		},
		{
			MisconceptionID: "missing_failure_factor",
			Value:           &missingFailureVal,
			ExpressionTeX:   missingFailureExpr,
			Description:     "Omits the failure factor (1-p)^(n-k)",
		},
		{
			MisconceptionID: "exactly_as_at_most",
			Value:           &atMostVal,
			ExpressionTeX:   atMostExpr,
			Description:     "Interprets 'exactly' as 'at most' P(X <= k)",
		},
		{
			MisconceptionID: "exactly_as_at_least",
			Value:           &atLeastVal,
			ExpressionTeX:   atLeastExpr,
			Description:     "Interprets 'exactly' as 'at least' P(X >= k)",
		},
		{
			MisconceptionID: "complement_event",
			Value:           &compVal,
			ExpressionTeX:   compExpr,
			Description:     "Computes complement probability 1 - P(X = k)",
		},
		{
			MisconceptionID: "experiment_as_single_trial",
			Value:           &singleTrialVal,
			ExpressionTeX:   singleTrialExpr,
			Description:     "Confuses experiment probability with single-trial probability p",
		},
		{
			MisconceptionID: "target_as_trial_count",
			Value:           &targetTrialsVal,
			ExpressionTeX:   targetTrialsExpr,
			Description:     "Uses requested success count k as trial count n",
		},
		{
			MisconceptionID: "count_as_probability",
			Value:           &countAsProbVal,
			ExpressionTeX:   countAsProbExpr,
			Description:     "Uses count k directly as a probability",
		},
	}

	return &BinomialDerivation{
		N:                    n,
		P:                    p,
		K:                    k,
		Event:                event,
		CanonicalProbability: canonProb,
		CanonicalRational:    canonRat,
		Mean:                 dist.Mean(),
		Variance:             dist.Variance(),
		StdDev:               dist.StdDev(),
		EventTeX:             eventTeX,
		ExpressionTeX:        exprTeX,
		CalculationTeX:       calcTeX,
		Distractors:          distractors,
	}, nil
}

// CheckDistractorCollisions identifies if any distractor is numerically indistinguishable
// from the canonical probability under the specified policy tolerance.
func (d *BinomialDerivation) CheckDistractorCollisions(policy domain.NumericPolicy) []string {
	var collisions []string
	allowedTol := math.Max(policy.AbsoluteTolerance, policy.RelativeTolerance*math.Abs(d.CanonicalProbability))

	for _, distractor := range d.Distractors {
		if distractor.Value == nil {
			continue
		}
		diff := math.Abs(*distractor.Value - d.CanonicalProbability)
		if diff <= allowedTol {
			collisions = append(collisions, fmt.Sprintf(
				"distractor %q value %.6g collides with canonical probability %.6g (diff: %.6g <= tol: %.6g)",
				distractor.MisconceptionID, *distractor.Value, d.CanonicalProbability, diff, allowedTol,
			))
		}
	}
	return collisions
}
