package bank

import (
	"fmt"
	"math"
	"sync"

	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

// Family defines the mathematical and parameter contract for a class of problems.
type Family interface {
	ID() string
	SupportedRuleVersions() []int
	ValidateParameters(ruleVersion int, params map[string]interface{}) error
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

	required := map[string]bool{"n": true, "p": true, "k": true}
	for k := range params {
		if !required[k] {
			return fmt.Errorf("binomial_pmf unknown parameter %q; only n, p, k allowed", k)
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

	// Validate k
	rawK, ok := params["k"]
	if !ok {
		return fmt.Errorf("binomial_pmf missing required parameter \"k\"")
	}
	kFloat, ok := toFloat(rawK)
	if !ok || math.Floor(kFloat) != kFloat || kFloat < 0 {
		return fmt.Errorf("binomial_pmf parameter \"k\" must be a non-negative integer, got %v", rawK)
	}
	k := int(kFloat)

	if k > n {
		return fmt.Errorf("binomial_pmf parameter \"k\" (%d) cannot exceed total trials \"n\" (%d)", k, n)
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
		"operation":      true,
		"p_a":            true,
		"p_b":            true,
		"p_intersection": true,
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
	}
	if !validOps[op] {
		return fmt.Errorf("set_probability invalid operation %q", op)
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
	return []int{1}
}

func (l *LinearCombinationFamily) ValidateParameters(ruleVersion int, params map[string]interface{}) error {
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
		"var_x":     true,
		"var_y":     true,
		"cov_xy":    true,
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

	rawA, ok := params["a"]
	if !ok {
		return fmt.Errorf("linear_combination missing required parameter \"a\"")
	}
	_, ok = toFloat(rawA)
	if !ok {
		return fmt.Errorf("linear_combination parameter \"a\" must be a number, got %v", rawA)
	}

	rawB, ok := params["b"]
	if !ok {
		return fmt.Errorf("linear_combination missing required parameter \"b\"")
	}
	_, ok = toFloat(rawB)
	if !ok {
		return fmt.Errorf("linear_combination parameter \"b\" must be a number, got %v", rawB)
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
	}
	if rawVarY, ok := params["var_y"]; ok {
		varY, ok = toFloat(rawVarY)
		if !ok || varY < 0.0 {
			return fmt.Errorf("linear_combination parameter \"var_y\" must be non-negative, got %v", rawVarY)
		}
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
		"lambda": true, "x": true, "x1": true, "x2": true, "tail": true,
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
	})
	return defaultRegistry
}
