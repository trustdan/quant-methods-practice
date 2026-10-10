package worksheets

import (
	"fmt"
	"strings"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/drill"
)

type PublicQuestion struct {
	ID              string                 `json:"id"`
	TemplateID      string                 `json:"template_id"`
	TemplateVersion int                    `json:"template_version"`
	Seed            int64                  `json:"seed"`
	Title           string                 `json:"title"`
	Scenario        string                 `json:"scenario_markdown"`
	Assumptions     []string               `json:"assumptions"`
	Parameters      map[string]interface{} `json:"parameters"`
}
type PublicItem struct {
	Key           string                    `json:"key"`
	QuestionID    string                    `json:"question_id"`
	StageID       string                    `json:"stage_id"`
	Kind          domain.StageKind          `json:"kind"`
	Prompt        string                    `json:"prompt_markdown"`
	Options       []drill.PublicOptionView  `json:"options"`
	NumericPolicy *domain.NumericPolicy     `json:"numeric_policy,omitempty"`
	Status        drill.StageProgressStatus `json:"status"`
	Draft         *domain.SubmittedAnswer   `json:"draft_answer,omitempty"`
	Attempts      []domain.StageAttempt     `json:"attempts"`
	Expected      *domain.ExpectedAnswer    `json:"expected_answer,omitempty"`
	Explanation   string                    `json:"explanation_markdown,omitempty"`
}
type View struct {
	ID        string           `json:"id"`
	Revision  int              `json:"revision"`
	Mode      string           `json:"mode"`
	Status    string           `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Evidence  string           `json:"evidence"`
	Questions []PublicQuestion `json:"questions"`
	Items     []PublicItem     `json:"items"`
	Dataset   *Dataset         `json:"dataset,omitempty"`
}

func Public(r *Record) View {
	v := View{ID: r.ID, Revision: r.Revision, Mode: r.Mode, Status: r.Status, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, Evidence: "Supported full-form practice; kept separate from independent mastery and transfer.", Dataset: r.Dataset, Questions: []PublicQuestion{}, Items: []PublicItem{}}
	if r.Dataset != nil {
		v.Evidence = "Supported reviewed-case practice: answer keys were shown during case review. Kept separate from independent mastery and transfer."
	}
	for _, q := range r.Questions {
		v.Questions = append(v.Questions, PublicQuestion{q.ID, q.TemplateID, q.TemplateVersion, q.Seed, q.Title, q.ScenarioMarkdown, q.Assumptions, q.Parameters})
	}
	for _, item := range r.Items {
		stage := item.State
		p := PublicItem{Key: item.Key, QuestionID: item.QuestionID, StageID: stage.Instance.ID, Kind: stage.Instance.Kind, Prompt: stage.Instance.PromptMarkdown, Status: stage.Status, Draft: stage.DraftAnswer, Attempts: stage.Attempts, NumericPolicy: stage.Instance.NumericPolicy, Options: []drill.PublicOptionView{}}
		for _, opt := range stage.Instance.Options {
			p.Options = append(p.Options, drill.PublicOptionView{ID: opt.ID, TextMarkdown: opt.TextMarkdown})
		}
		if stage.Status == drill.StageStatusCompleted {
			answer := stage.Instance.ExpectedAnswer
			p.Expected = &answer
			p.Explanation = stage.Instance.ExplanationMarkdown
		}
		v.Items = append(v.Items, p)
	}
	return v
}

// Markdown contains the public worksheet and only feedback already earned by
// submission. Raw HTML is neutralized for consumers that render Markdown.
func Markdown(r *Record) string {
	v := Public(r)
	var out strings.Builder
	fmt.Fprintf(&out, "# Practice worksheet\n\n%s\n\nStatus: %s. Revision: %d. Created: %s.\n\n", v.Evidence, v.Status, v.Revision, v.CreatedAt.Format(time.RFC3339))
	if v.Dataset != nil {
		fmt.Fprintf(&out, "Dataset reviewed by %s. Source/row meaning: %s\n\n", v.Dataset.Reviewer, v.Dataset.SourceNote)
	}
	for _, q := range v.Questions {
		fmt.Fprintf(&out, "## %s\n\n%s\n\nTemplate: %s v%d. Seed: %d.\n\n", q.Title, q.Scenario, q.TemplateID, q.TemplateVersion, q.Seed)
		for _, a := range q.Assumptions {
			fmt.Fprintf(&out, "- %s\n", a)
		}
		out.WriteString("\n")
		for _, item := range v.Items {
			if item.QuestionID != q.ID {
				continue
			}
			fmt.Fprintf(&out, "### %s\n\n%s\n\n", item.StageID, item.Prompt)
			for _, opt := range item.Options {
				fmt.Fprintf(&out, "- %s: %s\n", opt.ID, opt.TextMarkdown)
			}
			out.WriteString("\n")
			if item.NumericPolicy != nil {
				fmt.Fprintf(&out, "Numeric policy v%d: absolute tolerance %g, relative tolerance %g.\n\n", item.NumericPolicy.Version, item.NumericPolicy.AbsoluteTolerance, item.NumericPolicy.RelativeTolerance)
			}
			out.WriteString("Answer: ____________________\n\n")
			for _, att := range item.Attempts {
				answer := att.SubmittedAnswer.NumericRaw
				if att.SubmittedAnswer.Kind == domain.StageKindChoice {
					answer = att.SubmittedAnswer.OptionID
				}
				fmt.Fprintf(&out, "Attempt %d (%s): %s. Correct: %t. Assistance: %v.\n\n%s\n\n", att.AttemptNumber, att.CreatedAt.Format(time.RFC3339), answer, att.IsCorrect, att.Assistance, att.FeedbackMarkdown)
			}
		}
	}
	if v.Dataset != nil {
		out.WriteString("## Recorded experiments\n\nexperiment_id,successes\n\n")
		for _, row := range v.Dataset.Rows {
			fmt.Fprintf(&out, "%s,%d\n", row.ExperimentID, row.Successes)
		}
	}
	return strings.NewReplacer("<", "&lt;", ">", "&gt;").Replace(out.String())
}
