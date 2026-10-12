package bank_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/drill"
)

var finalWaveSlugs = []string{
	"twosample-read-excel-output", "regout-compare-models-fit", "mreg-adjusted-r2-weak-predictor", "mreg-overall-f-vs-individual-t", "mreg-dummy-variable", "regprob-nonlinearity", "regprob-autocorrelation",
	"desc-zscore-empirical-rule", "desc-weighted-mean-price", "desc-quartiles-iqr-fences", "drv-compare-projects-risk", "drv-insurance-break-even", "drv-cumulative-from-table", "lincomb-weekly-demand-sum-vs-scale", "lincomb-portfolio-correlated", "twosample-paired-training", "twosample-independent-means-layouts", "twosample-two-proportion-ads", "regprob-outlier-influence", "regprob-omitted-variable-causation",
}

func finalWaveTemplate(t *testing.T, slug string) *domain.QuestionTemplate {
	t.Helper()
	for _, dir := range []string{"approved", "drafts"} {
		path := filepath.Join("..", "..", "curriculum", dir, slug+".json")
		if _, err := os.Stat(path); err == nil {
			tmpl, err := bank.ValidateTemplateFile(path, nil)
			if err != nil {
				t.Fatal(err)
			}
			return tmpl
		}
	}
	t.Fatalf("missing final-wave template %s", slug)
	return nil
}

func TestFinalWavesCanonicalGradingAndSnapshots(t *testing.T) {
	for _, slug := range finalWaveSlugs {
		t.Run(slug, func(t *testing.T) {
			tmpl := finalWaveTemplate(t, slug)
			s, err := drill.NewSessionManager(nil).CreateSession(tmpl, 73)
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(s.ToPublicView())
			if err != nil {
				t.Fatal(err)
			}
			for _, secret := range []string{"numeric_value", "calculation", "expected_answer", "hint_markdown", "explanation_markdown"} {
				if strings.Contains(string(data), `"`+secret+`"`) {
					t.Fatalf("unresolved public view leaks %s", secret)
				}
			}
			for i, original := range s.Stages {
				// JSON round-trip models immutable persistence of the selected options
				// and exact answer. Every distractor must remain wrong on replay.
				data, err := json.Marshal(original.Instance)
				if err != nil {
					t.Fatal(err)
				}
				var snapshot domain.StageInstance
				if err = json.Unmarshal(data, &snapshot); err != nil {
					t.Fatal(err)
				}
				for _, o := range snapshot.Options {
					state := drill.StageState{Instance: snapshot, Status: drill.StageStatusActive}
					att, invalid, _, err := drill.GradeSubmission("test", "instance", &state, domain.SubmittedAnswer{Kind: domain.StageKindChoice, OptionID: o.ID}, time.Now)
					if err != nil || invalid {
						t.Fatalf("stage %d option %s failed: %v %v", i, o.ID, err, invalid)
					}
					if att.IsCorrect != (o.ID == snapshot.ExpectedAnswer.OptionID) {
						t.Fatalf("wrong replay grade: %s", o.ID)
					}
				}
			}
		})
	}
}

func TestCalculationAdmissionRejectsKeyAndParameterDrift(t *testing.T) {
	mutations := []struct {
		name   string
		mutate func(*domain.QuestionTemplate)
	}{
		{"wrong selected key", func(q *domain.QuestionTemplate) { q.Stages[0].ExpectedAnswer.OptionID = "misconception_a" }},
		{"changed input", func(q *domain.QuestionTemplate) { q.Parameters["x"] = 120 }},
		{"misleading label", func(q *domain.QuestionTemplate) { q.Stages[0].Options[0].TextMarkdown = "$99.0000$" }},
		{"numeric collision", func(q *domain.QuestionTemplate) {
			v := 2.0
			q.Stages[0].Options[1].NumericValue = &v
			q.Stages[0].Options[1].TextMarkdown = "$2.0000$"
		}},
		{"unknown calculation", func(q *domain.QuestionTemplate) { q.Stages[0].Calculation = "invented" }},
		{"removed binding", func(q *domain.QuestionTemplate) { q.Stages[0].Calculation = "" }},
		{"missing numeric value", func(q *domain.QuestionTemplate) { q.Stages[0].Options[0].NumericValue = nil }},
		{"missing causal hint", func(q *domain.QuestionTemplate) { q.Stages[0].Options[1].HintMarkdown = nil }},
		{"answer parameter injection", func(q *domain.QuestionTemplate) { q.Parameters["z"] = 2 }},
	}
	for _, tc := range mutations {
		t.Run(tc.name, func(t *testing.T) {
			q := finalWaveTemplate(t, "desc-zscore-empirical-rule")
			tc.mutate(q)
			if err := bank.ValidateTemplate(q, nil); err == nil {
				t.Fatal("corrupted template admitted")
			}
		})
	}
}

func TestFinalWaveRegressionFitFixtures(t *testing.T) {
	// Values independently obtained from SSE/df and 1-(SSE/df)/(SST/(n-1)).
	cases := []struct {
		slug, quantity string
		want           float64
	}{
		{"regout-compare-models-fit", "se", 4.47213595499958},
		{"regout-compare-models-fit", "adjusted_r2_b", .6777777777777778},
		{"mreg-adjusted-r2-weak-predictor", "r2_b", .702},
		{"mreg-adjusted-r2-weak-predictor", "adjusted_r2_b", .6676153846153846},
		{"mreg-dummy-variable", "prediction", 58},
		{"mreg-overall-f-vs-individual-t", "f", 36},
	}
	for _, tc := range cases {
		q := finalWaveTemplate(t, tc.slug)
		found := false
		for _, s := range q.Stages {
			if s.Calculation != tc.quantity {
				continue
			}
			found = true
			for _, o := range s.Options {
				if o.ID == s.ExpectedAnswer.OptionID && o.TextMarkdown != fmt.Sprintf("$%.4f$", tc.want) {
					t.Errorf("%s/%s fixture differs", tc.slug, tc.quantity)
				}
			}
		}
		if !found {
			t.Errorf("missing quantity %s/%s", tc.slug, tc.quantity)
		}
	}
}
