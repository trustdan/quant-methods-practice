// Package siwc implements the "Sign in with ChatGPT" local/open-source
// token-sharing OAuth flow (public client, PKCE S256, dynamic registration).
//
// Protocol values were reverified on 2026-10-05 against
// https://developers.openai.com/siwc/token-sharing-open-source and
// https://auth.openai.com/.well-known/openid-configuration.
package siwc

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const (
	// DynamicClientID is the registration entry point. It is never saved or
	// used for token exchange; the callback supplies the issued client ID.
	DynamicClientID = "dynamic_agent_client"
	// AgentName is this app's own name, sent only on initial registration.
	AgentName = "Quant Methods Practice"
	// CallbackPath is fixed; only the loopback port may vary between attempts.
	CallbackPath = "/auth/callback"
	// RequestedScopes is the documented scope set for plan inference.
	RequestedScopes = "openid profile email offline_access resource.invoke chatgpt.tokens.use.direct"
	// PlanScope must be granted before any plan inference is attempted.
	PlanScope = "chatgpt.tokens.use.direct"
)

// Endpoints holds the OAuth/OIDC endpoints. Tests substitute a fake server.
type Endpoints struct {
	Issuer    string
	Authorize string
	Token     string
	Revoke    string
	JWKS      string
	Resource  string
}

// DefaultEndpoints returns OpenAI's published endpoints.
func DefaultEndpoints() Endpoints {
	return Endpoints{
		Issuer:    "https://auth.openai.com",
		Authorize: "https://auth.openai.com/api/accounts/authorize",
		Token:     "https://auth.openai.com/api/accounts/oauth/token",
		Revoke:    "https://auth.openai.com/api/accounts/oauth/revoke",
		JWKS:      "https://auth.openai.com/.well-known/jwks.json",
		Resource:  "https://api.openai.com/v1",
	}
}

// randomToken returns n random bytes encoded as unpadded base64url.
func randomToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand unavailable: %v", err))
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// NewPKCEVerifier returns a 43-character RFC 7636 code verifier.
func NewPKCEVerifier() string { return randomToken(32) }

// PKCEChallengeS256 derives the S256 code challenge for a verifier.
func PKCEChallengeS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func constantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// JWK is the subset of an RSA JSON Web Key needed for RS256 verification.
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKSet is a published key set.
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

func (k JWK) publicKey() (*rsa.PublicKey, error) {
	if k.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported key type %q", k.Kty)
	}
	nb, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("invalid modulus: %w", err)
	}
	eb, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("invalid exponent: %w", err)
	}
	e := new(big.Int).SetBytes(eb)
	if !e.IsInt64() || e.Int64() < 3 || e.Int64() > 1<<31-1 {
		return nil, errors.New("invalid exponent")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nb), E: int(e.Int64())}, nil
}

// IDClaims are the validated ID token claims this app uses.
type IDClaims struct {
	Issuer   string          `json:"iss"`
	Subject  string          `json:"sub"`
	Audience json.RawMessage `json:"aud"`
	Expiry   int64           `json:"exp"`
	IssuedAt int64           `json:"iat"`
	Nonce    string          `json:"nonce"`
	Email    string          `json:"email"`
	Name     string          `json:"name"`
}

func (c IDClaims) hasAudience(clientID string) bool {
	var single string
	if json.Unmarshal(c.Audience, &single) == nil {
		return single == clientID
	}
	var many []string
	if json.Unmarshal(c.Audience, &many) == nil {
		for _, a := range many {
			if a == clientID {
				return true
			}
		}
	}
	return false
}

// IDTokenExpectations are the values an ID token must match.
// An empty Nonce skips the nonce check (refresh responses carry none).
type IDTokenExpectations struct {
	Issuer   string
	ClientID string
	Nonce    string
	Now      time.Time
}

// ErrIDToken wraps every ID token validation failure.
var ErrIDToken = errors.New("id token rejected")

// VerifyIDToken checks the RS256 signature against keys, then issuer,
// audience, expiry and nonce, returning the validated claims.
func VerifyIDToken(raw string, keys JWKSet, want IDTokenExpectations) (*IDClaims, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("%w: malformed token", ErrIDToken)
	}
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: malformed header", ErrIDToken)
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("%w: malformed header", ErrIDToken)
	}
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("%w: unsupported alg %q", ErrIDToken, header.Alg)
	}

	var pub *rsa.PublicKey
	for _, k := range keys.Keys {
		if k.Kid == header.Kid {
			if pub, err = k.publicKey(); err != nil {
				return nil, fmt.Errorf("%w: %v", ErrIDToken, err)
			}
			break
		}
	}
	if pub == nil {
		return nil, fmt.Errorf("%w: unknown signing key", ErrIDToken)
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("%w: malformed signature", ErrIDToken)
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, digest[:], sig); err != nil {
		return nil, fmt.Errorf("%w: bad signature", ErrIDToken)
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: malformed payload", ErrIDToken)
	}
	var claims IDClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("%w: malformed payload", ErrIDToken)
	}

	switch {
	case claims.Issuer != want.Issuer:
		return nil, fmt.Errorf("%w: issuer mismatch", ErrIDToken)
	case !claims.hasAudience(want.ClientID):
		return nil, fmt.Errorf("%w: audience mismatch", ErrIDToken)
	case claims.Expiry == 0 || !want.Now.Before(time.Unix(claims.Expiry, 0)):
		return nil, fmt.Errorf("%w: expired", ErrIDToken)
	case want.Nonce != "" && !constantTimeEqual(claims.Nonce, want.Nonce):
		return nil, fmt.Errorf("%w: nonce mismatch", ErrIDToken)
	case claims.Subject == "":
		return nil, fmt.Errorf("%w: missing subject", ErrIDToken)
	}
	return &claims, nil
}
