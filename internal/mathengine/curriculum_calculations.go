package mathengine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"
)

// InclusiveQuantile uses position (n-1)*p in the sorted sample, with linear
// interpolation, matching Excel PERCENTILE.INC and QUARTILE.INC.
func InclusiveQuantile(data []float64, p float64) (float64, error) {
	if len(data) == 0 || math.IsNaN(p) || p < 0 || p > 1 {
		return 0, fmt.Errorf("inclusive quantile requires data and p in [0,1]")
	}
	if _, err := Mean(data); err != nil {
		return 0, err
	}
	x := append([]float64(nil), data...)
	sort.Float64s(x)
	h := float64(len(x)-1) * p
	i := int(math.Floor(h))
	if i == len(x)-1 {
		return x[i], nil
	}
	return x[i] + (h-float64(i))*(x[i+1]-x[i]), nil
}

// CurriculumCalculations is the rule-v2, fixed-reference calculation contract.
// Only input parameters are accepted. No supplied answers or formula strings
// are evaluated. All returned quantities are canonical engine results.
func CurriculumCalculations(family string, params map[string]interface{}) (map[string]float64, error) {
	op, _ := params["operation"].(string)
	decode := func(dst interface{}) error {
		b, err := json.Marshal(params) // also rejects nonfinite inputs
		if err != nil {
			return err
		}
		d := json.NewDecoder(bytes.NewReader(b))
		d.DisallowUnknownFields()
		return d.Decode(dst)
	}
	result := map[string]float64{}
	switch family + "/" + op {
	case "descriptive_stats/z_score":
		var p struct {
			Operation   string
			X, Mean, SD *float64
		}
		if err := decode(&p); err != nil {
			return nil, err
		}
		if p.X == nil || p.Mean == nil || p.SD == nil {
			return nil, fmt.Errorf("z_score requires x, mean, sd")
		}
		z, err := ZScore(*p.X, *p.Mean, *p.SD)
		if err != nil {
			return nil, err
		}
		result["z"] = z
	case "descriptive_stats/weighted_mean":
		var p struct {
			Operation     string
			Data, Weights []float64
		}
		if err := decode(&p); err != nil {
			return nil, err
		}
		v, err := WeightedMean(p.Data, p.Weights)
		if err != nil {
			return nil, err
		}
		result["mean"] = v
	case "descriptive_stats/quartiles_inc":
		var p struct {
			Operation string
			Data      []float64
		}
		if err := decode(&p); err != nil {
			return nil, err
		}
		q1, err := InclusiveQuantile(p.Data, 0.25)
		if err != nil {
			return nil, err
		}
		q3, err := InclusiveQuantile(p.Data, 0.75)
		if err != nil {
			return nil, err
		}
		iqr := q3 - q1
		result = map[string]float64{"q1": q1, "q3": q3, "iqr": iqr, "lower_fence": q1 - 1.5*iqr, "upper_fence": q3 + 1.5*iqr}
	case "discrete_rv/pmf", "discrete_rv/compare_risk":
		var p struct {
			Operation             string
			Values, Probabilities []float64
			ValuesB               []float64 `json:"values_b"`
			ProbabilitiesB        []float64 `json:"probabilities_b"`
			Lower, Upper, K       *float64
		}
		if err := decode(&p); err != nil {
			return nil, err
		}
		pmf := func(values, probabilities []float64) (*DiscreteRV, error) {
			seen := map[float64]bool{}
			sum := 0.0
			for _, x := range values {
				if seen[x] {
					return nil, fmt.Errorf("duplicate PMF outcome")
				}
				seen[x] = true
			}
			for _, prob := range probabilities {
				if prob < 0 || prob > 1 {
					return nil, fmt.Errorf("PMF probability outside [0,1]")
				}
				sum += prob
			}
			if math.Abs(sum-1) > 1e-10 {
				return nil, fmt.Errorf("PMF probabilities must sum to one")
			}
			return NewDiscreteRV(values, probabilities)
		}
		rv, err := pmf(p.Values, p.Probabilities)
		if err != nil {
			return nil, err
		}
		result["mean"] = rv.ExpectedValue()
		result["variance"] = rv.Variance()
		result["sd"] = rv.StdDev()
		if p.K != nil {
			result["geq"] = rv.ProbabilityGeq(*p.K)
		}
		if (p.Lower == nil) != (p.Upper == nil) {
			return nil, fmt.Errorf("interval requires both lower and upper")
		}
		if p.Lower != nil {
			if *p.Lower > *p.Upper {
				return nil, fmt.Errorf("reversed interval")
			}
			result["between"] = rv.ProbabilityBetween(*p.Lower, *p.Upper)
		}
		if op == "compare_risk" {
			b, err := pmf(p.ValuesB, p.ProbabilitiesB)
			if err != nil {
				return nil, err
			}
			result["mean_b"] = b.ExpectedValue()
			result["sd_b"] = b.StdDev()
		} else if len(p.ValuesB) > 0 || len(p.ProbabilitiesB) > 0 {
			return nil, fmt.Errorf("second PMF requires compare_risk")
		}
	case "linear_combination/iid_sum_vs_scale":
		var p struct {
			Operation string
			N         int
			MeanX     *float64 `json:"mean_x"`
			VarX      *float64 `json:"var_x"`
		}
		if err := decode(&p); err != nil {
			return nil, err
		}
		if p.N < 1 || p.MeanX == nil || p.VarX == nil || *p.VarX < 0 {
			return nil, fmt.Errorf("iid sum requires n>=1, mean_x and var_x>=0")
		}
		n := float64(p.N)
		scaled, err := AffineVariance(n, *p.VarX)
		if err != nil {
			return nil, err
		}
		result = map[string]float64{"mean_sum": AffineMean(n, *p.MeanX, 0), "var_sum": n * (*p.VarX), "var_scaled": scaled, "sd_sum": math.Sqrt(n * (*p.VarX)), "sd_scaled": math.Sqrt(scaled)}
	case "linear_combination/portfolio":
		var p struct {
			Operation string
			A, B, Rho *float64
			MeanX     *float64 `json:"mean_x"`
			MeanY     *float64 `json:"mean_y"`
			VarX      *float64 `json:"var_x"`
			VarY      *float64 `json:"var_y"`
		}
		if err := decode(&p); err != nil {
			return nil, err
		}
		if p.A == nil || p.B == nil || p.Rho == nil || p.MeanX == nil || p.MeanY == nil || p.VarX == nil || p.VarY == nil {
			return nil, fmt.Errorf("portfolio requires weights, means, variances and rho")
		}
		if *p.Rho < -1 || *p.Rho > 1 {
			return nil, fmt.Errorf("rho must be in [-1,1]")
		}
		cov, err := CovarianceFromCorrelation(*p.Rho, *p.VarX, *p.VarY)
		if err != nil {
			return nil, err
		}
		v, err := LinearCombVariance(*p.A, *p.VarX, *p.B, *p.VarY, cov)
		if err != nil {
			return nil, err
		}
		result = map[string]float64{"mean": LinearCombMean(*p.A, *p.MeanX, *p.B, *p.MeanY, 0), "covariance": cov, "variance": v, "sd": math.Sqrt(v)}
	case "two_sample_test/paired", "two_sample_test/pooled", "two_sample_test/proportions":
		var p struct {
			Operation                string
			Before, After            []float64
			N1, N2, X1, X2           int
			Mean1, Mean2, Var1, Var2 *float64
			Alpha                    float64
		}
		if err := decode(&p); err != nil {
			return nil, err
		}
		if p.Alpha <= 0 || p.Alpha >= 1 {
			return nil, fmt.Errorf("alpha must be in (0,1)")
		}
		var stat float64
		var dist *TDistribution
		switch op {
		case "paired":
			if len(p.Before) != len(p.After) || len(p.Before) < 2 {
				return nil, fmt.Errorf("paired arrays must match with n>=2")
			}
			d := make([]float64, len(p.Before))
			for i := range d {
				d[i] = p.After[i] - p.Before[i]
			}
			mean, err := Mean(d)
			if err != nil {
				return nil, err
			}
			sd, err := SampleStdDev(d)
			if err != nil {
				return nil, err
			}
			r, err := ComputePairedTTest(mean, sd, len(d), 0)
			if err != nil {
				return nil, err
			}
			stat = r.TStat
			dist, _ = NewTDistribution(float64(r.DF))
			result = map[string]float64{"mean_diff": mean, "sd_diff": sd, "se": r.SEDiff, "df": float64(r.DF), "t": stat}
		case "pooled":
			if p.Mean1 == nil || p.Mean2 == nil || p.Var1 == nil || p.Var2 == nil {
				return nil, fmt.Errorf("pooled test requires means and variances")
			}
			r, err := ComputeTwoSampleTTestPooled(*p.Mean1, *p.Var1, p.N1, *p.Mean2, *p.Var2, p.N2, 0)
			if err != nil {
				return nil, err
			}
			stat = r.TStat
			dist, _ = NewTDistribution(float64(r.DF))
			result = map[string]float64{"pooled_variance": r.PooledVariance, "se": r.SE, "df": float64(r.DF), "t": stat}
		case "proportions":
			r, err := ComputeTwoSampleProportionTest(p.X1, p.N1, p.X2, p.N2)
			if err != nil {
				return nil, err
			}
			if float64(p.N1)*r.PooledP < 10 || float64(p.N1)*(1-r.PooledP) < 10 || float64(p.N2)*r.PooledP < 10 || float64(p.N2)*(1-r.PooledP) < 10 {
				return nil, fmt.Errorf("pooled expected success/failure counts must be >=10")
			}
			stat = r.ZStat
			norm := StandardNormal()
			result = map[string]float64{"pooled_p": r.PooledP, "se": r.SE, "z": stat, "p_left": norm.CDF(stat), "p_right": norm.Survival(stat), "p_two": 2 * norm.Survival(math.Abs(stat))}
		}
		if dist != nil {
			one, err := dist.Quantile(1 - p.Alpha)
			if err != nil {
				return nil, err
			}
			two, err := dist.Quantile(1 - p.Alpha/2)
			if err != nil {
				return nil, err
			}
			result["p_left"] = dist.CDF(stat)
			result["p_right"] = dist.Survival(stat)
			result["p_two"] = 2 * dist.Survival(math.Abs(stat))
			result["critical_one"] = one
			result["critical_two"] = two
		}
	case "regression_output/model", "regression_output/compare_models":
		var p struct {
			Operation string
			Model     RegressionOutputParams
			ModelB    *RegressionOutputParams `json:"model_b"`
			X         []float64
		}
		if err := decode(&p); err != nil {
			return nil, err
		}
		a, err := EvaluateRegressionOutput(p.Model)
		if err != nil {
			return nil, err
		}
		result = map[string]float64{"r2": a.RSquared, "adjusted_r2": a.AdjRSquared, "se": a.StdErrEst, "f": a.FStat, "p_f": a.SignificanceF}
		for i, c := range a.Predictors {
			result[fmt.Sprintf("b%d", i+1)] = c.Coeff
			result[fmt.Sprintf("p%d", i+1)] = c.PValue
		}
		if len(p.X) > 0 {
			v, err := a.Predict(p.X)
			if err != nil {
				return nil, err
			}
			result["prediction"] = v
		}
		if op == "compare_models" {
			if p.ModelB == nil {
				return nil, fmt.Errorf("comparison requires model_b")
			}
			b, err := EvaluateRegressionOutput(*p.ModelB)
			if err != nil {
				return nil, err
			}
			if a.N != b.N || math.Abs(a.SST-b.SST) > 1e-8 {
				return nil, fmt.Errorf("fit comparison requires same n and total response SS")
			}
			result["r2_b"] = b.RSquared
			result["adjusted_r2_b"] = b.AdjRSquared
			result["se_b"] = b.StdErrEst
		} else if p.ModelB != nil {
			return nil, fmt.Errorf("model_b requires compare_models")
		}
	default:
		return nil, fmt.Errorf("unsupported curriculum calculation %q/%q", family, op)
	}
	for key, value := range result {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("nonfinite result for %s", key)
		}
	}
	return result, nil
}
