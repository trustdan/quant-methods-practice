package worksheets

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/candidates"
	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/drill"
)

func reference(t *testing.T) *domain.QuestionTemplate {
	t.Helper()
	b, err := bank.LoadActiveBank("../../curriculum/approved", nil)
	if err != nil {
		t.Fatal(err)
	}
	tmpl, ok := b.Get("binomial_fair_coin_exactly_two")
	if !ok {
		t.Fatal("missing reference")
	}
	return tmpl
}
func correct(r *Record) map[string]domain.SubmittedAnswer {
	answers := map[string]domain.SubmittedAnswer{}
	for _, item := range r.Items {
		if item.State.Status == drill.StageStatusCompleted {
			continue
		}
		a := item.State.Instance.ExpectedAnswer
		sub := domain.SubmittedAnswer{Kind: item.State.Instance.Kind, OptionID: a.OptionID}
		if a.Value != nil {
			sub.NumericRaw = fmt.Sprintf("%.17g", *a.Value)
		}
		answers[item.Key] = sub
	}
	return answers
}
func TestCSVBoundaries(t *testing.T) {
	valid := "experiment_id,successes\r\nrun_1,2\r\nrun_2,4\r\n"
	if d, err := ParseCSV(valid); err != nil || len(d.Rows) != 2 {
		t.Fatalf("%v %v", d, err)
	}
	for _, raw := range []string{"", "experiment_id,successes\n", "id,successes\na,2", "experiment_id,successes,extra\na,2,x", "experiment_id,successes\na,2\na,3", "experiment_id,successes\n=SUM(A1),2", "experiment_id,successes\n@cmd,2", "experiment_id,successes\n../path,2", "experiment_id,successes\na,=2", "experiment_id,successes\na,2.0", "experiment_id,successes\na,5", "experiment_id,successes\na,-1", "experiment_id,successes\na, 2", "experiment_id,successes\na,2,3", "experiment_id,successes\n\"a\ncmd\",2", "experiment_id,successes\n\xff,2", strings.Repeat("a", MaxCSVBytes+1)} {
		if _, err := ParseCSV(raw); err == nil {
			t.Fatalf("unsafe CSV accepted: %.80q", raw)
		}
	}
	var raw strings.Builder
	raw.WriteString("experiment_id,successes\n")
	for i := 0; i < MaxRows; i++ {
		fmt.Fprintf(&raw, "run_%d,2\n", i)
	}
	if _, err := ParseCSV(raw.String()); err != nil {
		t.Fatal(err)
	}
	raw.WriteString("overflow,2\n")
	if _, err := ParseCSV(raw.String()); err == nil {
		t.Fatal("row cap bypassed")
	}
}

func TestFullFormMatchesProgressiveForEveryApprovedQuestion(t *testing.T) {
	b, err := bank.LoadActiveBank("../../curriculum/approved", nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	for _, tmpl := range b.List() {
		t.Run(tmpl.ID, func(t *testing.T) {
			r, err := New("worksheet_test", tmpl, 43, nil, now)
			if err != nil {
				t.Fatal(err)
			}
			answers := correct(r)
			next, err := Apply(r, Command{ID: "submit", Revision: 1, Type: "submit", Answers: answers}, now)
			if err != nil {
				t.Fatal(err)
			}
			if next.Status != "completed" {
				t.Fatal(next.Status)
			}
			for i, item := range r.Items {
				stage := item.State
				att, invalid, _, err := drill.GradeSubmission(r.ID, item.QuestionID, &stage, answers[item.Key], func() time.Time { return now })
				got := next.Items[i].State.Attempts[0]
				if err != nil || invalid || !att.IsCorrect || got.IsCorrect != att.IsCorrect || got.FeedbackMarkdown != att.FeedbackMarkdown {
					t.Fatalf("different progressive/full result: %+v vs %+v", got, att)
				}
				if len(got.Assistance) != 1 || got.Assistance[0] != domain.AssistanceFullSolution {
					t.Fatal("unsupported independent evidence")
				}
			}
		})
	}
}

func TestValidationIsAtomicAndPublicKeysStayHidden(t *testing.T) {
	r, err := New("worksheet_test", reference(t), 42, nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(r)
	answers := correct(r)
	for _, item := range r.Items {
		if item.State.Instance.Kind == domain.StageKindNumeric {
			answers[item.Key] = domain.SubmittedAnswer{Kind: domain.StageKindNumeric, NumericRaw: "1/0"}
		}
	}
	if _, err = Apply(r, Command{ID: "invalid", Revision: 1, Type: "submit", Answers: answers}, time.Now()); err == nil {
		t.Fatal("invalid field consumed attempts")
	}
	after, _ := json.Marshal(r)
	if string(before) != string(after) {
		t.Fatal("input record mutated on validation failure")
	}
	raw, _ := json.Marshal(Public(r))
	if strings.Contains(string(raw), "expected_answer") || strings.Contains(string(raw), "explanation_markdown") || strings.Contains(string(raw), "hint_markdown") {
		t.Fatal("unearned key leaked")
	}
	if strings.Contains(Markdown(r), "Correct!") {
		t.Fatal("worksheet export reveals key")
	}
	answers = correct(r)
	delete(answers, r.Items[0].Key)
	if _, err = Apply(r, Command{ID: "partial", Revision: 1, Type: "submit", Answers: answers}, time.Now()); err == nil {
		t.Fatal("partial full solution accepted")
	}
	answers["forged"] = domain.SubmittedAnswer{Kind: domain.StageKindChoice, OptionID: "x"}
	if _, err = Apply(r, Command{ID: "extra", Revision: 1, Type: "save_draft", Answers: answers}, time.Now()); err == nil {
		t.Fatal("unknown field accepted")
	}
}

func TestOwnNumericKeyHintRetryAndSnapshotIsolation(t *testing.T) {
	p := candidates.Local(1)
	tmpl, err := candidates.Build("candidate_test", p)
	if err != nil {
		t.Fatal(err)
	}
	tmpl.Status = domain.StatusActive
	tmpl.Approval = &domain.ApprovalRecord{Reviewer: "test fixture"}
	r, err := New("worksheet_test", tmpl, 7, nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	tmpl.Title = "changed bank wording"
	*tmpl.Stages[5].ExpectedAnswer.Value = .99
	answers := correct(r)
	numericKey := ""
	for _, item := range r.Items {
		if item.State.Instance.Kind == domain.StageKindNumeric {
			numericKey = item.Key
			answers[item.Key] = domain.SubmittedAnswer{Kind: domain.StageKindNumeric, NumericRaw: "0"}
		}
	}
	first, err := Apply(r, Command{ID: "first", Revision: 1, Type: "submit", Answers: answers}, time.Now())
	if err != nil || first.Status != "retry" {
		t.Fatalf("%v %v", first, err)
	}
	if first.Questions[0].Title == tmpl.Title {
		t.Fatal("snapshot aliases bank")
	}
	second, err := Apply(first, Command{ID: "second", Revision: 2, Type: "submit", Answers: map[string]domain.SubmittedAnswer{numericKey: {Kind: domain.StageKindNumeric, NumericRaw: "0"}}}, time.Now())
	if err != nil || second.Status != "completed" {
		t.Fatalf("%v %v", second, err)
	}
	for _, item := range second.Items {
		if item.Key == numericKey {
			state := item.State
			feedback := state.LastFeedback
			if !state.Revealed || !strings.Contains(feedback, "0.3114624023") || strings.Contains(feedback, "0.375") || len(state.Attempts) != 2 {
				t.Fatalf("wrong reveal: %+v", state)
			}
			if state.Attempts[1].Assistance[1] != domain.AssistanceHint || state.Attempts[1].Assistance[2] != domain.AssistanceRetry {
				t.Fatal("lost assistance")
			}
		}
	}
}

func TestDatasetCasesSeparateDescriptionFromHypotheticalModel(t *testing.T) {
	d, err := ParseCSV("experiment_id,successes\na,2\nb,0\nc,2\nd,4\ne,1\n")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = New("worksheet_dataset", reference(t), 42, d, time.Now()); err == nil {
		t.Fatal("unreviewed dataset accepted")
	}
	d.Reviewer = "test reviewer"
	d.SourceNote = "synthetic fixture: one row = four tosses; count heads"
	r, err := New("worksheet_dataset", reference(t), 42, d, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	d.Rows[0].Successes = 0
	if r.Dataset.Rows[0].Successes != 2 || len(r.Questions) != 2 {
		t.Fatal("dataset snapshot alias or missing model comparison")
	}
	answers := correct(r)
	for _, item := range r.Items {
		if item.State.Instance.ID == "observed_value" {
			answers[item.Key] = domain.SubmittedAnswer{Kind: domain.StageKindNumeric, NumericRaw: "2/5"}
		}
	}
	next, err := Apply(r, Command{ID: "grade_dataset", Revision: 1, Type: "submit", Answers: answers}, time.Now())
	if err != nil || next.Status != "completed" {
		t.Fatalf("%v %v", next, err)
	}
	if len(next.Items) != 11 {
		t.Fatalf("expected observed+theoretical fields: %d", len(next.Items))
	}
	if r.Dataset.Approval == nil || r.Dataset.Approval.Reviewer != d.Reviewer || d.Approval != nil {
		t.Fatal("missing or aliased semantic approval provenance")
	}
	// A method/assumption error must receive the appropriate causal hint.
	for _, item := range r.Items {
		if item.State.Instance.ID == "assumptions" {
			stage := item.State
			att, invalid, _, err := drill.GradeSubmission(r.ID, item.QuestionID, &stage, domain.SubmittedAnswer{Kind: domain.StageKindChoice, OptionID: "option_1"}, time.Now)
			if err != nil || invalid || att.IsCorrect || !strings.Contains(att.FeedbackMarkdown, "cannot prove") {
				t.Fatal("wrong inference received no causal hint")
			}
		}
	}
}
