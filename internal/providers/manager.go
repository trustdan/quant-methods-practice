package providers

import (
	"context"
	"fmt"
	"sync"

	"github.com/trustdan/quant-methods-practice/internal/auth"
	"github.com/trustdan/quant-methods-practice/internal/siwc"
	"github.com/trustdan/quant-methods-practice/internal/tutor"
)

// ProviderManager coordinates all providers, active routes, models, and budgets.
type ProviderManager struct {
	mu           sync.RWMutex
	vault        auth.Vault
	budget       *BudgetTracker
	catalog      *CatalogCache
	activeRoute  string
	activeModels map[string]string
	services     map[string]tutor.TutorService
	discoverers  map[string]ModelDiscoverer
	chatgpt      *siwc.Client
}

// NewProviderManager initializes the ProviderManager with all adapters.
func NewProviderManager(vault auth.Vault, budget *BudgetTracker, catalog *CatalogCache) *ProviderManager {
	if vault == nil {
		vault = auth.NewMemoryVault()
	}
	if budget == nil {
		budget = NewBudgetTracker(DefaultMaxRequestsPerSession)
	}
	if catalog == nil {
		catalog = NewCatalogCache()
	}

	pm := &ProviderManager{
		vault:        vault,
		budget:       budget,
		catalog:      catalog,
		activeRoute:  auth.RouteOffline,
		activeModels: make(map[string]string),
		services:     make(map[string]tutor.TutorService),
		discoverers:  make(map[string]ModelDiscoverer),
	}

	// Register adapters
	anthropic := NewAnthropicAdapter(vault, budget)
	gemini := NewGeminiAdapter(vault, budget)
	openai := NewOpenAIAdapter(vault, budget)
	pm.chatgpt = siwc.NewClient(vault)
	chatgpt := NewChatGPTPlanAdapter(pm.chatgpt, budget)

	pm.services[auth.RouteAnthropic] = anthropic
	pm.discoverers[auth.RouteAnthropic] = anthropic

	pm.services[auth.RouteGemini] = gemini
	pm.discoverers[auth.RouteGemini] = gemini

	pm.services[auth.RouteOpenAI] = openai
	pm.discoverers[auth.RouteOpenAI] = openai

	pm.services[auth.RouteChatGPT] = chatgpt
	pm.discoverers[auth.RouteChatGPT] = chatgpt

	// Default active models
	pm.activeModels[auth.RouteOffline] = "offline-curriculum"
	pm.activeModels[auth.RouteAnthropic] = anthropic.DefaultModel
	pm.activeModels[auth.RouteGemini] = gemini.DefaultModel
	pm.activeModels[auth.RouteOpenAI] = openai.DefaultModel
	pm.activeModels[auth.RouteChatGPT] = ""

	return pm
}

// RegisterService allows registering or overriding a service (e.g. OfflineTutor).
func (pm *ProviderManager) RegisterService(route string, svc tutor.TutorService) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.services[route] = svc
}

// GetActiveRoute returns the currently active route name.
func (pm *ProviderManager) GetActiveRoute() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.activeRoute
}

// GetActiveModel returns the active model for the active route.
func (pm *ProviderManager) GetActiveModel() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.activeModels[pm.activeRoute]
}

// SetActive updates the active route and model.
func (pm *ProviderManager) SetActive(route, model string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	validRoutes := map[string]bool{
		auth.RouteOffline:   true,
		auth.RouteAnthropic: true,
		auth.RouteGemini:    true,
		auth.RouteOpenAI:    true,
		auth.RouteChatGPT:   true,
	}

	if !validRoutes[route] {
		return fmt.Errorf("invalid provider route: %q", route)
	}
	if route == auth.RouteChatGPT {
		acct, ok := pm.chatgpt.SelectedAccount()
		if !ok || !acct.PlanGranted || acct.NeedsReauth {
			return fmt.Errorf("sign in with a ChatGPT account that grants plan usage before activating this route")
		}
		if model == "" && pm.activeModels[route] == "" {
			return fmt.Errorf("select a ChatGPT plan model (refresh the model list first)")
		}
	}

	pm.activeRoute = route
	if model != "" {
		pm.activeModels[route] = model
		// Update adapter default model if applicable
		if svc, ok := pm.services[route]; ok {
			switch adapter := svc.(type) {
			case *AnthropicAdapter:
				adapter.DefaultModel = model
			case *GeminiAdapter:
				adapter.DefaultModel = model
			case *OpenAIAdapter:
				adapter.DefaultModel = model
			case *ChatGPTPlanAdapter:
				adapter.DefaultModel = model
			}
		}
	}

	return nil
}

// GetSummaries returns status summaries for all known provider routes.
func (pm *ProviderManager) GetSummaries() []ProviderSummary {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	routes := []string{
		auth.RouteOffline,
		auth.RouteAnthropic,
		auth.RouteGemini,
		auth.RouteOpenAI,
		auth.RouteChatGPT,
	}

	names := map[string]string{
		auth.RouteOffline:   "Offline Reviewed (Default)",
		auth.RouteAnthropic: "Anthropic Claude",
		auth.RouteGemini:    "Google Gemini",
		auth.RouteOpenAI:    "OpenAI API",
		auth.RouteChatGPT:   "ChatGPT Plan",
	}

	requiresKey := map[string]bool{
		auth.RouteOffline:   false,
		auth.RouteAnthropic: true,
		auth.RouteGemini:    true,
		auth.RouteOpenAI:    true,
		auth.RouteChatGPT:   false,
	}

	summaries := make([]ProviderSummary, 0, len(routes))
	for _, r := range routes {
		models, _ := pm.catalog.GetModels(r)
		sum := ProviderSummary{
			Route:       r,
			Name:        names[r],
			Active:      r == pm.activeRoute,
			ActiveModel: pm.activeModels[r],
			ModelsCount: len(models),
			RequiresKey: requiresKey[r],
		}

		switch r {
		case auth.RouteOffline:
			sum.Configured, sum.AuthKind, sum.Billing = true, "none", "none"
			sum.Source = string(auth.SourceNone)
		case auth.RouteChatGPT:
			// Never read through the generic vault status: that would mask the
			// account store blob as if it were an API key.
			sum.AuthKind, sum.Billing, sum.Source = "oauth", "chatgpt_plan", string(auth.SourceNone)
			if acct, ok := pm.chatgpt.SelectedAccount(); ok {
				sum.AccountLabel = acct.Label
				sum.Configured = acct.PlanGranted && !acct.NeedsReauth
				sum.Source = string(auth.SourceVault)
			}
		default:
			st := pm.vault.Status(r)
			sum.Configured, sum.Source, sum.MaskedKey = st.Configured, string(st.Source), st.MaskedKey
			sum.AuthKind, sum.Billing = "api_key", "api_usage"
		}
		summaries = append(summaries, sum)
	}

	return summaries
}

// GetModels returns cached models for the specified route.
func (pm *ProviderManager) GetModels(route string) ([]ModelInfo, bool) {
	return pm.catalog.GetModels(route)
}

// AddCustomModel registers a custom unverified model ID for the route.
func (pm *ProviderManager) AddCustomModel(route, modelID string) ModelInfo {
	return pm.catalog.AddCustomModel(route, modelID)
}

// RefreshModels contacts the provider API to refresh the dynamic model catalog.
func (pm *ProviderManager) RefreshModels(ctx context.Context, route string) ([]ModelInfo, error) {
	pm.mu.RLock()
	disc, ok := pm.discoverers[route]
	pm.mu.RUnlock()

	if !ok || disc == nil {
		if route == auth.RouteOffline {
			models, _ := pm.catalog.GetModels(route)
			return models, nil
		}
		return nil, fmt.Errorf("model discovery not supported for route %q", route)
	}

	var key string
	if route != auth.RouteChatGPT { // the plan route authenticates with its own OAuth tokens
		var err error
		key, err = pm.vault.Get(route)
		if err != nil || key == "" {
			return nil, fmt.Errorf("cannot refresh models: API key not configured for %s", route)
		}
	}

	models, err := disc.DiscoverModels(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(models) > 0 {
		pm.catalog.SetModels(route, models)
	}

	return models, nil
}

// GetProvider returns the TutorService implementation for the route.
func (pm *ProviderManager) GetProvider(route string) (tutor.TutorService, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	svc, ok := pm.services[route]
	if !ok || svc == nil {
		return nil, fmt.Errorf("provider service %q not found", route)
	}
	return svc, nil
}

// GetActiveProvider returns the TutorService for the currently active route.
func (pm *ProviderManager) GetActiveProvider() (tutor.TutorService, error) {
	pm.mu.RLock()
	route := pm.activeRoute
	pm.mu.RUnlock()

	return pm.GetProvider(route)
}

// Vault returns the underlying Vault.
func (pm *ProviderManager) Vault() auth.Vault {
	return pm.vault
}

// ChatGPT returns the Sign in with ChatGPT client.
func (pm *ProviderManager) ChatGPT() *siwc.Client {
	return pm.chatgpt
}

// Budget returns the underlying BudgetTracker.
func (pm *ProviderManager) Budget() *BudgetTracker {
	return pm.budget
}
