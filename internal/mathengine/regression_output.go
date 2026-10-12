package mathengine

import (
	"fmt"
	"math"
)

// OutputPredictorInput represents the raw input for a regression coefficient (intercept or slope).
type OutputPredictorInput struct {
	Name   string  `json:"name"`
	Coeff  float64 `json:"coeff"`
	StdErr float64 `json:"std_err"`
}

// RegressionOutputParams specifies the inputs for constructing an Excel-style regression output model.
type RegressionOutputParams struct {
	N               int                    `json:"n"`
	K               int                    `json:"k"`
	SSR             float64                `json:"ssr"`
	SSE             float64                `json:"sse"`
	SST             *float64               `json:"sst,omitempty"`
	ConfidenceLevel float64                `json:"confidence_level"` // e.g. 0.95 (defaults to 0.95 if <= 0)
	Intercept       OutputPredictorInput   `json:"intercept"`
	Predictors      []OutputPredictorInput `json:"predictors"`
}

// EvaluatedCoefficient represents an evaluated regression coefficient row with t-stat, p-value, and confidence interval.
type EvaluatedCoefficient struct {
	Name      string  `json:"name"`
	Coeff     float64 `json:"coeff"`
	StdErr    float64 `json:"std_err"`
	TStat     float64 `json:"t_stat"`
	PValue    float64 `json:"p_value"`
	LowerCI   float64 `json:"lower_ci"`
	UpperCI   float64 `json:"upper_ci"`
	ExcelPVal string  `json:"excel_p_val"`
	ExcelCI   string  `json:"excel_ci"`
}

// RegressionOutputTable represents a fully evaluated Excel-style regression output summary.
type RegressionOutputTable struct {
	N               int                    `json:"n"`
	K               int                    `json:"k"`
	ConfidenceLevel float64                `json:"confidence_level"`
	Alpha           float64                `json:"alpha"`
	TCritical       float64                `json:"t_critical"` // Two-tailed t* for df_res
	MultipleR       float64                `json:"multiple_r"`
	RSquared        float64                `json:"r_squared"`
	AdjRSquared     float64                `json:"adj_r_squared"`
	StdErrEst       float64                `json:"std_err_est"` // s_e = sqrt(MSE)
	DFReg           int                    `json:"df_reg"`      // k
	DFRes           int                    `json:"df_res"`      // n - k - 1
	DFTotal         int                    `json:"df_total"`    // n - 1
	SSR             float64                `json:"ssr"`
	SSE             float64                `json:"sse"`
	SST             float64                `json:"sst"`
	MSR             float64                `json:"msr"`
	MSE             float64                `json:"mse"`
	FStat           float64                `json:"f_stat"`
	SignificanceF   float64                `json:"significance_f"`
	Intercept       EvaluatedCoefficient   `json:"intercept"`
	Predictors      []EvaluatedCoefficient `json:"predictors"`
	ExcelF          string                 `json:"excel_f"`
	ExcelSigF       string                 `json:"excel_sig_f"`
	ExcelRSq        string                 `json:"excel_rsq"`
	ExcelStdErr     string                 `json:"excel_stderr"`
	AssociationNote string                 `json:"association_note"`
}

// EvaluateRegressionOutput evaluates an Excel-style regression table from given summary parameters.
func EvaluateRegressionOutput(params RegressionOutputParams) (*RegressionOutputTable, error) {
	if params.K < 1 || params.K > 4 {
		return nil, fmt.Errorf("number of predictors k must be between 1 and 4, got %d", params.K)
	}
	if len(params.Predictors) != params.K {
		return nil, fmt.Errorf("number of predictors provided (%d) does not match k=%d", len(params.Predictors), params.K)
	}

	minN := params.K + 3
	if params.N < minN {
		return nil, fmt.Errorf("sample size n must be >= %d for k=%d, got %d", minN, params.K, params.N)
	}

	if params.SSR < 0 {
		return nil, fmt.Errorf("SSR cannot be negative, got %v", params.SSR)
	}
	if params.SSE <= 0 {
		return nil, fmt.Errorf("SSE must be strictly positive, got %v", params.SSE)
	}

	sstCalc := params.SSR + params.SSE
	if params.SST != nil {
		diff := math.Abs(*params.SST - sstCalc)
		allowedTol := math.Max(1e-4, 1e-4*sstCalc)
		if diff > allowedTol {
			return nil, fmt.Errorf("SST (%v) does not equal SSR + SSE (%v + %v = %v, diff: %v)",
				*params.SST, params.SSR, params.SSE, sstCalc, diff)
		}
	}
	sst := sstCalc

	confLevel := params.ConfidenceLevel
	if confLevel <= 0 {
		confLevel = 0.95
	}
	if confLevel >= 1.0 {
		return nil, fmt.Errorf("confidence level must be strictly less than 1.0, got %v", confLevel)
	}
	alpha := 1.0 - confLevel

	dfReg := params.K
	dfRes := params.N - params.K - 1
	dfTotal := params.N - 1

	msr := params.SSR / float64(dfReg)
	mse := params.SSE / float64(dfRes)
	seEst := math.Sqrt(mse)

	fStat := msr / mse
	fDist, err := NewFDistribution(float64(dfReg), float64(dfRes))
	if err != nil {
		return nil, fmt.Errorf("failed to create F-distribution: %w", err)
	}
	sigF := fDist.Survival(fStat)

	rSquared := params.SSR / sst
	multR := math.Sqrt(rSquared)
	// Adjusted R^2 = 1 - (SSE / dfRes) / (SST / dfTotal) = 1 - (MSE / MST)
	mst := sst / float64(dfTotal)
	adjRSquared := 1.0 - (mse / mst)

	tDist, err := NewTDistribution(float64(dfRes))
	if err != nil {
		return nil, fmt.Errorf("failed to create t-distribution: %w", err)
	}
	tCrit, err := tDist.Quantile(1.0 - alpha/2.0)
	if err != nil {
		return nil, fmt.Errorf("failed to compute critical t*: %w", err)
	}

	evalCoeff := func(input OutputPredictorInput) (EvaluatedCoefficient, error) {
		if input.StdErr <= 0 {
			return EvaluatedCoefficient{}, fmt.Errorf("standard error for %q must be strictly positive, got %v", input.Name, input.StdErr)
		}
		tVal := input.Coeff / input.StdErr
		pVal := 2.0 * tDist.Survival(math.Abs(tVal))
		lower := input.Coeff - tCrit*input.StdErr
		upper := input.Coeff + tCrit*input.StdErr

		return EvaluatedCoefficient{
			Name:      input.Name,
			Coeff:     input.Coeff,
			StdErr:    input.StdErr,
			TStat:     tVal,
			PValue:    pVal,
			LowerCI:   lower,
			UpperCI:   upper,
			ExcelPVal: fmt.Sprintf("=T.DIST.2T(ABS(%.4g), %d)", tVal, dfRes),
			ExcelCI:   fmt.Sprintf("%.4g ± T.INV.2T(%.4g, %d) * %.4g", input.Coeff, alpha, dfRes, input.StdErr),
		}, nil
	}

	interceptName := params.Intercept.Name
	if interceptName == "" {
		interceptName = "Intercept"
	}
	params.Intercept.Name = interceptName
	evalIntercept, err := evalCoeff(params.Intercept)
	if err != nil {
		return nil, err
	}

	evalPredictors := make([]EvaluatedCoefficient, len(params.Predictors))
	for i, pred := range params.Predictors {
		if pred.Name == "" {
			pred.Name = fmt.Sprintf("X%d", i+1)
		}
		evalPred, err := evalCoeff(pred)
		if err != nil {
			return nil, err
		}
		evalPredictors[i] = evalPred
	}

	// Consistency check when k = 1: t1^2 should equal FStat
	if params.K == 1 {
		t1 := evalPredictors[0].TStat
		t1Sq := t1 * t1
		diff := math.Abs(t1Sq - fStat)
		// Relative tolerance comparison
		allowedDiff := math.Max(0.05, 0.01*fStat)
		if diff > allowedDiff {
			return nil, fmt.Errorf("consistency check failed for k=1: t_slope^2 (%v) does not match F (%v)", t1Sq, fStat)
		}
	}

	assocNote := "Association is not causation: a significant regression equation indicates linear correlation, not a direct causal mechanism. Unobserved confounding variables may drive both predictors and response."

	return &RegressionOutputTable{
		N:               params.N,
		K:               params.K,
		ConfidenceLevel: confLevel,
		Alpha:           alpha,
		TCritical:       tCrit,
		MultipleR:       multR,
		RSquared:        rSquared,
		AdjRSquared:     adjRSquared,
		StdErrEst:       seEst,
		DFReg:           dfReg,
		DFRes:           dfRes,
		DFTotal:         dfTotal,
		SSR:             params.SSR,
		SSE:             params.SSE,
		SST:             sst,
		MSR:             msr,
		MSE:             mse,
		FStat:           fStat,
		SignificanceF:   sigF,
		Intercept:       evalIntercept,
		Predictors:      evalPredictors,
		ExcelF:          fmt.Sprintf("=MSR/MSE (%.4g / %.4g)", msr, mse),
		ExcelSigF:       fmt.Sprintf("=F.DIST.RT(%.4g, %d, %d)", fStat, dfReg, dfRes),
		ExcelRSq:        "=SSR/SST",
		ExcelStdErr:     "=SQRT(MSE)",
		AssociationNote: assocNote,
	}, nil
}

// Predict calculates yHat = b0 + sum(b_j * x_j) for the given predictor values.
func (t *RegressionOutputTable) Predict(xValues []float64) (float64, error) {
	if len(xValues) != t.K {
		return 0, fmt.Errorf("expected %d predictor values for k=%d, got %d", t.K, t.K, len(xValues))
	}
	yHat := t.Intercept.Coeff
	for j, x := range xValues {
		yHat += t.Predictors[j].Coeff * x
	}
	return yHat, nil
}
