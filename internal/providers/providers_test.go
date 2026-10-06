package providers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/auth"
	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/tutor"
)

func TestBudgetTracker(t *testing.T) {
	b := NewBudgetTracker(3)

	st := b.GetStatus()
	if st.MaxRequestsPerSession != 3 || st.CurrentRequests != 0 || st.RemainingRequests != 3 || st.CapReached {
		t.Fatalf("unexpected initial budget status: %+v", st)
	}

	if err := b.RecordRequest(); err != nil {
		t.Fatalf("unexpected error on request 1: %v", err)
	}
	if err := b.RecordRequest(); err != nil {
		t.Fatalf("unexpected error on request 2: %v", err)
	}
	if err := b.RecordRequest(); err != nil {
		t.Fatalf("unexpected error on request 3: %v", err)
	}

	// 4th request must fail
	err := b.RecordRequest()
	if err == nil {
		t.Fatal("expected ErrBudgetExceeded on 4th request, got nil")
	}

	st = b.GetStatus()
	if !st.CapReached || st.RemainingRequests != 0 {
		t.Fatalf("expected cap reached: %+v", st)
	}

	b.RecordTokens(150)
	if b.GetStatus().EstimatedTokens != 150 {
		t.Fatalf("tokens not recorded")
	}

	b.Reset()
	if b.GetStatus().CurrentRequests != 0 || b.GetStatus().CapReached {
		t.Fatalf("reset failed")
	}
}

func TestCatalogCache(t *testing.T) {
	c := NewCatalogCache()

	models, isStale := c.GetModels(auth.RouteAnthropic)
	if isStale || len(models) == 0 {
		t.Fatalf("expected default anthropic models: len=%d, isStale=%v", len(models), isStale)
	}

	custom := c.AddCustomModel(auth.RouteGemini, "gemini-custom-exp")
	if custom.ID != "gemini-custom-exp" || !custom.IsCustom {
		t.Fatalf("custom model registration failed: %+v", custom)
	}

	geminiModels, _ := c.GetModels(auth.RouteGemini)
	if len(geminiModels) < 2 || geminiModels[0].ID != "gemini-custom-exp" {
		t.Fatalf("expected custom model at front of list: %+v", geminiModels)
	}
}

func TestAnthropicAdapterStreamingAndDiscovery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		if r.Header.Get("x-api-key") != "sk-ant-test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path == "/v1/messages" {
			w.Header().Set("Content-Type", "text/event-stream")
			flusher, _ := w.(http.Flusher)

			fmt.Fprintf(w, "data: {\"type\": \"content_block_delta\", \"delta\": {\"type\": \"text_delta\", \"text\": \"First step. \"}}\n\n")
			flusher.Flush()
			fmt.Fprintf(w, "data: {\"type\": \"content_block_delta\", \"delta\": {\"type\": \"text_delta\", \"text\": \"Second step.\"}}\n\n")
			flusher.Flush()
			fmt.Fprintf(w, "data: {\"type\": \"message_stop\"}\n\n")
			flusher.Flush()
			return
		}

		if r.URL.Path == "/v1/models" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"data": [{"id": "claude-3-5-sonnet-20241022", "display_name": "Claude 3.5 Sonnet"}]}`)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	v := auth.NewMemoryVault()
	_ = v.Set(auth.RouteAnthropic, "sk-ant-test-key")
	b := NewBudgetTracker(10)

	adapter := NewAnthropicAdapter(v, b)
	adapter.BaseURL = ts.URL

	// 1. Discovery
	models, err := adapter.DiscoverModels(context.Background(), "")
	if err != nil || len(models) != 1 || models[0].ID != "claude-3-5-sonnet-20241022" {
		t.Fatalf("model discovery failed: %v, %+v", err, models)
	}

	// 2. Streaming
	ch, err := adapter.Stream(context.Background(), tutor.TutorRequest{
		RequestID: "req_test_1",
		Action:    tutor.ActionExplain,
		Instance:  &domain.QuestionInstance{Title: "Coin Tossing"},
	})
	if err != nil {
		t.Fatalf("stream init failed: %v", err)
	}

	var chunks []string
	var completeText string
	for ev := range ch {
		if ev.Type == tutor.EventTextDelta {
			chunks = append(chunks, ev.Delta)
		} else if ev.Type == tutor.EventComplete {
			completeText = ev.Text
		}
	}

	if len(chunks) != 2 || completeText != "First step. Second step." {
		t.Fatalf("unexpected stream outcome: chunks=%v, complete=%q", chunks, completeText)
	}
}

func TestGeminiAdapterStreamingAndDiscovery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "AIzaSy-test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if strings.Contains(r.URL.Path, "streamGenerateContent") {
			w.Header().Set("Content-Type", "text/event-stream")
			flusher, _ := w.(http.Flusher)

			fmt.Fprintf(w, "data: {\"candidates\": [{\"content\": {\"parts\": [{\"text\": \"Binomial derivation: \"}]}}]}\n\n")
			flusher.Flush()
			fmt.Fprintf(w, "data: {\"candidates\": [{\"content\": {\"parts\": [{\"text\": \"P(X=2) = 3/8.\"}]}}]}\n\n")
			flusher.Flush()
			return
		}

		if r.URL.Path == "/v1beta/models" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"models": [{"name": "models/gemini-2.0-flash", "displayName": "Gemini 2.0 Flash", "supportedGenerationMethods": ["generateContent"]}]}`)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	v := auth.NewMemoryVault()
	_ = v.Set(auth.RouteGemini, "AIzaSy-test-key")
	b := NewBudgetTracker(10)

	adapter := NewGeminiAdapter(v, b)
	adapter.BaseURL = ts.URL

	// 1. Discovery
	models, err := adapter.DiscoverModels(context.Background(), "")
	if err != nil || len(models) != 1 || models[0].ID != "gemini-2.0-flash" {
		t.Fatalf("gemini discovery failed: %v, %+v", err, models)
	}

	// 2. Streaming
	ch, err := adapter.Stream(context.Background(), tutor.TutorRequest{
		RequestID: "req_test_gemini",
		Action:    tutor.ActionExplain,
	})
	if err != nil {
		t.Fatalf("gemini stream failed: %v", err)
	}

	var completeText string
	for ev := range ch {
		if ev.Type == tutor.EventComplete {
			completeText = ev.Text
		}
	}

	if completeText != "Binomial derivation: P(X=2) = 3/8." {
		t.Fatalf("unexpected complete text: %q", completeText)
	}
}

func TestOpenAIAdapterStreamingAndDiscovery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sk-test-openai" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path == "/v1/chat/completions" {
			w.Header().Set("Content-Type", "text/event-stream")
			flusher, _ := w.(http.Flusher)

			fmt.Fprintf(w, "data: {\"choices\": [{\"delta\": {\"content\": \"OpenAI step.\"}}]}\n\n")
			flusher.Flush()
			fmt.Fprintf(w, "data: [DONE]\n\n")
			flusher.Flush()
			return
		}

		if r.URL.Path == "/v1/models" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"data": [{"id": "gpt-4o"}, {"id": "o3-mini"}]}`)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	v := auth.NewMemoryVault()
	_ = v.Set(auth.RouteOpenAI, "sk-test-openai")
	b := NewBudgetTracker(10)

	adapter := NewOpenAIAdapter(v, b)
	adapter.BaseURL = ts.URL

	// 1. Discovery
	models, err := adapter.DiscoverModels(context.Background(), "")
	if err != nil || len(models) != 2 {
		t.Fatalf("openai discovery failed: %v, %+v", err, models)
	}

	// 2. Streaming
	ch, err := adapter.Stream(context.Background(), tutor.TutorRequest{
		RequestID: "req_test_openai",
		Action:    tutor.ActionExplain,
	})
	if err != nil {
		t.Fatalf("openai stream failed: %v", err)
	}

	var completeText string
	for ev := range ch {
		if ev.Type == tutor.EventComplete {
			completeText = ev.Text
		}
	}

	if completeText != "OpenAI step." {
		t.Fatalf("unexpected complete text: %q", completeText)
	}
}

func TestProviderManagerCoordination(t *testing.T) {
	v := auth.NewMemoryVault()
	b := NewBudgetTracker(20)
	pm := NewProviderManager(v, b, nil)

	// Offline is default active
	if pm.GetActiveRoute() != auth.RouteOffline {
		t.Fatalf("expected offline active route, got %s", pm.GetActiveRoute())
	}

	// Summaries
	summaries := pm.GetSummaries()
	if len(summaries) != 5 {
		t.Fatalf("expected 5 summaries, got %d", len(summaries))
	}

	var offlineSum *ProviderSummary
	for i := range summaries {
		if summaries[i].Route == auth.RouteOffline {
			offlineSum = &summaries[i]
		}
	}
	if offlineSum == nil || !offlineSum.Active || !offlineSum.Configured {
		t.Fatalf("offline summary invalid: %+v", offlineSum)
	}

	// Switch active provider
	err := pm.SetActive(auth.RouteGemini, "gemini-1.5-pro")
	if err != nil {
		t.Fatalf("failed to set active: %v", err)
	}
	if pm.GetActiveRoute() != auth.RouteGemini || pm.GetActiveModel() != "gemini-1.5-pro" {
		t.Fatalf("active route/model mismatch: route=%s, model=%s", pm.GetActiveRoute(), pm.GetActiveModel())
	}
}

func TestOfflineByDefaultZeroNetwork(t *testing.T) {
	// A server that will panic if called
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected network call made in offline mode: %s", r.URL.Path)
	}))
	defer ts.Close()

	v := auth.NewMemoryVault()
	b := NewBudgetTracker(20)
	pm := NewProviderManager(v, b, nil)

	// Set invalid URL pointing to ts on all adapters
	if a, ok := pm.services[auth.RouteAnthropic].(*AnthropicAdapter); ok {
		a.BaseURL = ts.URL
	}
	if g, ok := pm.services[auth.RouteGemini].(*GeminiAdapter); ok {
		g.BaseURL = ts.URL
	}
	if o, ok := pm.services[auth.RouteOpenAI].(*OpenAIAdapter); ok {
		o.BaseURL = ts.URL
	}

	// In offline mode, getting summaries, models, or active provider must not touch network
	_ = pm.GetSummaries()
	models, _ := pm.GetModels(auth.RouteOffline)
	if len(models) == 0 {
		t.Fatal("expected offline models")
	}

	if pm.GetActiveRoute() != auth.RouteOffline {
		t.Fatal("expected offline default")
	}
}

func TestRateLimitBackoffAndSafeError(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	v := auth.NewMemoryVault()
	_ = v.Set(auth.RouteAnthropic, "sk-ant-test-key")
	b := NewBudgetTracker(10)

	adapter := NewAnthropicAdapter(v, b)
	adapter.BaseURL = ts.URL

	ch, err := adapter.Stream(context.Background(), tutor.TutorRequest{
		RequestID: "req_rl_test",
		Action:    tutor.ActionExplain,
	})
	if err != nil {
		t.Fatalf("unexpected stream err: %v", err)
	}

	var hasRateLimitErr bool
	for ev := range ch {
		if ev.Type == tutor.EventError && strings.Contains(ev.Error, "rate limit") {
			hasRateLimitErr = true
		}
	}

	if !hasRateLimitErr {
		t.Fatal("expected rate limit error event")
	}
	if calls < 2 {
		t.Fatalf("expected retries on 429, got %d calls", calls)
	}
}
