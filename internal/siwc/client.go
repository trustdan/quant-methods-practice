package siwc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/trustdan/quant-methods-practice/internal/auth"
)

// vaultKey is the vault entry holding the host ID and per-account records.
// The vault encrypts it at rest (DPAPI on Windows, AES-GCM file elsewhere).
const vaultKey = "chatgpt_plan_accounts"

const (
	attemptTimeout  = 10 * time.Minute
	exchangeTimeout = 30 * time.Second
	refreshLeeway   = 60 * time.Second
)

var (
	// ErrNotSignedIn means no ChatGPT account is selected.
	ErrNotSignedIn = errors.New("no ChatGPT account is signed in")
	// ErrReauthRequired means the saved tokens are no longer usable.
	ErrReauthRequired = errors.New("ChatGPT sign-in expired; sign in again")
	// ErrPlanScopeMissing means the grant does not permit plan inference.
	ErrPlanScopeMissing = errors.New("this ChatGPT account did not grant plan usage (chatgpt.tokens.use.direct)")
)

// Account is one ChatGPT account registration. Each account keeps its own
// issued client ID. Never serialized to the frontend; see AccountSummary.
type Account struct {
	Key               string    `json:"key"`
	Subject           string    `json:"sub"`
	ClientID          string    `json:"client_id"`
	Email             string    `json:"email,omitempty"`
	Name              string    `json:"name,omitempty"`
	AccessToken       string    `json:"access_token,omitempty"`
	RefreshToken      string    `json:"refresh_token,omitempty"`
	IDToken           string    `json:"id_token,omitempty"`
	Scopes            []string  `json:"scopes"`
	ExpiresAt         time.Time `json:"expires_at"`
	EarliestRefreshAt time.Time `json:"earliest_refresh_at,omitempty"`
	NeedsReauth       bool      `json:"needs_reauth,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (a *Account) hasScope(s string) bool {
	for _, x := range a.Scopes {
		if x == s {
			return true
		}
	}
	return false
}

// AccountSummary is the redacted account view safe for the UI.
type AccountSummary struct {
	Key         string    `json:"key"`
	Label       string    `json:"label"`
	Email       string    `json:"email,omitempty"`
	PlanGranted bool      `json:"plan_granted"`
	NeedsReauth bool      `json:"needs_reauth"`
	Selected    bool      `json:"selected"`
	ExpiresAt   time.Time `json:"expires_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type storeData struct {
	HostID   string              `json:"host_id"`
	Selected string              `json:"selected,omitempty"`
	Accounts map[string]*Account `json:"accounts"`
}

// SignInStatus reports the current sign-in attempt to the UI.
type SignInStatus struct {
	AttemptID string          `json:"attempt_id,omitempty"`
	State     string          `json:"state"` // idle, pending, complete, failed, cancelled, expired
	Error     string          `json:"error,omitempty"`
	Account   *AccountSummary `json:"account,omitempty"`
}

type attempt struct {
	id          string
	state       string
	nonce       string
	verifier    string
	redirectURI string
	host        string
	clientID    string // DynamicClientID for new registration
	reauthKey   string
	status      SignInStatus
	srv         *http.Server
	cancel      context.CancelFunc
}

// Client runs the sign-in flow and supplies fresh access tokens.
type Client struct {
	Endpoints  Endpoints
	HTTPClient *http.Client
	Now        func() time.Time

	vault     auth.Vault
	mu        sync.Mutex
	refreshMu sync.Mutex
	current   *attempt
	jwks      *JWKSet
	jwksAt    time.Time
}

// NewClient creates a client persisting accounts in vault.
func NewClient(vault auth.Vault) *Client {
	return &Client{
		Endpoints:  DefaultEndpoints(),
		HTTPClient: &http.Client{Timeout: exchangeTimeout},
		Now:        func() time.Time { return time.Now().UTC() },
		vault:      vault,
	}
}

func (c *Client) load() (*storeData, error) {
	d := &storeData{Accounts: map[string]*Account{}}
	raw, err := c.vault.Get(vaultKey)
	if err != nil || raw == "" {
		return d, nil
	}
	if err := json.Unmarshal([]byte(raw), d); err != nil {
		return nil, fmt.Errorf("corrupt ChatGPT account store: %w", err)
	}
	if d.Accounts == nil {
		d.Accounts = map[string]*Account{}
	}
	return d, nil
}

func (c *Client) save(d *storeData) error {
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	return c.vault.Set(vaultKey, string(b))
}

func accountKey(sub string) string {
	sum := sha256.Sum256([]byte(sub))
	return hex.EncodeToString(sum[:6])
}

func summarize(a *Account, selected bool) AccountSummary {
	label := a.Email
	if label == "" {
		label = a.Name
	}
	if label == "" {
		label = "ChatGPT account " + a.Key
	}
	return AccountSummary{
		Key:         a.Key,
		Label:       label,
		Email:       a.Email,
		PlanGranted: a.hasScope(PlanScope),
		NeedsReauth: a.NeedsReauth || a.RefreshToken == "",
		Selected:    selected,
		ExpiresAt:   a.ExpiresAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

// Accounts lists redacted account summaries, sorted by label.
func (c *Client) Accounts() ([]AccountSummary, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, err := c.load()
	if err != nil {
		return nil, err
	}
	out := make([]AccountSummary, 0, len(d.Accounts))
	for _, a := range d.Accounts {
		out = append(out, summarize(a, a.Key == d.Selected))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out, nil
}

// SelectedAccount returns the selected account's summary, if any.
func (c *Client) SelectedAccount() (*AccountSummary, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, err := c.load()
	if err != nil {
		return nil, false
	}
	a, ok := d.Accounts[d.Selected]
	if !ok {
		return nil, false
	}
	s := summarize(a, true)
	return &s, true
}

// Select makes a saved account the one used for plan inference.
func (c *Client) Select(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, err := c.load()
	if err != nil {
		return err
	}
	if _, ok := d.Accounts[key]; !ok {
		return fmt.Errorf("unknown ChatGPT account %q", key)
	}
	d.Selected = key
	return c.save(d)
}

// BeginSignIn starts a fresh attempt and returns the authorization URL to
// open in the user's browser. Pass a saved account key to reauthorize that
// account with its issued client ID; empty registers a new account.
func (c *Client) BeginSignIn(reauthKey string) (string, SignInStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stopAttemptLocked("cancelled")

	d, err := c.load()
	if err != nil {
		return "", SignInStatus{}, err
	}
	if d.HostID == "" {
		d.HostID = "urn:uuid:" + uuid.NewString()
		if err := c.save(d); err != nil {
			return "", SignInStatus{}, fmt.Errorf("persist host identifier: %w", err)
		}
	}

	at := &attempt{
		id:       randomToken(12),
		state:    randomToken(32),
		nonce:    randomToken(32),
		verifier: NewPKCEVerifier(),
		clientID: DynamicClientID,
	}
	q := url.Values{}
	if reauthKey != "" {
		acct, ok := d.Accounts[reauthKey]
		if !ok {
			return "", SignInStatus{}, fmt.Errorf("unknown ChatGPT account %q", reauthKey)
		}
		at.clientID = acct.ClientID
		at.reauthKey = reauthKey
		if acct.IDToken != "" {
			q.Set("id_token_hint", acct.IDToken)
		}
	} else {
		q.Set("agent_name_hint", AgentName)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", SignInStatus{}, fmt.Errorf("open loopback callback listener: %w", err)
	}
	at.host = ln.Addr().String()
	at.redirectURI = "http://" + at.host + CallbackPath

	q.Set("client_id", at.clientID)
	q.Set("ext_agent_host_id", d.HostID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", at.redirectURI)
	q.Set("scope", RequestedScopes)
	q.Set("resource", c.Endpoints.Resource)
	q.Set("state", at.state)
	q.Set("nonce", at.nonce)
	q.Set("code_challenge_method", "S256")
	q.Set("code_challenge", PKCEChallengeS256(at.verifier))

	ctx, cancel := context.WithTimeout(context.Background(), attemptTimeout)
	at.cancel = cancel
	at.status = SignInStatus{AttemptID: at.id, State: "pending"}

	mux := http.NewServeMux()
	mux.HandleFunc(CallbackPath, func(w http.ResponseWriter, r *http.Request) {
		c.handleCallback(ctx, at, w, r)
	})
	at.srv = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = at.srv.Serve(ln) }()
	go func() {
		<-ctx.Done()
		c.mu.Lock()
		if at.status.State == "pending" {
			at.status.State = "expired"
			at.status.Error = "Sign-in timed out. Start again when ready."
		}
		c.mu.Unlock()
		// Shutdown lets an in-flight callback page finish writing.
		sctx, scancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer scancel()
		_ = at.srv.Shutdown(sctx)
	}()

	c.current = at
	return c.Endpoints.Authorize + "?" + q.Encode(), at.status, nil
}

// Status reports the latest attempt.
func (c *Client) Status() SignInStatus {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.current == nil {
		return SignInStatus{State: "idle"}
	}
	return c.current.status
}

// CancelSignIn abandons a pending attempt and closes its listener.
func (c *Client) CancelSignIn() SignInStatus {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stopAttemptLocked("cancelled")
	if c.current == nil {
		return SignInStatus{State: "idle"}
	}
	return c.current.status
}

func (c *Client) stopAttemptLocked(state string) {
	if c.current == nil {
		return
	}
	if c.current.status.State == "pending" {
		c.current.status.State = state
	}
	c.current.cancel()
}

const callbackPage = `<!doctype html><html><head><meta charset="utf-8"><title>%s</title></head>
<body style="font-family:system-ui,sans-serif;max-width:32rem;margin:4rem auto;padding:0 1rem">
<h1 style="font-size:1.25rem">%s</h1><p>%s</p></body></html>`

func writeCallbackPage(w http.ResponseWriter, code int, title, msg string) {
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'")
	h.Set("Cache-Control", "no-store")
	h.Set("Referrer-Policy", "no-referrer")
	w.WriteHeader(code)
	esc := func(s string) string {
		return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
	}
	fmt.Fprintf(w, callbackPage, esc(title), esc(title), esc(msg))
}

func (c *Client) handleCallback(ctx context.Context, at *attempt, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.Host != at.host {
		writeCallbackPage(w, http.StatusBadRequest, "Invalid request", "This callback only accepts the sign-in redirect.")
		return
	}
	q := r.URL.Query()

	c.mu.Lock()
	pending := c.current == at && at.status.State == "pending"
	stateOK := constantTimeEqual(q.Get("state"), at.state)
	c.mu.Unlock()

	// A mismatched state is never allowed to consume or abort the attempt.
	if !pending || !stateOK {
		writeCallbackPage(w, http.StatusBadRequest, "Sign-in not accepted",
			"This sign-in response does not match an active request. Return to Quant Methods Practice and start again.")
		return
	}

	acct, err := c.completeCallback(ctx, at, q)

	c.mu.Lock()
	if err != nil {
		at.status.State = "failed"
		at.status.Error = err.Error()
	} else {
		at.status.State = "complete"
		at.status.Account = acct
	}
	c.mu.Unlock()
	at.cancel()

	if err != nil {
		writeCallbackPage(w, http.StatusOK, "Sign-in failed", err.Error()+" You can close this tab.")
		return
	}
	writeCallbackPage(w, http.StatusOK, "Signed in to ChatGPT",
		"Signed in as "+acct.Label+". You can close this tab and return to Quant Methods Practice.")
}

func (c *Client) completeCallback(ctx context.Context, at *attempt, q url.Values) (*AccountSummary, error) {
	if e := q.Get("error"); e != "" {
		if d := q.Get("error_description"); d != "" {
			return nil, fmt.Errorf("authorization denied (%s): %s", e, d)
		}
		return nil, fmt.Errorf("authorization denied (%s)", e)
	}
	code := q.Get("code")
	if code == "" {
		return nil, errors.New("authorization response had no code")
	}

	clientID := at.clientID
	if clientID == DynamicClientID {
		clientID = q.Get("client_id")
		if clientID == "" || clientID == DynamicClientID {
			return nil, errors.New("registration did not return an issued client ID")
		}
	}

	ctx, cancel := context.WithTimeout(ctx, exchangeTimeout)
	defer cancel()
	tok, err := c.postToken(ctx, url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code":          {code},
		"code_verifier": {at.verifier},
		"redirect_uri":  {at.redirectURI},
		"resource":      {c.Endpoints.Resource},
	})
	if err != nil {
		return nil, err
	}
	if tok.IDToken == "" || tok.AccessToken == "" {
		return nil, errors.New("token response missing id_token or access_token")
	}

	claims, err := c.verifyIDToken(ctx, tok.IDToken, clientID, at.nonce)
	if err != nil {
		return nil, err
	}

	scopeStr := tok.Scope
	if scopeStr == "" {
		scopeStr = q.Get("scope")
	}
	scopes := strings.Fields(scopeStr)
	planGranted := false
	for _, s := range scopes {
		planGranted = planGranted || s == PlanScope
	}
	if !planGranted {
		return nil, ErrPlanScopeMissing
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	d, err := c.load()
	if err != nil {
		return nil, err
	}
	if c.current != at || at.status.State != "pending" {
		return nil, errors.New("sign-in was cancelled before it completed; nothing was saved")
	}
	key := accountKey(claims.Subject)
	if at.reauthKey != "" && at.reauthKey != key {
		return nil, errors.New("signed in to a different ChatGPT account than the one being renewed; nothing was replaced")
	}

	now := c.Now()
	acct, exists := d.Accounts[key]
	if !exists {
		acct = &Account{Key: key, Subject: claims.Subject, CreatedAt: now}
		d.Accounts[key] = acct
	}
	acct.ClientID = clientID
	acct.Email = claims.Email
	acct.Name = claims.Name
	acct.AccessToken = tok.AccessToken
	acct.RefreshToken = tok.RefreshToken
	acct.IDToken = tok.IDToken
	acct.Scopes = scopes
	acct.ExpiresAt = now.Add(time.Duration(tok.ExpiresIn) * time.Second)
	acct.EarliestRefreshAt = tok.earliestRefresh()
	acct.NeedsReauth = false
	acct.UpdatedAt = now
	d.Selected = key
	if err := c.save(d); err != nil {
		return nil, fmt.Errorf("save ChatGPT credentials: %w", err)
	}
	s := summarize(acct, true)
	return &s, nil
}

func (c *Client) verifyIDToken(ctx context.Context, raw, clientID, nonce string) (*IDClaims, error) {
	keys, err := c.keys(ctx, false)
	if err != nil {
		return nil, err
	}
	want := IDTokenExpectations{Issuer: c.Endpoints.Issuer, ClientID: clientID, Nonce: nonce, Now: c.Now()}
	claims, err := VerifyIDToken(raw, *keys, want)
	if err != nil && strings.Contains(err.Error(), "unknown signing key") {
		// Keys rotate; refetch once before rejecting.
		if keys, err = c.keys(ctx, true); err != nil {
			return nil, err
		}
		claims, err = VerifyIDToken(raw, *keys, want)
	}
	return claims, err
}

func (c *Client) keys(ctx context.Context, force bool) (*JWKSet, error) {
	c.mu.Lock()
	if !force && c.jwks != nil && c.Now().Sub(c.jwksAt) < time.Hour {
		k := c.jwks
		c.mu.Unlock()
		return k, nil
	}
	c.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Endpoints.JWKS, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch signing keys: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch signing keys: status %d", resp.StatusCode)
	}
	var set JWKSet
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&set); err != nil {
		return nil, fmt.Errorf("decode signing keys: %w", err)
	}
	c.mu.Lock()
	c.jwks, c.jwksAt = &set, c.Now()
	c.mu.Unlock()
	return &set, nil
}

type tokenResponse struct {
	AccessToken       string          `json:"access_token"`
	RefreshToken      string          `json:"refresh_token"`
	IDToken           string          `json:"id_token"`
	TokenType         string          `json:"token_type"`
	ExpiresIn         int64           `json:"expires_in"`
	Scope             string          `json:"scope"`
	EarliestRefreshAt json.RawMessage `json:"earliest_refresh_at"`
}

// earliestRefresh accepts Unix seconds or an RFC 3339 string; the reference
// lists the field without fixing its encoding.
func (t tokenResponse) earliestRefresh() time.Time {
	raw := strings.Trim(string(t.EarliestRefreshAt), `"`)
	if raw == "" || raw == "null" {
		return time.Time{}
	}
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return time.Unix(n, 0).UTC()
	}
	if ts, err := time.Parse(time.RFC3339, raw); err == nil {
		return ts.UTC()
	}
	return time.Time{}
}

// OAuthError is a token endpoint error with its documented code.
type OAuthError struct {
	Status int
	Code   string
	Desc   string
}

func (e *OAuthError) Error() string {
	if e.Desc != "" {
		return fmt.Sprintf("ChatGPT token request failed (%s): %s", e.Code, e.Desc)
	}
	return fmt.Sprintf("ChatGPT token request failed (%s, status %d)", e.Code, e.Status)
}

// reauthCodes are the documented refresh errors that require a new sign-in.
var reauthCodes = map[string]bool{
	"invalid_grant": true, "invalid_refresh_token": true, "token_expired": true,
	"refresh_token_expired": true, "refresh_token_invalidated": true, "refresh_token_reused": true,
}

func parseOAuthError(status int, body []byte) *OAuthError {
	var flat struct {
		Error            json.RawMessage `json:"error"`
		ErrorDescription string          `json:"error_description"`
	}
	e := &OAuthError{Status: status, Code: "unknown_error"}
	if json.Unmarshal(body, &flat) != nil {
		return e
	}
	var code string
	if json.Unmarshal(flat.Error, &code) == nil && code != "" {
		e.Code, e.Desc = code, flat.ErrorDescription
		return e
	}
	var nested struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if json.Unmarshal(flat.Error, &nested) == nil && nested.Code != "" {
		e.Code, e.Desc = nested.Code, nested.Message
	}
	return e
}

func (c *Client) postToken(ctx context.Context, form url.Values) (*tokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoints.Token, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ChatGPT token request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, parseOAuthError(resp.StatusCode, body)
	}
	var tok tokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	return &tok, nil
}

// AccessToken returns a valid access token for the selected account,
// refreshing it when it is within a minute of expiry. It never falls back
// to any other credential.
func (c *Client) AccessToken(ctx context.Context) (string, error) {
	c.refreshMu.Lock()
	defer c.refreshMu.Unlock()

	c.mu.Lock()
	d, err := c.load()
	if err != nil {
		c.mu.Unlock()
		return "", err
	}
	acct, ok := d.Accounts[d.Selected]
	if !ok {
		c.mu.Unlock()
		return "", ErrNotSignedIn
	}
	snapshot := *acct
	c.mu.Unlock()

	switch {
	case !snapshot.hasScope(PlanScope):
		return "", ErrPlanScopeMissing
	case snapshot.NeedsReauth:
		return "", ErrReauthRequired
	case snapshot.AccessToken != "" && c.Now().Add(refreshLeeway).Before(snapshot.ExpiresAt):
		return snapshot.AccessToken, nil
	case snapshot.RefreshToken == "":
		c.markReauth(snapshot.Key)
		return "", ErrReauthRequired
	}

	tok, err := c.postToken(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {snapshot.ClientID},
		"refresh_token": {snapshot.RefreshToken},
		"resource":      {c.Endpoints.Resource},
	})
	if err != nil {
		var oe *OAuthError
		if errors.As(err, &oe) && reauthCodes[oe.Code] {
			c.markReauth(snapshot.Key)
			return "", ErrReauthRequired
		}
		return "", err
	}
	if tok.AccessToken == "" {
		return "", errors.New("refresh response missing access_token")
	}
	if tok.IDToken != "" {
		claims, err := c.verifyIDToken(ctx, tok.IDToken, snapshot.ClientID, "")
		if err != nil {
			return "", err
		}
		if accountKey(claims.Subject) != snapshot.Key {
			c.markReauth(snapshot.Key)
			return "", errors.New("refreshed identity does not match the saved ChatGPT account")
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	d, err = c.load()
	if err != nil {
		return "", err
	}
	acct, ok = d.Accounts[snapshot.Key]
	if !ok {
		return "", ErrNotSignedIn // signed out while refreshing
	}
	now := c.Now()
	acct.AccessToken = tok.AccessToken
	if tok.RefreshToken != "" {
		acct.RefreshToken = tok.RefreshToken // rotation: old token is now spent
	}
	if tok.IDToken != "" {
		acct.IDToken = tok.IDToken
	}
	if tok.Scope != "" {
		acct.Scopes = strings.Fields(tok.Scope)
	}
	acct.ExpiresAt = now.Add(time.Duration(tok.ExpiresIn) * time.Second)
	acct.EarliestRefreshAt = tok.earliestRefresh()
	acct.UpdatedAt = now
	if err := c.save(d); err != nil {
		return "", fmt.Errorf("save refreshed ChatGPT credentials: %w", err)
	}
	if !acct.hasScope(PlanScope) {
		return "", ErrPlanScopeMissing
	}
	return acct.AccessToken, nil
}

// MarkSelectedReauth flags the selected account after the API rejects its
// credentials (subscription_sharing_invalid_user).
func (c *Client) MarkSelectedReauth() {
	c.mu.Lock()
	d, err := c.load()
	c.mu.Unlock()
	if err == nil && d.Selected != "" {
		c.markReauth(d.Selected)
	}
}

func (c *Client) markReauth(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, err := c.load()
	if err != nil {
		return
	}
	if a, ok := d.Accounts[key]; ok {
		a.AccessToken, a.RefreshToken = "", ""
		a.NeedsReauth = true
		a.UpdatedAt = c.Now()
		_ = c.save(d)
	}
}

// SignOut revokes the account's tokens, then removes its local record.
// The record is removed even if revocation fails; revoked reports whether
// the server confirmed revocation.
func (c *Client) SignOut(ctx context.Context, key string) (revoked bool, err error) {
	c.refreshMu.Lock()
	defer c.refreshMu.Unlock()

	c.mu.Lock()
	d, err := c.load()
	if err != nil {
		c.mu.Unlock()
		return false, err
	}
	acct, ok := d.Accounts[key]
	if !ok {
		c.mu.Unlock()
		return false, fmt.Errorf("unknown ChatGPT account %q", key)
	}
	snapshot := *acct
	c.mu.Unlock()

	revoked = true
	for _, t := range []struct{ tok, hint string }{
		{snapshot.RefreshToken, "refresh_token"},
		{snapshot.AccessToken, "access_token"},
	} {
		if t.tok == "" {
			continue
		}
		if c.revoke(ctx, snapshot.ClientID, t.tok, t.hint) != nil {
			revoked = false
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if d, err = c.load(); err != nil {
		return revoked, err
	}
	delete(d.Accounts, key)
	if d.Selected == key {
		d.Selected = ""
	}
	return revoked, c.save(d)
}

func (c *Client) revoke(ctx context.Context, clientID, token, hint string) error {
	form := url.Values{"client_id": {clientID}, "token": {token}, "token_type_hint": {hint}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoints.Revoke, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revocation status %d", resp.StatusCode)
	}
	return nil
}
