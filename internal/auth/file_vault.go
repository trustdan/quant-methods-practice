package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type vaultPayload struct {
	Version   int                  `json:"version"`
	Secrets   map[string]string    `json:"secrets"`
	Times     map[string]time.Time `json:"times"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// FileVault provides an encrypted on-disk credential vault using AES-256-GCM.
type FileVault struct {
	filePath string
	key      []byte
	mu       sync.RWMutex
	cached   *vaultPayload
}

// NewFileVault creates or opens a FileVault at filePath, deriving a 256-bit key from seed.
func NewFileVault(filePath string, keySeed []byte) (*FileVault, error) {
	if filePath == "" {
		return nil, errors.New("filePath cannot be empty")
	}

	// Derive 32-byte key using SHA-256
	h := sha256.New()
	h.Write(keySeed)
	key := h.Sum(nil)

	fv := &FileVault{
		filePath: filePath,
		key:      key,
	}

	// Attempt to load existing file if present
	if _, err := os.Stat(filePath); err == nil {
		if err := fv.load(); err != nil {
			return nil, fmt.Errorf("failed to load encrypted vault: %w", err)
		}
	} else {
		fv.cached = &vaultPayload{
			Version:   1,
			Secrets:   make(map[string]string),
			Times:     make(map[string]time.Time),
			UpdatedAt: time.Now().UTC(),
		}
	}

	return fv, nil
}

func (v *FileVault) load() error {
	data, err := os.ReadFile(v.filePath)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		v.cached = &vaultPayload{
			Version:   1,
			Secrets:   make(map[string]string),
			Times:     make(map[string]time.Time),
			UpdatedAt: time.Now().UTC(),
		}
		return nil
	}

	// Format: nonce (12 bytes) + ciphertext + GCM tag
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decryption failed (corrupt file or invalid key): %w", err)
	}

	var payload vaultPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return fmt.Errorf("invalid vault payload: %w", err)
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

func (v *FileVault) save() error {
	if v.cached == nil {
		return nil
	}

	v.cached.UpdatedAt = time.Now().UTC()
	plaintext, err := json.Marshal(v.cached)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(v.key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Ensure directory exists
	dir := filepath.Dir(v.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Write with restrictive permissions (0600)
	tmpFile := v.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, ciphertext, 0600); err != nil {
		return err
	}
	return os.Rename(tmpFile, v.filePath)
}

func (v *FileVault) Get(route string) (string, error) {
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

func (v *FileVault) Set(route string, secret string) error {
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

func (v *FileVault) Delete(route string) error {
	return v.Set(route, "")
}

func (v *FileVault) Status(route string) CredentialStatus {
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

func (v *FileVault) AllStatuses() []CredentialStatus {
	routes := []string{RouteAnthropic, RouteGemini, RouteOpenAI}
	statuses := make([]CredentialStatus, 0, len(routes))
	for _, r := range routes {
		statuses = append(statuses, v.Status(r))
	}
	return statuses
}
