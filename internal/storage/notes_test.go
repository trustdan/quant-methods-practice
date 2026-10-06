package storage

import (
	"context"
	"testing"
	"time"
)

func TestSaveAndGetExplanation(t *testing.T) {
	db, err := OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := RunMigrations(ctx, db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	fixedTime := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	store := NewStore(db, func() time.Time { return fixedTime })

	note := &SavedExplanation{
		ID:               "note_binomial_pmf_01",
		RawMarkdown:      "## Binomial PMF\nFormula: $P(X=2) = \\binom{4}{2} (0.5)^2 (0.5)^2 = 0.375$",
		OriginInstanceID: "inst_test_01",
		OriginStageID:    "calculate_exact_prob",
		Topic:            "Binomial Distribution",
		ProviderInfo: ProviderInfo{
			Title:          "Worked Calculation for Coin Tosses",
			Concepts:       []string{"binomial_pmf", "binomial_combination"},
			Provider:       "offline",
			Model:          "offline-curriculum",
			Route:          "offline",
			ThreadID:       "thread_123",
			AdvisoryStatus: "Advisory note for self-study",
		},
	}

	if err := store.SaveExplanation(ctx, note); err != nil {
		t.Fatalf("SaveExplanation failed: %v", err)
	}

	fetched, err := store.GetExplanation(ctx, note.ID)
	if err != nil {
		t.Fatalf("GetExplanation failed: %v", err)
	}

	if fetched.ID != note.ID {
		t.Errorf("expected ID %q, got %q", note.ID, fetched.ID)
	}
	if fetched.RawMarkdown != note.RawMarkdown {
		t.Errorf("expected RawMarkdown %q, got %q", note.RawMarkdown, fetched.RawMarkdown)
	}
	if fetched.Topic != note.Topic {
		t.Errorf("expected Topic %q, got %q", note.Topic, fetched.Topic)
	}
	if fetched.ProviderInfo.Title != "Worked Calculation for Coin Tosses" {
		t.Errorf("expected Title %q, got %q", "Worked Calculation for Coin Tosses", fetched.ProviderInfo.Title)
	}
	if len(fetched.ProviderInfo.Concepts) != 2 || fetched.ProviderInfo.Concepts[0] != "binomial_pmf" {
		t.Errorf("expected concepts [binomial_pmf, ...], got %v", fetched.ProviderInfo.Concepts)
	}
	if fetched.CreatedAt.UTC() != fixedTime {
		t.Errorf("expected CreatedAt %v, got %v", fixedTime, fetched.CreatedAt.UTC())
	}
}

func TestSaveExplanationIdempotency(t *testing.T) {
	db, err := OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := RunMigrations(ctx, db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	currentTime := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	store := NewStore(db, func() time.Time { return currentTime })

	note := &SavedExplanation{
		ID:               "note_idempotent_test",
		RawMarkdown:      "Initial version",
		OriginInstanceID: "inst_01",
		OriginStageID:    "stage_01",
		Topic:            "Probability",
		ProviderInfo: ProviderInfo{
			Title: "Initial Title",
		},
	}

	if err := store.SaveExplanation(ctx, note); err != nil {
		t.Fatalf("First SaveExplanation failed: %v", err)
	}

	// Update note with same ID
	currentTime = currentTime.Add(5 * time.Minute)
	note.RawMarkdown = "Updated version with more detail"
	note.ProviderInfo.Title = "Updated Title"

	if err := store.SaveExplanation(ctx, note); err != nil {
		t.Fatalf("Second SaveExplanation failed: %v", err)
	}

	fetched, err := store.GetExplanation(ctx, note.ID)
	if err != nil {
		t.Fatalf("GetExplanation failed: %v", err)
	}
	if fetched.RawMarkdown != "Updated version with more detail" {
		t.Errorf("expected updated markdown, got %q", fetched.RawMarkdown)
	}
	if fetched.ProviderInfo.Title != "Updated Title" {
		t.Errorf("expected updated title, got %q", fetched.ProviderInfo.Title)
	}

	// Ensure there is only 1 record in table
	all, err := store.ListExplanations(ctx, "", "")
	if err != nil {
		t.Fatalf("ListExplanations failed: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected exactly 1 explanation after update, got %d", len(all))
	}
}

func TestListExplanationsFiltering(t *testing.T) {
	db, err := OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := RunMigrations(ctx, db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	store := NewStore(db, nil)

	notes := []*SavedExplanation{
		{
			ID:               "n1",
			RawMarkdown:      "Coin tossing experiment with $n=4, p=0.5$",
			OriginInstanceID: "inst1",
			OriginStageID:    "s1",
			Topic:            "Binomial Distribution",
			ProviderInfo: ProviderInfo{
				Title:    "Binomial Basics",
				Concepts: []string{"binomial_pmf"},
			},
		},
		{
			ID:               "n2",
			RawMarkdown:      "Arrival rate $\\lambda=3$ per hour in Poisson process",
			OriginInstanceID: "inst2",
			OriginStageID:    "s2",
			Topic:            "Poisson Distribution",
			ProviderInfo: ProviderInfo{
				Title:    "Call Center Arrivals",
				Concepts: []string{"poisson_rate", "exponential_timing"},
			},
		},
		{
			ID:               "n3",
			RawMarkdown:      "Addition rule $P(A \\cup B) = P(A) + P(B) - P(A \\cap B)$",
			OriginInstanceID: "inst3",
			OriginStageID:    "s3",
			Topic:            "Foundations",
			ProviderInfo: ProviderInfo{
				Title:    "Union of Events",
				Concepts: []string{"set_union", "inclusion_exclusion"},
			},
		},
	}

	for _, n := range notes {
		if err := store.SaveExplanation(ctx, n); err != nil {
			t.Fatalf("SaveExplanation failed for %s: %v", n.ID, err)
		}
	}

	// 1. List all
	all, err := store.ListExplanations(ctx, "", "")
	if err != nil {
		t.Fatalf("ListExplanations failed: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 notes, got %d", len(all))
	}

	// 2. Filter by topic "poisson"
	poissonNotes, err := store.ListExplanations(ctx, "", "poisson")
	if err != nil {
		t.Fatalf("ListExplanations by topic failed: %v", err)
	}
	if len(poissonNotes) != 1 || poissonNotes[0].ID != "n2" {
		t.Errorf("expected note n2 for topic poisson, got %v", poissonNotes)
	}

	// 3. Search query in concept "inclusion_exclusion"
	incNotes, err := store.ListExplanations(ctx, "inclusion", "")
	if err != nil {
		t.Fatalf("ListExplanations by concept query failed: %v", err)
	}
	if len(incNotes) != 1 || incNotes[0].ID != "n3" {
		t.Errorf("expected note n3 for concept query, got %v", incNotes)
	}

	// 4. Search query in text "lambda"
	lambdaNotes, err := store.ListExplanations(ctx, "lambda", "")
	if err != nil {
		t.Fatalf("ListExplanations by text query failed: %v", err)
	}
	if len(lambdaNotes) != 1 || lambdaNotes[0].ID != "n2" {
		t.Errorf("expected note n2 for text query, got %v", lambdaNotes)
	}
}

func TestDeleteExplanation(t *testing.T) {
	db, err := OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := RunMigrations(ctx, db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	store := NewStore(db, nil)

	note := &SavedExplanation{
		ID:               "to_delete",
		RawMarkdown:      "Temporary note",
		OriginInstanceID: "inst1",
		OriginStageID:    "s1",
		Topic:            "Test",
	}

	if err := store.SaveExplanation(ctx, note); err != nil {
		t.Fatalf("SaveExplanation failed: %v", err)
	}

	if err := store.DeleteExplanation(ctx, "to_delete"); err != nil {
		t.Fatalf("DeleteExplanation failed: %v", err)
	}

	_, err = store.GetExplanation(ctx, "to_delete")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after deletion, got %v", err)
	}
}

func TestTutorDraftSaveGetClear(t *testing.T) {
	db, err := OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := RunMigrations(ctx, db, nil); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	store := NewStore(db, nil)

	draft := &TutorDraft{
		ID:           "draft_sess_stage_01",
		ContextJSON:  `{"prompt":"Explain binomial moments"}`,
		RecoveryText: "The mean of a Binomial distribution is $\\mu = np$...",
	}

	if err := store.SaveTutorDraft(ctx, draft); err != nil {
		t.Fatalf("SaveTutorDraft failed: %v", err)
	}

	fetched, err := store.GetTutorDraft(ctx, draft.ID)
	if err != nil {
		t.Fatalf("GetTutorDraft failed: %v", err)
	}
	if fetched.RecoveryText != draft.RecoveryText {
		t.Errorf("expected RecoveryText %q, got %q", draft.RecoveryText, fetched.RecoveryText)
	}

	if err := store.ClearTutorDraft(ctx, draft.ID); err != nil {
		t.Fatalf("ClearTutorDraft failed: %v", err)
	}

	_, err = store.GetTutorDraft(ctx, draft.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after clear, got %v", err)
	}
}
