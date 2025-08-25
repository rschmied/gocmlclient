// Package auth provides auth service
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// AuthProvider implements TokenProvider using username/password authentication
type AuthProvider struct {
	baseURL     string
	username    string
	password    string
	presetToken string // Optional: use this token once before falling back to username/password

	client *http.Client
}

// AuthConfig configures the username/password provider
type AuthConfig struct {
	BaseURL     string
	Username    string
	Password    string
	PresetToken string // Optional: token to use before authentication

	// HTTP client configuration
	Timeout time.Duration
	// InsecureSkipVerify bool
	HTTPclient *http.Client
}

// NewAuthProvider creates a new username/password token provider
func NewAuthProvider(config AuthConfig) *AuthProvider {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	// this panics if there's no client provided
	_ = config.HTTPclient

	return &AuthProvider{
		baseURL:     config.BaseURL,
		username:    config.Username,
		password:    config.Password,
		presetToken: config.PresetToken,
		client:      config.HTTPclient,
	}
}

// FetchToken implements TokenProvider
func (p *AuthProvider) FetchToken(ctx context.Context) (string, time.Time, error) {
	// if we have a preset token, use it once, while it's valid
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
func (p *AuthProvider) authenticateWithPassword(ctx context.Context) (string, time.Time, error) {
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
func (p *AuthProvider) SetPresetToken(token string) {
	slog.Debug("Setting preset token")
	p.presetToken = token
}

// UpdateCredentials updates the username and password
func (p *AuthProvider) UpdateCredentials(username, password string) {
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
