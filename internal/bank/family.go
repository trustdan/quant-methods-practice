package bank

import (
	"fmt"
	"math"
	"sync"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

// Family defines the mathematical and parameter contract for a class of problems.
type Family interface {
	ID() string
	SupportedRuleVersions() []int
	ValidateParameters(ruleVersion int, params map[string]interface{}) error
}

// TemplateValidator allows a family to enforce family-specific question template structures.
type TemplateValidator interface {
	ValidateTemplate(tmpl *domain.QuestionTemplate) error
}

// Registry stores supported mathematical problem families.
type Registry struct {
	mu       sync.RWMutex
	families map[string]Family
}

// NewRegistry creates a new empty family registry.
func NewRegistry() *Registry {
	return &Registry{
		families: make(map[string]Family),
	}
}

// Register adds a family to the registry.
func (r *Registry) Register(f Family) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.families[f.ID()] = f
}

// Get retrieves a family by its stable ID.
func (r *Registry) Get(id string) (Family, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.families[id]
	return f, ok
}

// BinomialFamily implements the binomial probability distribution family.
type BinomialFamily struct{}

// NewBinomialFamily returns an instance of BinomialFamily.
func NewBinomialFamily() *BinomialFamily {
	return &BinomialFamily{}
}

func (b *BinomialFamily) ID() string {
	return "binomial_pmf"
}

func (b *BinomialFamily) SupportedRuleVersions() []int {
	return []int{1}
}

func (b *BinomialFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion != 1 {
		return fmt.Errorf("binomial_pmf unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"n":         true,
		"p":         true,
		"k":         true,
		"operation": true,
		"mean":      true,
		"sd":        true,
		"var":       true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("binomial_pmf unknown parameter %q; only n, p, k, operation, mean, sd, var allowed", k)
		}
	}

	// Validate n
	rawN, ok := params["n"]
	if !ok {
		return fmt.Errorf("binomial_pmf missing required parameter \"n\"")
	}
	nFloat, ok := toFloat(rawN)
	if !ok || math.Floor(nFloat) != nFloat || nFloat < 0 {
		return fmt.Errorf("binomial_pmf parameter \"n\" must be a non-negative integer, got %v", rawN)
	}
	n := int(nFloat)

	// Validate p
	rawP, ok := params["p"]
	if !ok {
		return fmt.Errorf("binomial_pmf missing required parameter \"p\"")
	}
	pFloat, ok := toFloat(rawP)
	if !ok || pFloat < 0.0 || pFloat > 1.0 {
		return fmt.Errorf("binomial_pmf parameter \"p\" must be in range [0, 1], got %v", rawP)
	}

	// Validate k or operation
	if rawK, ok := params["k"]; ok {
		kFloat, ok := toFloat(rawK)
		if !ok || math.Floor(kFloat) != kFloat || kFloat < 0 {
			return fmt.Errorf("binomial_pmf parameter \"k\" must be a non-negative integer, got %v", rawK)
		}
		k := int(kFloat)
		if k > n {
			return fmt.Errorf("binomial_pmf parameter \"k\" (%d) cannot exceed total trials \"n\" (%d)", k, n)
		}
	} else if rawOp, ok := params["operation"]; ok {
		op, ok := rawOp.(string)
		if !ok || (op != "mean_sd" && op != "mean" && op != "variance") {
			return fmt.Errorf("binomial_pmf invalid operation %q", rawOp)
		}
	} else {
		return fmt.Errorf("binomial_pmf missing required parameter \"k\" or \"operation\"")
	}

	if rawMean, ok := params["mean"]; ok {
		meanFloat, ok := toFloat(rawMean)
		if !ok || math.Abs(meanFloat-float64(n)*pFloat) > 1e-4 {
			return fmt.Errorf("binomial_pmf parameter \"mean\" must match n*p (%v)", float64(n)*pFloat)
		}
	}

	if rawSD, ok := params["sd"]; ok {
		sdFloat, ok := toFloat(rawSD)
		expectedSD := math.Sqrt(float64(n) * pFloat * (1.0 - pFloat))
		if !ok || math.Abs(sdFloat-expectedSD) > 1e-3 {
			return fmt.Errorf("binomial_pmf parameter \"sd\" must match sqrt(n*p*(1-p)) (%v)", expectedSD)
		}
	}

	return nil
}

// PoissonFamily implements the Poisson probability distribution family.
type PoissonFamily struct{}

// NewPoissonFamily returns an instance of PoissonFamily.
func NewPoissonFamily() *PoissonFamily {
	return &PoissonFamily{}
}

func (p *PoissonFamily) ID() string {
	return "poisson_pmf"
}

func (p *PoissonFamily) SupportedRuleVersions() []int {
	return []int{1}
}

func (p *PoissonFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion != 1 {
		return fmt.Errorf("poisson_pmf unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{"lambda": true, "k": true}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("poisson_pmf unknown parameter %q; only lambda, k allowed", k)
		}
	}

	rawLambda, ok := params["lambda"]
	if !ok {
		return fmt.Errorf("poisson_pmf missing required parameter \"lambda\"")
	}
	lambdaFloat, ok := toFloat(rawLambda)
	if !ok || lambdaFloat < 0.0 {
		return fmt.Errorf("poisson_pmf parameter \"lambda\" must be non-negative, got %v", rawLambda)
	}

	rawK, ok := params["k"]
	if !ok {
		return fmt.Errorf("poisson_pmf missing required parameter \"k\"")
	}
	kFloat, ok := toFloat(rawK)
	if !ok || math.Floor(kFloat) != kFloat || kFloat < 0 {
		return fmt.Errorf("poisson_pmf parameter \"k\" must be a non-negative integer, got %v", rawK)
	}

	return nil
}

// SetProbabilityFamily implements set operations and basic probability rules.
type SetProbabilityFamily struct{}

// NewSetProbabilityFamily returns an instance of SetProbabilityFamily.
func NewSetProbabilityFamily() *SetProbabilityFamily {
	return &SetProbabilityFamily{}
}

func (s *SetProbabilityFamily) ID() string {
	return "set_probability"
}

func (s *SetProbabilityFamily) SupportedRuleVersions() []int {
	return []int{1}
}

func (s *SetProbabilityFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion != 1 {
		return fmt.Errorf("set_probability unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"operation":        true,
		"p_a":              true,
		"p_b":              true,
		"p_b1":             true,
		"p_b2":             true,
		"p_intersection":   true,
		"cond_a_given_b1":  true,
		"cond_a_given_b2":  true,
		"target_partition": true,
		"p_total":          true,
		"posterior":        true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("set_probability unknown parameter %q", k)
		}
	}

	rawOp, ok := params["operation"]
	if !ok {
		return fmt.Errorf("set_probability missing required parameter \"operation\"")
	}
	op, ok := rawOp.(string)
	if !ok || op == "" {
		return fmt.Errorf("set_probability parameter \"operation\" must be non-empty string, got %v", rawOp)
	}

	validOps := map[string]bool{
		"union":                   true,
		"complement":              true,
		"conditional":             true,
		"disjoint_vs_independent": true,
		"multiplication":          true,
		"independence_check":      true,
		"bayes":                   true,
		"probability_tree":        true,
	}
	if !validOps[op] {
		return fmt.Errorf("set_probability invalid operation %q", op)
	}

	if op == "bayes" || op == "probability_tree" {
		rawB1, ok := params["p_b1"]
		if !ok {
			return fmt.Errorf("set_probability operation %q requires parameter \"p_b1\"", op)
		}
		rawB2, ok := params["p_b2"]
		if !ok {
			return fmt.Errorf("set_probability operation %q requires parameter \"p_b2\"", op)
		}
		pB1, ok1 := toFloat(rawB1)
		pB2, ok2 := toFloat(rawB2)
		if !ok1 || !ok2 || pB1 < 0.0 || pB2 < 0.0 || math.Abs((pB1+pB2)-1.0) > 1e-4 {
			return fmt.Errorf("set_probability p_b1 (%v) and p_b2 (%v) must be non-negative and sum to 1.0", rawB1, rawB2)
		}
		return nil
	}

	rawPA, ok := params["p_a"]
	if !ok {
		return fmt.Errorf("set_probability missing required parameter \"p_a\"")
	}
	pA, ok := toFloat(rawPA)
	if !ok || pA < 0.0 || pA > 1.0 {
		return fmt.Errorf("set_probability parameter \"p_a\" must be in [0, 1], got %v", rawPA)
	}

	if op == "complement" {
		return nil
	}

	rawPB, ok := params["p_b"]
	if !ok {
		return fmt.Errorf("set_probability operation %q requires parameter \"p_b\"", op)
	}
	pB, ok := toFloat(rawPB)
	if !ok || pB < 0.0 || pB > 1.0 {
		return fmt.Errorf("set_probability parameter \"p_b\" must be in [0, 1], got %v", rawPB)
	}

	if op == "independence_check" {
		if _, ok := params["p_intersection"]; !ok {
			return fmt.Errorf("set_probability operation \"independence_check\" requires parameter \"p_intersection\"")
		}
	}

	if rawPInter, ok := params["p_intersection"]; ok {
		pInter, ok := toFloat(rawPInter)
		if !ok || pInter < 0.0 || pInter > 1.0 {
			return fmt.Errorf("set_probability parameter \"p_intersection\" must be in [0, 1], got %v", rawPInter)
		}
		if err := mathengine.ValidateSetProbabilities(pA, pB, pInter); err != nil {
			return fmt.Errorf("set_probability parameter conflict: %w", err)
		}
	}

	return nil

}

// LinearCombinationFamily implements expectation and variance of sums / linear combinations.
type LinearCombinationFamily struct{}

// NewLinearCombinationFamily returns an instance of LinearCombinationFamily.
func NewLinearCombinationFamily() *LinearCombinationFamily {
	return &LinearCombinationFamily{}
}

func (l *LinearCombinationFamily) ID() string {
	return "linear_combination"
}

func (l *LinearCombinationFamily) SupportedRuleVersions() []int {
	return []int{1, 2}
}

func (l *LinearCombinationFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion == 2 {
		_, err := mathengine.CurriculumCalculations(l.ID(), params)
		return err
	}
	if ruleVersion != 1 {
		return fmt.Errorf("linear_combination unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"operation": true,
		"a":         true,
		"b":         true,
		"c":         true,
		"e_x":       true,
		"e_y":       true,
		"mean_x":    true,
		"mean_y":    true,
		"var_x":     true,
		"var_y":     true,
		"sd_x":      true,
		"sd_y":      true,
		"cov_xy":    true,
		"mean":      true,
		"var":       true,
		"sd":        true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("linear_combination unknown parameter %q", k)
		}
	}

	rawOp, ok := params["operation"]
	if !ok {
		return fmt.Errorf("linear_combination missing required parameter \"operation\"")
	}
	op, ok := rawOp.(string)
	if !ok || op == "" {
		return fmt.Errorf("linear_combination parameter \"operation\" must be non-empty string, got %v", rawOp)
	}

	if op == "affine" {
		rawA, ok := params["a"]
		if !ok {
			return fmt.Errorf("linear_combination operation \"affine\" requires parameter \"a\"")
		}
		if _, ok := toFloat(rawA); !ok {
			return fmt.Errorf("linear_combination parameter \"a\" must be a number, got %v", rawA)
		}
	} else if op == "sum_independent" || op == "diff_independent" {
		// a and b are optional (implicitly 1 and 1 or 1 and -1)
	} else {
		rawA, ok := params["a"]
		if !ok {
			return fmt.Errorf("linear_combination missing required parameter \"a\"")
		}
		if _, ok := toFloat(rawA); !ok {
			return fmt.Errorf("linear_combination parameter \"a\" must be a number, got %v", rawA)
		}

		rawB, ok := params["b"]
		if !ok {
			return fmt.Errorf("linear_combination missing required parameter \"b\"")
		}
		if _, ok := toFloat(rawB); !ok {
			return fmt.Errorf("linear_combination parameter \"b\" must be a number, got %v", rawB)
		}
	}

	// Validate variances and covariance if provided
	var varX, varY, covXY float64
	hasVarX, hasVarY, hasCov := false, false, false

	if rawVarX, ok := params["var_x"]; ok {
		varX, ok = toFloat(rawVarX)
		if !ok || varX < 0.0 {
			return fmt.Errorf("linear_combination parameter \"var_x\" must be non-negative, got %v", rawVarX)
		}
		hasVarX = true
	} else if rawSDX, ok := params["sd_x"]; ok {
		sdX, ok := toFloat(rawSDX)
		if !ok || sdX < 0.0 {
			return fmt.Errorf("linear_combination parameter \"sd_x\" must be non-negative, got %v", rawSDX)
		}
		varX = sdX * sdX
		hasVarX = true
	}

	if rawVarY, ok := params["var_y"]; ok {
		varY, ok = toFloat(rawVarY)
		if !ok || varY < 0.0 {
			return fmt.Errorf("linear_combination parameter \"var_y\" must be non-negative, got %v", rawVarY)
		}
		hasVarY = true
	} else if rawSDY, ok := params["sd_y"]; ok {
		sdY, ok := toFloat(rawSDY)
		if !ok || sdY < 0.0 {
			return fmt.Errorf("linear_combination parameter \"sd_y\" must be non-negative, got %v", rawSDY)
		}
		varY = sdY * sdY
		hasVarY = true
	}

	if rawCov, ok := params["cov_xy"]; ok {
		covXY, ok = toFloat(rawCov)
		if !ok {
			return fmt.Errorf("linear_combination parameter \"cov_xy\" must be a number, got %v", rawCov)
		}
		hasCov = true
	}

	if hasVarX && hasVarY && hasCov {
		maxCov := math.Sqrt(varX * varY)
		if math.Abs(covXY) > maxCov+1e-9 {
			return fmt.Errorf("linear_combination cov_xy=%v violates Cauchy-Schwarz bound [-%v, %v]", covXY, maxCov, maxCov)
		}
	}

	return nil

}

// NormalFamily implements the continuous normal distribution family.
type NormalFamily struct{}

func NewNormalFamily() *NormalFamily {
	return &NormalFamily{}
}

func (n *NormalFamily) ID() string {
	return "normal_dist"
}

func (n *NormalFamily) SupportedRuleVersions() []int {
	return []int{1}
}

func (n *NormalFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion != 1 {
		return fmt.Errorf("normal_dist unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"mu": true, "sigma": true, "x": true, "x1": true, "x2": true, "tail": true,
		"p": true, "k": true, "mu_b": true, "sigma_b": true, "x_b": true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("normal_dist unknown parameter %q", k)
		}
	}

	rawMu, ok := params["mu"]
	if !ok {
		return fmt.Errorf("normal_dist missing required parameter \"mu\"")
	}
	_, ok = toFloat(rawMu)
	if !ok {
		return fmt.Errorf("normal_dist parameter \"mu\" must be a number, got %v", rawMu)
	}

	rawSigma, ok := params["sigma"]
	if !ok {
		return fmt.Errorf("normal_dist missing required parameter \"sigma\"")
	}
	sigma, ok := toFloat(rawSigma)
	if !ok || sigma <= 0 {
		return fmt.Errorf("normal_dist parameter \"sigma\" must be strictly positive, got %v", rawSigma)
	}

	if rawX1, hasX1 := params["x1"]; hasX1 {
		x1, ok1 := toFloat(rawX1)
		if rawX2, hasX2 := params["x2"]; hasX2 {
			x2, ok2 := toFloat(rawX2)
			if ok1 && ok2 && x1 > x2 {
				return fmt.Errorf("normal_dist parameter \"x1\" (%v) cannot exceed \"x2\" (%v)", x1, x2)
			}
		}
	}

	if rawP, hasP := params["p"]; hasP {
		pVal, ok := toFloat(rawP)
		if !ok || pVal <= 0 || pVal >= 1 {
			return fmt.Errorf("normal_dist parameter \"p\" must be in open interval (0, 1), got %v", rawP)
		}
	}

	if rawK, hasK := params["k"]; hasK {
		kVal, ok := toFloat(rawK)
		if !ok || kVal <= 0 {
			return fmt.Errorf("normal_dist parameter \"k\" must be strictly positive, got %v", rawK)
		}
	}

	if rawMuB, hasMuB := params["mu_b"]; hasMuB {
		if _, ok := toFloat(rawMuB); !ok {
			return fmt.Errorf("normal_dist parameter \"mu_b\" must be a number, got %v", rawMuB)
		}
	}

	if rawSigmaB, hasSigmaB := params["sigma_b"]; hasSigmaB {
		sB, ok := toFloat(rawSigmaB)
		if !ok || sB <= 0 {
			return fmt.Errorf("normal_dist parameter \"sigma_b\" must be strictly positive, got %v", rawSigmaB)
		}
	}

	if rawXB, hasXB := params["x_b"]; hasXB {
		if _, ok := toFloat(rawXB); !ok {
			return fmt.Errorf("normal_dist parameter \"x_b\" must be a number, got %v", rawXB)
		}
	}

	return nil
}

// UniformFamily implements the continuous uniform distribution family.
type UniformFamily struct{}

func NewUniformFamily() *UniformFamily {
	return &UniformFamily{}
}

func (u *UniformFamily) ID() string {
	return "uniform_dist"
}

func (u *UniformFamily) SupportedRuleVersions() []int {
	return []int{1}
}

func (u *UniformFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion != 1 {
		return fmt.Errorf("uniform_dist unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"a": true, "b": true, "x": true, "x1": true, "x2": true, "tail": true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("uniform_dist unknown parameter %q", k)
		}
	}

	rawA, ok := params["a"]
	if !ok {
		return fmt.Errorf("uniform_dist missing required parameter \"a\"")
	}
	a, ok := toFloat(rawA)
	if !ok {
		return fmt.Errorf("uniform_dist parameter \"a\" must be a number, got %v", rawA)
	}

	rawB, ok := params["b"]
	if !ok {
		return fmt.Errorf("uniform_dist missing required parameter \"b\"")
	}
	b, ok := toFloat(rawB)
	if !ok {
		return fmt.Errorf("uniform_dist parameter \"b\" must be a number, got %v", rawB)
	}

	if a >= b {
		return fmt.Errorf("uniform_dist lower bound \"a\" (%v) must be strictly less than upper bound \"b\" (%v)", a, b)
	}

	return nil
}

// ExponentialFamily implements the continuous exponential distribution family.
type ExponentialFamily struct{}

func NewExponentialFamily() *ExponentialFamily {
	return &ExponentialFamily{}
}

func (e *ExponentialFamily) ID() string {
	return "exponential_dist"
}

func (e *ExponentialFamily) SupportedRuleVersions() []int {
	return []int{1}
}

func (e *ExponentialFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion != 1 {
		return fmt.Errorf("exponential_dist unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"lambda": true, "mean": true, "x": true, "x1": true, "x2": true, "tail": true,
		"p": true, "s": true, "t": true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("exponential_dist unknown parameter %q", k)
		}
	}

	rawLambda, ok := params["lambda"]
	if !ok {
		return fmt.Errorf("exponential_dist missing required parameter \"lambda\"")
	}
	lambda, ok := toFloat(rawLambda)
	if !ok || lambda <= 0 {
		return fmt.Errorf("exponential_dist parameter \"lambda\" must be strictly positive, got %v", rawLambda)
	}

	if rawMean, hasMean := params["mean"]; hasMean {
		meanVal, ok := toFloat(rawMean)
		if !ok || meanVal <= 0 {
			return fmt.Errorf("exponential_dist parameter \"mean\" must be strictly positive, got %v", rawMean)
		}
	}

	if rawP, hasP := params["p"]; hasP {
		pVal, ok := toFloat(rawP)
		if !ok || pVal <= 0 || pVal >= 1 {
			return fmt.Errorf("exponential_dist parameter \"p\" must be in open interval (0, 1), got %v", rawP)
		}
	}

	if rawS, hasS := params["s"]; hasS {
		sVal, ok := toFloat(rawS)
		if !ok || sVal < 0 {
			return fmt.Errorf("exponential_dist parameter \"s\" must be non-negative, got %v", rawS)
		}
	}

	if rawT, hasT := params["t"]; hasT {
		tVal, ok := toFloat(rawT)
		if !ok || tVal <= 0 {
			return fmt.Errorf("exponential_dist parameter \"t\" must be strictly positive, got %v", rawT)
		}
	}

	if rawX, hasX := params["x"]; hasX {
		xVal, ok := toFloat(rawX)
		if !ok || xVal < 0 {
			return fmt.Errorf("exponential_dist parameter \"x\" must be non-negative, got %v", rawX)
		}
	}

	rawX1, hasX1 := params["x1"]
	rawX2, hasX2 := params["x2"]
	if hasX1 || hasX2 {
		if !hasX1 || !hasX2 {
			return fmt.Errorf("exponential_dist interval requires both \"x1\" and \"x2\"")
		}
		x1, ok1 := toFloat(rawX1)
		x2, ok2 := toFloat(rawX2)
		if !ok1 || !ok2 {
			return fmt.Errorf("exponential_dist interval bounds \"x1\" and \"x2\" must be numbers")
		}
		if x1 > x2 {
			return fmt.Errorf("exponential_dist lower bound \"x1\" (%v) cannot exceed upper bound \"x2\" (%v)", x1, x2)
		}
	}

	return nil
}

// SamplingDistributionFamily implements the sampling distribution family.
type SamplingDistributionFamily struct{}

func NewSamplingDistributionFamily() *SamplingDistributionFamily {
	return &SamplingDistributionFamily{}
}

func (s *SamplingDistributionFamily) ID() string {
	return "sampling_distribution"
}

func (s *SamplingDistributionFamily) SupportedRuleVersions() []int {
	return []int{1}
}

func (s *SamplingDistributionFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion != 1 {
		return fmt.Errorf("sampling_distribution unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"stat": true, "mu": true, "sigma": true, "p": true, "n": true,
		"pop_size": true, "pop_is_normal": true, "pop_shape": true,
		"x_bar": true, "p_hat": true, "x": true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("sampling_distribution unknown parameter %q", k)
		}
	}

	rawN, ok := params["n"]
	if !ok {
		return fmt.Errorf("sampling_distribution missing required parameter \"n\"")
	}
	n, ok := toFloat(rawN)
	if !ok || n <= 0 || math.Floor(n) != n {
		return fmt.Errorf("sampling_distribution parameter \"n\" must be a positive integer, got %v", rawN)
	}

	if rawPopSize, ok := params["pop_size"]; ok {
		popSize, ok := toFloat(rawPopSize)
		if !ok || popSize <= 0 || math.Floor(popSize) != popSize {
			return fmt.Errorf("sampling_distribution parameter \"pop_size\" must be a positive integer, got %v", rawPopSize)
		}
		if int(n) > int(popSize) {
			return fmt.Errorf("sampling_distribution sample size n (%d) cannot exceed pop_size (%d)", int(n), int(popSize))
		}
	}

	if rawPopNormal, ok := params["pop_is_normal"]; ok {
		if _, ok := rawPopNormal.(bool); !ok {
			return fmt.Errorf("sampling_distribution parameter \"pop_is_normal\" must be a boolean, got %v", rawPopNormal)
		}
	}

	if rawPopShape, ok := params["pop_shape"]; ok {
		if shapeStr, ok := rawPopShape.(string); !ok || shapeStr == "" {
			return fmt.Errorf("sampling_distribution parameter \"pop_shape\" must be a non-empty string, got %v", rawPopShape)
		}
	}

	rawStat, hasStat := params["stat"]
	stat := "mean"
	if hasStat {
		if sStr, ok := rawStat.(string); ok {
			stat = sStr
		}
	}

	if stat == "proportion" {
		rawP, ok := params["p"]
		if !ok {
			return fmt.Errorf("sampling_distribution proportion requires parameter \"p\"")
		}
		p, ok := toFloat(rawP)
		if !ok || p <= 0 || p >= 1 {
			return fmt.Errorf("sampling_distribution parameter \"p\" must be in (0, 1), got %v", rawP)
		}

		if rawPHat, ok := params["p_hat"]; ok {
			pHat, ok := toFloat(rawPHat)
			if !ok || pHat <= 0 || pHat >= 1 {
				return fmt.Errorf("sampling_distribution parameter \"p_hat\" must be in (0, 1), got %v", rawPHat)
			}
		}
	} else {
		// mean
		rawMu, ok := params["mu"]
		if !ok {
			return fmt.Errorf("sampling_distribution missing parameter \"mu\"")
		}
		if _, ok := toFloat(rawMu); !ok {
			return fmt.Errorf("sampling_distribution parameter \"mu\" must be a number, got %v", rawMu)
		}

		rawSigma, ok := params["sigma"]
		if !ok {
			return fmt.Errorf("sampling_distribution missing parameter \"sigma\"")
		}
		sigma, ok := toFloat(rawSigma)
		if !ok || sigma <= 0 {
			return fmt.Errorf("sampling_distribution parameter \"sigma\" must be strictly positive, got %v", rawSigma)
		}

		if rawXBar, ok := params["x_bar"]; ok {
			if _, ok := toFloat(rawXBar); !ok {
				return fmt.Errorf("sampling_distribution parameter \"x_bar\" must be a number, got %v", rawXBar)
			}
		}

		if rawX, ok := params["x"]; ok {
			if _, ok := toFloat(rawX); !ok {
				return fmt.Errorf("sampling_distribution parameter \"x\" must be a number, got %v", rawX)
			}
		}
	}

	return nil
}

// ConfidenceIntervalFamily implements the confidence interval family.
type ConfidenceIntervalFamily struct{}

func NewConfidenceIntervalFamily() *ConfidenceIntervalFamily {
	return &ConfidenceIntervalFamily{}
}

func (c *ConfidenceIntervalFamily) ID() string {
	return "confidence_interval"
}

func (c *ConfidenceIntervalFamily) SupportedRuleVersions() []int {
	return []int{1}
}

func (c *ConfidenceIntervalFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion != 1 {
		return fmt.Errorf("confidence_interval unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"parameter": true, "point_estimate": true, "n": true, "confidence_level": true,
		"sigma": true, "s": true, "successes": true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("confidence_interval unknown parameter %q", k)
		}
	}

	rawN, ok := params["n"]
	if !ok {
		return fmt.Errorf("confidence_interval missing required parameter \"n\"")
	}
	n, ok := toFloat(rawN)
	if !ok || n <= 0 || math.Floor(n) != n {
		return fmt.Errorf("confidence_interval parameter \"n\" must be a positive integer, got %v", rawN)
	}

	rawCL, ok := params["confidence_level"]
	if !ok {
		return fmt.Errorf("confidence_interval missing required parameter \"confidence_level\"")
	}
	cl, ok := toFloat(rawCL)
	if !ok || cl <= 0 || cl >= 1 {
		return fmt.Errorf("confidence_interval parameter \"confidence_level\" must be in (0, 1), got %v", rawCL)
	}

	return nil
}

// HypothesisTestFamily implements the hypothesis testing family.
type HypothesisTestFamily struct{}

func NewHypothesisTestFamily() *HypothesisTestFamily {
	return &HypothesisTestFamily{}
}

func (h *HypothesisTestFamily) ID() string {
	return "hypothesis_test"
}

func (h *HypothesisTestFamily) SupportedRuleVersions() []int {
	return []int{1}
}

func (h *HypothesisTestFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion != 1 {
		return fmt.Errorf("hypothesis_test unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"test_type": true, "null_value": true, "sample_stat": true, "n": true,
		"sigma": true, "s": true, "tail": true, "alpha": true, "successes": true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("hypothesis_test unknown parameter %q", k)
		}
	}

	rawNull, ok := params["null_value"]
	if !ok {
		return fmt.Errorf("hypothesis_test missing required parameter \"null_value\"")
	}
	if _, ok := toFloat(rawNull); !ok {
		return fmt.Errorf("hypothesis_test parameter \"null_value\" must be a number, got %v", rawNull)
	}

	rawN, ok := params["n"]
	if !ok {
		return fmt.Errorf("hypothesis_test missing required parameter \"n\"")
	}
	n, ok := toFloat(rawN)
	if !ok || n <= 0 || math.Floor(n) != n {
		return fmt.Errorf("hypothesis_test parameter \"n\" must be a positive integer, got %v", rawN)
	}

	rawAlpha, ok := params["alpha"]
	if !ok {
		return fmt.Errorf("hypothesis_test missing required parameter \"alpha\"")
	}
	alpha, ok := toFloat(rawAlpha)
	if !ok || alpha <= 0 || alpha >= 1 {
		return fmt.Errorf("hypothesis_test parameter \"alpha\" must be in (0, 1), got %v", rawAlpha)
	}

	return nil
}

// LinearRegressionFamily implements the simple linear regression family.
type LinearRegressionFamily struct{}

func NewLinearRegressionFamily() *LinearRegressionFamily {
	return &LinearRegressionFamily{}
}

func (l *LinearRegressionFamily) ID() string {
	return "linear_regression"
}

func (l *LinearRegressionFamily) SupportedRuleVersions() []int {
	return []int{1}
}

func (l *LinearRegressionFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion != 1 {
		return fmt.Errorf("linear_regression unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"n": true, "slope": true, "intercept": true, "r": true, "r_squared": true, "x_pred": true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("linear_regression unknown parameter %q", k)
		}
	}

	rawSlope, ok := params["slope"]
	if !ok {
		return fmt.Errorf("linear_regression missing required parameter \"slope\"")
	}
	if _, ok := toFloat(rawSlope); !ok {
		return fmt.Errorf("linear_regression parameter \"slope\" must be a number, got %v", rawSlope)
	}

	rawIntercept, ok := params["intercept"]
	if !ok {
		return fmt.Errorf("linear_regression missing required parameter \"intercept\"")
	}
	if _, ok := toFloat(rawIntercept); !ok {
		return fmt.Errorf("linear_regression parameter \"intercept\" must be a number, got %v", rawIntercept)
	}

	return nil
}

func toFloat(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	default:
		return 0, false
	}
}

func toFloatSlice(val interface{}) ([]float64, bool) {
	switch slice := val.(type) {
	case []float64:
		return slice, true
	case []interface{}:
		res := make([]float64, len(slice))
		for i, item := range slice {
			f, ok := toFloat(item)
			if !ok {
				return nil, false
			}
			res[i] = f
		}
		return res, true
	default:
		return nil, false
	}
}

// ConceptCheckFamily implements conceptual questions with choice-only stages and verified distractors.
type ConceptCheckFamily struct{}

// NewConceptCheckFamily returns an instance of ConceptCheckFamily.
func NewConceptCheckFamily() *ConceptCheckFamily {
	return &ConceptCheckFamily{}
}

func (c *ConceptCheckFamily) ID() string {
	return "concept_check"
}

func (c *ConceptCheckFamily) SupportedRuleVersions() []int {
	return []int{1}
}

func (c *ConceptCheckFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion != 1 {
		return fmt.Errorf("concept_check unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"topic":     true,
		"concept":   true,
		"objective": true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("concept_check unknown parameter %q", k)
		}
	}
	return nil
}

func (c *ConceptCheckFamily) ValidateTemplate(tmpl *domain.QuestionTemplate) error {
	for i, stage := range tmpl.Stages {
		if stage.Kind != domain.StageKindChoice {
			return fmt.Errorf("concept_check stage[%d] %q must be choice kind, got %q", i, stage.ID, stage.Kind)
		}
		if stage.NumericPolicy != nil {
			return fmt.Errorf("concept_check stage[%d] %q numeric_policy must be null", i, stage.ID)
		}
		if stage.ExpectedAnswer.Kind != domain.AnswerKindChoice {
			return fmt.Errorf("concept_check stage[%d] %q expected_answer.kind must be choice", i, stage.ID)
		}

		expectedID := stage.ExpectedAnswer.OptionID
		foundExpected := false
		for _, opt := range stage.Options {
			if opt.ID == expectedID {
				foundExpected = true
				if opt.MisconceptionID != nil && *opt.MisconceptionID != "" {
					return fmt.Errorf("concept_check stage %q correct option %q must have null misconception_id", stage.ID, opt.ID)
				}
				if opt.HintMarkdown != nil && *opt.HintMarkdown != "" {
					return fmt.Errorf("concept_check stage %q correct option %q must have null hint_markdown", stage.ID, opt.ID)
				}
			} else {
				// Distractor must have non-empty misconception and hint
				if opt.MisconceptionID == nil || *opt.MisconceptionID == "" {
					return fmt.Errorf("concept_check stage %q distractor option %q must have non-empty misconception_id", stage.ID, opt.ID)
				}
				if opt.HintMarkdown == nil || *opt.HintMarkdown == "" {
					return fmt.Errorf("concept_check stage %q distractor option %q must have non-empty hint_markdown", stage.ID, opt.ID)
				}
			}
		}
		if !foundExpected {
			return fmt.Errorf("concept_check stage %q expected_answer option_id %q not in options", stage.ID, expectedID)
		}
		if stage.ExplanationMarkdown == "" {
			return fmt.Errorf("concept_check stage %q explanation_markdown cannot be empty", stage.ID)
		}
	}
	return nil
}

// RegressionOutputFamily implements the Excel-style regression output summary family.
type RegressionOutputFamily struct{}

// NewRegressionOutputFamily returns an instance of RegressionOutputFamily.
func NewRegressionOutputFamily() *RegressionOutputFamily {
	return &RegressionOutputFamily{}
}

func (r *RegressionOutputFamily) ID() string {
	return "regression_output"
}

func (r *RegressionOutputFamily) SupportedRuleVersions() []int {
	return []int{1, 2}
}

func (r *RegressionOutputFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion == 2 {
		_, err := mathengine.CurriculumCalculations(r.ID(), params)
		return err
	}
	if ruleVersion != 1 {
		return fmt.Errorf("regression_output unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"n":                true,
		"k":                true,
		"ssr":              true,
		"sse":              true,
		"sst":              true,
		"confidence_level": true,
		"alpha":            true,
		"b0":               true,
		"se_b0":            true,
		"intercept_name":   true,
		"intercept":        true,
		"predictors":       true,
		"b1":               true,
		"se_b1":            true,
		"x1_name":          true,
		"x1_pred":          true,
		"b2":               true,
		"se_b2":            true,
		"x2_name":          true,
		"x2_pred":          true,
		"b3":               true,
		"se_b3":            true,
		"x3_name":          true,
		"x3_pred":          true,
		"b4":               true,
		"se_b4":            true,
		"x4_name":          true,
		"x4_pred":          true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("regression_output unknown parameter %q", k)
		}
	}

	rawN, ok := params["n"]
	if !ok {
		return fmt.Errorf("regression_output missing required parameter \"n\"")
	}
	nFloat, ok := toFloat(rawN)
	if !ok || math.Floor(nFloat) != nFloat || nFloat <= 0 {
		return fmt.Errorf("regression_output parameter \"n\" must be a positive integer, got %v", rawN)
	}
	n := int(nFloat)

	rawK, ok := params["k"]
	if !ok {
		return fmt.Errorf("regression_output missing required parameter \"k\"")
	}
	kFloat, ok := toFloat(rawK)
	if !ok || math.Floor(kFloat) != kFloat || kFloat < 1 || kFloat > 4 {
		return fmt.Errorf("regression_output parameter \"k\" must be an integer between 1 and 4, got %v", rawK)
	}
	k := int(kFloat)

	rawSSR, ok := params["ssr"]
	if !ok {
		return fmt.Errorf("regression_output missing required parameter \"ssr\"")
	}
	ssr, ok := toFloat(rawSSR)
	if !ok || ssr < 0 {
		return fmt.Errorf("regression_output parameter \"ssr\" must be non-negative, got %v", rawSSR)
	}

	rawSSE, ok := params["sse"]
	if !ok {
		return fmt.Errorf("regression_output missing required parameter \"sse\"")
	}
	sse, ok := toFloat(rawSSE)
	if !ok || sse <= 0 {
		return fmt.Errorf("regression_output parameter \"sse\" must be strictly positive, got %v", rawSSE)
	}

	var sstPtr *float64
	if rawSST, hasSST := params["sst"]; hasSST {
		sstVal, ok := toFloat(rawSST)
		if !ok || sstVal <= 0 {
			return fmt.Errorf("regression_output parameter \"sst\" must be positive, got %v", rawSST)
		}
		sstPtr = &sstVal
	}

	confLevel := 0.95
	if rawCL, hasCL := params["confidence_level"]; hasCL {
		clVal, ok := toFloat(rawCL)
		if !ok || clVal <= 0 || clVal >= 1 {
			return fmt.Errorf("regression_output parameter \"confidence_level\" must be in (0, 1), got %v", rawCL)
		}
		confLevel = clVal
	} else if rawAlpha, hasAlpha := params["alpha"]; hasAlpha {
		alphaVal, ok := toFloat(rawAlpha)
		if !ok || alphaVal <= 0 || alphaVal >= 1 {
			return fmt.Errorf("regression_output parameter \"alpha\" must be in (0, 1), got %v", rawAlpha)
		}
		confLevel = 1.0 - alphaVal
	}

	intercept := mathengine.OutputPredictorInput{Name: "Intercept"}
	if rawIntName, hasIntName := params["intercept_name"]; hasIntName {
		if s, ok := rawIntName.(string); ok && s != "" {
			intercept.Name = s
		}
	}
	if rawB0, hasB0 := params["b0"]; hasB0 {
		b0, ok := toFloat(rawB0)
		if !ok {
			return fmt.Errorf("regression_output parameter \"b0\" must be a number, got %v", rawB0)
		}
		intercept.Coeff = b0
	} else {
		return fmt.Errorf("regression_output missing parameter \"b0\"")
	}
	if rawSeB0, hasSeB0 := params["se_b0"]; hasSeB0 {
		seB0, ok := toFloat(rawSeB0)
		if !ok || seB0 <= 0 {
			return fmt.Errorf("regression_output parameter \"se_b0\" must be strictly positive, got %v", rawSeB0)
		}
		intercept.StdErr = seB0
	} else {
		return fmt.Errorf("regression_output missing parameter \"se_b0\"")
	}

	predictors := make([]mathengine.OutputPredictorInput, k)
	for j := 1; j <= k; j++ {
		bKey := fmt.Sprintf("b%d", j)
		seKey := fmt.Sprintf("se_b%d", j)
		nameKey := fmt.Sprintf("x%d_name", j)

		rawBj, ok := params[bKey]
		if !ok {
			return fmt.Errorf("regression_output missing parameter %q for predictor %d", bKey, j)
		}
		bj, ok := toFloat(rawBj)
		if !ok {
			return fmt.Errorf("regression_output parameter %q must be a number, got %v", bKey, rawBj)
		}

		rawSeBj, ok := params[seKey]
		if !ok {
			return fmt.Errorf("regression_output missing parameter %q for predictor %d", seKey, j)
		}
		seBj, ok := toFloat(rawSeBj)
		if !ok || seBj <= 0 {
			return fmt.Errorf("regression_output parameter %q must be strictly positive, got %v", seKey, rawSeBj)
		}

		predName := fmt.Sprintf("X%d", j)
		if rawName, hasName := params[nameKey]; hasName {
			if s, ok := rawName.(string); ok && s != "" {
				predName = s
			}
		}

		predictors[j-1] = mathengine.OutputPredictorInput{
			Name:   predName,
			Coeff:  bj,
			StdErr: seBj,
		}
	}

	evalParams := mathengine.RegressionOutputParams{
		N:               n,
		K:               k,
		SSR:             ssr,
		SSE:             sse,
		SST:             sstPtr,
		ConfidenceLevel: confLevel,
		Intercept:       intercept,
		Predictors:      predictors,
	}

	if _, err := mathengine.EvaluateRegressionOutput(evalParams); err != nil {
		return fmt.Errorf("regression_output mathematical consistency error: %w", err)
	}

	return nil
}

// SampleSizeFamily implements sample size planning for means, proportions, and standard errors.
type SampleSizeFamily struct{}

// NewSampleSizeFamily returns an instance of SampleSizeFamily.
func NewSampleSizeFamily() *SampleSizeFamily {
	return &SampleSizeFamily{}
}

func (s *SampleSizeFamily) ID() string {
	return "sample_size"
}

func (s *SampleSizeFamily) SupportedRuleVersions() []int {
	return []int{1}
}

func (s *SampleSizeFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion != 1 {
		return fmt.Errorf("sample_size unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"target": true, "sigma": true, "margin_of_error": true, "confidence_level": true,
		"p": true, "target_se": true, "n": true, "n1": true, "n2": true, "moe_ratio": true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("sample_size unknown parameter %q", k)
		}
	}

	rawTarget, ok := params["target"]
	if !ok {
		return fmt.Errorf("sample_size missing required parameter \"target\"")
	}
	target, ok := rawTarget.(string)
	if !ok {
		return fmt.Errorf("sample_size parameter \"target\" must be a string, got %v", rawTarget)
	}

	switch target {
	case "mean":
		rawSigma, ok := params["sigma"]
		if !ok {
			return fmt.Errorf("sample_size target \"mean\" requires parameter \"sigma\"")
		}
		sigma, ok := toFloat(rawSigma)
		if !ok || sigma <= 0 {
			return fmt.Errorf("sample_size parameter \"sigma\" must be strictly positive, got %v", rawSigma)
		}

		rawMOE, ok := params["margin_of_error"]
		if !ok {
			return fmt.Errorf("sample_size target \"mean\" requires parameter \"margin_of_error\"")
		}
		moe, ok := toFloat(rawMOE)
		if !ok || moe <= 0 {
			return fmt.Errorf("sample_size parameter \"margin_of_error\" must be strictly positive, got %v", rawMOE)
		}

		rawCL, ok := params["confidence_level"]
		if !ok {
			return fmt.Errorf("sample_size target \"mean\" requires parameter \"confidence_level\"")
		}
		cl, ok := toFloat(rawCL)
		if !ok || cl <= 0 || cl >= 1 {
			return fmt.Errorf("sample_size parameter \"confidence_level\" must be in (0, 1), got %v", rawCL)
		}

	case "proportion":
		rawMOE, ok := params["margin_of_error"]
		if !ok {
			return fmt.Errorf("sample_size target \"proportion\" requires parameter \"margin_of_error\"")
		}
		moe, ok := toFloat(rawMOE)
		if !ok || moe <= 0 {
			return fmt.Errorf("sample_size parameter \"margin_of_error\" must be strictly positive, got %v", rawMOE)
		}

		rawCL, ok := params["confidence_level"]
		if !ok {
			return fmt.Errorf("sample_size target \"proportion\" requires parameter \"confidence_level\"")
		}
		cl, ok := toFloat(rawCL)
		if !ok || cl <= 0 || cl >= 1 {
			return fmt.Errorf("sample_size parameter \"confidence_level\" must be in (0, 1), got %v", rawCL)
		}

		if rawP, hasP := params["p"]; hasP {
			p, ok := toFloat(rawP)
			if !ok || p <= 0 || p >= 1 {
				return fmt.Errorf("sample_size parameter \"p\" must be in (0, 1), got %v", rawP)
			}
		}

	case "target_se":
		rawSigma, ok := params["sigma"]
		if !ok {
			return fmt.Errorf("sample_size target \"target_se\" requires parameter \"sigma\"")
		}
		sigma, ok := toFloat(rawSigma)
		if !ok || sigma <= 0 {
			return fmt.Errorf("sample_size parameter \"sigma\" must be strictly positive, got %v", rawSigma)
		}

		rawTSE, ok := params["target_se"]
		if !ok {
			return fmt.Errorf("sample_size target \"target_se\" requires parameter \"target_se\"")
		}
		tse, ok := toFloat(rawTSE)
		if !ok || tse <= 0 {
			return fmt.Errorf("sample_size parameter \"target_se\" must be strictly positive, got %v", rawTSE)
		}

	case "comparison", "scaling":
		// valid general comparison cases
	default:
		return fmt.Errorf("sample_size unsupported target %q", target)
	}

	return nil
}

// DescriptiveStatsFamily implements descriptive statistics (mean, median, SD, CV, etc.).
type DescriptiveStatsFamily struct{}

func NewDescriptiveStatsFamily() *DescriptiveStatsFamily {
	return &DescriptiveStatsFamily{}
}

func (d *DescriptiveStatsFamily) ID() string {
	return "descriptive_stats"
}

func (d *DescriptiveStatsFamily) SupportedRuleVersions() []int {
	return []int{1, 2}
}

func (d *DescriptiveStatsFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion == 2 {
		_, err := mathengine.CurriculumCalculations(d.ID(), params)
		return err
	}
	if ruleVersion != 1 {
		return fmt.Errorf("descriptive_stats unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"operation":           true,
		"data":                true,
		"mean":                true,
		"median":              true,
		"mode":                true,
		"range":               true,
		"sample_variance":     true,
		"sample_sd":           true,
		"population_variance": true,
		"population_sd":       true,
		"cv_percent":          true,
		"cv":                  true,
		"mean_a":              true,
		"sd_a":                true,
		"cv_a":                true,
		"mean_b":              true,
		"sd_b":                true,
		"cv_b":                true,
		"x":                   true,
		"z":                   true,
		"weights":             true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("descriptive_stats unknown parameter %q", k)
		}
	}

	rawOp, ok := params["operation"]
	if !ok {
		return fmt.Errorf("descriptive_stats missing required parameter \"operation\"")
	}
	op, ok := rawOp.(string)
	if !ok || op == "" {
		return fmt.Errorf("descriptive_stats parameter \"operation\" must be non-empty string, got %v", rawOp)
	}

	if rawData, ok := params["data"]; ok {
		slice, ok := toFloatSlice(rawData)
		if !ok || len(slice) < 2 {
			return fmt.Errorf("descriptive_stats parameter \"data\" must be a list of at least 2 numbers, got %v", rawData)
		}
	}

	return nil
}

// DiscreteRVFamily implements discrete random variables, PMF tables, and expected value/variance.
type DiscreteRVFamily struct{}

func NewDiscreteRVFamily() *DiscreteRVFamily {
	return &DiscreteRVFamily{}
}

func (d *DiscreteRVFamily) ID() string {
	return "discrete_rv"
}

func (d *DiscreteRVFamily) SupportedRuleVersions() []int {
	return []int{1, 2}
}

func (d *DiscreteRVFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion == 2 {
		_, err := mathengine.CurriculumCalculations(d.ID(), params)
		return err
	}
	if ruleVersion != 1 {
		return fmt.Errorf("discrete_rv unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"operation":      true,
		"values":         true,
		"probabilities":  true,
		"missing_index":  true,
		"missing_prob":   true,
		"expected_value": true,
		"e_x2":           true,
		"variance":       true,
		"sd":             true,
		"k":              true,
		"prob":           true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("discrete_rv unknown parameter %q", k)
		}
	}

	rawOp, ok := params["operation"]
	if !ok {
		return fmt.Errorf("discrete_rv missing required parameter \"operation\"")
	}
	op, ok := rawOp.(string)
	if !ok || op == "" {
		return fmt.Errorf("discrete_rv parameter \"operation\" must be non-empty string, got %v", rawOp)
	}

	if rawVals, ok := params["values"]; ok {
		vals, ok := toFloatSlice(rawVals)
		if !ok || len(vals) == 0 {
			return fmt.Errorf("discrete_rv parameter \"values\" must be a non-empty list of numbers")
		}
	}

	if rawProbs, ok := params["probabilities"]; ok {
		probs, ok := toFloatSlice(rawProbs)
		if !ok || len(probs) == 0 {
			return fmt.Errorf("discrete_rv parameter \"probabilities\" must be a non-empty list of numbers")
		}
	}

	return nil
}

// TwoSampleTestFamily implements paired t-tests, independent two-sample t-tests, and two-proportion z-tests.
type TwoSampleTestFamily struct{}

func NewTwoSampleTestFamily() *TwoSampleTestFamily {
	return &TwoSampleTestFamily{}
}

func (t *TwoSampleTestFamily) ID() string {
	return "two_sample_test"
}

func (t *TwoSampleTestFamily) SupportedRuleVersions() []int {
	return []int{1, 2}
}

func (t *TwoSampleTestFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
	if ruleVersion == 2 {
		_, err := mathengine.CurriculumCalculations(t.ID(), params)
		return err
	}
	if ruleVersion != 1 {
		return fmt.Errorf("two_sample_test unsupported rule_version %d (supported: 1)", ruleVersion)
	}

	allowed := map[string]bool{
		"operation":           true,
		"test_type":           true,
		"n":                   true,
		"mean_diff":           true,
		"sd_diff":             true,
		"hyp_diff":            true,
		"n1":                  true,
		"n2":                  true,
		"mean1":               true,
		"mean2":               true,
		"var1":                true,
		"var2":                true,
		"pooled_var":          true,
		"x1":                  true,
		"x2":                  true,
		"p1_hat":              true,
		"p2_hat":              true,
		"pooled_p":            true,
		"df":                  true,
		"t_stat":              true,
		"z_stat":              true,
		"p_value_one_tail":    true,
		"p_value_two_tail":    true,
		"t_critical_one_tail": true,
		"t_critical_two_tail": true,
		"alpha":               true,
		"decision":            true,
	}
	for k := range params {
		if !allowed[k] {
			return fmt.Errorf("two_sample_test unknown parameter %q", k)
		}
	}

	rawOp, ok := params["operation"]
	if !ok {
		return fmt.Errorf("two_sample_test missing required parameter \"operation\"")
	}
	op, ok := rawOp.(string)
	if !ok || op == "" {
		return fmt.Errorf("two_sample_test parameter \"operation\" must be non-empty string, got %v", rawOp)
	}

	return nil
}

var (
	defaultRegistryOnce sync.Once
	defaultRegistry     *Registry
)

// DefaultRegistry returns the singleton registry with standard supported families registered.
func DefaultRegistry() *Registry {
	defaultRegistryOnce.Do(func() {
		defaultRegistry = NewRegistry()
		defaultRegistry.Register(NewBinomialFamily())
		defaultRegistry.Register(NewPoissonFamily())
		defaultRegistry.Register(NewSetProbabilityFamily())
		defaultRegistry.Register(NewLinearCombinationFamily())
		defaultRegistry.Register(NewNormalFamily())
		defaultRegistry.Register(NewUniformFamily())
		defaultRegistry.Register(NewExponentialFamily())
		defaultRegistry.Register(NewSamplingDistributionFamily())
		defaultRegistry.Register(NewConfidenceIntervalFamily())
		defaultRegistry.Register(NewHypothesisTestFamily())
		defaultRegistry.Register(NewLinearRegressionFamily())
		defaultRegistry.Register(NewConceptCheckFamily())
		defaultRegistry.Register(NewRegressionOutputFamily())
		defaultRegistry.Register(NewSampleSizeFamily())
		defaultRegistry.Register(NewDescriptiveStatsFamily())
		defaultRegistry.Register(NewDiscreteRVFamily())
		defaultRegistry.Register(NewTwoSampleTestFamily())
	})
	return defaultRegistry
}
