package mathengine

import (
	"fmt"
	"math"
)

// DataPoint represents a paired bivariate observation (x, y).
type DataPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ResidualPoint represents an observation paired with its fitted value and residual.
type ResidualPoint struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	YHat     float64 `json:"y_hat"`
	Residual float64 `json:"residual"`
}

// LinearRegressionResult contains the complete OLS regression model and diagnostic statistics.
type LinearRegressionResult struct {
	N             int             `json:"n"`
	MeanX         float64         `json:"mean_x"`
	MeanY         float64         `json:"mean_y"`
	SSxx          float64         `json:"ss_xx"`
	SSyy          float64         `json:"ss_yy"`
	SSxy          float64         `json:"ss_xy"`
	Slope         float64         `json:"slope"`         // b1
	Intercept     float64         `json:"intercept"`     // b0
	Correlation   float64         `json:"correlation"`   // r
	RSquared      float64         `json:"r_squared"`     // R^2
	SST           float64         `json:"sst"`           // Total sum of squares
	SSR           float64         `json:"ssr"`           // Regression sum of squares
	SSE           float64         `json:"sse"`           // Error sum of squares
	StdErrEst     float64         `json:"std_err_est"`   // s_e = sqrt(SSE / (n-2))
	StdErrSlope   float64         `json:"std_err_slope"` // s_{b1} = s_e / sqrt(SSxx)
	TSlope        float64         `json:"t_slope"`       // t = b1 / s_{b1}
	PSlope        float64         `json:"p_slope"`       // Two-tailed p-value for slope
	DegreesOfFree int             `json:"degrees_of_freedom"`
	Residuals     []ResidualPoint `json:"residuals"`
	Equation      string          `json:"equation"`
	Association   string          `json:"association_note"`
	ExcelSlope    string          `json:"excel_slope"`
	ExcelInt      string          `json:"excel_intercept"`
	ExcelCorrel   string          `json:"excel_correl"`
	ExcelRSq      string          `json:"excel_rsq"`
}

// ComputeLinearRegression computes Ordinary Least Squares (OLS) regression for paired data points.
func ComputeLinearRegression(points []DataPoint) (*LinearRegressionResult, error) {
	n := len(points)
	if n < 3 {
		return nil, fmt.Errorf("linear regression requires at least 3 points, got %d", n)
	}

	sumX := 0.0
	sumY := 0.0
	for _, p := range points {
		if math.IsNaN(p.X) || math.IsInf(p.X, 0) || math.IsNaN(p.Y) || math.IsInf(p.Y, 0) {
			return nil, fmt.Errorf("data points must be finite numbers")
		}
		sumX += p.X
		sumY += p.Y
	}
	meanX := sumX / float64(n)
	meanY := sumY / float64(n)

	ssXX := 0.0
	ssYY := 0.0
	ssXY := 0.0

	for _, p := range points {
		dx := p.X - meanX
		dy := p.Y - meanY
		ssXX += dx * dx
		ssYY += dy * dy
		ssXY += dx * dy
	}

	if ssXX <= 1e-12 {
		return nil, fmt.Errorf("variance of X is zero or near zero (all X values are identical); cannot fit slope")
	}

	slope := ssXY / ssXX
	intercept := meanY - slope*meanX

	var r float64
	if ssYY > 1e-12 {
		r = ssXY / math.Sqrt(ssXX*ssYY)
		// Clamp numerical float rounding to [-1, 1]
		if r > 1.0 {
			r = 1.0
		} else if r < -1.0 {
			r = -1.0
		}
	} else {
		r = 0.0
	}

	ssr := slope * ssXY
	if ssr < 0 {
		ssr = 0
	}
	sst := ssYY
	sse := sst - ssr
	if sse < 0 {
		sse = 0
	}

	var rSquared float64
	if sst > 1e-12 {
		rSquared = ssr / sst
	} else {
		rSquared = 1.0
	}

	df := n - 2
	seEst := math.Sqrt(sse / float64(df))
	seSlope := seEst / math.Sqrt(ssXX)

	var tSlope float64
	var pSlope float64
	if seSlope > 1e-12 {
		tSlope = slope / seSlope
		tDist, err := NewTDistribution(float64(df))
		if err == nil {
			pSlope = 2.0 * tDist.Survival(math.Abs(tSlope))
		}
	} else {
		tSlope = 0.0
		pSlope = 1.0
	}

	residuals := make([]ResidualPoint, n)
	for i, p := range points {
		yHat := intercept + slope*p.X
		residuals[i] = ResidualPoint{
			X:        p.X,
			Y:        p.Y,
			YHat:     math.Round(yHat*10000) / 10000,
			Residual: math.Round((p.Y-yHat)*10000) / 10000,
		}
	}

	eqSign := "+"
	absInt := intercept
	if intercept < 0 {
		eqSign = "-"
		absInt = -intercept
	}
	equation := fmt.Sprintf("y_hat = %.4g*x %s %.4g", slope, eqSign, absInt)

	assocNote := "Association is not causation: a high correlation or R^2 indicates a linear relationship, not that X causes Y. Confounding variables or reverse causation may exist."

	return &LinearRegressionResult{
		N:             n,
		MeanX:         meanX,
		MeanY:         meanY,
		SSxx:          ssXX,
		SSyy:          ssYY,
		SSxy:          ssXY,
		Slope:         slope,
		Intercept:     intercept,
		Correlation:   r,
		RSquared:      rSquared,
		SST:           sst,
		SSR:           ssr,
		SSE:           sse,
		StdErrEst:     seEst,
		StdErrSlope:   seSlope,
		TSlope:        tSlope,
		PSlope:        pSlope,
		DegreesOfFree: df,
		Residuals:     residuals,
		Equation:      equation,
		Association:   assocNote,
		ExcelSlope:    "=SLOPE(known_y's, known_x's)",
		ExcelInt:      "=INTERCEPT(known_y's, known_x's)",
		ExcelCorrel:   "=CORREL(known_x's, known_y's)",
		ExcelRSq:      "=RSQ(known_y's, known_x's)",
	}, nil
}

// Predict calculates the point prediction yHat and standard error of prediction at xNew.
func (reg *LinearRegressionResult) Predict(xNew float64) (float64, float64) {
	yHat := reg.Intercept + reg.Slope*xNew
	dx := xNew - reg.MeanX
	// Standard error of prediction for an individual new value:
	// s_pred = s_e * sqrt(1 + 1/n + (xNew - meanX)^2 / SSxx)
	sePred := reg.StdErrEst * math.Sqrt(1.0+1.0/float64(reg.N)+dx*dx/reg.SSxx)
	return yHat, sePred
}
