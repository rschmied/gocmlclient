package auth

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// TokenProvider defines the interface for token acquisition
type TokenProvider interface {
	// FetchToken retrieves a new authentication token
	FetchToken(ctx context.Context) (token string, expiry time.Time, err error)
}

// Manager handles authentication token lifecycle
type Manager struct {
	provider TokenProvider

	// token state (protected by mutex)
	mu     sync.RWMutex
	token  string
	expiry time.Time

	// provider configuration
	refreshBuffer time.Duration // how early to refresh before expiry
}

// Config configures the auth manager
type Config struct {
	RefreshBuffer time.Duration
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		RefreshBuffer: 30 * time.Second,
	}
}

// NewManager creates a new authentication manager
func NewManager(provider TokenProvider, config Config) *Manager {
	if config.RefreshBuffer == 0 {
		config.RefreshBuffer = DefaultConfig().RefreshBuffer
	}

	return &Manager{
		provider:      provider,
		refreshBuffer: config.RefreshBuffer,
	}
}

// GetToken returns a valid authentication token. if the current token is
// expired or about to expire, it will automatically refresh
func (m *Manager) GetToken(ctx context.Context) (string, error) {
	m.mu.RLock()
	if m.token != "" && m.isTokenValid() {
		token := m.token
		m.mu.RUnlock()
		return token, nil
	}
	m.mu.RUnlock()
	return m.refreshToken(ctx)
}

// ForceRefresh forces a token refresh regardless of current token state
func (m *Manager) ForceRefresh(ctx context.Context) (string, error) {
	return m.refreshToken(ctx)
}

// InvalidateToken marks the current token as invalid (e.g. 401)
func (m *Manager) InvalidateToken() {
	m.mu.Lock()
	defer m.mu.Unlock()

	slog.Debug("Invalidating current token")
	m.token = ""
	m.expiry = time.Time{}
}

// HasValidToken returns true if the manager has a valid token
func (m *Manager) HasValidToken() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.token != "" && m.isTokenValid()
}

// TokenExpiry returns the current token's expiry time
func (m *Manager) TokenExpiry() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.expiry
}

// refreshToken acquires a new token from the provider
func (m *Manager) refreshToken(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check pattern: another goroutine might have refreshed while we waited
	if m.token != "" && m.isTokenValid() {
		return m.token, nil
	}

	slog.Debug("Refreshing authentication token")

	token, expiry, err := m.provider.FetchToken(ctx)
	if err != nil {
		slog.Error("Failed to fetch token", "error", err)
		return "", fmt.Errorf("fetch token: %w", err)
	}

	if token == "" {
		return "", fmt.Errorf("provider returned empty token")
	}

	// Validate expiry time
	if expiry.Before(time.Now()) {
		slog.Warn("Provider returned already-expired token", "expiry", expiry)
	}

	m.token = token
	m.expiry = expiry

	slog.Debug("Token refreshed successfully",
		"expiry", expiry,
		"valid_for", time.Until(expiry),
	)

	return token, nil
}

// isTokenValid checks if the current token is still valid
// Must be called with at least a read lock held
func (m *Manager) isTokenValid() bool {
	if m.token == "" {
		return false
	}

	// Consider token invalid if it expires within the refresh buffer
	refreshTime := m.expiry.Add(-m.refreshBuffer)
	return time.Now().Before(refreshTime)
}

// Stats returns authentication statistics
func (m *Manager) Stats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return Stats{
		HasToken:    m.token != "",
		TokenExpiry: m.expiry,
		IsValid:     m.isTokenValid(),
		TimeUntilRefresh: func() time.Duration {
			if m.token == "" {
				return 0
			}
			refreshTime := m.expiry.Add(-m.refreshBuffer)
			if time.Now().After(refreshTime) {
				return 0
			}
			return time.Until(refreshTime)
		}(),
	}
}

// Stats contains authentication manager statistics
type Stats struct {
	HasToken         bool
	TokenExpiry      time.Time
	IsValid          bool
	TimeUntilRefresh time.Duration
}
