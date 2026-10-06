package drill

import (
	"fmt"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

var stageLabels = map[string]string{
	// Binomial
	"define_variable":        "Target",
	"choose_model":           "Model",
	"choose_distribution":    "Model",
	"check_parameters":       "Conditions",
	"translate_event":        "Event",
	"define_event":           "Event",
	"event_translation":      "Event",
	"build_expression":       "Expression",
	"calculate_probability":  "Calculate",
	"calculate_cumulative":   "Calculate",
	"calculate_zero_success": "Calculate",
	"interpret_probability":  "Interpret",
	"complement_check":       "Complement",
	"interpret_result":       "Interpret",
	// Poisson
	"model_selection":          "Model",
	"parameter_identification": "Parameters",
	"calculate_pmf":            "Calculate",
	"moments_check":            "Moments",
	"event_formula":            "Formula",
	"interpret_at_least_one":   "Complement",
	"rate_parameter":           "Rate",
	"calculate_one_event":      "Calculate",
	"mean_vs_mode":             "Properties",
	// Set Probability
	"identify_formula":           "Formula",
	"calculate_union":            "Calculate",
	"disjoint_check":             "Disjoint",
	"complement_definition":      "Definition",
	"calculate_complement":       "Calculate",
	"conditional_formula":        "Formula",
	"calculate_conditional":      "Calculate",
	"independence_determination": "Independence",
	// Linear combinations
	"expectation_linearity": "Linearity",
	"variance_additivity":   "Variance",
	"calculate_variance":    "Calculate",
	"covariance_role":       "Covariance",
	"constant_rule":         "Constant",
	"squared_coefficients":  "Scaling",
	"covariance_impact":     "Covariance",
}

// StageLabel returns the pedagogical label for a stage.
func StageLabel(stageID string, fallbackNumber int) string {
	if label, ok := stageLabels[stageID]; ok {
		return label
	}
	return fmt.Sprintf("Stage %d", fallbackNumber)
}

// BuildRecap constructs the full drill summary and canonical mathematical derivation.
func BuildRecap(s *DrillSession) (*DrillRecap, error) {
	recap := &DrillRecap{
		Title:            s.QuestionInstance.Title,
		ScenarioMarkdown: s.QuestionInstance.ScenarioMarkdown,
		TotalStages:      len(s.Stages),
		StageSummaries:   make([]StageSummary, 0, len(s.Stages)),
	}

	// Derive canonical binomial problem if parameters are available
	params := s.QuestionInstance.Parameters
	if nVal, okN := getIntParam(params, "n"); okN {
		if pVal, okP := getFloatParam(params, "p"); okP {
			if kVal, okK := getIntParam(params, "k"); okK {
				deriv, err := mathengine.DeriveBinomialProblem(nVal, pVal, kVal, mathengine.Exactly(kVal))
				if err == nil {
					recap.CanonicalDerivation = deriv
				}
			}
		}
	}

	for i, stage := range s.Stages {
		summary := StageSummary{
			StageNumber:            i + 1,
			StageID:                stage.Instance.ID,
			Label:                  StageLabel(stage.Instance.ID, i+1),
			PromptMarkdown:         stage.Instance.PromptMarkdown,
			ExplanationMarkdown:    stage.Instance.ExplanationMarkdown,
			MisconceptionTriggered: stage.MisconceptionID,
		}

		// Determine outcome
		if stage.FirstTryCorrect {
			summary.Outcome = "first_try"
			recap.FirstTryCount++
		} else if stage.SolvedOnRetry {
			summary.Outcome = "retry"
			recap.RetryCount++
		} else if stage.Revealed {
			summary.Outcome = "revealed"
			recap.RevealedCount++
		} else {
			summary.Outcome = "pending"
		}

		// Format learner answer from last attempt
		if len(stage.Attempts) > 0 {
			lastAtt := stage.Attempts[len(stage.Attempts)-1]
			summary.LearnerAnswer = formatAnswer(stage.Instance, lastAtt.SubmittedAnswer)
		} else {
			summary.LearnerAnswer = "(None submitted)"
		}

		// Format canonical answer
		summary.CanonicalAnswer = formatExpectedAnswer(stage.Instance)

		recap.StageSummaries = append(recap.StageSummaries, summary)
	}

	return recap, nil
}

func formatAnswer(inst domain.StageInstance, sub domain.SubmittedAnswer) string {
	if sub.Kind == domain.StageKindChoice {
		return getOptionText(inst.Options, sub.OptionID)
	}
	if sub.Kind == domain.StageKindNumeric {
		return sub.NumericRaw
	}
	return ""
}

func formatExpectedAnswer(inst domain.StageInstance) string {
	if inst.Kind == domain.StageKindChoice {
		return getOptionText(inst.Options, inst.ExpectedAnswer.OptionID)
	}
	if inst.Kind == domain.StageKindNumeric && inst.ExpectedAnswer.Value != nil {
		return fmt.Sprintf("%.4g (3/8 or 37.5%%)", *inst.ExpectedAnswer.Value)
	}
	return ""
}

func getIntParam(params map[string]any, key string) (int, bool) {
	val, ok := params[key]
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

func getFloatParam(params map[string]any, key string) (float64, bool) {
	val, ok := params[key]
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	default:
		return 0, false
	}
}
