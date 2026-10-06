package auth

import (
	"strings"
	"time"
)

// Supported provider routes
const (
	RouteOffline   = "offline"
	RouteAnthropic = "anthropic"
	RouteGemini    = "gemini"
	RouteOpenAI    = "openai"
	// RouteChatGPT is the ChatGPT-plan sign-in route, billed to the user's
	// ChatGPT subscription and distinct from the OpenAI API-key route.
	RouteChatGPT = "chatgpt"
)

// CredentialSource indicates where a credential originated.
type CredentialSource string

const (
	SourceNone        CredentialSource = "none"
	SourceEnvironment CredentialSource = "environment"
	SourceVault       CredentialSource = "vault"
	SourceSession     CredentialSource = "session"
)

// CredentialStatus describes the redacted status of a credential for a given route.
// Invariant: The raw API key is NEVER included in this struct.
type CredentialStatus struct {
	Route      string           `json:"route"`
	Configured bool             `json:"configured"`
	Source     CredentialSource `json:"source"`
	MaskedKey  string           `json:"masked_key,omitempty"`
	UpdatedAt  *time.Time       `json:"updated_at,omitempty"`
}

// Vault manages protected backend-only credential storage.
// Invariant: Credentials are stored on the local backend and never echoed back.
type Vault interface {
	Get(route string) (string, error)
	Set(route string, secret string) error
	Delete(route string) error
	Status(route string) CredentialStatus
	AllStatuses() []CredentialStatus
}

// MaskKey produces a safe display representation of an API key,
// showing only the identifying prefix and the final 4 characters.
// It guarantees that secrets are never exposed in UI or logs.
func MaskKey(key string) string {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return ""
	}
	n := len(trimmed)
	if n <= 8 {
		return "••••••••"
	}

	// For standard API keys, preserve prefix if recognizable
	prefixLen := 4
	if strings.HasPrefix(trimmed, "sk-ant-") {
		prefixLen = 6 // e.g. "sk-ant"
	} else if strings.HasPrefix(trimmed, "sk-") {
		prefixLen = 2 // e.g. "sk"
	} else if strings.HasPrefix(trimmed, "AIzaSy") {
		prefixLen = 6 // e.g. "AIzaSy"
	}

	prefix := strings.TrimSuffix(trimmed[:prefixLen], "-")
	return prefix + "..." + trimmed[n-4:]
}
