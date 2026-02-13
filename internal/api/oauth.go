package api

import (
"context"
"crypto/rand"
"crypto/sha256"
"encoding/base64"
"encoding/json"
"fmt"
"net"
"net/http"
"net/url"
"os"
"path/filepath"
"strings"
"time"
)

// TokenStore persists OAuth tokens to disk.
type TokenStore struct {
Path string
}

// Token represents an OAuth2 token.
type Token struct {
AccessToken  string    `json:"access_token"`
RefreshToken string    `json:"refresh_token,omitempty"`
TokenType    string    `json:"token_type"`
ExpiresAt    time.Time `json:"expires_at"`
Scope        string    `json:"scope,omitempty"`
}

// TokenFile maps service names to tokens.
type TokenFile map[string]*Token

func (t *Token) IsExpired() bool {
return time.Now().After(t.ExpiresAt)
}

// DefaultTokenStore returns the XDG-compliant token store path.
func DefaultTokenStore() (*TokenStore, error) {
dir := os.Getenv("XDG_CONFIG_HOME")
if dir == "" {
home, err := os.UserHomeDir()
if err != nil {
return nil, fmt.Errorf("failed to determine home directory: %w", err)
}
dir = filepath.Join(home, ".config")
}
return &TokenStore{Path: filepath.Join(dir, "salja", "tokens.json")}, nil
}

func (s *TokenStore) Load() (TokenFile, error) {
f, err := os.OpenFile(s.Path, os.O_RDONLY, 0600)
if err != nil {
if os.IsNotExist(err) {
return make(TokenFile), nil
}
return nil, err
}
defer f.Close()
if err := lockFile(f, false); err != nil {
return nil, fmt.Errorf("failed to lock token file: %w", err)
}
defer unlockFile(f)
data, err := os.ReadFile(s.Path)
if err != nil {
return nil, err
}
var tf TokenFile
if err := json.Unmarshal(data, &tf); err != nil {
return nil, fmt.Errorf("invalid token file: %w", err)
}
return tf, nil
}

func (s *TokenStore) Save(tf TokenFile) error {
if err := os.MkdirAll(filepath.Dir(s.Path), 0700); err != nil {
return err
}
f, err := os.OpenFile(s.Path, os.O_WRONLY|os.O_CREATE, 0600)
if err != nil {
return err
}
defer f.Close()
if err := lockFile(f, true); err != nil {
return fmt.Errorf("failed to lock token file: %w", err)
}
defer unlockFile(f)
data, err := json.MarshalIndent(tf, "", "  ")
if err != nil {
return err
}
return os.WriteFile(s.Path, data, 0600)
}

func (s *TokenStore) Get(service string) (*Token, error) {
tf, err := s.Load()
if err != nil {
return nil, err
}
t, ok := tf[service]
if !ok {
return nil, fmt.Errorf("no token for service %q — run: salja auth login %s", service, service)
}
return t, nil
}

func (s *TokenStore) Set(service string, token *Token) error {
tf, err := s.Load()
if err != nil {
return err
}
tf[service] = token
return s.Save(tf)
}

func (s *TokenStore) Delete(service string) error {
tf, err := s.Load()
if err != nil {
return err
}
delete(tf, service)
return s.Save(tf)
}

// PKCEConfig holds OAuth2 PKCE configuration.
type PKCEConfig struct {
ClientID     string
AuthURL      string
TokenURL     string
RedirectURI  string
Scopes       []string
}

// PKCEFlow performs the OAuth2 Authorization Code flow with PKCE.
type PKCEFlow struct {
Config PKCEConfig
}

func NewPKCEFlow(cfg PKCEConfig) *PKCEFlow {
return &PKCEFlow{Config: cfg}
}

func (f *PKCEFlow) generateVerifier() (string, string, error) {
b := make([]byte, 32)
if _, err := rand.Read(b); err != nil {
return "", "", err
}
verifier := base64.RawURLEncoding.EncodeToString(b)
h := sha256.Sum256([]byte(verifier))
challenge := base64.RawURLEncoding.EncodeToString(h[:])
return verifier, challenge, nil
}

// Authorize opens a browser-based auth flow and returns the token.
func (f *PKCEFlow) Authorize(ctx context.Context) (*Token, error) {
verifier, challenge, err := f.generateVerifier()
if err != nil {
return nil, fmt.Errorf("PKCE verifier generation failed: %w", err)
}

// Start local callback server
listener, err := net.Listen("tcp", "127.0.0.1:0")
if err != nil {
return nil, fmt.Errorf("failed to start callback server: %w", err)
}
defer listener.Close()

port := listener.Addr().(*net.TCPAddr).Port
redirectURI := f.Config.RedirectURI
if redirectURI == "" {
redirectURI = fmt.Sprintf("http://127.0.0.1:%d/callback", port)
}

state, err := generateState()
if err != nil {
return nil, fmt.Errorf("failed to generate state: %w", err)
}
authURL := fmt.Sprintf("%s?client_id=%s&response_type=code&redirect_uri=%s&scope=%s&state=%s&code_challenge=%s&code_challenge_method=S256",
f.Config.AuthURL,
url.QueryEscape(f.Config.ClientID),
url.QueryEscape(redirectURI),
url.QueryEscape(strings.Join(f.Config.Scopes, " ")),
url.QueryEscape(state),
url.QueryEscape(challenge),
)

fmt.Fprintf(os.Stderr, "Open this URL in your browser to authorize:\n\n  %s\n\nWaiting for callback...\n", authURL)

codeCh := make(chan string, 1)
errCh := make(chan error, 1)

mux := http.NewServeMux()
mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
if r.URL.Query().Get("state") != state {
errCh <- fmt.Errorf("state mismatch")
http.Error(w, "State mismatch", http.StatusBadRequest)
return
}
if errMsg := r.URL.Query().Get("error"); errMsg != "" {
errCh <- fmt.Errorf("auth error: %s — %s", errMsg, r.URL.Query().Get("error_description"))
fmt.Fprintf(w, "Authorization failed: %s", errMsg)
return
}
code := r.URL.Query().Get("code")
codeCh <- code
fmt.Fprint(w, "Authorization successful! You can close this tab.")
})

server := &http.Server{Handler: mux}
go server.Serve(listener)
defer server.Shutdown(ctx)

var code string
select {
case code = <-codeCh:
case err := <-errCh:
return nil, err
case <-ctx.Done():
return nil, ctx.Err()
}

return f.exchangeCode(ctx, code, verifier, redirectURI)
}

func (f *PKCEFlow) exchangeCode(ctx context.Context, code, verifier, redirectURI string) (*Token, error) {
data := url.Values{
"grant_type":    {"authorization_code"},
"code":          {code},
"redirect_uri":  {redirectURI},
"client_id":     {f.Config.ClientID},
"code_verifier": {verifier},
}

req, err := http.NewRequestWithContext(ctx, "POST", f.Config.TokenURL, strings.NewReader(data.Encode()))
if err != nil {
return nil, err
}
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

resp, err := http.DefaultClient.Do(req)
if err != nil {
return nil, fmt.Errorf("token exchange failed: %w", err)
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
return nil, fmt.Errorf("token exchange returned HTTP %d", resp.StatusCode)
}

var tokenResp struct {
AccessToken  string `json:"access_token"`
RefreshToken string `json:"refresh_token"`
TokenType    string `json:"token_type"`
ExpiresIn    int    `json:"expires_in"`
Scope        string `json:"scope"`
}
if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
return nil, fmt.Errorf("failed to decode token response: %w", err)
}

return &Token{
AccessToken:  tokenResp.AccessToken,
RefreshToken: tokenResp.RefreshToken,
TokenType:    tokenResp.TokenType,
ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
Scope:        tokenResp.Scope,
}, nil
}

// RefreshAccessToken uses a refresh token to get a new access token.
func (f *PKCEFlow) RefreshAccessToken(ctx context.Context, refreshToken string) (*Token, error) {
data := url.Values{
"grant_type":    {"refresh_token"},
"refresh_token": {refreshToken},
"client_id":     {f.Config.ClientID},
}

req, err := http.NewRequestWithContext(ctx, "POST", f.Config.TokenURL, strings.NewReader(data.Encode()))
if err != nil {
return nil, err
}
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

resp, err := http.DefaultClient.Do(req)
if err != nil {
return nil, err
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
return nil, fmt.Errorf("token refresh returned HTTP %d", resp.StatusCode)
}

var tokenResp struct {
AccessToken  string `json:"access_token"`
RefreshToken string `json:"refresh_token"`
TokenType    string `json:"token_type"`
ExpiresIn    int    `json:"expires_in"`
}
if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
return nil, err
}

rt := tokenResp.RefreshToken
if rt == "" {
rt = refreshToken
}

return &Token{
AccessToken:  tokenResp.AccessToken,
RefreshToken: rt,
TokenType:    tokenResp.TokenType,
ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
}, nil
}

// TokenRefresher can refresh expired tokens.
type TokenRefresher struct {
	Flow  *PKCEFlow
	Store *TokenStore
	Service string
}

// EnsureValid refreshes the token if expired and a refresh token is available.
func (r *TokenRefresher) EnsureValid(ctx context.Context, token *Token) (*Token, error) {
	if !token.IsExpired() {
		return token, nil
	}
	if token.RefreshToken == "" {
		return nil, fmt.Errorf("token for %s is expired and no refresh token available; run: salja auth login %s", r.Service, r.Service)
	}
	newToken, err := r.Flow.RefreshAccessToken(ctx, token.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh %s token: %w", r.Service, err)
	}
	if r.Store != nil {
		_ = r.Store.Set(r.Service, newToken)
	}
	return newToken, nil
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
