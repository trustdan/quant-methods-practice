//go:build !windows

package auth

import "errors"

// DPAPIVault fallback stub on non-Windows platforms.
type DPAPIVault struct {
	*FileVault
}

func NewDPAPIVault(filePath string) (*DPAPIVault, error) {
	// Fallback to FileVault with standard machine seed
	fv, err := NewFileVault(filePath, []byte("quant-methods-practice-default-key"))
	if err != nil {
		return nil, err
	}
	return &DPAPIVault{FileVault: fv}, nil
}
