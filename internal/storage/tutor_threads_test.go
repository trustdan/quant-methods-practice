package storage

import (
	"context"
	"testing"
	"time"
)

func TestTutorThreadsStorage(t *testing.T) {
	db, err := OpenMemory()
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := RunMigrations(ctx, db, nil); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	clockTime := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	store := NewStore(db, func() time.Time { return clockTime })

	// 1. Create a thread
	thread := &TutorThread{
		ID:         "th_test_1",
		SessionID:  "sess_123",
		InstanceID: "inst_binom_1",
		StageID:    "stage_calc",
		Provider:   "local",
		Model:      "llama-3-8b",
		Title:      "Understanding Binomial Independence",
	}

	if err := store.CreateTutorThread(ctx, thread); err != nil {
		t.Fatalf("CreateTutorThread failed: %v", err)
	}

	// 2. Fetch the thread
	fetched, err := store.GetTutorThread(ctx, "th_test_1")
	if err != nil {
		t.Fatalf("GetTutorThread failed: %v", err)
	}
	if fetched.ID != "th_test_1" || fetched.Title != "Understanding Binomial Independence" {
		t.Errorf("unexpected thread data: %+v", fetched)
	}
	if len(fetched.Messages) != 0 {
		t.Errorf("expected 0 messages initially, got %d", len(fetched.Messages))
	}

	// 3. Append messages
	clockTime = clockTime.Add(1 * time.Minute)
	userMsg := &TutorMessage{
		ID:        "msg_u1",
		ThreadID:  "th_test_1",
		Role:      "user",
		Content:   "Why does the formula multiply by n choose k?",
		CreatedAt: clockTime,
	}
	if err := store.AppendTutorMessage(ctx, userMsg); err != nil {
		t.Fatalf("AppendTutorMessage userMsg failed: %v", err)
	}

	clockTime = clockTime.Add(1 * time.Minute)
	asstMsg := &TutorMessage{
		ID:        "msg_a1",
		ThreadID:  "th_test_1",
		Role:      "assistant",
		Content:   "Because there are multiple sequences of outcomes that yield exactly k successes.",
		CreatedAt: clockTime,
	}
	if err := store.AppendTutorMessage(ctx, asstMsg); err != nil {
		t.Fatalf("AppendTutorMessage asstMsg failed: %v", err)
	}

	// 4. Verify thread now contains both messages in chronological order
	fetchedWithMsgs, err := store.GetTutorThread(ctx, "th_test_1")
	if err != nil {
		t.Fatalf("GetTutorThread with messages failed: %v", err)
	}
	if len(fetchedWithMsgs.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(fetchedWithMsgs.Messages))
	}
	if fetchedWithMsgs.Messages[0].Role != "user" || fetchedWithMsgs.Messages[1].Role != "assistant" {
		t.Errorf("unexpected message roles: %+v", fetchedWithMsgs.Messages)
	}

	// 5. List threads
	threads, err := store.ListTutorThreads(ctx, "inst_binom_1", "stage_calc")
	if err != nil {
		t.Fatalf("ListTutorThreads failed: %v", err)
	}
	if len(threads) != 1 || threads[0].ID != "th_test_1" {
		t.Errorf("unexpected list result: %+v", threads)
	}

	// Filter by different stage should return empty
	emptyThreads, err := store.ListTutorThreads(ctx, "inst_binom_1", "stage_other")
	if err != nil {
		t.Fatalf("ListTutorThreads empty failed: %v", err)
	}
	if len(emptyThreads) != 0 {
		t.Errorf("expected 0 threads for different stage, got %d", len(emptyThreads))
	}

	// 6. Delete thread and verify messages are cleared
	if err := store.DeleteTutorThread(ctx, "th_test_1"); err != nil {
		t.Fatalf("DeleteTutorThread failed: %v", err)
	}

	_, err = store.GetTutorThread(ctx, "th_test_1")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after deletion, got %v", err)
	}

	msgs, err := store.GetTutorMessages(ctx, "th_test_1")
	if err != nil {
		t.Fatalf("GetTutorMessages after deletion failed: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages after cascade deletion, got %d", len(msgs))
	}
}
