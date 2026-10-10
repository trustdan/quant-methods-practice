package drill

import (
	"fmt"
	"strings"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

// GradeSubmission evaluates learner input against the stage contract.
// If the input is syntactically invalid, it returns invalidInput = true with a diagnostic
// and does NOT consume an attempt.
func GradeSubmission(
	sessionID string,
	instanceID string,
	stage *StageState,
	sub domain.SubmittedAnswer,
	clock func() time.Time,
) (attempt domain.StageAttempt, invalidInput bool, diagnostic string, err error) {
	if clock == nil {
		clock = time.Now
	}

	attemptNum := len(stage.Attempts) + 1

	switch stage.Instance.Kind {
	case domain.StageKindChoice:
		if sub.Kind != domain.StageKindChoice || sub.OptionID == "" {
			return domain.StageAttempt{}, true, "Please select an option before submitting.", nil
		}

		var matchedOption *domain.Option
		for i := range stage.Instance.Options {
			if stage.Instance.Options[i].ID == sub.OptionID {
				matchedOption = &stage.Instance.Options[i]
				break
			}
		}

		if matchedOption == nil {
			return domain.StageAttempt{}, true, fmt.Sprintf("Option %q is not valid for this stage.", sub.OptionID), nil
		}

		isCorrect := (matchedOption.ID == stage.Instance.ExpectedAnswer.OptionID)

		att := domain.StageAttempt{
			ID:              fmt.Sprintf("att_%s_%d", stage.Instance.ID, attemptNum),
			SessionID:       sessionID,
			InstanceID:      instanceID,
			StageID:         stage.Instance.ID,
			AttemptNumber:   attemptNum,
			SubmittedAnswer: sub,
			IsCorrect:       isCorrect,
			CreatedAt:       clock(),
		}

		if isCorrect {
			if attemptNum == 1 {
				att.FeedbackMarkdown = fmt.Sprintf("Correct! %s", stage.Instance.ExplanationMarkdown)
			} else {
				att.FeedbackMarkdown = fmt.Sprintf("Correct on retry! %s", stage.Instance.ExplanationMarkdown)
			}
		} else {
			if matchedOption.MisconceptionID != nil {
				stage.MisconceptionID = matchedOption.MisconceptionID
			}

			causalHint := ""
			if matchedOption.HintMarkdown != nil && *matchedOption.HintMarkdown != "" {
				causalHint = *matchedOption.HintMarkdown
			} else {
				causalHint = "Review the problem conditions and reconsider your selection."
			}

			if attemptNum == 1 {
				att.FeedbackMarkdown = fmt.Sprintf("Not quite. %s", causalHint)
				stage.ActiveHint = causalHint
			} else {
				// Second error: reveal solution
				correctText := getOptionText(stage.Instance.Options, stage.Instance.ExpectedAnswer.OptionID)
				att.FeedbackMarkdown = fmt.Sprintf("Incorrect. The correct choice is: **%s**.\n\n%s", correctText, stage.Instance.ExplanationMarkdown)
			}
		}

		return att, false, "", nil

	case domain.StageKindNumeric:
		if sub.Kind != domain.StageKindNumeric || strings.TrimSpace(sub.NumericRaw) == "" {
			return domain.StageAttempt{}, true, "Please enter a numeric probability before submitting.", nil
		}

		if stage.Instance.ExpectedAnswer.Value == nil {
			return domain.StageAttempt{}, false, "", fmt.Errorf("numeric stage %q missing expected value", stage.Instance.ID)
		}

		var allowedForms []domain.NumericForm
		if stage.Instance.NumericPolicy != nil {
			allowedForms = stage.Instance.NumericPolicy.AllowedForms
		}

		parsed, parseErr := mathengine.ParseNumericInput(sub.NumericRaw, allowedForms)
		if parseErr != nil {
			return domain.StageAttempt{}, true, parseErr.Error(), nil
		}

		sub.NormalizedValue = &parsed.Value
		form := parsed.Form
		sub.Form = &form

		var policy domain.NumericPolicy
		if stage.Instance.NumericPolicy != nil {
			policy = *stage.Instance.NumericPolicy
		} else {
			policy = domain.NumericPolicy{
				Version:           1,
				AbsoluteTolerance: 1e-6,
				RelativeTolerance: 1e-6,
			}
		}

		gradeRes := mathengine.GradeNumericAnswer(parsed, *stage.Instance.ExpectedAnswer.Value, policy)

		att := domain.StageAttempt{
			ID:              fmt.Sprintf("att_%s_%d", stage.Instance.ID, attemptNum),
			SessionID:       sessionID,
			InstanceID:      instanceID,
			StageID:         stage.Instance.ID,
			AttemptNumber:   attemptNum,
			SubmittedAnswer: sub,
			IsCorrect:       gradeRes.Correct,
			CreatedAt:       clock(),
		}

		if gradeRes.Correct {
			if attemptNum == 1 {
				att.FeedbackMarkdown = fmt.Sprintf("Correct! %s", stage.Instance.ExplanationMarkdown)
			} else {
				att.FeedbackMarkdown = fmt.Sprintf("Correct on retry! %s", stage.Instance.ExplanationMarkdown)
			}
		} else {
			hint := ""
			if strings.Contains(gradeRes.Reason, "bare percentage") {
				hint = fmt.Sprintf("Entered %.6g. Did you mean %.6g%%? Bare numbers are interpreted as absolute probabilities; probabilities must be between 0 and 1.", parsed.Value, parsed.Value)
			} else {
				hint = "Check which outcomes belong to the event, the model conditions, and the arithmetic. Use the quantities in this problem."
				if stage.Instance.ID == "observed_value" {
					hint = "Count experiments with exactly two heads, then divide by the number of recorded experiments. The denominator is experiments, not individual tosses."
				}
			}

			if attemptNum == 1 {
				att.FeedbackMarkdown = fmt.Sprintf("Not quite. %s", hint)
				stage.ActiveHint = hint
			} else {
				att.FeedbackMarkdown = fmt.Sprintf("Incorrect. The expected value is **%.10g %s**.\n\n%s", *stage.Instance.ExpectedAnswer.Value, stage.Instance.ExpectedAnswer.Units, stage.Instance.ExplanationMarkdown)
			}
		}

		return att, false, "", nil

	default:
		return domain.StageAttempt{}, false, "", fmt.Errorf("unsupported stage kind %q", stage.Instance.Kind)
	}
}

func getOptionText(options []domain.Option, optionID string) string {
	for _, opt := range options {
		if opt.ID == optionID {
			return opt.TextMarkdown
		}
	}
	return optionID
}
