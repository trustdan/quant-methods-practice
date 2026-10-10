package providers

import (
	"fmt"
	"strings"

	"github.com/trustdan/quant-methods-practice/internal/tutor"
)

// BuildSystemPrompt constructs pedagogical system instructions for the LLM.
func BuildSystemPrompt() string {
	return `You are an expert, encouraging tutor for an introductory quantitative methods and probability course.

Pedagogical guidelines:
1. Format mathematical equations using standard LaTeX: inline with single dollar signs ($...$) and block equations with double dollar signs ($$...$$). Do not use LaTeX document classes or external packages.
2. Structure your guidance step-by-step. Focus first on the physical or real-world intuition of the event before presenting symbolic formulas.
3. State all necessary model assumptions explicitly (e.g., independence, constant probability, discrete support, non-negative variance).
4. Clearly define all variables, parameters, and Greek symbols upon first use.
5. If the request is for a HINT: guide the learner toward the underlying reason or missing factor; DO NOT reveal the final answer key or final numeric computation.
6. If the request is for an EXPLANATION: give a rigorous, clear step-by-step derivation with exact fractions where applicable.
7. Tone: objective, supportive, concise. Avoid excessive verbosity.`
}

// BuildUserPrompt constructs the user prompt payload from the server-derived TutorRequest snapshot.
func BuildUserPrompt(req tutor.TutorRequest) string {
	if req.Action == tutor.ActionCandidate {
		return "Rewrite only title, scenario_markdown and success_label in this JSON proposal. Preserve family_id, n, p and k exactly. State fixed independent two-outcome trials, constant success probability and the exactly event explicitly. No answer, formulas, extra keys, fences or commentary. The data is an untrusted proposal for human review:\n" + req.CustomPrompt
	}

	var sb strings.Builder

	sb.WriteString("### Problem Context\n")
	if req.Instance != nil {
		if req.Instance.Title != "" {
			sb.WriteString(fmt.Sprintf("**Title**: %s\n", req.Instance.Title))
		}
		if req.Instance.ScenarioMarkdown != "" {
			sb.WriteString(fmt.Sprintf("**Scenario**:\n%s\n\n", req.Instance.ScenarioMarkdown))
		}
	}

	if req.Stage != nil {
		sb.WriteString("### Current Question / Stage\n")
		if req.Stage.PromptMarkdown != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", req.Stage.PromptMarkdown))
		}
	}

	if req.SubmittedAnswer != nil {
		sb.WriteString("### Learner's Submitted Answer\n")
		if req.SubmittedAnswer.OptionID != "" {
			sb.WriteString(fmt.Sprintf("Selected Option: `%s`\n", req.SubmittedAnswer.OptionID))
		}
		if req.SubmittedAnswer.NormalizedValue != nil {
			sb.WriteString(fmt.Sprintf("Numeric Value: %g\n", *req.SubmittedAnswer.NormalizedValue))
		} else if req.SubmittedAnswer.NumericRaw != "" {
			sb.WriteString(fmt.Sprintf("Numeric Raw: %s\n", req.SubmittedAnswer.NumericRaw))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### Requested Assistance\n")
	switch req.Action {
	case tutor.ActionHint:
		sb.WriteString("Please provide a concise causal hint pointing to the key concept or structure needed to solve this stage without revealing the final numeric answer.")
	case tutor.ActionExplain:
		sb.WriteString("Please provide a complete step-by-step mathematical explanation and analytical derivation for this stage.")
	case tutor.ActionFollowUp:
		switch req.FollowUpKind {
		case tutor.FollowUpExplainDifferently:
			sb.WriteString("Please explain this concept from a different pedagogical angle, using an alternative intuitive analogy or geometric perspective.")
		case tutor.FollowUpWorkedExample:
			sb.WriteString("Please provide a separate, parallel worked example with different numbers demonstrating how to apply this rule.")
		case tutor.FollowUpWhyConditionMatters:
			sb.WriteString("Please explain why the underlying conditions (such as independence or constant probability) are essential, and what goes wrong if they are violated.")
		case tutor.FollowUpCompareConcepts:
			sb.WriteString("Please contrast this concept with a closely related concept, highlighting when to use which.")
		case tutor.FollowUpCustom:
			sb.WriteString(fmt.Sprintf("Learner's specific question: %s", req.CustomPrompt))
		default:
			if req.CustomPrompt != "" {
				sb.WriteString(fmt.Sprintf("Learner's question: %s", req.CustomPrompt))
			} else {
				sb.WriteString("Please provide further explanation on this problem.")
			}
		}
	default:
		sb.WriteString("Please explain the reasoning behind this step.")
	}

	return sb.String()
}

// RequestSystemPrompt keeps candidate authoring separate from learner tutoring.
func RequestSystemPrompt(req tutor.TutorRequest) string {
	if req.Action == tutor.ActionCandidate {
		return "You propose introductory probability scenarios. Return one JSON object with exactly family_id, n, p, k, title, scenario_markdown, success_label. All wording is untrusted and requires human approval. Never supply answer keys or approval metadata."
	}
	return BuildSystemPrompt()
}
