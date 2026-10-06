package auth

import (
	"errors"
	"sync"
	"time"
)

// MemoryVault provides an in-memory credential vault suitable for session-only
// storage, testing, and environments where persistent storage is disabled.
type MemoryVault struct {
	mu      sync.RWMutex
	secrets map[string]string
	times   map[string]time.Time
}

// NewMemoryVault initializes an empty MemoryVault.
func NewMemoryVault() *MemoryVault {
	return &MemoryVault{
		secrets: make(map[string]string),
		times:   make(map[string]time.Time),
	}
}

func (v *MemoryVault) Get(route string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	val, ok := v.secrets[route]
	if !ok || val == "" {
		return "", errors.New("credential not found")
	}
	return val, nil
}

func (v *MemoryVault) Set(route string, secret string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if secret == "" {
		delete(v.secrets, route)
		delete(v.times, route)
		return nil
	}

	v.secrets[route] = secret
	v.times[route] = time.Now().UTC()
	return nil
}

func (v *MemoryVault) Delete(route string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	delete(v.secrets, route)
	delete(v.times, route)
	return nil
}

func (v *MemoryVault) Status(route string) CredentialStatus {
	v.mu.RLock()
	defer v.mu.RUnlock()

	secret, ok := v.secrets[route]
	if !ok || secret == "" {
		return CredentialStatus{
			Route:      route,
			Configured: false,
			Source:     SourceNone,
		}
	}

	t := v.times[route]
	return CredentialStatus{
		Route:      route,
		Configured: true,
		Source:     SourceSession,
		MaskedKey:  MaskKey(secret),
		UpdatedAt:  &t,
	}
}

func (v *MemoryVault) AllStatuses() []CredentialStatus {
	routes := []string{RouteAnthropic, RouteGemini, RouteOpenAI}
	statuses := make([]CredentialStatus, 0, len(routes))
	for _, r := range routes {
		statuses = append(statuses, v.Status(r))
	}
	return statuses
}
