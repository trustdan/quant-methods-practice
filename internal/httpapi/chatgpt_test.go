package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/auth"
	"github.com/trustdan/quant-methods-practice/internal/providers"
	"github.com/trustdan/quant-methods-practice/internal/siwc"
)

func newChatGPTTestServer(t *testing.T) http.Handler {
	t.Helper()
	vault := auth.NewMemoryVault()
	pm := providers.NewProviderManager(vault, providers.NewBudgetTracker(20), providers.NewCatalogCache())
	srv, err := NewServer(Config{Addr: "127.0.0.1:0", Vault: vault, ProviderManager: pm})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pm.ChatGPT().CancelSignIn() })
	return srv.httpServer.Handler
}

func doLocal(h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Host = "127.0.0.1"
	if method != http.MethodGet {
		req.Header.Set("Origin", "http://127.0.0.1")
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestChatGPTRouteRejectsPastedTokens(t *testing.T) {
	h := newChatGPTTestServer(t)
	rec := doLocal(h, http.MethodPost, "/api/providers/chatgpt/credentials", map[string]string{"api_key": "eyJ-session-token"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("pasted token status = %d", rec.Code)
	}
	if rec := doLocal(h, http.MethodGet, "/api/providers/chatgpt/accounts", nil); rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("accounts after rejected paste = %d %s", rec.Code, rec.Body.String())
	}
}

func TestChatGPTActivationRequiresSignIn(t *testing.T) {
	h := newChatGPTTestServer(t)
	rec := doLocal(h, http.MethodPost, "/api/providers/active", map[string]string{"route": "chatgpt", "model": "plan-a"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("activation without sign-in = %d", rec.Code)
	}
}

func TestChatGPTSignInStartAndCancel(t *testing.T) {
	h := newChatGPTTestServer(t)

	rec := doLocal(h, http.MethodPost, "/api/providers/chatgpt/signin", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("start = %d %s", rec.Code, rec.Body.String())
	}
	var start chatGPTSignInResponseDTO
	_ = json.NewDecoder(rec.Body).Decode(&start)
	u, err := url.Parse(start.AuthorizeURL)
	if err != nil || u.Host != "auth.openai.com" || start.Status.State != "pending" {
		t.Fatalf("start response = %+v", start)
	}
	q := u.Query()
	if q.Get("client_id") != siwc.DynamicClientID || q.Get("code_challenge_method") != "S256" ||
		!strings.HasPrefix(q.Get("redirect_uri"), "http://127.0.0.1:") {
		t.Fatalf("authorize params = %v", q)
	}

	// The app's own API never serves the OAuth callback.
	if rec := doLocal(h, http.MethodGet, "/auth/callback?code=x&state="+q.Get("state"), nil); strings.Contains(rec.Body.String(), "Signed in") {
		t.Fatal("main server must not accept OAuth callbacks")
	}

	rec = doLocal(h, http.MethodPost, "/api/providers/chatgpt/signin/cancel", nil)
	var st siwc.SignInStatus
	_ = json.NewDecoder(rec.Body).Decode(&st)
	if st.State != "cancelled" {
		t.Fatalf("cancel = %+v", st)
	}
	rec = doLocal(h, http.MethodGet, "/api/providers/chatgpt/signin", nil)
	_ = json.NewDecoder(rec.Body).Decode(&st)
	if st.State != "cancelled" {
		t.Fatalf("status after cancel = %+v", st)
	}
}
