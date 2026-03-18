// Package auth provides auth service
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rschmied/gocmlclient/internal/httputil"
	"github.com/rschmied/gocmlclient/internal/logging"
)

// AuthProvider implements TokenProvider using username/password authentication
type AuthProvider struct {
	baseURL     string
	username    string
	password    string
	presetToken string // Optional: use this token once before falling back to username/password
	clientID    string
	clientUUID  string
	version     string

	client *http.Client
}

// AuthConfig configures the username/password provider
type AuthConfig struct {
	BaseURL     string
	Username    string
	Password    string
	PresetToken string
	Client      *http.Client
	ClientID    string
	ClientUUID  string
	Version     string
	Timeout     time.Duration
}

// NewAuthProvider creates a new username/password token provider
func NewAuthProvider(config AuthConfig) *AuthProvider {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	// this panics if there's no client provided
	_ = config.Client

	// Set default timeout on client if not set
	if config.Client.Timeout == 0 {
		config.Client.Timeout = config.Timeout
	}

	return &AuthProvider{
		baseURL:     config.BaseURL,
		username:    config.Username,
		password:    config.Password,
		presetToken: config.PresetToken,
		clientID:    config.ClientID,
		clientUUID:  config.ClientUUID,
		version:     config.Version,
		client:      config.Client,
	}
}

// FetchToken implements TokenProvider
func (p *AuthProvider) FetchToken(ctx context.Context) (string, time.Time, error) {
	// if we have a preset token, use it once, while it's valid
	if p.presetToken != "" {
		logging.Debug("Using preset token")
		token := p.presetToken
		p.presetToken = "" // Clear it so we don't reuse it

		// Preset tokens typically have a default expiry
		expiry := time.Now().Add(8 * time.Hour)
		return token, expiry, nil
	}

	if strings.TrimSpace(p.username) == "" || strings.TrimSpace(p.password) == "" {
		return "", time.Time{}, fmt.Errorf("authentication requires username/password but none configured")
	}

	return p.authenticateWithPassword(ctx)
}

// authenticateWithPassword performs username/password authentication
func (p *AuthProvider) authenticateWithPassword(ctx context.Context) (string, time.Time, error) {
	logging.Debug("Authenticating with username/password", "username", p.username)

	// Prepare request body
	reqBody := authRequest{
		Username: p.username,
		Password: p.password,
	}

	// Use shared request building logic
	req, err := httputil.BuildRequest(ctx, p.baseURL, "POST", "/api/v0/auth_extended", nil, reqBody)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("build auth request: %w", err)
	}
	httputil.ApplyClientIdentityHeaders(req.Header, p.clientID, p.clientUUID, p.version)

	// Execute request
	logging.Debug("Sending authentication request", "url", req.URL.String())
	res, err := p.client.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("auth request failed: %w", err)
	}
	defer res.Body.Close() //nolint:errcheck

	// Handle authentication failure
	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		logging.Error("Authentication failed",
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

	logging.Debug("Authentication successful",
		"username", authRes.Username,
		"admin", authRes.Admin,
		"expiry", expiry,
	)

	return authRes.Token, expiry, nil
}

// SetPresetToken sets a token to use on the next FetchToken call
// Useful for scenarios where you have a valid token but need to initialize the provider
func (p *AuthProvider) SetPresetToken(token string) {
	logging.Debug("Setting preset token")
	p.presetToken = token
}

// UpdateCredentials updates the username and password
func (p *AuthProvider) UpdateCredentials(username, password string) {
	logging.Debug("Updating credentials", "username", username)
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
