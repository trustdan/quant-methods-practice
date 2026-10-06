package providers

import (
	"sync"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/auth"
)

const CatalogStaleDuration = 24 * time.Hour

// CatalogEntry stores cached models with timestamp.
type CatalogEntry struct {
	Models    []ModelInfo
	FetchedAt time.Time
}

// CatalogCache provides local capability and model caching.
type CatalogCache struct {
	mu      sync.RWMutex
	entries map[string]CatalogEntry
}

// NewCatalogCache initializes a catalog cache populated with curated defaults.
func NewCatalogCache() *CatalogCache {
	c := &CatalogCache{
		entries: make(map[string]CatalogEntry),
	}
	c.initDefaults()
	return c
}

func (c *CatalogCache) initDefaults() {
	now := time.Now().UTC()

	// Offline default
	c.entries[auth.RouteOffline] = CatalogEntry{
		Models: []ModelInfo{
			{
				ID:                "offline-curriculum",
				Name:              "Offline Verified Curriculum",
				Provider:          auth.RouteOffline,
				SupportsStreaming: true,
				ContextWindow:     8192,
				Description:       "Deterministic step-by-step mathematical derivations and causal hints from verified course bank.",
				IsDefault:         true,
			},
		},
		FetchedAt: now,
	}

	// Anthropic defaults
	c.entries[auth.RouteAnthropic] = CatalogEntry{
		Models: []ModelInfo{
			{
				ID:                "claude-3-5-sonnet-20241022",
				Name:              "Claude 3.5 Sonnet",
				Provider:          auth.RouteAnthropic,
				SupportsStreaming: true,
				ContextWindow:     200000,
				Description:       "Anthropic's most capable model for mathematical and causal reasoning.",
				IsDefault:         true,
			},
			{
				ID:                "claude-3-5-haiku-20241022",
				Name:              "Claude 3.5 Haiku",
				Provider:          auth.RouteAnthropic,
				SupportsStreaming: true,
				ContextWindow:     200000,
				Description:       "Fast and responsive Anthropic model for everyday statistical guidance.",
			},
			{
				ID:                "claude-3-opus-20240229",
				Name:              "Claude 3 Opus",
				Provider:          auth.RouteAnthropic,
				SupportsStreaming: true,
				ContextWindow:     200000,
				Description:       "Comprehensive analysis and step-by-step proofs.",
			},
		},
		FetchedAt: now,
	}

	// Gemini defaults
	c.entries[auth.RouteGemini] = CatalogEntry{
		Models: []ModelInfo{
			{
				ID:                "gemini-2.0-flash",
				Name:              "Gemini 2.0 Flash",
				Provider:          auth.RouteGemini,
				SupportsStreaming: true,
				ContextWindow:     1048576,
				Description:       "Google's fast, high-rate next-gen model with multimodal understanding.",
				IsDefault:         true,
			},
			{
				ID:                "gemini-1.5-pro",
				Name:              "Gemini 1.5 Pro",
				Provider:          auth.RouteGemini,
				SupportsStreaming: true,
				ContextWindow:     2097152,
				Description:       "Highly capable model with extensive context and deep reasoning.",
			},
			{
				ID:                "gemini-1.5-flash",
				Name:              "Gemini 1.5 Flash",
				Provider:          auth.RouteGemini,
				SupportsStreaming: true,
				ContextWindow:     1048576,
				Description:       "Lightweight, fast model optimized for quick explanations.",
			},
		},
		FetchedAt: now,
	}

	// OpenAI defaults
	c.entries[auth.RouteOpenAI] = CatalogEntry{
		Models: []ModelInfo{
			{
				ID:                "gpt-4o",
				Name:              "GPT-4o",
				Provider:          auth.RouteOpenAI,
				SupportsStreaming: true,
				ContextWindow:     128000,
				Description:       "Flagship omni model with strong quantitative and pedagogical capabilities.",
				IsDefault:         true,
			},
			{
				ID:                "gpt-4o-mini",
				Name:              "GPT-4o mini",
				Provider:          auth.RouteOpenAI,
				SupportsStreaming: true,
				ContextWindow:     128000,
				Description:       "Fast, lightweight, and cost-effective model.",
			},
			{
				ID:                "o3-mini",
				Name:              "o3-mini",
				Provider:          auth.RouteOpenAI,
				SupportsStreaming: true,
				ContextWindow:     200000,
				Description:       "Specialized reasoning model for STEM and mathematical deduction.",
			},
		},
		FetchedAt: now,
	}
}

// GetModels returns the models list and whether the catalog is stale (> 24h).
func (c *CatalogCache) GetModels(route string) ([]ModelInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[route]
	if !ok {
		return []ModelInfo{}, true
	}

	isStale := time.Since(entry.FetchedAt) > CatalogStaleDuration
	return entry.Models, isStale
}

// SetModels saves discovered models to the cache.
func (c *CatalogCache) SetModels(route string, models []ModelInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[route] = CatalogEntry{
		Models:    models,
		FetchedAt: time.Now().UTC(),
	}
}

// AddCustomModel registers an explicit user-specified model ID.
func (c *CatalogCache) AddCustomModel(route, modelID string) ModelInfo {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := c.entries[route]
	for _, m := range entry.Models {
		if m.ID == modelID {
			return m
		}
	}

	custom := ModelInfo{
		ID:                modelID,
		Name:              modelID + " (Custom)",
		Provider:          route,
		SupportsStreaming: true,
		ContextWindow:     8192,
		Description:       "User-specified unverified model.",
		IsCustom:          true,
	}

	entry.Models = append([]ModelInfo{custom}, entry.Models...)
	c.entries[route] = entry
	return custom
}
