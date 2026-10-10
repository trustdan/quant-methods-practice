package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/candidates"
	"github.com/trustdan/quant-methods-practice/internal/storage"
	"github.com/trustdan/quant-methods-practice/internal/tutor"
)

func TestCandidateLifecycleAndBoundary(t *testing.T) {
	db, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = storage.RunMigrations(context.Background(), db, nil); err != nil {
		t.Fatal(err)
	}
	store := storage.NewStore(db, nil)
	srv, err := NewServer(Config{Bank: bank.NewBank(), Store: store})
	if err != nil {
		t.Fatal(err)
	}
	call := func(method, path string, body any, cookie *http.Cookie, origin string) *httptest.ResponseRecorder {
		data, _ := json.Marshal(body)
		req := httptest.NewRequest(method, "http://127.0.0.1"+path, bytes.NewReader(data))
		req.Header.Set("Origin", origin)
		if cookie != nil {
			req.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		srv.httpServer.Handler.ServeHTTP(w, req)
		return w
	}
	if w := call("GET", "/api/candidates", nil, nil, ""); w.Code != 401 {
		t.Fatalf("unauthenticated: %d", w.Code)
	}
	w := call("POST", "/api/local-session", map[string]string{"bootstrap_token": srv.BootstrapToken()}, nil, "http://127.0.0.1")
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	cookie := w.Result().Cookies()[0]
	for _, origin := range []string{"https://evil.example", "http://127.0.0.1:9999", ""} {
		if w := call("POST", "/api/candidates", map[string]any{"mode": "local"}, cookie, origin); w.Code != 403 {
			t.Fatalf("origin %q: %d", origin, w.Code)
		}
	}
	w = call("POST", "/api/candidates", map[string]any{"mode": "local", "seed": 1, "approval": true}, cookie, "http://127.0.0.1")
	if w.Code != 400 {
		t.Fatal("unknown field accepted")
	}
	w = call("POST", "/api/candidates", map[string]any{"mode": "local", "seed": 1}, cookie, "http://127.0.0.1")
	if w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	var rec candidates.Record
	if err = json.Unmarshal(w.Body.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	if srv.bank.Count() != 0 {
		t.Fatal("draft entered bank")
	}
	review := map[string]any{"action": "approve", "expected_revision": 1, "reviewer": "human", "notes": "reviewed every stage"}
	path := "/api/candidates/" + rec.ID + "/review"
	if w = call("POST", path, review, cookie, "http://127.0.0.1"); w.Code != 400 {
		t.Fatal("semantic check bypassed")
	}
	review["semantic_confirmed"] = true
	if w = call("POST", path, review, cookie, "http://127.0.0.1"); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if srv.bank.Count() != 1 {
		t.Fatal("approved candidate not active")
	}
	if w = call("POST", path, review, cookie, "http://127.0.0.1"); w.Code != 409 {
		t.Fatal("duplicate approval accepted")
	}
	w = call("GET", "/api/candidates/export", nil, cookie, "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), rec.ID) {
		t.Fatal("approved export missing")
	}
	// Startup reloads only approved records, not generated or rejected records.
	restart, err := NewServer(Config{Bank: bank.NewBank(), Store: store})
	if err != nil {
		t.Fatal(err)
	}
	if restart.bank.Count() != 1 {
		t.Fatal("approval not restored after restart")
	}
	review["action"] = "retire"
	review["expected_revision"] = 2
	if w = call("POST", path, review, cookie, "http://127.0.0.1"); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if srv.bank.Count() != 0 || len(srv.bank.ListByModule("module_2")) != 0 {
		t.Fatal("retired candidate still selectable")
	}
	restart, err = NewServer(Config{Bank: bank.NewBank(), Store: store})
	if err != nil {
		t.Fatal(err)
	}
	if restart.bank.Count() != 0 {
		t.Fatal("retirement not restored")
	}
}

type candidateStream struct {
	events []tutor.TutorEvent
	wait   bool
}

func (f candidateStream) ProviderID() string { return "fake" }
func (f candidateStream) Capabilities() tutor.ProviderCapabilities {
	return tutor.ProviderCapabilities{}
}
func (f candidateStream) Stream(ctx context.Context, req tutor.TutorRequest) (<-chan tutor.TutorEvent, error) {
	ch := make(chan tutor.TutorEvent, len(f.events))
	for _, e := range f.events {
		ch <- e
	}
	if f.wait {
		go func() { <-ctx.Done(); close(ch) }()
	} else {
		close(ch)
	}
	return ch, nil
}
func TestCandidateAIUntrustedOutputAndCancellation(t *testing.T) {
	base := candidates.Local(0)
	data, _ := json.Marshal(base)
	success := []tutor.TutorEvent{{Type: tutor.EventTextDelta, Delta: string(data)}, {Type: tutor.EventComplete}}
	if _, err := generateCandidateWording(context.Background(), candidateStream{events: success}, base); err != nil {
		t.Fatal(err)
	}
	changed := base
	changed.P = .2
	other, _ := json.Marshal(changed)
	for _, events := range [][]tutor.TutorEvent{
		{{Type: tutor.EventTextDelta, Delta: string(other)}, {Type: tutor.EventComplete}},
		{{Type: tutor.EventTextDelta, Delta: string(data)}},
		{{Type: tutor.EventTextDelta, Delta: strings.Repeat("x", 12001)}},
		{{Type: tutor.EventFallback}},
		{{Type: tutor.EventTextDelta, Delta: strings.TrimSuffix(string(data), "}") + `,"answer":0.99}`}, {Type: tutor.EventComplete}},
	} {
		if _, err := generateCandidateWording(context.Background(), candidateStream{events: events}, base); err == nil {
			t.Fatal("untrusted/incomplete output accepted")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := generateCandidateWording(ctx, candidateStream{wait: true}, base); err == nil {
		t.Fatal("cancel ignored")
	}
}
