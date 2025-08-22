package auth

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// UsernamePasswordProvider implements TokenProvider using username/password authentication
type UsernamePasswordProvider struct {
	baseURL     string
	username    string
	password    string
	presetToken string // Optional: use this token once before falling back to username/password

	client *http.Client
}

// UsernamePasswordConfig configures the username/password provider
type UsernamePasswordConfig struct {
	BaseURL     string
	Username    string
	Password    string
	PresetToken string // Optional: token to use before authentication

	// HTTP client configuration
	Timeout            time.Duration
	InsecureSkipVerify bool
}

// NewUsernamePasswordProvider creates a new username/password token provider
func NewUsernamePasswordProvider(config UsernamePasswordConfig) *UsernamePasswordProvider {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	// Create HTTP client for authentication (no auth middleware)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
		},
		Proxy: http.ProxyFromEnvironment,
	}

	client := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}

	return &UsernamePasswordProvider{
		baseURL:     config.BaseURL,
		username:    config.Username,
		password:    config.Password,
		presetToken: config.PresetToken,
		client:      client,
	}
}

// FetchToken implements TokenProvider
func (p *UsernamePasswordProvider) FetchToken(ctx context.Context) (string, time.Time, error) {
	// If we have a preset token, use it once
	if p.presetToken != "" {
		slog.Debug("Using preset token")
		token := p.presetToken
		p.presetToken = "" // Clear it so we don't reuse it

		// Preset tokens typically have a default expiry
		expiry := time.Now().Add(8 * time.Hour)
		return token, expiry, nil
	}

	return p.authenticateWithPassword(ctx)
}

// authenticateWithPassword performs username/password authentication
func (p *UsernamePasswordProvider) authenticateWithPassword(ctx context.Context) (string, time.Time, error) {
	slog.Debug("Authenticating with username/password", "username", p.username)

	// Prepare request body
	reqBody := authRequest{
		Username: p.username,
		Password: p.password,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("marshal auth request: %w", err)
	}

	// Build auth URL
	authURL, err := url.JoinPath(p.baseURL, "/api/v0/auth_extended")
	if err != nil {
		return "", time.Time{}, fmt.Errorf("build auth URL: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	slog.Debug("Sending authentication request", "url", authURL)
	res, err := p.client.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("auth request failed: %w", err)
	}
	defer res.Body.Close()

	// Handle authentication failure
	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		slog.Error("Authentication failed",
			"status", res.StatusCode,
			"body", string(body),
			"username", p.username,
		)
		return "", time.Time{}, fmt.Errorf("authentication failed: %s", res.Status)
	}

	// Parse response
	var authRes authResponse
	if err := json.NewDecoder(res.Body).Decode(&authRes); err != nil {
		return "", time.Time{}, fmt.Errorf("decode auth response: %w", err)
	}

	if authRes.Token == "" {
		return "", time.Time{}, fmt.Errorf("empty token in auth response")
	}

	// Default expiry if not provided by server
	expiry := time.Now().Add(8 * time.Hour)

	slog.Debug("Authentication successful",
		"username", authRes.Username,
		"admin", authRes.Admin,
		"expiry", expiry,
	)

	return authRes.Token, expiry, nil
}

// SetPresetToken sets a token to use on the next FetchToken call
// Useful for scenarios where you have a valid token but need to initialize the provider
func (p *UsernamePasswordProvider) SetPresetToken(token string) {
	slog.Debug("Setting preset token")
	p.presetToken = token
}

// UpdateCredentials updates the username and password
func (p *UsernamePasswordProvider) UpdateCredentials(username, password string) {
	slog.Debug("Updating credentials", "username", username)
	p.username = username
	p.password = password
}

// authRequest represents the authentication request body
type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// authResponse represents the authentication response
type authResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Admin    bool   `json:"admin"`
}

// StaticTokenProvider implements TokenProvider for static/long-lived tokens
type StaticTokenProvider struct {
	token  string
	expiry time.Time
}

// NewStaticTokenProvider creates a provider for static tokens
func NewStaticTokenProvider(token string, expiry time.Time) *StaticTokenProvider {
	return &StaticTokenProvider{
		token:  token,
		expiry: expiry,
	}
}

// FetchToken implements TokenProvider
func (p *StaticTokenProvider) FetchToken(ctx context.Context) (string, time.Time, error) {
	slog.Debug("Using static token")
	return p.token, p.expiry, nil
}

// UpdateToken updates the static token and expiry
func (p *StaticTokenProvider) UpdateToken(token string, expiry time.Time) {
	slog.Debug("Updating static token", "expiry", expiry)
	p.token = token
	p.expiry = expiry
}
