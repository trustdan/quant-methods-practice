package mathengine

import (
	"fmt"
	"math"
)

// TDistribution represents Student's t-distribution with nu degrees of freedom.
type TDistribution struct {
	Nu float64 // Degrees of freedom (nu > 0)
}

// NewTDistribution constructs a Student's t-distribution with nu degrees of freedom.
func NewTDistribution(nu float64) (*TDistribution, error) {
	if math.IsNaN(nu) || math.IsInf(nu, 0) || nu <= 0 {
		return nil, fmt.Errorf("degrees of freedom nu must be strictly positive, got %v", nu)
	}
	return &TDistribution{Nu: nu}, nil
}

// PDF returns the probability density function value at t:
// f(t; nu) = Gamma((nu+1)/2) / (sqrt(pi*nu) * Gamma(nu/2)) * (1 + t^2/nu)^(-(nu+1)/2)
func (t *TDistribution) PDF(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0
	}
	nu := t.Nu
	// Compute in log-gamma space for numerical stability
	lg1, _ := math.Lgamma(0.5 * (nu + 1.0))
	lg2, _ := math.Lgamma(0.5 * nu)
	lnNorm := lg1 - lg2 - 0.5*math.Log(math.Pi*nu)
	lnKernel := -0.5 * (nu + 1.0) * math.Log1p(x*x/nu)
	return math.Exp(lnNorm + lnKernel)
}

// CDF returns the cumulative distribution function P(T <= t).
// Utilizes the regularized incomplete beta function I_x(a, b).
func (t *TDistribution) CDF(x float64) float64 {
	if math.IsNaN(x) {
		return 0
	}
	if math.IsInf(x, -1) {
		return 0
	}
	if math.IsInf(x, 1) {
		return 1
	}
	if x == 0 {
		return 0.5
	}

	nu := t.Nu
	// For Student's t:
	// x_beta = nu / (nu + t^2)
	// P(T > |t|) = I_{x_beta}(nu/2, 0.5)
	xBeta := nu / (nu + x*x)
	ib := RegularizedIncompleteBeta(xBeta, 0.5*nu, 0.5)

	if x > 0 {
		return 1.0 - 0.5*ib
	}
	return 0.5 * ib
}

// Survival returns P(T >= t) = 1 - CDF(t).
func (t *TDistribution) Survival(x float64) float64 {
	if math.IsNaN(x) {
		return 0
	}
	if math.IsInf(x, -1) {
		return 1
	}
	if math.IsInf(x, 1) {
		return 0
	}
	if x == 0 {
		return 0.5
	}

	nu := t.Nu
	xBeta := nu / (nu + x*x)
	ib := RegularizedIncompleteBeta(xBeta, 0.5*nu, 0.5)

	if x > 0 {
		return 0.5 * ib
	}
	return 1.0 - 0.5*ib
}

// Quantile returns the inverse CDF t such that CDF(t) = p for p in (0, 1).
func (t *TDistribution) Quantile(p float64) (float64, error) {
	if p <= 0 || p >= 1 {
		return 0, fmt.Errorf("quantile probability p must be in open interval (0, 1), got %v", p)
	}
	if p == 0.5 {
		return 0.0, nil
	}
	// Symmetry: t_p = -t_{1-p}
	if p < 0.5 {
		q, err := t.Quantile(1.0 - p)
		if err != nil {
			return 0, err
		}
		return -q, nil
	}

	nu := t.Nu

	// As nu -> inf, t distribution converges to standard normal
	z := standardNormalQuantile(p)
	if nu >= 1000 {
		return z, nil
	}

	// Cornish-Fisher expansion initial approximation
	z2 := z * z
	z3 := z * z2
	z5 := z3 * z2
	c1 := 0.25 * (z3 + z)
	c2 := (5.0*z5 + 16.0*z3 + 3.0*z) / 96.0
	c3 := (3.0*z5*z2 + 19.0*z5 + 17.0*z3 - 15.0*z) / 384.0
	cur := z + c1/nu + c2/(nu*nu) + c3/(nu*nu*nu)
	if cur <= 0 {
		cur = z
	}

	// Newton-Raphson / Halley refinement bounded with bracket
	low := 0.0
	high := math.Max(cur*2.0, 10.0)

	for iter := 0; iter < 25; iter++ {
		cdf := t.CDF(cur)
		diff := cdf - p
		if math.Abs(diff) < 1e-12 {
			break
		}
		if diff > 0 {
			high = cur
		} else {
			low = cur
		}

		pdf := t.PDF(cur)
		if pdf > 1e-100 {
			ratio := diff / pdf
			// Halley correction
			f2overf1 := -(nu + 1.0) * cur / (nu + cur*cur)
			denom := 1.0 - 0.5*ratio*f2overf1
			step := ratio / denom
			next := cur - step
			if next > low && next < high {
				cur = next
				if math.Abs(step) < 1e-12 {
					break
				}
				continue
			}
		}
		// Bisection fallback
		cur = 0.5 * (low + high)
	}

	return cur, nil
}

// Mean returns the expected value of T: 0 if nu > 1, undefined (NaN) if nu <= 1.
func (t *TDistribution) Mean() float64 {
	if t.Nu > 1.0 {
		return 0.0
	}
	return math.NaN()
}

// Variance returns the variance of T: nu / (nu - 2) if nu > 2, +Inf if 1 < nu <= 2, NaN if nu <= 1.
func (t *TDistribution) Variance() float64 {
	if t.Nu > 2.0 {
		return t.Nu / (t.Nu - 2.0)
	}
	if t.Nu > 1.0 {
		return math.Inf(1)
	}
	return math.NaN()
}

// StdDev returns the standard deviation if nu > 2.
func (t *TDistribution) StdDev() float64 {
	v := t.Variance()
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return v
	}
	return math.Sqrt(v)
}

// ExcelCDF returns the Excel formula for P(T <= t): =T.DIST(x, df, TRUE).
func (t *TDistribution) ExcelCDF(x float64) string {
	return fmt.Sprintf("=T.DIST(%.4g, %.4g, TRUE)", x, t.Nu)
}

// Excel2TTail returns the Excel formula for two-tailed probability P(|T| >= |t|): =T.DIST.2T(ABS(x), df).
func (t *TDistribution) Excel2TTail(x float64) string {
	return fmt.Sprintf("=T.DIST.2T(ABS(%.4g), %.4g)", x, t.Nu)
}

// ExcelQuantile2T returns the Excel formula for two-tailed critical value: =T.INV.2T(alpha, df).
func (t *TDistribution) ExcelQuantile2T(alpha float64) string {
	return fmt.Sprintf("=T.INV.2T(%.4g, %.4g)", alpha, t.Nu)
}

// RegularizedIncompleteBeta evaluates I_x(a, b) = B(x; a, b) / B(a, b) for x in [0, 1] and a, b > 0.
func RegularizedIncompleteBeta(x, a, b float64) float64 {
	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}

	// Symmetry transformation if x > (a + 1) / (a + b + 2)
	if x > (a+1.0)/(a+b+2.0) {
		return 1.0 - RegularizedIncompleteBeta(1.0-x, b, a)
	}

	lga, _ := math.Lgamma(a)
	lgb, _ := math.Lgamma(b)
	lgab, _ := math.Lgamma(a + b)
	lnBeta := lga + lgb - lgab

	lnPrefactor := a*math.Log(x) + b*math.Log1p(-x) - lnBeta
	prefactor := math.Exp(lnPrefactor) / a

	return prefactor * betaContinuedFraction(a, b, x)
}

func betaContinuedFraction(a, b, x float64) float64 {
	const maxIter = 200
	const fpMin = 1e-30
	const eps = 1e-14

	qab := a + b
	qap := a + 1.0
	qam := a - 1.0

	c := 1.0
	d := 1.0 - qab*x/qap
	if math.Abs(d) < fpMin {
		d = fpMin
	}
	d = 1.0 / d
	h := d

	for m := 1; m <= maxIter; m++ {
		mf := float64(m)
		m2 := 2.0 * mf

		// Even step (2*m)
		aa := mf * (b - mf) * x / ((qam + m2) * (a + m2))
		d = 1.0 + aa*d
		if math.Abs(d) < fpMin {
			d = fpMin
		}
		c = 1.0 + aa/c
		if math.Abs(c) < fpMin {
			c = fpMin
		}
		d = 1.0 / d
		h *= d * c

		// Odd step (2*m + 1)
		aa = -(a + mf) * (qab + mf) * x / ((a + m2) * (qap + m2))
		d = 1.0 + aa*d
		if math.Abs(d) < fpMin {
			d = fpMin
		}
		c = 1.0 + aa/c
		if math.Abs(c) < fpMin {
			c = fpMin
		}
		d = 1.0 / d
		del := d * c
		h *= del

		if math.Abs(del-1.0) < eps {
			break
		}
	}
	return h
}
