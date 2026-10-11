package httpapi

import (
	"fmt"
	"math"
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
