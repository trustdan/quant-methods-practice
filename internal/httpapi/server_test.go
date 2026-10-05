package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func newTestServer(t *testing.T) *Server {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><body>Test UI</body></html>"),
		},
		"assets/app.js": &fstest.MapFile{
			Data: []byte("console.log('test');"),
		},
	}

	srv, err := NewServer(Config{
		Addr:     "127.0.0.1:0",
		AssetsFS: mockFS,
		Version:  "0.1.0-test",
	})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	return srv
}

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.Host = "127.0.0.1"
	rec := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var data map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&data); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}

	if data["status"] != "ok" || data["version"] != "0.1.0-test" {
		t.Fatalf("unexpected health data: %v", data)
	}

	// Verify security headers
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("missing or incorrect X-Content-Type-Options header")
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("missing or incorrect X-Frame-Options header")
	}
	if rec.Header().Get("Content-Security-Policy") == "" {
		t.Errorf("missing Content-Security-Policy header")
	}
}

func TestLocalSessionExchange(t *testing.T) {
	srv := newTestServer(t)
	bootstrapToken := srv.BootstrapToken()

	// 1. Valid exchange
	body, _ := json.Marshal(map[string]string{"bootstrap_token": bootstrapToken})
	req := httptest.NewRequest(http.MethodPost, "/api/local-session", bytes.NewReader(body))
	req.Host = "127.0.0.1"
	rec := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on first exchange, got %d", rec.Code)
	}

	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "quant_session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil || !sessionCookie.HttpOnly {
		t.Fatalf("expected HttpOnly quant_session cookie")
	}

	// 2. Second exchange with same bootstrap token must fail (single-use)
	req2 := httptest.NewRequest(http.MethodPost, "/api/local-session", bytes.NewReader(body))
	req2.Host = "127.0.0.1"
	rec2 := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on reused bootstrap token, got %d", rec2.Code)
	}
}

func TestHostHeaderRejection(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.Host = "attacker.example.com"
	rec := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for non-loopback host, got %d", rec.Code)
	}
}

func TestCrossOriginMutationRejection(t *testing.T) {
	srv := newTestServer(t)

	body, _ := json.Marshal(map[string]string{"bootstrap_token": "foo"})
	req := httptest.NewRequest(http.MethodPost, "/api/local-session", bytes.NewReader(body))
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://malicious-site.com")
	rec := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for disallowed origin, got %d", rec.Code)
	}
}

func TestStaticAndSPAFallback(t *testing.T) {
	srv := newTestServer(t)

	// Static file request
	req1 := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	req1.Host = "127.0.0.1"
	rec1 := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200 for static asset, got %d", rec1.Code)
	}

	// SPA route fallback (e.g. /practice/session-1) should serve index.html
	req2 := httptest.NewRequest(http.MethodGet, "/practice/session-1", nil)
	req2.Host = "127.0.0.1"
	rec2 := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 for SPA route, got %d", rec2.Code)
	}
	if !bytes.Contains(rec2.Body.Bytes(), []byte("Test UI")) {
		t.Fatalf("expected index.html body on SPA fallback")
	}
}
