//go:build windows

package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// DPAPIVault uses Windows Data Protection API (DPAPI) to encrypt credentials
// using the current user's logon credentials.
type DPAPIVault struct {
	filePath string
	mu       sync.RWMutex
	cached   *vaultPayload
}

// NewDPAPIVault creates or opens a DPAPI-backed vault at filePath.
func NewDPAPIVault(filePath string) (*DPAPIVault, error) {
	if filePath == "" {
		return nil, errors.New("filePath cannot be empty")
	}

	v := &DPAPIVault{
		filePath: filePath,
	}

	if _, err := os.Stat(filePath); err == nil {
		if err := v.load(); err != nil {
			return nil, fmt.Errorf("failed to load DPAPI vault: %w", err)
		}
	} else {
		v.cached = &vaultPayload{
			Version:   1,
			Secrets:   make(map[string]string),
			Times:     make(map[string]time.Time),
			UpdatedAt: time.Now().UTC(),
		}
	}

	return v, nil
}

func dpapiEncrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, nil
	}

	var inBlob windows.DataBlob
	inBlob.Size = uint32(len(plaintext))
	inBlob.Data = &plaintext[0]

	var outBlob windows.DataBlob
	err := windows.CryptProtectData(&inBlob, nil, nil, 0, nil, 0, &outBlob)
	if err != nil {
		return nil, fmt.Errorf("CryptProtectData failed: %w", err)
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(outBlob.Data)))

	ciphertext := make([]byte, outBlob.Size)
	copy(ciphertext, unsafe.Slice(outBlob.Data, outBlob.Size))
	return ciphertext, nil
}

func dpapiDecrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, nil
	}

	var inBlob windows.DataBlob
	inBlob.Size = uint32(len(ciphertext))
	inBlob.Data = &ciphertext[0]

	var outBlob windows.DataBlob
	err := windows.CryptUnprotectData(&inBlob, nil, nil, 0, nil, 0, &outBlob)
	if err != nil {
		return nil, fmt.Errorf("CryptUnprotectData failed: %w", err)
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(outBlob.Data)))

	plaintext := make([]byte, outBlob.Size)
	copy(plaintext, unsafe.Slice(outBlob.Data, outBlob.Size))
	return plaintext, nil
}

func (v *DPAPIVault) load() error {
	ciphertext, err := os.ReadFile(v.filePath)
	if err != nil {
		return err
	}
	if len(ciphertext) == 0 {
		v.cached = &vaultPayload{
			Version:   1,
			Secrets:   make(map[string]string),
			Times:     make(map[string]time.Time),
			UpdatedAt: time.Now().UTC(),
		}
		return nil
	}

	plaintext, err := dpapiDecrypt(ciphertext)
	if err != nil {
		return err
	}

	var payload vaultPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return fmt.Errorf("invalid DPAPI payload: %w", err)
	}

	if payload.Secrets == nil {
		payload.Secrets = make(map[string]string)
	}
	if payload.Times == nil {
		payload.Times = make(map[string]time.Time)
	}

	v.cached = &payload
	return nil
}

func (v *DPAPIVault) save() error {
	if v.cached == nil {
		return nil
	}

	v.cached.UpdatedAt = time.Now().UTC()
	plaintext, err := json.Marshal(v.cached)
	if err != nil {
		return err
	}

	ciphertext, err := dpapiEncrypt(plaintext)
	if err != nil {
		return err
	}

	dir := filepath.Dir(v.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	tmpFile := v.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, ciphertext, 0600); err != nil {
		return err
	}
	return os.Rename(tmpFile, v.filePath)
}

func (v *DPAPIVault) Get(route string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.cached == nil {
		return "", errors.New("vault not loaded")
	}

	secret, ok := v.cached.Secrets[route]
	if !ok || secret == "" {
		return "", errors.New("credential not found")
	}
	return secret, nil
}

func (v *DPAPIVault) Set(route string, secret string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.cached == nil {
		v.cached = &vaultPayload{
			Version: 1,
			Secrets: make(map[string]string),
			Times:   make(map[string]time.Time),
		}
	}

	if secret == "" {
		delete(v.cached.Secrets, route)
		delete(v.cached.Times, route)
	} else {
		v.cached.Secrets[route] = secret
		v.cached.Times[route] = time.Now().UTC()
	}

	return v.save()
}

func (v *DPAPIVault) Delete(route string) error {
	return v.Set(route, "")
}

func (v *DPAPIVault) Status(route string) CredentialStatus {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.cached == nil {
		return CredentialStatus{Route: route, Configured: false, Source: SourceNone}
	}

	secret, ok := v.cached.Secrets[route]
	if !ok || secret == "" {
		return CredentialStatus{Route: route, Configured: false, Source: SourceNone}
	}

	t, hasTime := v.cached.Times[route]
	var pTime *time.Time
	if hasTime {
		pTime = &t
	}

	return CredentialStatus{
		Route:      route,
		Configured: true,
		Source:     SourceVault,
		MaskedKey:  MaskKey(secret),
		UpdatedAt:  pTime,
	}
}

func (v *DPAPIVault) AllStatuses() []CredentialStatus {
	routes := []string{RouteAnthropic, RouteGemini, RouteOpenAI}
	statuses := make([]CredentialStatus, 0, len(routes))
	for _, r := range routes {
		statuses = append(statuses, v.Status(r))
	}
	return statuses
}
