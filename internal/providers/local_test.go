package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/auth"
	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/tutor"
)

func TestLocalLLMAdapterLoopbackEnforcement(t *testing.T) {
	adapter := NewLocalLLMAdapter("")
	if adapter.ProviderID() != auth.RouteLocal {
		t.Errorf("expected provider ID %q, got %q", auth.RouteLocal, adapter.ProviderID())
	}
	if !adapter.Capabilities().IsOffline {
		t.Errorf("expected IsOffline to be true for local adapter")
	}

	// Valid loopback addresses
	validURLs := []string{
		"http://localhost:1234",
		"http://127.0.0.1:11434",
		"http://[::1]:8080",
		"http://localhost:5000/v1",
	}
	for _, u := range validURLs {
		if err := adapter.SetBaseURL(u); err != nil {
			t.Errorf("expected valid loopback URL %q to be accepted, got error: %v", u, err)
		}
	}

	// Invalid external addresses (should be rejected for privacy/offline safety)
	invalidURLs := []string{
		"http://api.openai.com",
		"https://remote-server.com:1234",
		"http://192.168.1.50:1234",
		"http://10.0.0.1:11434",
	}
	for _, u := range invalidURLs {
		if err := adapter.SetBaseURL(u); err == nil {
			t.Errorf("expected non-loopback URL %q to be rejected, but it was accepted", u)
		}
	}
}

func TestLocalLLMAdapterDiscovery(t *testing.T) {
	// Mock local server (e.g. LM Studio / Ollama)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "meta-llama-3-8b-instruct", "object": "model"},
				{"id": "mistral-7b-instruct-v0.3", "object": "model"},
			},
		})
	}))
	defer srv.Close()

	adapter := NewLocalLLMAdapter(srv.URL)

	ctx := context.Background()
	models, err := adapter.DiscoverModels(ctx, "")
	if err != nil {
		t.Fatalf("DiscoverModels failed: %v", err)
	}

	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}
	if models[0].ID != "meta-llama-3-8b-instruct" || models[1].ID != "mistral-7b-instruct-v0.3" {
		t.Errorf("unexpected models: %+v", models)
	}
	if adapter.DefaultModel != "meta-llama-3-8b-instruct" {
		t.Errorf("expected DefaultModel to be updated to first discovered model, got %q", adapter.DefaultModel)
	}
}

func TestLocalLLMAdapterStream(t *testing.T) {
	// Mock local streaming completions
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}

		var req openAIChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(req.Messages) == 0 {
			http.Error(w, "empty messages", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}

		// Stream tokens
		tokens := []string{"To ", "find ", "the ", "binomial ", "probability, ", "identify ", "n, ", "p, ", "and ", "k."}
		for _, tok := range tokens {
			chunk := map[string]any{
				"choices": []map[string]any{
					{
						"delta": map[string]string{
							"content": tok,
						},
					},
				},
			}
			chunkBytes, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "data: %s\n\n", string(chunkBytes))
			flusher.Flush()
		}

		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer srv.Close()

	adapter := NewLocalLLMAdapter(srv.URL)
	adapter.DefaultModel = "llama-3"

	tutorReq := tutor.TutorRequest{
		RequestID:    "req_stream_test",
		Action:       tutor.ActionChat,
		CustomPrompt: "How do I calculate binomial probability?",
		History: []tutor.ChatMessage{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hello, what topic are you working on?"},
		},
		Instance: &domain.QuestionInstance{
			Title: "Coin Toss Experiment",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	streamChan, err := adapter.Stream(ctx, tutorReq)
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}

	var events []tutor.TutorEvent
	var fullText strings.Builder

	for ev := range streamChan {
		events = append(events, ev)
		if ev.Type == tutor.EventTextDelta {
			fullText.WriteString(ev.Delta)
		}
	}

	if len(events) < 3 {
		t.Fatalf("expected at least 3 events (started, text deltas, complete), got %d", len(events))
	}

	if events[0].Type != tutor.EventStarted {
		t.Errorf("expected first event to be %q, got %q", tutor.EventStarted, events[0].Type)
	}

	lastEv := events[len(events)-1]
	if lastEv.Type != tutor.EventComplete {
		t.Errorf("expected last event to be %q, got %q", tutor.EventComplete, lastEv.Type)
	}

	expected := "To find the binomial probability, identify n, p, and k."
	if fullText.String() != expected {
		t.Errorf("expected full text %q, got %q", expected, fullText.String())
	}
}

func TestLocalLLMAdapterConnectionFailure(t *testing.T) {
	// Point to an unused port on localhost
	adapter := NewLocalLLMAdapter("http://localhost:59999")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := adapter.DiscoverModels(ctx, "")
	if err == nil {
		t.Errorf("expected error connecting to non-existent server, got nil")
	} else if !strings.Contains(err.Error(), "cannot connect to local LLM server") {
		t.Errorf("expected connection error message, got %v", err)
	}
}
