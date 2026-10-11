package httpapi

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

type CurvePoint struct {
	X   float64 `json:"x"`
	PDF float64 `json:"pdf"`
	CDF float64 `json:"cdf"`
}

type NormalVisualResponse struct {
	Family            string             `json:"family"`
	Mu                float64            `json:"mu"`
	Sigma             float64            `json:"sigma"`
	Tail              string             `json:"tail"`
	CutoffX           *float64           `json:"cutoff_x,omitempty"`
	CutoffX1          *float64           `json:"cutoff_x1,omitempty"`
	CutoffX2          *float64           `json:"cutoff_x2,omitempty"`
	ZScore            *float64           `json:"z_score,omitempty"`
	ZScore1           *float64           `json:"z_score_1,omitempty"`
	ZScore2           *float64           `json:"z_score_2,omitempty"`
	ShadedProbability float64            `json:"shaded_probability"`
	DensityAtCutoff   *float64           `json:"density_at_cutoff,omitempty"`
	ExcelFormula      string             `json:"excel_formula"`
	Summary           string             `json:"summary"`
	EmpiricalRule     map[string]float64 `json:"empirical_rule"`
	Points            []CurvePoint       `json:"points"`
}

type UniformVisualResponse struct {
	Family            string       `json:"family"`
	A                 float64      `json:"a"`
	B                 float64      `json:"b"`
	Tail              string       `json:"tail"`
	CutoffX           *float64     `json:"cutoff_x,omitempty"`
	CutoffX1          *float64     `json:"cutoff_x1,omitempty"`
	CutoffX2          *float64     `json:"cutoff_x2,omitempty"`
	HeightDensity     float64      `json:"height_density"`
	ShadedProbability float64      `json:"shaded_probability"`
	Mean              float64      `json:"mean"`
	Variance          float64      `json:"variance"`
	Summary           string       `json:"summary"`
	Points            []CurvePoint `json:"points"`
}

type ExponentialVisualResponse struct {
	Family            string       `json:"family"`
	Lambda            float64      `json:"lambda"`
	Mean              float64      `json:"mean"`
	Tail              string       `json:"tail"`
	CutoffX           *float64     `json:"cutoff_x,omitempty"`
	CutoffX1          *float64     `json:"cutoff_x1,omitempty"`
	CutoffX2          *float64     `json:"cutoff_x2,omitempty"`
	ShadedProbability float64      `json:"shaded_probability"`
	DensityAtCutoff   *float64     `json:"density_at_cutoff,omitempty"`
	ExcelFormula      string       `json:"excel_formula"`
	MemorylessNote    string       `json:"memoryless_note"`
	Summary           string       `json:"summary"`
	Points            []CurvePoint `json:"points"`
}

type DiscreteBar struct {
	K             int     `json:"k"`
	PMF           float64 `json:"pmf"`
	CDF           float64 `json:"cdf"`
	IsHighlighted bool    `json:"is_highlighted"`
}

type DiscreteVisualResponse struct {
	Family           string         `json:"family"`
	Parameters       map[string]any `json:"parameters"`
	Mean             float64        `json:"mean"`
	Variance         float64        `json:"variance"`
	Event            string         `json:"event"`
	TotalProbability float64        `json:"total_probability"`
	ExcelFormula     string         `json:"excel_formula"`
	Summary          string         `json:"summary"`
	Bars             []DiscreteBar  `json:"bars"`
}

type VennVisualResponse struct {
	PA                float64            `json:"p_a"`
	PB                float64            `json:"p_b"`
	PIntersection     float64            `json:"p_intersection"`
	Region            string             `json:"region"`
	Regions           map[string]float64 `json:"regions"`
	ShadedProbability float64            `json:"shaded_probability"`
	Formula           string             `json:"formula"`
	Summary           string             `json:"summary"`
}

type SamplingBar struct {
	BinMid             float64 `json:"bin_mid"`
	Count              int     `json:"count"`
	Density            float64 `json:"density"`
	TheoreticalDensity float64 `json:"theoretical_density"`
}

type SamplingVisualResponse struct {
	Family          string        `json:"family"`
	PopDistribution string        `json:"pop_distribution"`
	PopMean         float64       `json:"pop_mean"`
	PopStdDev       float64       `json:"pop_std_dev"`
	SampleSize      int           `json:"sample_size"`
	Trials          int           `json:"trials"`
	StandardError   float64       `json:"standard_error"`
	EmpiricalMean   float64       `json:"empirical_mean"`
	EmpiricalStdErr float64       `json:"empirical_std_err"`
	Bars            []SamplingBar `json:"bars"`
	NormalCurve     []CurvePoint  `json:"normal_curve"`
	CLTNote         string        `json:"clt_note"`
	ExcelFormula    string        `json:"excel_formula"`
	Summary         string        `json:"summary"`
}

type ConfidenceIntervalVisualResponse struct {
	Family            string                         `json:"family"`
	ConfidenceLevel   float64                        `json:"confidence_level"`
	Alpha             float64                        `json:"alpha"`
	TrueMean          float64                        `json:"true_mean"`
	Sigma             float64                        `json:"sigma"`
	SampleSize        int                            `json:"sample_size"`
	NumIntervals      int                            `json:"num_intervals"`
	Intervals         []mathengine.SimulatedInterval `json:"intervals"`
	CapturedCount     int                            `json:"captured_count"`
	EmpiricalCoverage float64                        `json:"empirical_coverage"`
	ExcelFormula      string                         `json:"excel_formula"`
	Summary           string                         `json:"summary"`
}

type HypothesisVisualResponse struct {
	Family          string       `json:"family"`
	TestType        string       `json:"test_type"`
	Tail            string       `json:"tail"`
	NullHypothesis  string       `json:"null_hypothesis"`
	AltHypothesis   string       `json:"alt_hypothesis"`
	NullValue       float64      `json:"null_value"`
	SampleStat      float64      `json:"sample_stat"`
	SampleSize      int          `json:"sample_size"`
	StandardError   float64      `json:"standard_error"`
	TestStatistic   float64      `json:"test_statistic"`
	DegreesOfFree   *float64     `json:"degrees_of_freedom,omitempty"`
	PValue          float64      `json:"p_value"`
	Alpha           float64      `json:"alpha"`
	CriticalValue   float64      `json:"critical_value"`
	CriticalValue2  *float64     `json:"critical_value_2,omitempty"`
	RejectNull      bool         `json:"reject_null"`
	Decision        string       `json:"decision"`
	RejectionRegion string       `json:"rejection_region"`
	ExcelFormula    string       `json:"excel_formula"`
	Summary         string       `json:"summary"`
	Points          []CurvePoint `json:"points"`
}

type RegressionVisualResponse struct {
	Family          string                     `json:"family"`
	DatasetName     string                     `json:"dataset_name"`
	N               int                        `json:"n"`
	Points          []mathengine.DataPoint     `json:"points"`
	Slope           float64                    `json:"slope"`
	Intercept       float64                    `json:"intercept"`
	Correlation     float64                    `json:"correlation"`
	RSquared        float64                    `json:"r_squared"`
	StdErrEst       float64                    `json:"std_err_est"`
	Equation        string                     `json:"equation"`
	AssociationNote string                     `json:"association_note"`
	ExcelFormula    string                     `json:"excel_formula"`
	Residuals       []mathengine.ResidualPoint `json:"residuals"`
	LineStartX      float64                    `json:"line_start_x"`
	LineStartY      float64                    `json:"line_start_y"`
	LineEndX        float64                    `json:"line_end_x"`
	LineEndY        float64                    `json:"line_end_y"`
	PredictedX      *float64                   `json:"predicted_x,omitempty"`
	PredictedY      *float64                   `json:"predicted_y,omitempty"`
	Summary         string                     `json:"summary"`
}

func (s *Server) handleVisuals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/visuals")
	path = strings.TrimPrefix(path, "/")

	switch path {
	case "normal":
		s.handleVisualsNormal(w, r)
	case "uniform":
		s.handleVisualsUniform(w, r)
	case "exponential":
		s.handleVisualsExponential(w, r)
	case "discrete":
		s.handleVisualsDiscrete(w, r)
	case "venn":
		s.handleVisualsVenn(w, r)
	case "sampling":
		s.handleVisualsSampling(w, r)
	case "confidence_interval":
		s.handleVisualsCI(w, r)
	case "hypothesis_test":
		s.handleVisualsHypothesis(w, r)
	case "regression":
		s.handleVisualsRegression(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleVisualsNormal(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	mu := parseQueryFloat(q.Get("mu"), 0.0)
	sigma := parseQueryFloat(q.Get("sigma"), 1.0)
	if sigma <= 0 {
		http.Error(w, "sigma must be strictly positive", http.StatusBadRequest)
		return
	}

	norm, err := mathengine.NewNormalDistribution(mu, sigma)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tail := q.Get("tail")
	if tail == "" {
		tail = "lower"
	}

	numSamples := int(parseQueryFloat(q.Get("samples"), 120))
	if numSamples < 30 {
		numSamples = 30
	} else if numSamples > 300 {
		numSamples = 300
	}

	xMin := mu - 3.8*sigma
	xMax := mu + 3.8*sigma
	step := (xMax - xMin) / float64(numSamples-1)

	points := make([]CurvePoint, numSamples)
	for i := 0; i < numSamples; i++ {
		x := xMin + float64(i)*step
		points[i] = CurvePoint{
			X:   math.Round(x*10000) / 10000,
			PDF: norm.PDF(x),
			CDF: norm.CDF(x),
		}
	}

	resp := NormalVisualResponse{
		Family: "normal",
		Mu:     mu,
		Sigma:  sigma,
		Tail:   tail,
		EmpiricalRule: map[string]float64{
			"one_sigma":   0.6827,
			"two_sigma":   0.9545,
			"three_sigma": 0.9973,
		},
		Points: points,
	}

	switch tail {
	case "upper":
		x := parseQueryFloat(q.Get("x"), mu+sigma)
		resp.CutoffX = &x
		z := norm.ZScore(x)
		resp.ZScore = &z
		prob := norm.Survival(x)
		resp.ShadedProbability = prob
		d := norm.PDF(x)
		resp.DensityAtCutoff = &d
		if mu == 0 && sigma == 1 {
			resp.ExcelFormula = fmt.Sprintf("=1 - NORM.S.DIST(%.4g, TRUE)", x)
		} else {
			resp.ExcelFormula = fmt.Sprintf("=1 - NORM.DIST(%.4g, %.4g, %.4g, TRUE)", x, mu, sigma)
		}
		resp.Summary = fmt.Sprintf("Normal N(mu=%.4g, sigma=%.4g): Upper tail P(X >= %.4g) = %.4g (Z = %.2f). Curve height f(%.4g) = %.4g represents probability density.", mu, sigma, x, prob, z, x, d)

	case "interval":
		x1 := parseQueryFloat(q.Get("x1"), mu-sigma)
		x2 := parseQueryFloat(q.Get("x2"), mu+sigma)
		if x1 > x2 {
			x1, x2 = x2, x1
		}
		resp.CutoffX1 = &x1
		resp.CutoffX2 = &x2
		z1 := norm.ZScore(x1)
		z2 := norm.ZScore(x2)
		resp.ZScore1 = &z1
		resp.ZScore2 = &z2
		prob := norm.CDF(x2) - norm.CDF(x1)
		resp.ShadedProbability = prob
		if mu == 0 && sigma == 1 {
			resp.ExcelFormula = fmt.Sprintf("=NORM.S.DIST(%.4g, TRUE) - NORM.S.DIST(%.4g, TRUE)", x2, x1)
		} else {
			resp.ExcelFormula = fmt.Sprintf("=NORM.DIST(%.4g, %.4g, %.4g, TRUE) - NORM.DIST(%.4g, %.4g, %.4g, TRUE)", x2, mu, sigma, x1, mu, sigma)
		}
		resp.Summary = fmt.Sprintf("Normal N(mu=%.4g, sigma=%.4g): Interval P(%.4g <= X <= %.4g) = %.4g (Z1 = %.2f, Z2 = %.2f).", mu, sigma, x1, x2, prob, z1, z2)

	case "two_tailed":
		crit := math.Abs(parseQueryFloat(q.Get("x"), 1.96*sigma))
		x1 := mu - crit
		x2 := mu + crit
		resp.CutoffX1 = &x1
		resp.CutoffX2 = &x2
		z1 := norm.ZScore(x1)
		z2 := norm.ZScore(x2)
		resp.ZScore1 = &z1
		resp.ZScore2 = &z2
		prob := norm.CDF(x1) + norm.Survival(x2)
		resp.ShadedProbability = prob
		resp.ExcelFormula = fmt.Sprintf("=2 * (1 - NORM.DIST(%.4g, %.4g, %.4g, TRUE))", x2, mu, sigma)
		resp.Summary = fmt.Sprintf("Normal N(mu=%.4g, sigma=%.4g): Two-tailed region (|Z| >= %.2f) = %.4g.", mu, sigma, math.Abs(z2), prob)

	default: // "lower"
		x := parseQueryFloat(q.Get("x"), mu)
		resp.CutoffX = &x
		z := norm.ZScore(x)
		resp.ZScore = &z
		prob := norm.CDF(x)
		resp.ShadedProbability = prob
		d := norm.PDF(x)
		resp.DensityAtCutoff = &d
		if mu == 0 && sigma == 1 {
			resp.ExcelFormula = fmt.Sprintf("=NORM.S.DIST(%.4g, TRUE)", x)
		} else {
			resp.ExcelFormula = fmt.Sprintf("=NORM.DIST(%.4g, %.4g, %.4g, TRUE)", x, mu, sigma)
		}
		resp.Summary = fmt.Sprintf("Normal N(mu=%.4g, sigma=%.4g): Lower tail P(X <= %.4g) = %.4g (Z = %.2f). Curve height f(%.4g) = %.4g represents probability density.", mu, sigma, x, prob, z, x, d)
	}

	writeJSON(w, resp)
}

func (s *Server) handleVisualsUniform(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	a := parseQueryFloat(q.Get("a"), 0.0)
	b := parseQueryFloat(q.Get("b"), 10.0)
	if a >= b {
		http.Error(w, "lower bound a must be strictly less than upper bound b", http.StatusBadRequest)
		return
	}

	dist, err := mathengine.NewUniformDistribution(a, b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tail := q.Get("tail")
	if tail == "" {
		tail = "interval"
	}

	pad := (b - a) * 0.2
	xMin := a - pad
	xMax := b + pad
	numSamples := 100
	step := (xMax - xMin) / float64(numSamples-1)

	points := make([]CurvePoint, numSamples)
	for i := 0; i < numSamples; i++ {
		x := xMin + float64(i)*step
		points[i] = CurvePoint{
			X:   math.Round(x*1000) / 1000,
			PDF: dist.PDF(x),
			CDF: dist.CDF(x),
		}
	}

	resp := UniformVisualResponse{
		Family:        "uniform",
		A:             a,
		B:             b,
		Tail:          tail,
		HeightDensity: dist.PDF(0.5 * (a + b)),
		Mean:          dist.Mean(),
		Variance:      dist.Variance(),
		Points:        points,
	}

	switch tail {
	case "lower":
		x := parseQueryFloat(q.Get("x"), 0.5*(a+b))
		resp.CutoffX = &x
		prob := dist.CDF(x)
		resp.ShadedProbability = prob
		resp.Summary = fmt.Sprintf("Uniform U(%.4g, %.4g): Lower tail P(X <= %.4g) = %.4g. Constant height f(x) = %.4g.", a, b, x, prob, resp.HeightDensity)
	case "upper":
		x := parseQueryFloat(q.Get("x"), 0.5*(a+b))
		resp.CutoffX = &x
		prob := dist.Survival(x)
		resp.ShadedProbability = prob
		resp.Summary = fmt.Sprintf("Uniform U(%.4g, %.4g): Upper tail P(X >= %.4g) = %.4g.", a, b, x, prob)
	default:
		x1 := parseQueryFloat(q.Get("x1"), a+0.25*(b-a))
		x2 := parseQueryFloat(q.Get("x2"), a+0.75*(b-a))
		if x1 > x2 {
			x1, x2 = x2, x1
		}
		resp.CutoffX1 = &x1
		resp.CutoffX2 = &x2
		prob, _ := dist.Interval(x1, x2)
		resp.ShadedProbability = prob
		resp.Summary = fmt.Sprintf("Uniform U(%.4g, %.4g): Interval P(%.4g <= X <= %.4g) = %.4g = (%.4g - %.4g)/(%.4g - %.4g).", a, b, x1, x2, prob, math.Min(x2, b), math.Max(x1, a), b, a)
	}

	writeJSON(w, resp)
}

func (s *Server) handleVisualsExponential(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	lambda := parseQueryFloat(q.Get("lambda"), 1.0)
	if lambda <= 0 {
		http.Error(w, "rate lambda must be strictly positive", http.StatusBadRequest)
		return
	}

	dist, err := mathengine.NewExponentialDistribution(lambda)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tail := q.Get("tail")
	if tail == "" {
		tail = "upper"
	}

	xMax := 4.5 / lambda
	numSamples := 100
	step := xMax / float64(numSamples-1)

	points := make([]CurvePoint, numSamples)
	for i := 0; i < numSamples; i++ {
		x := float64(i) * step
		points[i] = CurvePoint{
			X:   math.Round(x*1000) / 1000,
			PDF: dist.PDF(x),
			CDF: dist.CDF(x),
		}
	}

	resp := ExponentialVisualResponse{
		Family:         "exponential",
		Lambda:         lambda,
		Mean:           dist.Mean(),
		Tail:           tail,
		MemorylessNote: "Memoryless property: P(X > s + t | X > s) = P(X > t) = exp(-lambda * t). Future lifetime does not depend on elapsed time.",
		Points:         points,
	}

	switch tail {
	case "lower":
		x := parseQueryFloat(q.Get("x"), dist.Mean())
		resp.CutoffX = &x
		prob := dist.CDF(x)
		d := dist.PDF(x)
		resp.ShadedProbability = prob
		resp.DensityAtCutoff = &d
		resp.ExcelFormula = fmt.Sprintf("=EXPON.DIST(%.4g, %.4g, TRUE)", x, lambda)
		resp.Summary = fmt.Sprintf("Exponential (lambda=%.4g, mean=%.4g): Cumulative P(X <= %.4g) = 1 - e^(-%.4g * %.4g) = %.4g.", lambda, dist.Mean(), x, lambda, x, prob)
	case "interval":
		x1 := parseQueryFloat(q.Get("x1"), 0.5*dist.Mean())
		x2 := parseQueryFloat(q.Get("x2"), 1.5*dist.Mean())
		if x1 > x2 {
			x1, x2 = x2, x1
		}
		resp.CutoffX1 = &x1
		resp.CutoffX2 = &x2
		prob, _ := dist.Interval(x1, x2)
		resp.ShadedProbability = prob
		resp.ExcelFormula = fmt.Sprintf("=EXPON.DIST(%.4g, %.4g, TRUE) - EXPON.DIST(%.4g, %.4g, TRUE)", x2, lambda, x1, lambda)
		resp.Summary = fmt.Sprintf("Exponential (lambda=%.4g): Interval P(%.4g <= X <= %.4g) = %.4g.", lambda, x1, x2, prob)
	default: // upper
		x := parseQueryFloat(q.Get("x"), dist.Mean())
		resp.CutoffX = &x
		prob := dist.Survival(x)
		d := dist.PDF(x)
		resp.ShadedProbability = prob
		resp.DensityAtCutoff = &d
		resp.ExcelFormula = fmt.Sprintf("=1 - EXPON.DIST(%.4g, %.4g, TRUE)", x, lambda)
		resp.Summary = fmt.Sprintf("Exponential (lambda=%.4g): Survival P(X >= %.4g) = e^(-%.4g * %.4g) = %.4g.", lambda, x, lambda, x, prob)
	}

	writeJSON(w, resp)
}

func (s *Server) handleVisualsDiscrete(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	family := q.Get("family")
	if family == "" {
		family = "binomial"
	}
	event := q.Get("event")
	if event == "" {
		event = "equal"
	}

	switch family {
	case "binomial":
		n := int(parseQueryFloat(q.Get("n"), 10))
		p := parseQueryFloat(q.Get("p"), 0.5)
		if n < 0 || n > 100 || p < 0 || p > 1 {
			http.Error(w, "invalid binomial parameters (n in [0, 100], p in [0, 1])", http.StatusBadRequest)
			return
		}
		binom, err := mathengine.NewBinomial(n, p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		k := int(parseQueryFloat(q.Get("k"), float64(n)/2.0))
		k1 := int(parseQueryFloat(q.Get("k1"), 0))
		k2 := int(parseQueryFloat(q.Get("k2"), float64(n)))

		bars := make([]DiscreteBar, n+1)
		var totalProb float64

		for i := 0; i <= n; i++ {
			pmf := binom.PMF(i)
			cdf := binom.CDF(i)
			isHigh := isDiscreteHighlighted(i, event, k, k1, k2)
			if isHigh {
				totalProb += pmf
			}
			bars[i] = DiscreteBar{
				K:             i,
				PMF:           pmf,
				CDF:           cdf,
				IsHighlighted: isHigh,
			}
		}

		var excelFormula string
		switch event {
		case "equal":
			excelFormula = fmt.Sprintf("=BINOM.DIST(%d, %d, %.4g, FALSE)", k, n, p)
		case "at_most":
			excelFormula = fmt.Sprintf("=BINOM.DIST(%d, %d, %.4g, TRUE)", k, n, p)
		case "at_least":
			excelFormula = fmt.Sprintf("=1 - BINOM.DIST(%d, %d, %.4g, TRUE)", k-1, n, p)
		case "between":
			excelFormula = fmt.Sprintf("=BINOM.DIST(%d, %d, %.4g, TRUE) - BINOM.DIST(%d, %d, %.4g, TRUE)", k2, n, p, k1-1, n, p)
		}

		writeJSON(w, DiscreteVisualResponse{
			Family: "binomial",
			Parameters: map[string]any{
				"n": n,
				"p": p,
			},
			Mean:             binom.Mean(),
			Variance:         binom.Variance(),
			Event:            event,
			TotalProbability: totalProb,
			ExcelFormula:     excelFormula,
			Summary:          fmt.Sprintf("Binomial(n=%d, p=%.4g): Event %s has probability %.4g. Mean = %.2f, Variance = %.2f.", n, p, event, totalProb, binom.Mean(), binom.Variance()),
			Bars:             bars,
		})

	case "poisson":
		lambda := parseQueryFloat(q.Get("lambda"), 4.0)
		if lambda < 0 || lambda > 50 {
			http.Error(w, "invalid poisson lambda in [0, 50]", http.StatusBadRequest)
			return
		}
		poiss, err := mathengine.NewPoisson(lambda)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		k := int(parseQueryFloat(q.Get("k"), lambda))
		k1 := int(parseQueryFloat(q.Get("k1"), 0))
		k2 := int(parseQueryFloat(q.Get("k2"), lambda*2))

		maxK := int(math.Max(15, lambda+4*math.Sqrt(math.Max(1, lambda))))
		if maxK > 60 {
			maxK = 60
		}

		bars := make([]DiscreteBar, maxK+1)
		var totalProb float64

		for i := 0; i <= maxK; i++ {
			pmf := poiss.PMF(i)
			cdf := poiss.CDF(i)
			isHigh := isDiscreteHighlighted(i, event, k, k1, k2)
			if isHigh {
				totalProb += pmf
			}
			bars[i] = DiscreteBar{
				K:             i,
				PMF:           pmf,
				CDF:           cdf,
				IsHighlighted: isHigh,
			}
		}

		var excelFormula string
		switch event {
		case "equal":
			excelFormula = fmt.Sprintf("=POISSON.DIST(%d, %.4g, FALSE)", k, lambda)
		case "at_most":
			excelFormula = fmt.Sprintf("=POISSON.DIST(%d, %.4g, TRUE)", k, lambda)
		case "at_least":
			excelFormula = fmt.Sprintf("=1 - POISSON.DIST(%d, %.4g, TRUE)", k-1, lambda)
		default:
			excelFormula = fmt.Sprintf("=POISSON.DIST(%d, %.4g, TRUE) - POISSON.DIST(%d, %.4g, TRUE)", k2, lambda, k1-1, lambda)
		}

		writeJSON(w, DiscreteVisualResponse{
			Family: "poisson",
			Parameters: map[string]any{
				"lambda": lambda,
			},
			Mean:             poiss.Mean(),
			Variance:         poiss.Variance(),
			Event:            event,
			TotalProbability: totalProb,
			ExcelFormula:     excelFormula,
			Summary:          fmt.Sprintf("Poisson(lambda=%.4g): Event %s has probability %.4g. Mean = Variance = %.2f.", lambda, event, totalProb, poiss.Mean()),
			Bars:             bars,
		})

	default:
		http.Error(w, "unknown discrete family", http.StatusBadRequest)
	}
}

func isDiscreteHighlighted(i int, event string, k, k1, k2 int) bool {
	switch event {
	case "equal":
		return i == k
	case "at_most":
		return i <= k
	case "at_least":
		return i >= k
	case "between":
		return i >= k1 && i <= k2
	default:
		return i == k
	}
}

func (s *Server) handleVisualsVenn(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	pA := parseQueryFloat(q.Get("p_a"), 0.5)
	pB := parseQueryFloat(q.Get("p_b"), 0.4)
	pInter := parseQueryFloat(q.Get("p_intersection"), 0.2)

	if err := mathengine.ValidateSetProbabilities(pA, pB, pInter); err != nil {
		http.Error(w, fmt.Sprintf("invalid set probabilities: %v", err), http.StatusBadRequest)
		return
	}

	region := q.Get("region")
	if region == "" {
		region = "intersection"
	}

	pAOnly := pA - pInter
	pBOnly := pB - pInter
	pUnion := pA + pB - pInter
	pNeither := 1.0 - pUnion

	regions := map[string]float64{
		"a_only":       math.Round(pAOnly*10000) / 10000,
		"b_only":       math.Round(pBOnly*10000) / 10000,
		"intersection": math.Round(pInter*10000) / 10000,
		"neither":      math.Round(pNeither*10000) / 10000,
		"union":        math.Round(pUnion*10000) / 10000,
		"complement_a": math.Round((1.0-pA)*10000) / 10000,
		"complement_b": math.Round((1.0-pB)*10000) / 10000,
	}

	var shadedProb float64
	var formula string
	switch region {
	case "a_only":
		shadedProb = pAOnly
		formula = "P(A \\setminus B) = P(A) - P(A \\cap B)"
	case "b_only":
		shadedProb = pBOnly
		formula = "P(B \\setminus A) = P(B) - P(A \\cap B)"
	case "union":
		shadedProb = pUnion
		formula = "P(A \\cup B) = P(A) + P(B) - P(A \\cap B)"
	case "neither":
		shadedProb = pNeither
		formula = "P((A \\cup B)^c) = 1 - P(A \\cup B)"
	case "complement_a":
		shadedProb = 1.0 - pA
		formula = "P(A^c) = 1 - P(A)"
	case "complement_b":
		shadedProb = 1.0 - pB
		formula = "P(B^c) = 1 - P(B)"
	default: // intersection
		region = "intersection"
		shadedProb = pInter
		formula = "P(A \\cap B)"
	}

	writeJSON(w, VennVisualResponse{
		PA:                pA,
		PB:                pB,
		PIntersection:     pInter,
		Region:            region,
		Regions:           regions,
		ShadedProbability: shadedProb,
		Formula:           formula,
		Summary:           fmt.Sprintf("Venn Diagram: P(A)=%.4g, P(B)=%.4g, P(A and B)=%.4g. Shaded region %s has probability %.4g.", pA, pB, pInter, region, shadedProb),
	})
}

func (s *Server) handleVisualsSampling(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	popDist := q.Get("pop_dist")
	if popDist == "" {
		popDist = "skewed"
	}

	n := int(parseQueryFloat(q.Get("n"), 25))
	if n < 2 {
		n = 2
	} else if n > 100 {
		n = 100
	}

	trials := int(parseQueryFloat(q.Get("trials"), 400))
	if trials < 50 {
		trials = 50
	} else if trials > 1000 {
		trials = 1000
	}

	seed := int64(parseQueryFloat(q.Get("seed"), 12345))
	rng := rand.New(rand.NewSource(seed))

	var popMean, popSD float64
	sampleGen := func() float64 {
		switch popDist {
		case "uniform":
			// U(0, 20): mu = 10, var = 400 / 12 = 33.333, sd = 5.7735
			popMean = 10.0
			popSD = 20.0 / math.Sqrt(12.0)
			return rng.Float64() * 20.0
		case "normal":
			// N(50, 10^2)
			popMean = 50.0
			popSD = 10.0
			return 50.0 + 10.0*rng.NormFloat64()
		case "exponential":
			// Exp(lambda = 0.2): mu = 5, sd = 5
			popMean = 5.0
			popSD = 5.0
			return -math.Log(1.0-rng.Float64()) / 0.2
		default: // "skewed"
			// Right-skewed Chi-square-like with df = 3: mu = 3, sd = sqrt(6) = 2.449
			popMean = 3.0
			popSD = math.Sqrt(6.0)
			z1 := rng.NormFloat64()
			z2 := rng.NormFloat64()
			z3 := rng.NormFloat64()
			return z1*z1 + z2*z2 + z3*z3
		}
	}

	// Warm up popMean and popSD calculation
	sampleGen()

	sampleMeans := make([]float64, trials)
	sumMeans := 0.0
	for i := 0; i < trials; i++ {
		sum := 0.0
		for j := 0; j < n; j++ {
			sum += sampleGen()
		}
		sm := sum / float64(n)
		sampleMeans[i] = sm
		sumMeans += sm
	}
	empMean := sumMeans / float64(trials)

	sumSqDiff := 0.0
	for _, sm := range sampleMeans {
		diff := sm - empMean
		sumSqDiff += diff * diff
	}
	empSE := math.Sqrt(sumSqDiff / float64(trials-1))
	theorSE := popSD / math.Sqrt(float64(n))

	// Histogram binning (20 bins)
	numBins := 20
	minX := popMean - 3.8*theorSE
	maxX := popMean + 3.8*theorSE
	binWidth := (maxX - minX) / float64(numBins)

	bins := make([]int, numBins)
	for _, sm := range sampleMeans {
		idx := int((sm - minX) / binWidth)
		if idx < 0 {
			idx = 0
		} else if idx >= numBins {
			idx = numBins - 1
		}
		bins[idx]++
	}

	normOverlay, _ := mathengine.NewNormalDistribution(popMean, theorSE)
	bars := make([]SamplingBar, numBins)
	for i := 0; i < numBins; i++ {
		mid := minX + (float64(i)+0.5)*binWidth
		density := float64(bins[i]) / (float64(trials) * binWidth)
		theorDens := 0.0
		if normOverlay != nil {
			theorDens = normOverlay.PDF(mid)
		}
		bars[i] = SamplingBar{
			BinMid:             math.Round(mid*1000) / 1000,
			Count:              bins[i],
			Density:            math.Round(density*10000) / 10000,
			TheoreticalDensity: math.Round(theorDens*10000) / 10000,
		}
	}

	// Normal curve points
	numCurvePts := 80
	curvePts := make([]CurvePoint, numCurvePts)
	step := (maxX - minX) / float64(numCurvePts-1)
	for i := 0; i < numCurvePts; i++ {
		x := minX + float64(i)*step
		curvePts[i] = CurvePoint{
			X:   math.Round(x*1000) / 1000,
			PDF: normOverlay.PDF(x),
			CDF: normOverlay.CDF(x),
		}
	}

	excelFormula := fmt.Sprintf("=NORM.DIST(x_bar, %.4g, %.4g / SQRT(%d), TRUE)", popMean, popSD, n)
	cltNote := fmt.Sprintf("Central Limit Theorem: Regardless of the population shape (%s), the distribution of sample means approaches N(%.4g, (%.4g / sqrt(%d))^2) = N(%.4g, %.4g^2) as n grows.", popDist, popMean, popSD, n, popMean, theorSE)

	writeJSON(w, SamplingVisualResponse{
		Family:          "sampling",
		PopDistribution: popDist,
		PopMean:         math.Round(popMean*1000) / 1000,
		PopStdDev:       math.Round(popSD*1000) / 1000,
		SampleSize:      n,
		Trials:          trials,
		StandardError:   math.Round(theorSE*10000) / 10000,
		EmpiricalMean:   math.Round(empMean*1000) / 1000,
		EmpiricalStdErr: math.Round(empSE*10000) / 10000,
		Bars:            bars,
		NormalCurve:     curvePts,
		CLTNote:         cltNote,
		ExcelFormula:    excelFormula,
		Summary:         fmt.Sprintf("Sampling Distribution: n = %d, %d trials from %s population. Theoretical SE = %.4g vs Empirical SE = %.4g.", n, trials, popDist, theorSE, empSE),
	})
}

func (s *Server) handleVisualsCI(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	confLevel := parseQueryFloat(q.Get("conf_level"), 0.95)
	if confLevel <= 0 || confLevel >= 1 {
		confLevel = 0.95
	}

	mu := parseQueryFloat(q.Get("mu"), 50.0)
	sigma := parseQueryFloat(q.Get("sigma"), 10.0)
	if sigma <= 0 {
		sigma = 10.0
	}

	n := int(parseQueryFloat(q.Get("n"), 16))
	if n < 2 {
		n = 2
	} else if n > 100 {
		n = 100
	}

	numIntervals := int(parseQueryFloat(q.Get("num_intervals"), 30))
	if numIntervals < 5 {
		numIntervals = 5
	} else if numIntervals > 100 {
		numIntervals = 100
	}

	seed := int64(parseQueryFloat(q.Get("seed"), 42))

	intervals, empiricalCov, err := mathengine.SimulateConfidenceIntervalCoverage(mu, sigma, n, confLevel, numIntervals, seed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	captured := 0
	for _, iv := range intervals {
		if iv.CapturesTruth {
			captured++
		}
	}

	alpha := 1.0 - confLevel
	excelFormula := fmt.Sprintf("=CONFIDENCE.T(%.4g, s, %d)", alpha, n)

	writeJSON(w, ConfidenceIntervalVisualResponse{
		Family:            "confidence_interval",
		ConfidenceLevel:   confLevel,
		Alpha:             alpha,
		TrueMean:          mu,
		Sigma:             sigma,
		SampleSize:        n,
		NumIntervals:      numIntervals,
		Intervals:         intervals,
		CapturedCount:     captured,
		EmpiricalCoverage: math.Round(empiricalCov*1000) / 1000,
		ExcelFormula:      excelFormula,
		Summary:           fmt.Sprintf("%.1f%% Confidence Interval Simulation: %d of %d simulated intervals captured true mu = %.4g (empirical coverage %.1f%%).", confLevel*100, captured, numIntervals, mu, empiricalCov*100),
	})
}

func (s *Server) handleVisualsHypothesis(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	testType := q.Get("test_type")
	if testType == "" {
		testType = "mean_t"
	}

	tailStr := q.Get("tail")
	var tail mathengine.TailDirection
	switch tailStr {
	case "left", "left_tailed":
		tail = mathengine.TailLeft
	case "right", "right_tailed":
		tail = mathengine.TailRight
	default:
		tail = mathengine.TailTwoSided
	}

	alpha := parseQueryFloat(q.Get("alpha"), 0.05)
	if alpha <= 0 || alpha >= 1 {
		alpha = 0.05
	}

	nullMu := parseQueryFloat(q.Get("null_mu"), 50.0)
	xBar := parseQueryFloat(q.Get("x_bar"), 53.0)
	sVal := parseQueryFloat(q.Get("s"), 8.0)
	if sVal <= 0 {
		sVal = 8.0
	}

	n := int(parseQueryFloat(q.Get("n"), 25))
	if n < 2 {
		n = 2
	}

	var res *mathengine.HypothesisTestResult
	var err error
	if testType == "mean_z" {
		res, err = mathengine.PerformMeanZTest(nullMu, xBar, sVal, n, tail, alpha)
	} else {
		res, err = mathengine.PerformMeanTTest(nullMu, xBar, sVal, n, tail, alpha)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate null distribution curve points on standard scale [-4, 4]
	numPoints := 100
	points := make([]CurvePoint, numPoints)
	step := 8.0 / float64(numPoints-1)

	var distPDF, distCDF func(float64) float64
	if testType == "mean_z" {
		norm := mathengine.StandardNormal()
		distPDF = norm.PDF
		distCDF = norm.CDF
	} else {
		tDist, _ := mathengine.NewTDistribution(float64(n - 1))
		distPDF = tDist.PDF
		distCDF = tDist.CDF
	}

	for i := 0; i < numPoints; i++ {
		x := -4.0 + float64(i)*step
		points[i] = CurvePoint{
			X:   math.Round(x*1000) / 1000,
			PDF: distPDF(x),
			CDF: distCDF(x),
		}
	}

	writeJSON(w, HypothesisVisualResponse{
		Family:          "hypothesis_test",
		TestType:        testType,
		Tail:            string(tail),
		NullHypothesis:  res.NullHypothesis,
		AltHypothesis:   res.AltHypothesis,
		NullValue:       res.NullValue,
		SampleStat:      res.SampleStat,
		SampleSize:      res.SampleSize,
		StandardError:   math.Round(res.StandardError*10000) / 10000,
		TestStatistic:   math.Round(res.TestStatistic*1000) / 1000,
		DegreesOfFree:   res.DegreesOfFree,
		PValue:          math.Round(res.PValue*10000) / 10000,
		Alpha:           alpha,
		CriticalValue:   math.Round(res.CriticalValue*1000) / 1000,
		CriticalValue2:  res.CriticalValue2,
		RejectNull:      res.RejectNull,
		Decision:        res.Decision,
		RejectionRegion: res.RejectionRegion,
		ExcelFormula:    res.ExcelFormula,
		Summary:         res.Explanation,
		Points:          points,
	})
}

func (s *Server) handleVisualsRegression(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	dataset := q.Get("dataset")
	if dataset == "" {
		dataset = "advertising_sales"
	}

	var points []mathengine.DataPoint
	switch dataset {
	case "study_exam":
		points = []mathengine.DataPoint{
			{X: 2.0, Y: 55.0},
			{X: 4.0, Y: 62.0},
			{X: 5.0, Y: 70.0},
			{X: 7.0, Y: 78.0},
			{X: 8.0, Y: 85.0},
			{X: 10.0, Y: 92.0},
			{X: 12.0, Y: 95.0},
		}
	default: // "advertising_sales"
		dataset = "advertising_sales"
		points = []mathengine.DataPoint{
			{X: 1.5, Y: 12.0},
			{X: 2.0, Y: 15.0},
			{X: 3.0, Y: 17.0},
			{X: 4.0, Y: 22.0},
			{X: 5.0, Y: 24.0},
			{X: 6.5, Y: 30.0},
			{X: 7.0, Y: 31.0},
			{X: 8.5, Y: 36.0},
		}
	}

	reg, err := mathengine.ComputeLinearRegression(points)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	minX := points[0].X
	maxX := points[0].X
	for _, p := range points {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
	}

	lineStartX := minX - 0.5
	lineStartY := reg.Intercept + reg.Slope*lineStartX
	lineEndX := maxX + 0.5
	lineEndY := reg.Intercept + reg.Slope*lineEndX

	var predX *float64
	var predY *float64
	if rawPred := q.Get("x_pred"); rawPred != "" {
		px := parseQueryFloat(rawPred, 5.0)
		py, _ := reg.Predict(px)
		predX = &px
		pyRound := math.Round(py*1000) / 1000
		predY = &pyRound
	}

	excelFormula := fmt.Sprintf("=SLOPE(Y_range, X_range) -> %.4g, =INTERCEPT(Y_range, X_range) -> %.4g", reg.Slope, reg.Intercept)

	writeJSON(w, RegressionVisualResponse{
		Family:          "regression",
		DatasetName:     dataset,
		N:               reg.N,
		Points:          points,
		Slope:           math.Round(reg.Slope*10000) / 10000,
		Intercept:       math.Round(reg.Intercept*10000) / 10000,
		Correlation:     math.Round(reg.Correlation*10000) / 10000,
		RSquared:        math.Round(reg.RSquared*10000) / 10000,
		StdErrEst:       math.Round(reg.StdErrEst*10000) / 10000,
		Equation:        reg.Equation,
		AssociationNote: reg.Association,
		ExcelFormula:    excelFormula,
		Residuals:       reg.Residuals,
		LineStartX:      math.Round(lineStartX*100) / 100,
		LineStartY:      math.Round(lineStartY*100) / 100,
		LineEndX:        math.Round(lineEndX*100) / 100,
		LineEndY:        math.Round(lineEndY*100) / 100,
		PredictedX:      predX,
		PredictedY:      predY,
		Summary:         fmt.Sprintf("Simple Linear Regression: %s. r = %.4g, R^2 = %.4g (%.1f%% of variation explained). %s", reg.Equation, reg.Correlation, reg.RSquared, reg.RSquared*100, reg.Association),
	})
}

func parseQueryFloat(raw string, defaultVal float64) float64 {
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return defaultVal
	}
	return v
}
