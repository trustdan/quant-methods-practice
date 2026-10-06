package auth

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// StandardVault coordinates persistent/session storage and backend environment variables.
//
// Invariants:
// 1. Vault secrets stay on the backend and are NEVER exposed to the frontend or logs.
// 2. Precedence: Explicitly set vault credentials take precedence over environment variables.
// 3. Environment variables are checked as fallback:
//   - Anthropic: ANTHROPIC_API_KEY
//   - Gemini:    GEMINI_API_KEY, GOOGLE_API_KEY
//   - OpenAI:    OPENAI_API_KEY
type StandardVault struct {
	inner Vault
	mu    sync.RWMutex
}

// NewStandardVault creates a StandardVault using the preferred platform storage.
func NewStandardVault(dataDir string) (*StandardVault, error) {
	if dataDir == "" {
		return &StandardVault{inner: NewMemoryVault()}, nil
	}

	vaultFile := filepath.Join(dataDir, "credentials.vault")

	if runtime.GOOS == "windows" {
		dpapi, err := NewDPAPIVault(vaultFile)
		if err == nil {
			return &StandardVault{inner: dpapi}, nil
		}
	}

	// Fallback to AES-256-GCM FileVault
	seed := []byte("quant-methods-practice-default-key-seed")
	fv, err := NewFileVault(vaultFile, seed)
	if err != nil {
		// If persistent file vault fails, fallback to MemoryVault
		return &StandardVault{inner: NewMemoryVault()}, nil
	}

	return &StandardVault{inner: fv}, nil
}

// NewStandardVaultWithInner allows wrapping a specific Vault implementation (e.g. for testing).
func NewStandardVaultWithInner(inner Vault) *StandardVault {
	if inner == nil {
		inner = NewMemoryVault()
	}
	return &StandardVault{inner: inner}
}

func getEnvKey(route string) string {
	switch route {
	case RouteAnthropic:
		return os.Getenv("ANTHROPIC_API_KEY")
	case RouteGemini:
		if k := os.Getenv("GEMINI_API_KEY"); k != "" {
			return k
		}
		return os.Getenv("GOOGLE_API_KEY")
	case RouteOpenAI:
		return os.Getenv("OPENAI_API_KEY")
	default:
		return ""
	}
}

func (v *StandardVault) Get(route string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// 1. Check inner vault first
	if val, err := v.inner.Get(route); err == nil && val != "" {
		return val, nil
	}

	// 2. Fallback to environment variable
	if envVal := getEnvKey(route); envVal != "" {
		return envVal, nil
	}

	return "", errors.New("credential not configured")
}

func (v *StandardVault) Set(route string, secret string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.inner.Set(route, secret)
}

func (v *StandardVault) Delete(route string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.inner.Delete(route)
}

func (v *StandardVault) Status(route string) CredentialStatus {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// 1. Check inner vault
	st := v.inner.Status(route)
	if st.Configured {
		return st
	}

	// 2. Fallback to environment
	if envVal := getEnvKey(route); envVal != "" {
		return CredentialStatus{
			Route:      route,
			Configured: true,
			Source:     SourceEnvironment,
			MaskedKey:  MaskKey(envVal),
		}
	}

	return CredentialStatus{
		Route:      route,
		Configured: false,
		Source:     SourceNone,
	}
}

func (v *StandardVault) AllStatuses() []CredentialStatus {
	routes := []string{RouteOffline, RouteAnthropic, RouteGemini, RouteOpenAI}
	statuses := make([]CredentialStatus, 0, len(routes))
	for _, r := range routes {
		if r == RouteOffline {
			statuses = append(statuses, CredentialStatus{
				Route:      RouteOffline,
				Configured: true,
				Source:     SourceNone,
			})
			continue
		}
		statuses = append(statuses, v.Status(r))
	}
	return statuses
}
