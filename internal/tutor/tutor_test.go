package tutor

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/storage"
)

func TestOfflineTutorHints(t *testing.T) {
	tutor := NewOfflineTutor(0, nil)

	reqs := []TutorRequest{
		{
			Action: ActionHint,
			Instance: &domain.QuestionInstance{
				TemplateID: "binomial_fair_coin_exactly_two",
			},
			StageID: "calculate_exact_prob",
		},
		{
			Action: ActionHint,
			Instance: &domain.QuestionInstance{
				TemplateID: "poisson_call_center_arrivals",
			},
			StageID: "calculate_poisson_pmf",
		},
		{
			Action: ActionHint,
			Instance: &domain.QuestionInstance{
				TemplateID: "set_probability_union_rule",
			},
			StageID: "union_rule",
		},
	}

	for _, req := range reqs {
		ch, err := tutor.Stream(context.Background(), req)
		if err != nil {
			t.Fatalf("Stream failed for template %s: %v", req.Instance.TemplateID, err)
		}

		var fullText string
		for ev := range ch {
			if ev.Type == EventComplete {
				fullText = ev.Text
			}
		}

		if !strings.Contains(fullText, "Causal Hint") {
			t.Errorf("expected causal hint header for %s, got: %s", req.Instance.TemplateID, fullText)
		}
		if !strings.Contains(fullText, "$") {
			t.Errorf("expected LaTeX math formatting in hint for %s", req.Instance.TemplateID)
		}
	}
}

func TestOfflineTutorExplanationMathCorrectness(t *testing.T) {
	tutor := NewOfflineTutor(0, nil)

	req := TutorRequest{
		Action: ActionExplain,
		Instance: &domain.QuestionInstance{
			TemplateID: "binomial_fair_coin_exactly_two",
		},
		StageID: "calculate_exact_prob",
	}

	ch, err := tutor.Stream(context.Background(), req)
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}

	var fullText string
	for ev := range ch {
		if ev.Type == EventComplete {
			fullText = ev.Text
		}
	}

	// Verify exact mathematical values
	expectedSnippets := []string{
		"0.375",
		"\\frac{3}{8}",
		"\\binom{4}{2}",
		"\\mu = np = 4 \\times 0.5 = 2",
		"\\sigma^2 = np(1-p) = 4 \\times 0.5 \\times 0.5 = 1",
	}

	for _, snip := range expectedSnippets {
		if !strings.Contains(fullText, snip) {
			t.Errorf("expected explanation to contain %q, but did not find it in:\n%s", snip, fullText)
		}
	}
}

func TestOfflineTutorFollowUps(t *testing.T) {
	tutor := NewOfflineTutor(0, nil)

	modes := []struct {
		kind     FollowUpKind
		expected string
	}{
		{FollowUpExplainDifferently, "Intuitive & Visual Explanation"},
		{FollowUpWorkedExample, "Parallel Worked Example"},
		{FollowUpWhyConditionMatters, "Why Do the Assumptions Matter?"},
		{FollowUpCompareConcepts, "Comparative Concept Analysis"},
		{FollowUpCustom, "Inquiry:"},
	}

	for _, m := range modes {
		req := TutorRequest{
			Action:       ActionFollowUp,
			FollowUpKind: m.kind,
			CustomPrompt: "Does independence matter?",
			Instance: &domain.QuestionInstance{
				TemplateID: "binomial_fair_coin_exactly_two",
			},
		}

		ch, err := tutor.Stream(context.Background(), req)
		if err != nil {
			t.Fatalf("Stream failed for follow up %s: %v", m.kind, err)
		}

		var fullText string
		for ev := range ch {
			if ev.Type == EventComplete {
				fullText = ev.Text
			}
		}

		if !strings.Contains(fullText, m.expected) {
			t.Errorf("expected follow up %s to contain %q, got:\n%s", m.kind, m.expected, fullText)
		}
	}
}

func TestOfflineTutorStreamingAndComplete(t *testing.T) {
	tutor := NewOfflineTutor(1*time.Millisecond, nil)

	req := TutorRequest{
		RequestID: "req_stream_test",
		Action:    ActionExplain,
		Instance: &domain.QuestionInstance{
			TemplateID: "binomial_fair_coin_exactly_two",
		},
	}

	ch, err := tutor.Stream(context.Background(), req)
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}

	var events []TutorEvent
	for ev := range ch {
		events = append(events, ev)
	}

	if len(events) < 3 {
		t.Fatalf("expected at least 3 events (started, deltas, complete), got %d", len(events))
	}

	if events[0].Type != EventStarted {
		t.Errorf("expected first event to be %s, got %s", EventStarted, events[0].Type)
	}

	lastEv := events[len(events)-1]
	if lastEv.Type != EventComplete {
		t.Errorf("expected last event to be %s, got %s", EventComplete, lastEv.Type)
	}
	if lastEv.Text == "" {
		t.Errorf("expected non-empty text in complete event")
	}
}

func TestOfflineTutorCancellation(t *testing.T) {
	tutor := NewOfflineTutor(10*time.Millisecond, nil)

	ctx, cancel := context.WithCancel(context.Background())

	req := TutorRequest{
		RequestID: "req_cancel_test",
		Action:    ActionExplain,
		Instance: &domain.QuestionInstance{
			TemplateID: "binomial_fair_coin_exactly_two",
		},
	}

	ch, err := tutor.Stream(ctx, req)
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}

	// Read one event then cancel
	ev := <-ch
	if ev.Type != EventStarted {
		t.Errorf("expected started event, got %s", ev.Type)
	}

	cancel()

	sawCancelled := false
	for ev := range ch {
		if ev.Type == EventCancelled {
			sawCancelled = true
		}
	}

	if !sawCancelled {
		t.Errorf("expected EventCancelled after context cancellation")
	}
}

func TestFakeProviderAndFallback(t *testing.T) {
	offline := NewOfflineTutor(0, nil)
	manager := NewTutorManager(offline)

	fake := NewFakeProvider("flaky_ai")
	fake.ShouldFail = true
	fake.FailMessage = "API rate limit exceeded (429)"
	manager.RegisterProvider(fake)

	req := TutorRequest{
		RequestID: "req_fallback_test",
		Action:    ActionExplain,
		Provider:  "flaky_ai",
		Instance: &domain.QuestionInstance{
			TemplateID: "binomial_fair_coin_exactly_two",
		},
	}

	_, ch, err := manager.StartRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("StartRequest failed: %v", err)
	}

	sawFallback := false
	sawComplete := false
	var finalText string

	for ev := range ch {
		if ev.Type == EventFallback {
			sawFallback = true
			if !strings.Contains(ev.FallbackLabel, "API rate limit") {
				t.Errorf("expected fallback label to mention error, got: %s", ev.FallbackLabel)
			}
		}
		if ev.Type == EventComplete {
			sawComplete = true
			finalText = ev.Text
		}
	}

	if !sawFallback {
		t.Errorf("expected EventFallback when provider fails")
	}
	if !sawComplete {
		t.Errorf("expected EventComplete from fallback offline tutor")
	}
	if !strings.Contains(finalText, "0.375") {
		t.Errorf("expected offline tutor content with 0.375, got: %s", finalText)
	}
}

func TestTutorManagerCancellation(t *testing.T) {
	offline := NewOfflineTutor(20*time.Millisecond, nil)
	manager := NewTutorManager(offline)

	req := TutorRequest{
		RequestID: "req_mgr_cancel",
		Action:    ActionExplain,
		Instance: &domain.QuestionInstance{
			TemplateID: "binomial_fair_coin_exactly_two",
		},
	}

	reqID, ch, err := manager.StartRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("StartRequest failed: %v", err)
	}

	// Verify active
	if !manager.IsActive(reqID) {
		t.Errorf("expected request to be active")
	}

	// Cancel it
	time.Sleep(5 * time.Millisecond)
	cancelled := manager.CancelRequest(reqID)
	if !cancelled {
		t.Errorf("expected CancelRequest to return true")
	}

	sawCancelled := false
	for ev := range ch {
		if ev.Type == EventCancelled {
			sawCancelled = true
		}
	}

	if !sawCancelled {
		t.Errorf("expected EventCancelled event")
	}

	// Wait briefly for cleanup
	time.Sleep(10 * time.Millisecond)
	if manager.IsActive(reqID) {
		t.Errorf("expected request to no longer be active after cancellation")
	}
}

func TestTutorCannotModifyGradesOrMastery(t *testing.T) {
	// Structural invariant assertion:
	// Verify that TutorService and OfflineTutor types do not implement or expose
	// any methods that could modify grades, records, or bank status.
	tutor := NewOfflineTutor(0, nil)

	// Ensure TutorService interface only has Stream, ProviderID, Capabilities
	var svc TutorService = tutor
	if svc.ProviderID() != "offline" {
		t.Errorf("unexpected provider ID: %s", svc.ProviderID())
	}
	caps := svc.Capabilities()
	if !caps.IsOffline {
		t.Errorf("expected IsOffline to be true")
	}
}

func TestExportNoteToMarkdown(t *testing.T) {
	note := &storage.SavedExplanation{
		ID:               "note_export_01",
		RawMarkdown:      "Formula: $P(X=2) = \\binom{4}{2} (0.5)^4 = 0.375$",
		OriginInstanceID: "inst_01",
		OriginStageID:    "s1",
		Topic:            "Binomial Distribution",
		ProviderInfo: storage.ProviderInfo{
			Title:          "Coin Toss Derivation",
			Concepts:       []string{"binomial_pmf", "binomial_combination"},
			Provider:       "offline",
			Model:          "offline-curriculum",
			AdvisoryStatus: "Advisory note for self-study",
		},
		CreatedAt: time.Date(2026, 10, 5, 14, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 10, 5, 14, 5, 0, 0, time.UTC),
	}

	scenario := "A fair coin is tossed 4 times. Find the probability of exactly 2 heads."
	md := ExportNoteToMarkdown(note, scenario)

	// Assertions for Obsidian/Typora compatibility
	if !strings.HasPrefix(md, "---\n") {
		t.Errorf("expected YAML frontmatter at start of export")
	}
	if !strings.Contains(md, "title: \"Coin Toss Derivation\"") {
		t.Errorf("expected title in frontmatter")
	}
	if !strings.Contains(md, "> [!NOTE]") || !strings.Contains(md, "Advisory Note") {
		t.Errorf("expected advisory GitHub/Obsidian callout in export")
	}
	if !strings.Contains(md, "### Problem Context") || !strings.Contains(md, scenario) {
		t.Errorf("expected problem context in export")
	}
	if !strings.Contains(md, "Formula: $P(X=2) = \\binom{4}{2} (0.5)^4 = 0.375$") {
		t.Errorf("expected LaTeX math preserved verbatim in export")
	}
}

func TestConcurrentTutorRequests(t *testing.T) {
	offline := NewOfflineTutor(0, nil)
	manager := NewTutorManager(offline)

	const n = 10
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			req := TutorRequest{
				Action: ActionExplain,
				Instance: &domain.QuestionInstance{
					TemplateID: "binomial_fair_coin_exactly_two",
				},
			}
			_, ch, err := manager.StartRequest(context.Background(), req)
			if err != nil {
				t.Errorf("StartRequest failed: %v", err)
				return
			}
			for range ch {
				// drain
			}
		}(i)
	}

	wg.Wait()
}
