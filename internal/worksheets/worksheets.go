package worksheets

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/drill"
)

var ErrConflict = errors.New("worksheet changed; reload before retrying")
var ErrInvalidInput = errors.New("invalid worksheet input")

func invalidInput(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, fmt.Sprintf(format, args...))
}

type Item struct {
	Key        string           `json:"key"`
	QuestionID string           `json:"question_id"`
	State      drill.StageState `json:"state"`
}
type Record struct {
	SourceTemplate *domain.QuestionTemplate  `json:"source_template"`
	ID             string                    `json:"id"`
	Revision       int                       `json:"revision"`
	Mode           string                    `json:"mode"`
	Status         string                    `json:"status"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
	Questions      []domain.QuestionInstance `json:"questions"`
	Items          []Item                    `json:"items"`
	Dataset        *Dataset                  `json:"dataset,omitempty"`
}
type Command struct {
	ID       string                            `json:"command_id"`
	Revision int                               `json:"expected_revision"`
	Type     string                            `json:"type"`
	Answers  map[string]domain.SubmittedAnswer `json:"answers"`
}

// New copies the approved problem and the seeded option order from the same
// instantiator used by progressive practice. No new bank content is activated.
func New(id string, tmpl *domain.QuestionTemplate, seed int64, dataset *Dataset, now time.Time) (*Record, error) {
	if tmpl == nil || tmpl.Status != domain.StatusActive || tmpl.Approval == nil {
		return nil, errors.New("choose an approved active question")
	}
	if dataset != nil && tmpl.ID != "binomial_fair_coin_exactly_two" {
		return nil, errors.New("dataset case requires the reviewed four-fair-coin reference")
	}
	s, err := drill.NewSessionManager(func() time.Time { return now.UTC() }).CreateSession(tmpl, seed)
	if err != nil {
		return nil, err
	}
	r := &Record{ID: id, Revision: 1, Mode: "full_solution", Status: "draft", CreatedAt: now.UTC(), UpdatedAt: now.UTC(), Questions: []domain.QuestionInstance{s.QuestionInstance}, SourceTemplate: tmpl}
	if dataset != nil {
		if strings.TrimSpace(dataset.Reviewer) == "" || strings.TrimSpace(dataset.SourceNote) == "" {
			return nil, errors.New("dataset review needs your name and a source/row-meaning note")
		}
		r.Mode = "dataset"
		copyDataset := *dataset
		copyDataset.Rows = append([]Row{}, dataset.Rows...)
		r.Dataset = &copyDataset
		// Approval is authored here, never accepted as a browser-supplied record.
		r.Dataset.Approval = &domain.ApprovalRecord{Reviewer: dataset.Reviewer, ReviewedAt: now.UTC(), Notes: "Explicit semantic confirmation of all case fields, options, hints, numeric keys and row meaning. Source/row note: " + dataset.SourceNote}
		q, err := empiricalQuestion(id, dataset, seed)
		if err != nil {
			return nil, err
		}
		r.Questions = append([]domain.QuestionInstance{q}, r.Questions...)
	}
	for _, q := range r.Questions {
		for _, stage := range q.Stages {
			r.Items = append(r.Items, Item{Key: q.ID + ":" + stage.ID, QuestionID: q.ID, State: drill.StageState{Instance: stage, Status: drill.StageStatusActive}})
		}
	}
	// Break every pointer/map/slice alias to the current bank and caller data.
	return clone(r)
}

func clone(r *Record) (*Record, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var copy Record
	err = json.Unmarshal(b, &copy)
	return &copy, err
}

// Apply validates the entire form before committing any attempts. The caller
// atomically persists this result together with the idempotent command result.
func Apply(old *Record, cmd Command, now time.Time) (*Record, error) {
	if !safeID.MatchString(cmd.ID) {
		return nil, invalidInput("command_id must be a safe identifier of 1–80 characters")
	}
	if cmd.Revision != old.Revision {
		return nil, ErrConflict
	}
	if cmd.Type != "save_draft" && cmd.Type != "submit" {
		return nil, invalidInput("unknown worksheet command")
	}
	if old.Status == "completed" {
		return nil, invalidInput("completed worksheets are read-only")
	}
	r, err := clone(old)
	if err != nil {
		return nil, err
	}
	valid := map[string]bool{}
	for _, item := range r.Items {
		if item.State.Status != drill.StageStatusCompleted {
			valid[item.Key] = true
		}
	}
	for key, answer := range cmd.Answers {
		if !valid[key] {
			return nil, invalidInput("unknown or completed field %q", key)
		}
		if len(answer.NumericRaw) > 200 || len(answer.OptionID) > 100 {
			return nil, invalidInput("answer too long")
		}
	}
	if cmd.Type == "submit" && len(cmd.Answers) != len(valid) {
		return nil, invalidInput("complete every unfinished field before submitting")
	}
	complete := true
	for i := range r.Items {
		item := &r.Items[i]
		stage := &item.State
		if stage.Status == drill.StageStatusCompleted {
			continue
		}
		answer, supplied := cmd.Answers[item.Key]
		answer.NormalizedValue = nil
		answer.Form = nil
		if cmd.Type == "save_draft" {
			stage.DraftAnswer = nil
			if supplied {
				if answer.Kind != stage.Instance.Kind {
					return nil, invalidInput("invalid answer kind for %s", stage.Instance.ID)
				}
				// Keep raw input only; canonical normalization is always computed at grading time.
				stage.DraftAnswer = &answer
			}
			continue
		}
		attempt, invalid, diagnostic, err := drill.GradeSubmission(r.ID, item.QuestionID, stage, answer, func() time.Time { return now.UTC() })
		if err != nil {
			return nil, err
		}
		if invalid {
			return nil, invalidInput("%s: %s", stage.Instance.ID, diagnostic)
		}
		attempt.Assistance = []domain.AssistanceType{domain.AssistanceFullSolution}
		if r.Dataset != nil {
			attempt.Assistance = append(attempt.Assistance, domain.AssistanceReference)
		}
		if len(stage.Attempts) > 0 {
			attempt.Assistance = append(attempt.Assistance, domain.AssistanceHint, domain.AssistanceRetry)
		}
		stage.Attempts = append(stage.Attempts, attempt)
		stage.DraftAnswer = nil
		stage.LastFeedback = attempt.FeedbackMarkdown
		if attempt.IsCorrect || len(stage.Attempts) == 2 {
			stage.Status = drill.StageStatusCompleted
			stage.FirstTryCorrect = attempt.IsCorrect && len(stage.Attempts) == 1
			stage.SolvedOnRetry = attempt.IsCorrect && len(stage.Attempts) == 2
			stage.Revealed = !attempt.IsCorrect
			if stage.Revealed {
				stage.Assistance = append(stage.Assistance, domain.AssistanceSolutionReveal)
			}
		} else {
			stage.Status = drill.StageStatusRetry
			complete = false
		}
		stage.Assistance = append(stage.Assistance, attempt.Assistance...)
	}
	if cmd.Type == "submit" {
		r.Status = "retry"
		if complete {
			r.Status = "completed"
		}
	}
	r.Revision++
	r.UpdatedAt = now.UTC()
	return r, nil
}

// PreviewDatasetQuestion exposes keys only for explicit author review; no bank or practice writes.
func PreviewDatasetQuestion(d *Dataset, seed int64) (domain.QuestionInstance, error) {
	return empiricalQuestion("dataset_preview", d, seed)
}

func empiricalQuestion(id string, d *Dataset, seed int64) (domain.QuestionInstance, error) {
	matching, value, err := d.Summary()
	if err != nil {
		return domain.QuestionInstance{}, err
	}
	hint := "The observational unit is a complete four-toss experiment. Count matching experiments, then divide by all recorded experiments."
	choice := func(stageID, prompt, correct string, options ...string) domain.StageInstance {
		stage := domain.StageInstance{ID: stageID, Kind: domain.StageKindChoice, PromptMarkdown: prompt, ExpectedAnswer: domain.ExpectedAnswer{Kind: domain.AnswerKindChoice, OptionID: correct}, ExplanationMarkdown: hint}
		for i, text := range options {
			stage.Options = append(stage.Options, domain.Option{ID: fmt.Sprintf("option_%d", i+1), TextMarkdown: text, HintMarkdown: &hint})
		}
		return stage
	}
	// Rotate choices deterministically so the correct answer does not occupy one fixed slot.
	stages := []domain.StageInstance{
		choice("method", "Which method describes the observed fraction of experiments with exactly two heads?", "option_2", "Use the theoretical binomial probability as the observed fraction.", "Count matching experiments and divide by all recorded experiments.", "Divide the number of heads by four times the number of experiments."),
		choice("assumptions", "What do these observations establish?", "option_3", "They prove each toss is independent.", "They prove every toss has probability 0.5 of heads.", "They describe the recorded experiments; fairness and independence remain model assumptions."),
		{ID: "observed_value", Kind: domain.StageKindNumeric, PromptMarkdown: "What fraction of recorded experiments have exactly two heads? Enter a decimal, percentage or fraction.", ExpectedAnswer: domain.ExpectedAnswer{Kind: domain.AnswerKindNumeric, Value: &value, Units: "probability"}, NumericPolicy: &domain.NumericPolicy{Version: 1, AbsoluteTolerance: 1e-6, RelativeTolerance: 1e-6, AllowedForms: []domain.NumericForm{domain.NumericFormDecimal, domain.NumericFormPercent, domain.NumericFormFraction}}, ExplanationMarkdown: fmt.Sprintf("There are %d matching experiments out of %d: $%d/%d = %.10g$. This is an observed relative frequency, not proof of a population probability.", matching, len(d.Rows), matching, len(d.Rows), value)},
		choice("interpretation", "How should you compare this recorded fraction with the theoretical binomial probability?", "option_1", "The recorded fraction can differ from the model probability; this alone does not validate or refute fairness and independence.", "The recorded fraction must equal the binomial probability exactly.", "The recorded fraction is the chance that every future experiment will have exactly two heads."),
	}
	// Explain the mistaken inference, not just the arithmetic, on conditions and
	// interpretation errors. These are fixed teaching rules, reviewed in preview.
	causalHints := map[string][]string{
		"method":         {"A model probability comes from assumptions. An observed fraction comes from counting the recorded experiments.", hint, "The event concerns complete experiments with exactly two heads, not heads pooled over individual tosses."},
		"assumptions":    {"Independence describes how tosses are generated together. A finite table of counts cannot prove that condition.", "Fairness specifies a single toss's probability. Recorded counts alone cannot prove it is exactly 0.5.", "Describe these data separately from assumptions about the process generating future tosses."},
		"interpretation": {"Keep a recorded proportion separate from a probability supplied by a hypothetical model.", "Finite recorded frequencies can vary even when a model is appropriate; equality is not required.", "A fraction of recorded experiments is not a guarantee about every future experiment."},
	}
	for i := range stages {
		for j := range stages[i].Options {
			text := causalHints[stages[i].ID][j]
			stages[i].Options[j].HintMarkdown = &text
		}
	}
	stages[1].ExplanationMarkdown = "These recorded counts describe observations. Fairness and independence are assumptions of the separate coin model; neither condition is proved by this file."
	stages[3].ExplanationMarkdown = "Observed relative frequency describes this finite collection of experiments. The theoretical probability describes a model; agreement or disagreement alone does not establish or refute its assumptions."
	for i := range stages {
		if n := len(stages[i].Options); n > 0 {
			shift := int(seed % int64(n))
			if shift < 0 {
				shift += n
			}
			opts := stages[i].Options
			stages[i].Options = append(append([]domain.Option{}, opts[shift:]...), opts[:shift]...)
		}
	}
	return domain.QuestionInstance{ID: id + "_empirical", TemplateID: "empirical_four_toss_frequency", TemplateVersion: 1, Seed: seed, Parameters: map[string]interface{}{"trials_per_experiment": 4, "target_successes": 2, "experiment_count": len(d.Rows), "frequency_rule_version": 1}, Title: "Recorded experiments: observed relative frequency", ScenarioMarkdown: "Each row records the number of heads in one complete four-toss experiment. Describe the recorded frequency of exactly two heads, then solve the separate fair-independent-coin model below.", Assumptions: []string{"One row is one complete four-toss experiment.", "The file describes heads counts, not individual tosses.", "Fairness and independence are hypothetical model assumptions, not conclusions from this CSV."}, SettingGroup: "empirical_four_toss_counts", Stages: stages}, nil
}
