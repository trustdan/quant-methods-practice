package candidates

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/domain"
)

func TestReviewedVariationsAndCanonicalKeys(t *testing.T) {
	expected := []float64{.375, .31146240234375, .3125, .301989888}
	for i, want := range expected {
		p := Local(int64(i))
		tmpl, err := Build("candidate_test", p)
		if err != nil {
			t.Fatal(err)
		}
		if tmpl.Status != domain.StatusDraft || tmpl.Approval != nil {
			t.Fatal("generation approved content")
		}
		got := *tmpl.Stages[5].ExpectedAnswer.Value
		if math.Abs(got-want) > 1e-14 {
			t.Fatalf("variation %d: got %g want %g", i, got, want)
		}
		if tmpl.SettingGroup != "binomial_fixed_independent_exact_count" {
			t.Fatal("cosmetic variation promoted transfer")
		}
		if err = bank.ValidateTemplate(tmpl, nil); err != nil {
			t.Fatal(err)
		}
	}
	if Local(-1) != Local(3) {
		t.Fatal("negative seeds must be deterministic")
	}
}
func TestCandidateStrictnessAndScope(t *testing.T) {
	data, _ := json.Marshal(Local(0))
	for _, invalid := range []string{string(data) + ` {}`, strings.TrimSuffix(string(data), "}") + `,"expected_answer":0.99}`, strings.TrimSuffix(string(data), "}") + `,"approval":{"reviewer":"AI"}}`} {
		if _, err := Decode([]byte(invalid)); err == nil {
			t.Fatalf("accepted %s", invalid)
		}
	}
	for _, change := range []func(*Proposal){func(p *Proposal) { p.FamilyID = "unknown" }, func(p *Proposal) { p.N = 1000000 }, func(p *Proposal) { p.P = math.NaN() }, func(p *Proposal) { p.P = .123 }, func(p *Proposal) { p.K = p.N }, func(p *Proposal) { p.Scenario = "" }} {
		p := Local(0)
		change(&p)
		if _, err := Build("candidate_test", p); err == nil {
			t.Fatalf("accepted invalid proposal %+v", p)
		}
	}
}
func TestSemanticReviewRequiredAndKeysRebuilt(t *testing.T) {
	p := Local(0)
	p.Scenario = "Draw four cards without replacement. Exactly two red cards." // valid numbers, wrong assumptions
	rec, err := NewRecord("candidate_test", p, Source{Mode: "manual"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	review := Review{Action: "approve", Reviewer: "human", Notes: "review findings"}
	if _, err = PrepareReview(rec, 1, review, time.Now()); err == nil {
		t.Fatal("approved without semantic attestation")
	}
	review.Action = "reject"
	rejected, err := PrepareReview(rec, 1, review, time.Now())
	if err != nil || rejected.Status != "rejected" {
		t.Fatalf("reject: %v", err)
	}
	review.Action = "approve"
	review.SemanticConfirmed = true
	if _, err = PrepareReview(rejected, 2, review, time.Now()); err != ErrConflict {
		t.Fatal("rejected version reactivated")
	}
	rec.Proposal = Local(0)
	forged := .99
	rec.Template.Stages[5].ExpectedAnswer.Value = &forged
	approved, err := PrepareReview(rec, 1, review, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if *approved.Template.Stages[5].ExpectedAnswer.Value != .375 {
		t.Fatal("proposal key became authoritative")
	}
	if rec.Status != "pending" || rec.Template.Status != domain.StatusDraft {
		t.Fatal("review mutated original")
	}
	if _, err = PrepareReview(approved, 1, review, time.Now()); err != ErrConflict {
		t.Fatal("stale review accepted")
	}
}
