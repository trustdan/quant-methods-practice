package providers

import (
	"context"
)

// ModelInfo describes an LLM model's metadata and capabilities.
type ModelInfo struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Provider          string `json:"provider"`
	SupportsStreaming bool   `json:"supports_streaming"`
	ContextWindow     int    `json:"context_window"`
	Description       string `json:"description"`
	IsCustom          bool   `json:"is_custom,omitempty"`
	IsDefault         bool   `json:"is_default,omitempty"`
}

// ProviderSummary describes the public state and configuration status of a provider route.
// Invariant: Secrets are NEVER included here.
type ProviderSummary struct {
	Route       string `json:"route"`
	Name        string `json:"name"`
	Configured  bool   `json:"configured"`
	Active      bool   `json:"active"`
	ActiveModel string `json:"active_model"`
	Source      string `json:"source"`
	MaskedKey   string `json:"masked_key,omitempty"`
	ModelsCount int    `json:"models_count"`
	RequiresKey bool   `json:"requires_key"`
	// AuthKind is "none", "api_key" or "oauth".
	AuthKind string `json:"auth_kind"`
	// Billing is "none", "api_usage" or "chatgpt_plan", shown so the learner
	// always knows which account is charged.
	Billing      string `json:"billing"`
	AccountLabel string `json:"account_label,omitempty"`
}

// BudgetStatus describes session request and token expenditure vs caps.
type BudgetStatus struct {
	MaxRequestsPerSession int  `json:"max_requests_per_session"`
	CurrentRequests       int  `json:"current_requests"`
	RemainingRequests     int  `json:"remaining_requests"`
	EstimatedTokens       int  `json:"estimated_tokens"`
	CapReached            bool `json:"cap_reached"`
}

// ModelDiscoverer is implemented by providers that support dynamic model catalog fetching.
type ModelDiscoverer interface {
	DiscoverModels(ctx context.Context, apiKey string) ([]ModelInfo, error)
}
