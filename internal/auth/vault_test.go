package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMaskKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"short", "••••••••"},
		{"12345678", "••••••••"},
		{"sk-ant-api03-abcdef1234", "sk-ant...1234"},
		{"sk-proj-1234567890abcdef", "sk...cdef"},
		{"AIzaSyAbcdEfghIjkl1234", "AIzaSy...1234"},
		{"random-api-key-9999", "rand...9999"},
	}

	for _, tc := range tests {
		actual := MaskKey(tc.input)
		if actual != tc.expected {
			t.Errorf("MaskKey(%q) = %q; expected %q", tc.input, actual, tc.expected)
		}
	}
}

func TestMemoryVault(t *testing.T) {
	v := NewMemoryVault()

	// Initial status
	st := v.Status(RouteAnthropic)
	if st.Configured || st.Source != SourceNone {
		t.Errorf("expected not configured, got %+v", st)
	}

	// Set credential
	err := v.Set(RouteAnthropic, "sk-ant-test-key-1234")
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	// Get credential
	val, err := v.Get(RouteAnthropic)
	if err != nil || val != "sk-ant-test-key-1234" {
		t.Errorf("expected key retrieved, got %q, err: %v", val, err)
	}

	// Status after set
	st = v.Status(RouteAnthropic)
	if !st.Configured || st.Source != SourceSession {
		t.Errorf("expected session configured, got %+v", st)
	}
	if st.MaskedKey != "sk-ant...1234" {
		t.Errorf("unexpected masked key: %s", st.MaskedKey)
	}

	// Delete
	err = v.Delete(RouteAnthropic)
	if err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	_, err = v.Get(RouteAnthropic)
	if err == nil {
		t.Error("expected error getting deleted key, got nil")
	}
}

func TestFileVault(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "test.vault")
	seed := []byte("test-encryption-seed")

	v, err := NewFileVault(vaultPath, seed)
	if err != nil {
		t.Fatalf("failed to create FileVault: %v", err)
	}

	err = v.Set(RouteOpenAI, "sk-proj-test-openai-9988")
	if err != nil {
		t.Fatalf("failed to set secret: %v", err)
	}

	// Reload from disk
	v2, err := NewFileVault(vaultPath, seed)
	if err != nil {
		t.Fatalf("failed to reload FileVault: %v", err)
	}

	val, err := v2.Get(RouteOpenAI)
	if err != nil || val != "sk-proj-test-openai-9988" {
		t.Fatalf("expected recovered secret, got %q, err: %v", val, err)
	}

	st := v2.Status(RouteOpenAI)
	if !st.Configured || st.Source != SourceVault {
		t.Errorf("expected vault source, got %+v", st)
	}

	// Wrong seed cannot decrypt
	_, err = NewFileVault(vaultPath, []byte("wrong-seed"))
	if err == nil {
		t.Error("expected error with wrong decryption seed, got nil")
	}
}

func TestStandardVaultPrecedence(t *testing.T) {
	mem := NewMemoryVault()
	sv := NewStandardVaultWithInner(mem)

	// Set env var
	_ = os.Setenv("ANTHROPIC_API_KEY", "sk-ant-from-env-0001")
	defer func() {
		_ = os.Unsetenv("ANTHROPIC_API_KEY")
	}()

	// 1. Initially should resolve from environment
	val, err := sv.Get(RouteAnthropic)
	if err != nil || val != "sk-ant-from-env-0001" {
		t.Fatalf("expected env key, got %q, err: %v", val, err)
	}
	st := sv.Status(RouteAnthropic)
	if !st.Configured || st.Source != SourceEnvironment {
		t.Errorf("expected environment source, got %+v", st)
	}

	// 2. Setting vault key overrides environment
	err = sv.Set(RouteAnthropic, "sk-ant-from-vault-9999")
	if err != nil {
		t.Fatalf("unexpected set error: %v", err)
	}

	val, err = sv.Get(RouteAnthropic)
	if err != nil || val != "sk-ant-from-vault-9999" {
		t.Fatalf("expected vault key override, got %q, err: %v", val, err)
	}
	st = sv.Status(RouteAnthropic)
	if !st.Configured || st.Source != SourceSession {
		t.Errorf("expected session source, got %+v", st)
	}

	// 3. Deleting vault key falls back to environment
	err = sv.Delete(RouteAnthropic)
	if err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	val, err = sv.Get(RouteAnthropic)
	if err != nil || val != "sk-ant-from-env-0001" {
		t.Fatalf("expected fallback to env key, got %q, err: %v", val, err)
	}
	st = sv.Status(RouteAnthropic)
	if !st.Configured || st.Source != SourceEnvironment {
		t.Errorf("expected fallback to env source, got %+v", st)
	}
}
