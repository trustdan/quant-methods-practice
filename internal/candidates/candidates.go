// Package candidates builds untrusted proposals. Only Go rules create answer keys.
package candidates

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

// Proposal deliberately has no formula, grade, approval, or expected-answer fields.
type Proposal struct {
	FamilyID string  `json:"family_id"`
	N        int     `json:"n"`
	P        float64 `json:"p"`
	K        int     `json:"k"`
	Title    string  `json:"title"`
	Scenario string  `json:"scenario_markdown"`
	Success  string  `json:"success_label"`
}
type Source struct {
	Mode  string `json:"mode"`
	Route string `json:"route"`
	Model string `json:"model"`
	Seed  int64  `json:"seed"`
}
type Review struct {
	Action            string    `json:"action"`
	Reviewer          string    `json:"reviewer"`
	Notes             string    `json:"notes"`
	SemanticConfirmed bool      `json:"semantic_confirmed"`
	Timestamp         time.Time `json:"timestamp"`
}
type Record struct {
	ID        string                   `json:"id"`
	Revision  int                      `json:"revision"`
	Status    string                   `json:"status"`
	Proposal  Proposal                 `json:"proposal"`
	Template  *domain.QuestionTemplate `json:"template"`
	Source    Source                   `json:"source"`
	CreatedAt time.Time                `json:"created_at"`
	Reviews   []Review                 `json:"reviews"`
}

var ErrConflict = errors.New("candidate changed or already reviewed; reload the preview")

func Decode(data []byte) (Proposal, error) {
	var p Proposal
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&p); err != nil {
		return p, fmt.Errorf("invalid candidate JSON: %w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return p, errors.New("candidate must contain exactly one JSON document")
	}
	return p, nil
}

// Local uses a finite, tested parameter set. Cosmetic variants retain one setting group.
func Local(seed int64) Proposal {
	combinations := []struct {
		n int
		p float64
		k int
	}{{4, .5, 2}, {8, .25, 2}, {6, .5, 3}, {10, .2, 2}}
	idx := seed % int64(len(combinations))
	if idx < 0 {
		idx += int64(len(combinations))
	}
	c := combinations[idx]
	return Proposal{FamilyID: "binomial_pmf", N: c.n, P: c.p, K: c.k,
		Title: fmt.Sprintf("Exactly %d successes in %d trials", c.k, c.n), Success: "success",
		Scenario: fmt.Sprintf("A simulated experiment has %d independent trials. Each trial has two outcomes, success or failure, and the probability of success is constant at %g. Find the probability of exactly %d successes.", c.n, c.p, c.k)}
}

func Build(id string, p Proposal) (*domain.QuestionTemplate, error) {
	if p.FamilyID != "binomial_pmf" {
		return nil, errors.New("candidate generation currently supports binomial_pmf only; other families require reviewed builders")
	}
	if p.N < 2 || p.N > 50 || p.K < 1 || p.K >= p.N || math.IsNaN(p.P) || math.IsInf(p.P, 0) || p.P <= 0 || p.P >= 1 {
		return nil, errors.New("use 2–50 trials, 0 < p < 1, and 1 <= k < n")
	}
	for _, f := range []struct {
		name, value string
		max         int
	}{{"title", p.Title, 200}, {"scenario", p.Scenario, 6000}, {"success label", p.Success, 120}} {
		if strings.TrimSpace(f.value) == "" || len(f.value) > f.max {
			return nil, fmt.Errorf("%s is required and must be at most %d bytes", f.name, f.max)
		}
	}
	allowed := false
	for seed := int64(0); seed < 4; seed++ {
		v := Local(seed)
		if p.N == v.N && p.P == v.P && p.K == v.K {
			allowed = true
		}
	}
	if !allowed {
		return nil, errors.New("parameters must use a reviewed variation: (4,0.5,2), (8,0.25,2), (6,0.5,3), or (10,0.2,2)")
	}
	deriv, err := mathengine.DeriveBinomialProblem(p.N, p.P, p.K, mathengine.Exactly(p.K))
	if err != nil {
		return nil, err
	}
	value := deriv.CanonicalProbability
	t := &domain.QuestionTemplate{SchemaVersion: 1, TemplateKind: domain.TemplateKindFixedReference, ID: id, Version: 1, Status: domain.StatusDraft, FamilyID: p.FamilyID, RuleVersion: 1, ModuleID: "module_2", Title: p.Title, ScenarioMarkdown: p.Scenario,
		Parameters:   map[string]any{"n": p.N, "p": p.P, "k": p.K},
		Assumptions:  []string{fmt.Sprintf("%d fixed trials", p.N), "Independent trials", "Two outcomes per trial", fmt.Sprintf("Constant success probability %g", p.P)},
		SettingGroup: "binomial_fixed_independent_exact_count", SourceRefs: []string{"Generic introductory candidate; wording requires human review"}}
	correct := func(id, text string) domain.Option { return domain.Option{ID: id, TextMarkdown: text} }
	wrong := func(id, text, mis, hint string) domain.Option {
		return domain.Option{ID: id, TextMarkdown: text, MisconceptionID: &mis, HintMarkdown: &hint}
	}
	choice := func(id, concept, prompt, explanation string, options ...domain.Option) {
		t.ConceptIDs = append(t.ConceptIDs, concept)
		t.Stages = append(t.Stages, domain.StageTemplate{ID: id, Kind: domain.StageKindChoice, PromptMarkdown: prompt, Options: options, ExpectedAnswer: domain.ExpectedAnswer{Kind: domain.AnswerKindChoice, OptionID: options[0].ID}, EvidenceConceptIDs: []string{concept}, ExplanationMarkdown: explanation})
	}
	choice("define_variable", "random_variable_definition", "What should $X$ represent?", fmt.Sprintf("$X$ counts trials with the outcome labeled %q. Its support is the integers from 0 through %d.", p.Success, p.N),
		correct("count", fmt.Sprintf("The number of trials with outcome %q among all %d trials", p.Success, p.N)),
		wrong("probability", "The probability of success on one trial", "parameter_as_random_variable", "Which quantity varies across repetitions of the whole experiment?"),
		wrong("trials", "The fixed number of trials", "fixed_count_as_random_variable", "The number of trials is fixed. What outcome count remains unknown?"))
	choice("choose_model", "distribution_selection", "Which distribution models this count?", "A binomial count uses a fixed number of independent two-outcome trials with constant success probability.",
		correct("binomial", "Binomial"), wrong("poisson", "Poisson", "poisson_for_fixed_trials", "Is exposure fixed, or is the number of trials fixed?"), wrong("normal", "Continuous normal", "continuous_for_discrete_count", "Can this count take fractional values?"))
	choice("check_parameters", "binomial_conditions", "Which conditions justify the model?", fmt.Sprintf("$n=%d$, $p=%g$; the story must state fixed, independent, two-outcome trials with constant success probability.", p.N, p.P),
		correct("conditions", fmt.Sprintf("%d independent two-outcome trials, each with success probability %g", p.N, p.P)), wrong("dependent", "Dependent trials are sufficient", "dependence_ignored", "Would a success change the probability of the next success?"), wrong("varying", "Success probability may vary between trials", "constant_p_ignored", "Does one common p describe every trial?"))
	choice("translate_event", "event_translation", fmt.Sprintf("Which event means exactly %d successes?", p.K), fmt.Sprintf("Exactly selects only $X=%d$.", p.K),
		correct("exactly", fmt.Sprintf("$X=%d$", p.K)), wrong("at_most", fmt.Sprintf("$X\\le %d$", p.K), "exactly_as_at_most", "Would a smaller count satisfy exactly?"), wrong("at_least", fmt.Sprintf("$X\\ge %d$", p.K), "exactly_as_at_least", "Would a larger count satisfy exactly?"))
	expression := fmt.Sprintf("\\binom{%d}{%d}(%g)^{%d}(1-%g)^{%d}", p.N, p.K, p.P, p.K, p.P, p.N-p.K)
	choice("build_expression", "binomial_expression", "Which expression counts every ordering and both outcomes?", "The combination counts positions of successes. Both success and failure factors are needed.",
		correct("complete", "$"+expression+"$"), wrong("one_order", fmt.Sprintf("$(%g)^{%d}(1-%g)^{%d}$", p.P, p.K, p.P, p.N-p.K), "missing_combination", "How many placements of the successes are possible?"), wrong("no_failures", fmt.Sprintf("$\\binom{%d}{%d}(%g)^{%d}$", p.N, p.K, p.P, p.K), "missing_failure_factor", "What must happen on the remaining trials?"))
	t.ConceptIDs = append(t.ConceptIDs, "probability_calculation")
	t.Stages = append(t.Stages, domain.StageTemplate{ID: "calculate_probability", Kind: domain.StageKindNumeric, PromptMarkdown: fmt.Sprintf("Calculate $P(X=%d)$. Use a decimal, explicit percent, or fraction. Absolute tolerance: 0.0000005; relative tolerance: 0.000001.", p.K), Options: []domain.Option{}, ExpectedAnswer: domain.ExpectedAnswer{Kind: domain.AnswerKindNumeric, Value: &value, Units: "probability"}, EvidenceConceptIDs: []string{"probability_calculation"}, ExplanationMarkdown: fmt.Sprintf("$%s = %.10g$. This is %.8g%%.", expression, value, 100*value), NumericPolicy: &domain.NumericPolicy{Version: 1, AbsoluteTolerance: 5e-7, RelativeTolerance: 1e-6, AllowedForms: []domain.NumericForm{domain.NumericFormDecimal, domain.NumericFormPercent, domain.NumericFormFraction}}})
	choice("interpret_probability", "probability_interpretation", "What does the calculated probability describe?", "This probability concerns the whole experiment, and does not guarantee the next result.",
		correct("experiment", fmt.Sprintf("The chance of exactly %d successes across all %d trials", p.K, p.N)), wrong("single", "The chance of success on one trial", "experiment_as_single_trial", "Is the requested event one trial or the whole experiment?"), wrong("guaranteed", "A guarantee about the next experiment", "probability_as_guarantee", "Is this probability equal to one?"))
	// Derivation succeeds before template validation; model-provided values never enter keys.
	if err := bank.ValidateTemplate(t, nil); err != nil {
		return nil, err
	}
	return t, nil
}

func PrepareReview(rec *Record, revision int, review Review, now time.Time) (*Record, error) {
	if rec.Revision != revision {
		return nil, ErrConflict
	}
	if strings.TrimSpace(review.Reviewer) == "" || len(review.Reviewer) > 200 || strings.TrimSpace(review.Notes) == "" || len(review.Notes) > 4000 {
		return nil, errors.New("reviewer and review notes are required (200/4000 bytes maximum)")
	}
	if (review.Action == "approve" || review.Action == "reject") && rec.Status != "pending" {
		return nil, ErrConflict
	}
	if review.Action == "retire" && rec.Status != "approved" {
		return nil, ErrConflict
	}
	if review.Action != "approve" && review.Action != "reject" && review.Action != "retire" {
		return nil, errors.New("unknown review action")
	}
	// Rebuild from the proposal, so a stored or supplied key cannot become authoritative.
	tmpl, err := Build(rec.ID, rec.Proposal)
	if err != nil {
		return nil, err
	}
	out := *rec
	tmpl.SourceRefs = append([]string{}, rec.Template.SourceRefs...)
	out.Template = tmpl
	out.Revision++
	out.Reviews = append(append([]Review{}, rec.Reviews...), review)
	out.Reviews[len(out.Reviews)-1].Timestamp = now.UTC()
	switch review.Action {
	case "approve":
		if !review.SemanticConfirmed {
			return nil, errors.New("confirm review of the scenario, assumptions, all stages, answers and hints before approval")
		}
		out.Status = "approved"
		tmpl.Status = domain.StatusActive
		tmpl.Approval = &domain.ApprovalRecord{Reviewer: review.Reviewer, ReviewedAt: now.UTC(), Notes: review.Notes}
	case "reject":
		out.Status = "rejected"
	case "retire":
		out.Status = "retired"
		tmpl.Status = domain.StatusRetired
		tmpl.Approval = rec.Template.Approval
	}
	if err := bank.ValidateTemplate(tmpl, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// NewRecord constructs an immutable draft with generation provenance, never an approval.
func NewRecord(id string, proposal Proposal, source Source, now time.Time) (*Record, error) {
	tmpl, err := Build(id, proposal)
	if err != nil {
		return nil, err
	}
	tmpl.SourceRefs = append(tmpl.SourceRefs, fmt.Sprintf("Generation: %s / %s / %s; seed %d", source.Mode, source.Route, source.Model, source.Seed))
	return &Record{ID: id, Revision: 1, Status: "pending", Proposal: proposal, Template: tmpl, Source: source, CreatedAt: now.UTC(), Reviews: []Review{}}, nil
}
