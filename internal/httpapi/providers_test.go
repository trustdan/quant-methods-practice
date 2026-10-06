package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/auth"
	"github.com/trustdan/quant-methods-practice/internal/providers"
)

func TestProvidersEndpoints(t *testing.T) {
	vault := auth.NewMemoryVault()
	budget := providers.NewBudgetTracker(20)
	catalog := providers.NewCatalogCache()
	pm := providers.NewProviderManager(vault, budget, catalog)

	srv, err := NewServer(Config{
		Addr:            "127.0.0.1:0",
		Vault:           vault,
		ProviderManager: pm,
	})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	handler := srv.httpServer.Handler

	// 1. GET /api/providers
	req := httptest.NewRequest(http.MethodGet, "/api/providers", nil)
	req.Host = "127.0.0.1"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var listResp providersListResponseDTO
	if err := json.NewDecoder(rec.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if listResp.ActiveRoute != auth.RouteOffline || len(listResp.Providers) != 5 {
		t.Fatalf("unexpected providers list response: %+v", listResp)
	}

	// 2. POST /api/providers/gemini/credentials
	rawKey := "AIzaSySecretGeminiKey1234"
	credBody, _ := json.Marshal(map[string]string{"api_key": rawKey})
	req = httptest.NewRequest(http.MethodPost, "/api/providers/gemini/credentials", bytes.NewReader(credBody))
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://127.0.0.1")
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 setting credentials, got %d: %s", rec.Code, rec.Body.String())
	}

	// Invariant: Response must NEVER contain the raw key!
	if strings.Contains(rec.Body.String(), rawKey) {
		t.Fatalf("SECURITY VIOLATION: Raw API key echoed in response body: %s", rec.Body.String())
	}

	var credStatus auth.CredentialStatus
	if err := json.NewDecoder(rec.Body).Decode(&credStatus); err != nil {
		t.Fatalf("failed to decode credential status: %v", err)
	}
	if !credStatus.Configured || credStatus.MaskedKey != "AIzaSy...1234" {
		t.Fatalf("unexpected cred status: %+v", credStatus)
	}

	// 3. POST /api/providers/active (Switch to Gemini)
	activeBody, _ := json.Marshal(map[string]string{
		"route": "gemini",
		"model": "gemini-2.0-flash",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/providers/active", bytes.NewReader(activeBody))
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://127.0.0.1")
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 setting active provider, got %d", rec.Code)
	}

	if pm.GetActiveRoute() != auth.RouteGemini || pm.GetActiveModel() != "gemini-2.0-flash" {
		t.Fatalf("active route mismatch: route=%s, model=%s", pm.GetActiveRoute(), pm.GetActiveModel())
	}

	// 4. GET /api/providers/gemini/models
	req = httptest.NewRequest(http.MethodGet, "/api/providers/gemini/models", nil)
	req.Host = "127.0.0.1"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 getting models, got %d", rec.Code)
	}

	var modelsResp modelsResponseDTO
	if err := json.NewDecoder(rec.Body).Decode(&modelsResp); err != nil {
		t.Fatalf("failed to decode models resp: %v", err)
	}
	if len(modelsResp.Models) == 0 {
		t.Fatal("expected cached gemini models")
	}

	// 5. POST /api/providers/gemini/disconnect
	req = httptest.NewRequest(http.MethodPost, "/api/providers/gemini/disconnect", nil)
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://127.0.0.1")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 disconnecting, got %d", rec.Code)
	}

	// Disconnecting active provider must safely fallback to offline
	if pm.GetActiveRoute() != auth.RouteOffline {
		t.Fatalf("expected fallback to offline after disconnect, got %s", pm.GetActiveRoute())
	}

	// 6. GET /api/providers/budget
	req = httptest.NewRequest(http.MethodGet, "/api/providers/budget", nil)
	req.Host = "127.0.0.1"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 getting budget, got %d", rec.Code)
	}

	var bSt providers.BudgetStatus
	if err := json.NewDecoder(rec.Body).Decode(&bSt); err != nil {
		t.Fatalf("failed to decode budget status: %v", err)
	}
	if bSt.MaxRequestsPerSession != 20 {
		t.Fatalf("expected 20 max requests, got %d", bSt.MaxRequestsPerSession)
	}
}
