package providers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/auth"
	"github.com/trustdan/quant-methods-practice/internal/tutor"
)

type fakePlanTokens struct {
	token  string
	err    error
	reauth atomic.Int32
}

func (f *fakePlanTokens) AccessToken(context.Context) (string, error) { return f.token, f.err }
func (f *fakePlanTokens) MarkSelectedReauth()                         { f.reauth.Add(1) }

func drain(ch <-chan tutor.TutorEvent) (text string, errMsg string) {
	for e := range ch {
		switch e.Type {
		case tutor.EventTextDelta:
			text += e.Delta
		case tutor.EventError:
			errMsg = e.Error
		}
	}
	return text, errMsg
}

func planAdapter(url string, tokens PlanTokenSource) *ChatGPTPlanAdapter {
	a := NewChatGPTPlanAdapter(tokens, NewBudgetTracker(20))
	a.BaseURL = url
	a.DefaultModel = "plan-model"
	return a
}

func TestChatGPTPlanRequestUsesOnlyPermittedFields(t *testing.T) {
	var body map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/responses" || r.Header.Get("Authorization") != "Bearer plan-token" {
			t.Errorf("unexpected request %s auth=%q", r.URL.Path, r.Header.Get("Authorization"))
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "event: response.output_text.delta\ndata: {\"type\":\"response.output_text.delta\",\"delta\":\"Think about \"}\n\n")
		io.WriteString(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"independence.\"}\n\n")
		io.WriteString(w, "data: {\"type\":\"response.completed\",\"response\":{}}\n\n")
	}))
	defer ts.Close()

	ch, err := planAdapter(ts.URL, &fakePlanTokens{token: "plan-token"}).Stream(context.Background(), tutor.TutorRequest{RequestID: "r1", Action: tutor.ActionHint})
	if err != nil {
		t.Fatal(err)
	}
	text, errMsg := drain(ch)
	if errMsg != "" || text != "Think about independence." {
		t.Fatalf("text=%q err=%q", text, errMsg)
	}

	keys := make([]string, 0, len(body))
	for k := range body {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if strings.Join(keys, ",") != "input,instructions,model,store,stream" {
		t.Fatalf("payload keys = %v; plan route forbids extra fields", keys)
	}
	if body["store"] != false || body["stream"] != true {
		t.Fatalf("store/stream = %v/%v", body["store"], body["stream"])
	}
	input, ok := body["input"].([]any)
	if !ok || len(input) != 1 || input[0].(map[string]any)["role"] != "user" {
		t.Fatalf("input must be an array of user messages without a system role: %v", body["input"])
	}
}

func TestChatGPTPlanStreamWithoutCompletedIsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"partial\"}\n\n")
	}))
	defer ts.Close()
	ch, _ := planAdapter(ts.URL, &fakePlanTokens{token: "t"}).Stream(context.Background(), tutor.TutorRequest{})
	if _, errMsg := drain(ch); !strings.Contains(errMsg, "before completion") {
		t.Fatalf("err = %q", errMsg)
	}
}

func TestChatGPTPlanUsageLimitIsNotRetried(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
		io.WriteString(w, `{"error":{"code":"subscription_sharing_usage_limit_exceeded","message":"limit"}}`)
	}))
	defer ts.Close()
	ch, _ := planAdapter(ts.URL, &fakePlanTokens{token: "t"}).Stream(context.Background(), tutor.TutorRequest{})
	_, errMsg := drain(ch)
	if calls.Load() != 1 || !strings.Contains(errMsg, "usage limit") || !strings.Contains(errMsg, "Usage") {
		t.Fatalf("calls=%d err=%q", calls.Load(), errMsg)
	}
}

func TestChatGPTPlanUnavailableRetriesBounded(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, `{"error":{"code":"subscription_sharing_usage_unavailable"}}`)
	}))
	defer ts.Close()
	ch, _ := planAdapter(ts.URL, &fakePlanTokens{token: "t"}).Stream(context.Background(), tutor.TutorRequest{})
	_, errMsg := drain(ch)
	if calls.Load() != 3 || !strings.Contains(errMsg, "temporarily unavailable") {
		t.Fatalf("calls=%d err=%q", calls.Load(), errMsg)
	}
}

func TestChatGPTPlanInvalidUserMarksReauth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":{"code":"subscription_sharing_invalid_user"}}`)
	}))
	defer ts.Close()
	tokens := &fakePlanTokens{token: "t"}
	ch, _ := planAdapter(ts.URL, tokens).Stream(context.Background(), tutor.TutorRequest{})
	if _, errMsg := drain(ch); !strings.Contains(errMsg, "Sign in again") || tokens.reauth.Load() != 1 {
		t.Fatalf("err=%q reauth=%d", errMsg, tokens.reauth.Load())
	}
}

func TestChatGPTPlanNeverFallsBackToAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-should-never-be-used")
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
	defer ts.Close()

	budget := NewBudgetTracker(20)
	a := planAdapter(ts.URL, &fakePlanTokens{err: errors.New("no ChatGPT account is signed in")})
	a.Budget = budget
	if _, err := a.Stream(context.Background(), tutor.TutorRequest{}); err == nil {
		t.Fatal("expected error without a plan token")
	}
	if _, err := a.DiscoverModels(context.Background(), "sk-should-never-be-used"); err == nil {
		t.Fatal("model discovery must not accept an API key")
	}
	if calls.Load() != 0 || budget.GetStatus().CurrentRequests != 0 {
		t.Fatalf("no network call or budget use allowed: calls=%d budget=%+v", calls.Load(), budget.GetStatus())
	}
}

func TestChatGPTPlanModelDiscoveryFiltersVisibility(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"models":[
			{"slug":"plan-a","display_name":"Plan A","visibility":"list"},
			{"slug":"plan-hidden","display_name":"Hidden","visibility":"hide"},
			{"slug":"plan-b","visibility":"list"}]}`)
	}))
	defer ts.Close()
	models, err := planAdapter(ts.URL, &fakePlanTokens{token: "t"}).DiscoverModels(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 || models[0].ID != "plan-a" || models[0].Name != "Plan A" || models[1].Name != "plan-b" {
		t.Fatalf("models = %+v", models)
	}
	for _, m := range models {
		if m.Provider != auth.RouteChatGPT {
			t.Fatalf("provider = %q", m.Provider)
		}
	}
}

func TestManagerKeepsPlanAndAPIRoutesDistinct(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-env-key-123456")
	pm := NewProviderManager(auth.NewStandardVaultWithInner(auth.NewMemoryVault()), nil, nil)

	if err := pm.SetActive(auth.RouteChatGPT, "plan-a"); err == nil {
		t.Fatal("plan route must not activate without a signed-in plan account")
	}
	if pm.GetActiveRoute() != auth.RouteOffline {
		t.Fatalf("active route changed to %q", pm.GetActiveRoute())
	}

	byRoute := map[string]ProviderSummary{}
	for _, s := range pm.GetSummaries() {
		byRoute[s.Route] = s
	}
	plan, api := byRoute[auth.RouteChatGPT], byRoute[auth.RouteOpenAI]
	if plan.Configured || plan.Billing != "chatgpt_plan" || plan.AuthKind != "oauth" || plan.MaskedKey != "" {
		t.Fatalf("plan summary = %+v (an OpenAI API key must not configure it)", plan)
	}
	if !api.Configured || api.Billing != "api_usage" || api.AuthKind != "api_key" {
		t.Fatalf("api summary = %+v", api)
	}
}
