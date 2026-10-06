package siwc

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/auth"
)

// fakeAuth imitates auth.openai.com for the documented flow.
type fakeAuth struct {
	t   *testing.T
	key *rsa.PrivateKey
	srv *httptest.Server
	now func() time.Time

	mu            sync.Mutex
	sub           string
	email         string
	issuedClient  string
	grantedScope  string
	badNonce      bool
	badAudience   bool
	expiredID     bool
	signWithOther bool
	refreshErr    string
	challenge     string
	nonce         string
	redirectURI   string
	lastAuthorize url.Values
	refreshCount  int
	revoked       []string
}

func newFakeAuth(t *testing.T, now func() time.Time) *fakeAuth {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeAuth{
		t: t, key: key, now: now,
		sub: "user-abc", email: "learner@example.com", issuedClient: "oaiapp_test123",
		grantedScope: RequestedScopes,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/authorize", f.authorize)
	mux.HandleFunc("/token", f.token)
	mux.HandleFunc("/revoke", f.revoke)
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(JWKSet{Keys: []JWK{{
			Kty: "RSA", Kid: "kid-1", Alg: "RS256",
			N: base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		}}})
	})
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeAuth) endpoints() Endpoints {
	return Endpoints{
		Issuer: "https://auth.openai.com", Authorize: f.srv.URL + "/authorize",
		Token: f.srv.URL + "/token", Revoke: f.srv.URL + "/revoke",
		JWKS: f.srv.URL + "/jwks", Resource: "https://api.openai.com/v1",
	}
}

func (f *fakeAuth) authorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f.mu.Lock()
	f.lastAuthorize = q
	f.challenge = q.Get("code_challenge")
	f.nonce = q.Get("nonce")
	f.redirectURI = q.Get("redirect_uri")
	issued, scope := f.issuedClient, f.grantedScope
	f.mu.Unlock()

	back := url.Values{"code": {"auth-code-1"}, "state": {q.Get("state")}, "scope": {scope}}
	if q.Get("client_id") == DynamicClientID {
		back.Set("client_id", issued)
	}
	http.Redirect(w, r, q.Get("redirect_uri")+"?"+back.Encode(), http.StatusFound)
}

func (f *fakeAuth) idToken(aud, nonce string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	exp := f.now().Add(time.Hour)
	if f.expiredID {
		exp = f.now().Add(-time.Minute)
	}
	if f.badAudience {
		aud = "oaiapp_someone_else"
	}
	if f.badNonce {
		nonce = "tampered"
	}
	header, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": "kid-1", "typ": "JWT"})
	claims, _ := json.Marshal(map[string]any{
		"iss": "https://auth.openai.com", "sub": f.sub, "aud": aud, "exp": exp.Unix(),
		"iat": f.now().Unix(), "nonce": nonce, "email": f.email,
	})
	signing := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(claims)
	key := f.key
	if f.signWithOther {
		key, _ = rsa.GenerateKey(rand.Reader, 2048)
	}
	sum := sha256.Sum256([]byte(signing))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
	if err != nil {
		f.t.Fatal(err)
	}
	return signing + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func (f *fakeAuth) token(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	w.Header().Set("Content-Type", "application/json")
	fail := func(code string) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
	}
	if r.Form.Get("resource") != "https://api.openai.com/v1" {
		fail("invalid_target")
		return
	}
	f.mu.Lock()
	issued, challenge, redirect, scope, nonce := f.issuedClient, f.challenge, f.redirectURI, f.grantedScope, f.nonce
	refreshErr := f.refreshErr
	f.mu.Unlock()

	switch r.Form.Get("grant_type") {
	case "authorization_code":
		switch {
		case r.Form.Get("client_id") != issued:
			fail("invalid_client")
			return
		case PKCEChallengeS256(r.Form.Get("code_verifier")) != challenge:
			fail("invalid_grant")
			return
		case r.Form.Get("redirect_uri") != redirect || r.Form.Get("code") != "auth-code-1":
			fail("invalid_grant")
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "access-1", "refresh_token": "refresh-1", "id_token": f.idToken(issued, nonce),
			"token_type": "Bearer", "expires_in": 3600, "scope": scope,
		})
	case "refresh_token":
		if refreshErr != "" {
			fail(refreshErr)
			return
		}
		f.mu.Lock()
		f.refreshCount++
		n := f.refreshCount
		f.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "access-r" + string(rune('0'+n)), "refresh_token": "refresh-r" + string(rune('0'+n)),
			"token_type": "Bearer", "expires_in": 3600, "scope": scope,
		})
	default:
		fail("unsupported_grant_type")
	}
}

func (f *fakeAuth) revoke(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	f.mu.Lock()
	f.revoked = append(f.revoked, r.Form.Get("token_type_hint")+":"+r.Form.Get("token"))
	f.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *clock) advance(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

func setup(t *testing.T) (*Client, *fakeAuth, *clock, auth.Vault) {
	clk := &clock{t: time.Now().UTC()}
	f := newFakeAuth(t, clk.now)
	v := auth.NewMemoryVault()
	c := NewClient(v)
	c.Endpoints = f.endpoints()
	c.Now = clk.now
	return c, f, clk, v
}

// browse plays the user's browser: follow the authorize redirect to the
// loopback callback and return the callback's HTTP status.
func browse(t *testing.T, authorizeURL string) int {
	t.Helper()
	resp, err := http.Get(authorizeURL)
	if err != nil {
		t.Fatalf("browser navigation failed: %v", err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

func signIn(t *testing.T, c *Client, reauth string) SignInStatus {
	t.Helper()
	u, _, err := c.BeginSignIn(reauth)
	if err != nil {
		t.Fatal(err)
	}
	browse(t, u)
	return c.Status()
}

func TestPKCEChallengeMatchesRFC7636Vector(t *testing.T) {
	got := PKCEChallengeS256("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk")
	if got != "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM" {
		t.Fatalf("S256 challenge = %q", got)
	}
	if v := NewPKCEVerifier(); len(v) < 43 || v == NewPKCEVerifier() {
		t.Fatalf("verifier must be >=43 chars and fresh: %q", v)
	}
}

func TestNewRegistrationSignInStoresIssuedClientAndSelectsAccount(t *testing.T) {
	c, f, _, v := setup(t)

	st := signIn(t, c, "")
	if st.State != "complete" || st.Account == nil || !st.Account.PlanGranted {
		t.Fatalf("sign-in status = %+v", st)
	}

	q := f.lastAuthorize
	checks := map[string]string{
		"client_id": DynamicClientID, "response_type": "code", "code_challenge_method": "S256",
		"scope": RequestedScopes, "resource": "https://api.openai.com/v1", "agent_name_hint": AgentName,
	}
	for k, want := range checks {
		if q.Get(k) != want {
			t.Errorf("authorize %s = %q, want %q", k, q.Get(k), want)
		}
	}
	if !strings.HasPrefix(q.Get("redirect_uri"), "http://127.0.0.1:") || !strings.HasSuffix(q.Get("redirect_uri"), CallbackPath) {
		t.Errorf("redirect_uri = %q", q.Get("redirect_uri"))
	}
	if !strings.HasPrefix(q.Get("ext_agent_host_id"), "urn:uuid:") || q.Get("state") == "" || q.Get("nonce") == "" {
		t.Errorf("missing host id/state/nonce: %v", q)
	}

	d, _ := c.load()
	acct := d.Accounts[d.Selected]
	if acct == nil || acct.ClientID != "oaiapp_test123" || acct.Subject != "user-abc" {
		t.Fatalf("stored account = %+v", acct)
	}
	tok, err := c.AccessToken(context.Background())
	if err != nil || tok != "access-1" {
		t.Fatalf("AccessToken = %q, %v", tok, err)
	}

	// Host identity persists across attempts; reauth uses the issued client ID.
	hostID := q.Get("ext_agent_host_id")
	signIn(t, c, acct.Key)
	if f.lastAuthorize.Get("ext_agent_host_id") != hostID {
		t.Error("host identifier changed between attempts")
	}
	if f.lastAuthorize.Get("client_id") != "oaiapp_test123" || f.lastAuthorize.Has("agent_name_hint") {
		t.Errorf("reauth must use issued client and omit agent_name_hint: %v", f.lastAuthorize)
	}
	if f.lastAuthorize.Get("id_token_hint") == "" {
		t.Error("reauth should send retained id_token_hint")
	}

	// Summaries never expose tokens.
	sums, _ := c.Accounts()
	b, _ := json.Marshal(sums)
	for _, secret := range []string{"access-1", "refresh-1", acct.IDToken} {
		if strings.Contains(string(b), secret) {
			t.Fatalf("account summary leaked a token: %s", b)
		}
	}
	_ = v
}

func TestCallbackStateMismatchDoesNotConsumeAttempt(t *testing.T) {
	c, _, _, _ := setup(t)
	authURL, _, err := c.BeginSignIn("")
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(authURL)
	redirect := u.Query().Get("redirect_uri")

	resp, err := http.Get(redirect + "?code=evil&state=forged")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("forged state status = %d", resp.StatusCode)
	}
	if st := c.Status(); st.State != "pending" {
		t.Fatalf("forged callback changed attempt state to %q", st.State)
	}
	browse(t, authURL)
	if st := c.Status(); st.State != "complete" {
		t.Fatalf("genuine callback after forgery: %+v", st)
	}
}

func TestSignInRejectsInvalidIDTokensAndMissingPlanScope(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*fakeAuth)
		want  string
	}{
		{"nonce mismatch", func(f *fakeAuth) { f.badNonce = true }, "nonce mismatch"},
		{"audience mismatch", func(f *fakeAuth) { f.badAudience = true }, "audience mismatch"},
		{"expired id token", func(f *fakeAuth) { f.expiredID = true }, "expired"},
		{"forged signature", func(f *fakeAuth) { f.signWithOther = true }, "bad signature"},
		{"plan scope not granted", func(f *fakeAuth) { f.grantedScope = "openid profile email offline_access" }, "chatgpt.tokens.use.direct"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, f, _, _ := setup(t)
			tc.setup(f)
			st := signIn(t, c, "")
			if st.State != "failed" || !strings.Contains(st.Error, tc.want) {
				t.Fatalf("status = %+v, want failure containing %q", st, tc.want)
			}
			if accts, _ := c.Accounts(); len(accts) != 0 {
				t.Fatalf("rejected sign-in saved an account: %+v", accts)
			}
			if _, err := c.AccessToken(context.Background()); !errors.Is(err, ErrNotSignedIn) {
				t.Fatalf("AccessToken err = %v, want ErrNotSignedIn", err)
			}
		})
	}
}

func TestAuthorizationErrorCallbackFails(t *testing.T) {
	c, _, _, _ := setup(t)
	authURL, _, _ := c.BeginSignIn("")
	u, _ := url.Parse(authURL)
	q := u.Query()
	resp, err := http.Get(q.Get("redirect_uri") + "?" + url.Values{
		"state": {q.Get("state")}, "error": {"access_denied"}, "error_description": {"User declined"},
	}.Encode())
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if st := c.Status(); st.State != "failed" || !strings.Contains(st.Error, "access_denied") {
		t.Fatalf("status = %+v", st)
	}
}

func TestRefreshRotatesTokensThenRequiresReauthOnReuse(t *testing.T) {
	c, f, clk, _ := setup(t)
	signIn(t, c, "")

	clk.advance(59*time.Minute + 30*time.Second) // inside the refresh leeway
	tok, err := c.AccessToken(context.Background())
	if err != nil || tok != "access-r1" {
		t.Fatalf("refresh: %q, %v", tok, err)
	}
	d, _ := c.load()
	if d.Accounts[d.Selected].RefreshToken != "refresh-r1" {
		t.Fatal("rotated refresh token not saved")
	}
	if tok2, _ := c.AccessToken(context.Background()); tok2 != "access-r1" || f.refreshCount != 1 {
		t.Fatalf("fresh token should be reused without refresh: %q (refreshes=%d)", tok2, f.refreshCount)
	}

	clk.advance(2 * time.Hour)
	f.refreshErr = "refresh_token_reused"
	if _, err := c.AccessToken(context.Background()); !errors.Is(err, ErrReauthRequired) {
		t.Fatalf("err = %v, want ErrReauthRequired", err)
	}
	sums, _ := c.Accounts()
	if len(sums) != 1 || !sums[0].NeedsReauth {
		t.Fatalf("account should remain listed but need reauth: %+v", sums)
	}
	d, _ = c.load()
	if a := d.Accounts[d.Selected]; a.RefreshToken != "" || a.AccessToken != "" {
		t.Fatal("invalid tokens were not cleared")
	}

	// Reauthorizing the same account restores it.
	f.refreshErr = ""
	if st := signIn(t, c, sums[0].Key); st.State != "complete" {
		t.Fatalf("reauth: %+v", st)
	}
	if _, err := c.AccessToken(context.Background()); err != nil {
		t.Fatalf("after reauth: %v", err)
	}
}

func TestReauthAsDifferentAccountReplacesNothing(t *testing.T) {
	c, f, _, _ := setup(t)
	first := signIn(t, c, "")
	f.sub = "user-other"
	st := signIn(t, c, first.Account.Key)
	if st.State != "failed" || !strings.Contains(st.Error, "different ChatGPT account") {
		t.Fatalf("status = %+v", st)
	}
	sums, _ := c.Accounts()
	if len(sums) != 1 || sums[0].Key != first.Account.Key {
		t.Fatalf("accounts = %+v", sums)
	}
}

func TestSignOutRevokesThenRemovesAccount(t *testing.T) {
	c, f, _, _ := setup(t)
	st := signIn(t, c, "")
	revoked, err := c.SignOut(context.Background(), st.Account.Key)
	if err != nil || !revoked {
		t.Fatalf("SignOut = %v, %v", revoked, err)
	}
	want := []string{"refresh_token:refresh-1", "access_token:access-1"}
	if strings.Join(f.revoked, ",") != strings.Join(want, ",") {
		t.Fatalf("revoked = %v", f.revoked)
	}
	if sums, _ := c.Accounts(); len(sums) != 0 {
		t.Fatalf("account not removed: %+v", sums)
	}
	if _, err := c.AccessToken(context.Background()); !errors.Is(err, ErrNotSignedIn) {
		t.Fatalf("err = %v", err)
	}
}

func TestCancelledAttemptClosesCallbackListener(t *testing.T) {
	c, _, _, _ := setup(t)
	authURL, _, _ := c.BeginSignIn("")
	u, _ := url.Parse(authURL)
	redirect := u.Query().Get("redirect_uri")

	if st := c.CancelSignIn(); st.State != "cancelled" {
		t.Fatalf("cancel status = %+v", st)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := (&http.Client{Timeout: 200 * time.Millisecond}).Get(redirect)
		if err != nil {
			return // listener closed
		}
		resp.Body.Close()
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("callback listener still accepting after cancel")
}
