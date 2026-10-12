package mathengine

import (
	"fmt"
	"math"
)

// ContinuousDistribution defines the common evaluation methods for continuous distributions.
type ContinuousDistribution interface {
	PDF(x float64) float64
	CDF(x float64) float64
	Survival(x float64) float64
	Mean() float64
	Variance() float64
	StdDev() float64
}

// NormalDistribution represents a continuous normal (Gaussian) distribution N(mu, sigma^2).
type NormalDistribution struct {
	Mu    float64
	Sigma float64
}

// NewNormalDistribution constructs a NormalDistribution with mean mu and standard deviation sigma > 0.
func NewNormalDistribution(mu, sigma float64) (*NormalDistribution, error) {
	if math.IsNaN(mu) || math.IsInf(mu, 0) {
		return nil, fmt.Errorf("normal distribution mu must be finite, got %v", mu)
	}
	if math.IsNaN(sigma) || math.IsInf(sigma, 0) || sigma <= 0 {
		return nil, fmt.Errorf("normal distribution sigma must be strictly positive, got %v", sigma)
	}
	return &NormalDistribution{Mu: mu, Sigma: sigma}, nil
}

// StandardNormal returns a standard normal distribution N(0, 1).
func StandardNormal() *NormalDistribution {
	return &NormalDistribution{Mu: 0, Sigma: 1}
}

// PDF returns the probability density function value at x: f(x) = (1 / (sigma * sqrt(2*pi))) * exp(-(x-mu)^2 / (2*sigma^2)).
func (n *NormalDistribution) PDF(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0
	}
	z := (x - n.Mu) / n.Sigma
	coef := 1.0 / (n.Sigma * math.Sqrt(2*math.Pi))
	return coef * math.Exp(-0.5*z*z)
}

// CDF returns the cumulative distribution function P(X <= x) = Phi((x - mu) / sigma).
func (n *NormalDistribution) CDF(x float64) float64 {
	if math.IsNaN(x) {
		return 0
	}
	if math.IsInf(x, -1) {
		return 0
	}
	if math.IsInf(x, 1) {
		return 1
	}
	z := (x - n.Mu) / n.Sigma
	// Phi(z) = 0.5 * (1 + erf(z / sqrt(2)))
	return 0.5 * (1.0 + math.Erf(z/math.Sqrt2))
}

// Survival returns the upper tail survival probability P(X >= x) = 1 - CDF(x) = 0.5 * erfc(z / sqrt(2)).
// Using erfc avoids catastrophic cancellation for large positive z.
func (n *NormalDistribution) Survival(x float64) float64 {
	if math.IsNaN(x) {
		return 0
	}
	if math.IsInf(x, -1) {
		return 1
	}
	if math.IsInf(x, 1) {
		return 0
	}
	z := (x - n.Mu) / n.Sigma
	return 0.5 * math.Erfc(z/math.Sqrt2)
}

// Interval returns P(x1 <= X <= x2) = CDF(x2) - CDF(x1).
func (n *NormalDistribution) Interval(x1, x2 float64) (float64, error) {
	if x1 > x2 {
		return 0, fmt.Errorf("interval lower bound x1 (%v) cannot exceed upper bound x2 (%v)", x1, x2)
	}
	return n.CDF(x2) - n.CDF(x1), nil
}

// ZScore returns the standardized z-score (x - mu) / sigma.
func (n *NormalDistribution) ZScore(x float64) float64 {
	return (x - n.Mu) / n.Sigma
}

// Quantile returns the inverse CDF x such that P(X <= x) = p for p in (0, 1).
// Implements Acklam's rational approximation for standard normal inverse CDF.
func (n *NormalDistribution) Quantile(p float64) (float64, error) {
	if p <= 0 || p >= 1 {
		return 0, fmt.Errorf("quantile probability p must be in open interval (0, 1), got %v", p)
	}
	z := standardNormalQuantile(p)
	return n.Mu + z*n.Sigma, nil
}

func (n *NormalDistribution) Mean() float64 {
	return n.Mu
}

func (n *NormalDistribution) Variance() float64 {
	return n.Sigma * n.Sigma
}

func (n *NormalDistribution) StdDev() float64 {
	return n.Sigma
}

// ExcelCDF returns the Excel formula for P(X <= x): =NORM.DIST(x, mu, sigma, TRUE).
func (n *NormalDistribution) ExcelCDF(x float64) string {
	if n.Mu == 0 && n.Sigma == 1 {
		return fmt.Sprintf("=NORM.S.DIST(%.4g, TRUE)", x)
	}
	return fmt.Sprintf("=NORM.DIST(%.4g, %.4g, %.4g, TRUE)", x, n.Mu, n.Sigma)
}

// ExcelPDF returns the Excel formula for density f(x): =NORM.DIST(x, mu, sigma, FALSE).
func (n *NormalDistribution) ExcelPDF(x float64) string {
	return fmt.Sprintf("=NORM.DIST(%.4g, %.4g, %.4g, FALSE)", x, n.Mu, n.Sigma)
}

// ExcelQuantile returns the Excel formula for quantile: =NORM.INV(p, mu, sigma).
func (n *NormalDistribution) ExcelQuantile(p float64) string {
	if n.Mu == 0 && n.Sigma == 1 {
		return fmt.Sprintf("=NORM.S.INV(%.4g)", p)
	}
	return fmt.Sprintf("=NORM.INV(%.4g, %.4g, %.4g)", p, n.Mu, n.Sigma)
}

// standardNormalQuantile computes Phi^-1(p) using Peter J. Acklam's algorithm.
// Absolute error is less than 1.15e-9 over (0, 1).
func standardNormalQuantile(p float64) float64 {
	// Coefficients in rational approximations
	a := [6]float64{
		-3.969683028665376e+01,
		2.209460984245205e+02,
		-2.759285104469687e+02,
		1.383577518672690e+02,
		-3.066479806614716e+01,
		2.506628277459239e+00,
	}

	b := [5]float64{
		-5.447609879822406e+01,
		1.615858368580409e+02,
		-1.556989798598866e+02,
		6.680131188771972e+01,
		-1.328068155288572e+01,
	}

	c := [6]float64{
		-7.784894002430293e-03,
		-3.223964580411365e-01,
		-2.400758277161838e+00,
		-2.549732539343734e+00,
		4.374664141464968e+00,
		2.938163982698783e+00,
	}

	d := [4]float64{
		7.784695709041462e-03,
		3.224671290700398e-01,
		2.445134137142996e+00,
		3.754408661907416e+00,
	}

	// Break points
	pLow := 0.02425
	pHigh := 1.0 - pLow

	var q, r float64

	if p < pLow {
		// Rational approximation for lower region:
		q = math.Sqrt(-2.0 * math.Log(p))
		return (((((c[0]*q+c[1])*q+c[2])*q+c[3])*q+c[4])*q + c[5]) /
			((((d[0]*q+d[1])*q+d[2])*q+d[3])*q + 1.0)
	} else if p <= pHigh {
		// Rational approximation for central region:
		q = p - 0.5
		r = q * q
		return (((((a[0]*r+a[1])*r+a[2])*r+a[3])*r+a[4])*r + a[5]) * q /
			(((((b[0]*r+b[1])*r+b[2])*r+b[3])*r+b[4])*r + 1.0)
	} else {
		// Rational approximation for upper region:
		q = math.Sqrt(-2.0 * math.Log(1.0-p))
		return -(((((c[0]*q+c[1])*q+c[2])*q+c[3])*q+c[4])*q + c[5]) /
			((((d[0]*q+d[1])*q+d[2])*q+d[3])*q + 1.0)
	}
}

// UniformDistribution represents a continuous uniform distribution U(a, b) with a < b.
type UniformDistribution struct {
	A float64
	B float64
}

// NewUniformDistribution constructs a UniformDistribution with parameters a < b.
func NewUniformDistribution(a, b float64) (*UniformDistribution, error) {
	if math.IsNaN(a) || math.IsInf(a, 0) || math.IsNaN(b) || math.IsInf(b, 0) {
		return nil, fmt.Errorf("uniform distribution bounds must be finite numbers, got a=%v, b=%v", a, b)
	}
	if a >= b {
		return nil, fmt.Errorf("uniform distribution lower bound a (%v) must be strictly less than upper bound b (%v)", a, b)
	}
	return &UniformDistribution{A: a, B: b}, nil
}

// PDF returns the probability density f(x) = 1 / (b - a) for a <= x <= b, 0 elsewhere.
func (u *UniformDistribution) PDF(x float64) float64 {
	if x < u.A || x > u.B {
		return 0
	}
	return 1.0 / (u.B - u.A)
}

// CDF returns P(X <= x).
func (u *UniformDistribution) CDF(x float64) float64 {
	if x <= u.A {
		return 0
	}
	if x >= u.B {
		return 1
	}
	return (x - u.A) / (u.B - u.A)
}

// Survival returns P(X >= x) = 1 - CDF(x).
func (u *UniformDistribution) Survival(x float64) float64 {
	if x <= u.A {
		return 1
	}
	if x >= u.B {
		return 0
	}
	return (u.B - x) / (u.B - u.A)
}

// Interval returns P(x1 <= X <= x2) = CDF(x2) - CDF(x1).
func (u *UniformDistribution) Interval(x1, x2 float64) (float64, error) {
	if x1 > x2 {
		return 0, fmt.Errorf("interval lower bound x1 (%v) cannot exceed upper bound x2 (%v)", x1, x2)
	}
	return u.CDF(x2) - u.CDF(x1), nil
}

// Quantile returns x such that P(X <= x) = p for p in [0, 1].
func (u *UniformDistribution) Quantile(p float64) (float64, error) {
	if p < 0 || p > 1 {
		return 0, fmt.Errorf("uniform quantile probability p must be in [0, 1], got %v", p)
	}
	return u.A + p*(u.B-u.A), nil
}

func (u *UniformDistribution) Mean() float64 {
	return 0.5 * (u.A + u.B)
}

func (u *UniformDistribution) Variance() float64 {
	d := u.B - u.A
	return (d * d) / 12.0
}

func (u *UniformDistribution) StdDev() float64 {
	return math.Sqrt(u.Variance())
}

// ExponentialDistribution represents a continuous exponential distribution with rate parameter lambda > 0.
// Expected value E[X] = 1/lambda; variance Var(X) = 1/lambda^2.
type ExponentialDistribution struct {
	Lambda float64
}

// NewExponentialDistribution constructs an ExponentialDistribution with rate parameter lambda > 0.
func NewExponentialDistribution(lambda float64) (*ExponentialDistribution, error) {
	if math.IsNaN(lambda) || math.IsInf(lambda, 0) || lambda <= 0 {
		return nil, fmt.Errorf("exponential distribution rate lambda must be strictly positive, got %v", lambda)
	}
	return &ExponentialDistribution{Lambda: lambda}, nil
}

// PDF returns f(x) = lambda * exp(-lambda * x) for x >= 0, 0 for x < 0.
func (e *ExponentialDistribution) PDF(x float64) float64 {
	if x < 0 {
		return 0
	}
	return e.Lambda * math.Exp(-e.Lambda*x)
}

// CDF returns P(X <= x) = 1 - exp(-lambda * x) for x >= 0, 0 for x < 0.
func (e *ExponentialDistribution) CDF(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return -math.Expm1(-e.Lambda * x)
}

// Survival returns upper tail probability P(X >= x) = exp(-lambda * x) for x >= 0, 1 for x < 0.
func (e *ExponentialDistribution) Survival(x float64) float64 {
	if x <= 0 {
		return 1
	}
	return math.Exp(-e.Lambda * x)
}

// Interval returns P(x1 <= X <= x2) = CDF(x2) - CDF(x1).
func (e *ExponentialDistribution) Interval(x1, x2 float64) (float64, error) {
	if x1 > x2 {
		return 0, fmt.Errorf("interval lower bound x1 (%v) cannot exceed upper bound x2 (%v)", x1, x2)
	}
	return e.CDF(x2) - e.CDF(x1), nil
}

// Quantile returns x such that P(X <= x) = p for p in [0, 1).
func (e *ExponentialDistribution) Quantile(p float64) (float64, error) {
	if p < 0 || p >= 1 {
		return 0, fmt.Errorf("exponential quantile probability p must be in [0, 1), got %v", p)
	}
	return -math.Log1p(-p) / e.Lambda, nil
}

// MemorylessSurvival returns P(X > s + t | X > s) = P(X > t) = exp(-lambda * t).
func (e *ExponentialDistribution) MemorylessSurvival(t float64) float64 {
	return e.Survival(t)
}

func (e *ExponentialDistribution) Mean() float64 {
	return 1.0 / e.Lambda
}

func (e *ExponentialDistribution) Variance() float64 {
	return 1.0 / (e.Lambda * e.Lambda)
}

func (e *ExponentialDistribution) StdDev() float64 {
	return 1.0 / e.Lambda
}

// ExcelCDF returns the Excel formula for cumulative exponential probability: =EXPON.DIST(x, lambda, TRUE).
func (e *ExponentialDistribution) ExcelCDF(x float64) string {
	return fmt.Sprintf("=EXPON.DIST(%.4g, %.4g, TRUE)", x, e.Lambda)
}

// ExcelPDF returns the Excel formula for exponential density: =EXPON.DIST(x, lambda, FALSE).
func (e *ExponentialDistribution) ExcelPDF(x float64) string {
	return fmt.Sprintf("=EXPON.DIST(%.4g, %.4g, FALSE)", x, e.Lambda)
}

// ExcelQuantile returns the Excel formula for exponential quantile: =-LN(1 - p) / lambda.
func (e *ExponentialDistribution) ExcelQuantile(p float64) string {
	return fmt.Sprintf("=-LN(1 - %.4g) / %.4g", p, e.Lambda)
}

// ContinuousEventKind specifies continuous event relations.
type ContinuousEventKind string

const (
	ContinuousEventEqual       ContinuousEventKind = "equal"
	ContinuousEventLessThan    ContinuousEventKind = "less_than"
	ContinuousEventAtMost      ContinuousEventKind = "at_most"
	ContinuousEventGreaterThan ContinuousEventKind = "greater_than"
	ContinuousEventAtLeast     ContinuousEventKind = "at_least"
	ContinuousEventBetween     ContinuousEventKind = "between"
)

// ContinuousEventResult contains the evaluation of a continuous event along with pedagogical explanations.
type ContinuousEventResult struct {
	Kind                ContinuousEventKind `json:"kind"`
	Probability         float64             `json:"probability"`
	DensityAtCutoff     float64             `json:"density_at_cutoff,omitempty"`
	Explanation         string              `json:"explanation"`
	ExcelFormula        string              `json:"excel_formula,omitempty"`
	IsSinglePointZero   bool                `json:"is_single_point_zero"`
	EndpointsMatterNote string              `json:"endpoints_matter_note"`
}

// EvaluateContinuousEvent evaluates a continuous event for any ContinuousDistribution.
// Enforces the core invariant that P(X = c) = 0 for continuous random variables,
// and endpoints do not alter probability: P(X <= c) = P(X < c).
func EvaluateContinuousEvent(dist ContinuousDistribution, kind ContinuousEventKind, x1, x2 float64) (*ContinuousEventResult, error) {
	endpointsNote := "In continuous distributions, endpoints do not change probability: P(X <= c) = P(X < c) because P(X = c) = 0 for any single point."

	switch kind {
	case ContinuousEventEqual:
		density := dist.PDF(x1)
		return &ContinuousEventResult{
			Kind:                ContinuousEventEqual,
			Probability:         0.0,
			DensityAtCutoff:     density,
			IsSinglePointZero:   true,
			Explanation:         fmt.Sprintf("P(X = %.4g) = 0. In continuous distributions, the probability of any single exact point is strictly zero. The curve height f(%.4g) = %.4g is the probability density, not probability.", x1, x1, density),
			EndpointsMatterNote: endpointsNote,
		}, nil

	case ContinuousEventLessThan, ContinuousEventAtMost:
		prob := dist.CDF(x1)
		density := dist.PDF(x1)
		var excelFormula string
		if norm, ok := dist.(*NormalDistribution); ok {
			excelFormula = norm.ExcelCDF(x1)
		} else if exp, ok := dist.(*ExponentialDistribution); ok {
			excelFormula = exp.ExcelCDF(x1)
		}
		return &ContinuousEventResult{
			Kind:                kind,
			Probability:         prob,
			DensityAtCutoff:     density,
			Explanation:         fmt.Sprintf("P(X <= %.4g) = %.4g, calculated by the cumulative distribution function (area under the curve to the left of %.4g).", x1, prob, x1),
			ExcelFormula:        excelFormula,
			EndpointsMatterNote: endpointsNote,
		}, nil

	case ContinuousEventGreaterThan, ContinuousEventAtLeast:
		prob := dist.Survival(x1)
		density := dist.PDF(x1)
		var excelFormula string
		if norm, ok := dist.(*NormalDistribution); ok {
			excelFormula = fmt.Sprintf("1 - %s", norm.ExcelCDF(x1))
		} else if exp, ok := dist.(*ExponentialDistribution); ok {
			excelFormula = fmt.Sprintf("1 - %s", exp.ExcelCDF(x1))
		}
		return &ContinuousEventResult{
			Kind:                kind,
			Probability:         prob,
			DensityAtCutoff:     density,
			Explanation:         fmt.Sprintf("P(X >= %.4g) = %.4g, representing the upper-tail survival probability (area to the right of %.4g).", x1, prob, x1),
			ExcelFormula:        excelFormula,
			EndpointsMatterNote: endpointsNote,
		}, nil

	case ContinuousEventBetween:
		if x1 > x2 {
			return nil, fmt.Errorf("between event lower bound x1 (%v) cannot exceed upper bound x2 (%v)", x1, x2)
		}
		prob := dist.CDF(x2) - dist.CDF(x1)
		var excelFormula string
		if norm, ok := dist.(*NormalDistribution); ok {
			excelFormula = fmt.Sprintf("%s - %s", norm.ExcelCDF(x2), norm.ExcelCDF(x1))
		} else if exp, ok := dist.(*ExponentialDistribution); ok {
			excelFormula = fmt.Sprintf("%s - %s", exp.ExcelCDF(x2), exp.ExcelCDF(x1))
		}
		return &ContinuousEventResult{
			Kind:                ContinuousEventBetween,
			Probability:         prob,
			Explanation:         fmt.Sprintf("P(%.4g <= X <= %.4g) = F(%.4g) - F(%.4g) = %.4g (area under curve between bounds).", x1, x2, x2, x1, prob),
			ExcelFormula:        excelFormula,
			EndpointsMatterNote: endpointsNote,
		}, nil

	default:
		return nil, fmt.Errorf("unknown continuous event kind %q", kind)
	}
}
