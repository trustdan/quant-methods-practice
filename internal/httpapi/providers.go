package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/trustdan/quant-methods-practice/internal/auth"
	"github.com/trustdan/quant-methods-practice/internal/providers"
)

type setCredentialsRequestDTO struct {
	APIKey string `json:"api_key"`
}

type setActiveProviderDTO struct {
	Route string `json:"route"`
	Model string `json:"model,omitempty"`
}

type addCustomModelDTO struct {
	ModelID string `json:"model_id"`
}

type providersListResponseDTO struct {
	Providers   []providers.ProviderSummary `json:"providers"`
	ActiveRoute string                      `json:"active_route"`
	ActiveModel string                      `json:"active_model"`
	Budget      providers.BudgetStatus      `json:"budget"`
}

type modelsResponseDTO struct {
	Route   string                `json:"route"`
	Models  []providers.ModelInfo `json:"models"`
	IsStale bool                  `json:"is_stale"`
}

func (s *Server) handleProviders(w http.ResponseWriter, r *http.Request) {
	if s.providerManager == nil {
		http.Error(w, "provider manager unavailable", http.StatusServiceUnavailable)
		return
	}

	subPath := strings.TrimPrefix(r.URL.Path, "/api/providers")
	subPath = strings.TrimPrefix(subPath, "/")

	// 1. Root /api/providers
	if subPath == "" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		resp := providersListResponseDTO{
			Providers:   s.providerManager.GetSummaries(),
			ActiveRoute: s.providerManager.GetActiveRoute(),
			ActiveModel: s.providerManager.GetActiveModel(),
			Budget:      s.providerManager.Budget().GetStatus(),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	// 2. /api/providers/active
	if subPath == "active" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var dto setActiveProviderDTO
		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := s.providerManager.SetActive(dto.Route, dto.Model); err != nil {
			http.Error(w, fmt.Sprintf("failed to set active provider: %v", err), http.StatusBadRequest)
			return
		}

		// Persist active provider choice in user preferences if store available
		if s.config.Store != nil {
			ctx := r.Context()
			prefMap := make(map[string]any)
			if raw, err := s.config.Store.GetSettings(ctx, "user_preferences"); err == nil && raw != "" {
				_ = json.Unmarshal([]byte(raw), &prefMap)
			}
			prefMap["active_provider_route"] = s.providerManager.GetActiveRoute()
			prefMap["active_provider_model"] = s.providerManager.GetActiveModel()
			if prefBytes, err := json.Marshal(prefMap); err == nil {
				_ = s.config.Store.SaveSettings(ctx, "user_preferences", string(prefBytes))
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":       "ok",
			"active_route": s.providerManager.GetActiveRoute(),
			"active_model": s.providerManager.GetActiveModel(),
		})
		return
	}

	// 3. /api/providers/budget or /api/providers/budget/reset
	if strings.HasPrefix(subPath, "budget") {
		if subPath == "budget" {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(s.providerManager.Budget().GetStatus())
			return
		}
		if subPath == "budget/reset" {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			s.providerManager.Budget().Reset()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(s.providerManager.Budget().GetStatus())
			return
		}
	}

	// 4. Sub-routes per provider: /api/providers/{route}/...
	parts := strings.Split(subPath, "/")
	route := parts[0]

	if route == auth.RouteChatGPT && s.handleChatGPT(w, r, parts[1:]) {
		return
	}

	// /api/providers/{route}/credentials
	if len(parts) >= 2 && parts[1] == "credentials" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var dto setCredentialsRequestDTO
		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		key := strings.TrimSpace(dto.APIKey)
		// Clear body DTO from memory immediately after reading
		dto.APIKey = ""

		if key == "" {
			http.Error(w, "API key cannot be empty", http.StatusBadRequest)
			return
		}

		if err := s.providerManager.Vault().Set(route, key); err != nil {
			http.Error(w, fmt.Sprintf("failed to save credential: %v", err), http.StatusInternalServerError)
			return
		}

		// Return REDACTED status only. Invariant: NEVER echo back the secret.
		st := s.providerManager.Vault().Status(route)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(st)
		return
	}

	// /api/providers/{route}/disconnect
	if len(parts) >= 2 && parts[1] == "disconnect" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := s.providerManager.Vault().Delete(route); err != nil {
			http.Error(w, fmt.Sprintf("failed to disconnect credential: %v", err), http.StatusInternalServerError)
			return
		}

		// If disconnected provider was active, revert to offline
		if s.providerManager.GetActiveRoute() == route {
			_ = s.providerManager.SetActive(auth.RouteOffline, "offline-curriculum")
		}

		st := s.providerManager.Vault().Status(route)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(st)
		return
	}

	// /api/providers/{route}/models
	if len(parts) == 2 && parts[1] == "models" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		models, isStale := s.providerManager.GetModels(route)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(modelsResponseDTO{
			Route:   route,
			Models:  models,
			IsStale: isStale,
		})
		return
	}

	// /api/providers/{route}/models/refresh
	if len(parts) == 3 && parts[1] == "models" && parts[2] == "refresh" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		models, err := s.providerManager.RefreshModels(r.Context(), route)
		if err != nil {
			http.Error(w, fmt.Sprintf("model discovery failed: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(modelsResponseDTO{
			Route:   route,
			Models:  models,
			IsStale: false,
		})
		return
	}

	// /api/providers/{route}/models/custom
	if len(parts) == 3 && parts[1] == "models" && parts[2] == "custom" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var dto addCustomModelDTO
		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil || strings.TrimSpace(dto.ModelID) == "" {
			http.Error(w, "invalid custom model ID", http.StatusBadRequest)
			return
		}

		custom := s.providerManager.AddCustomModel(route, strings.TrimSpace(dto.ModelID))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(custom)
		return
	}

	http.NotFound(w, r)
}
