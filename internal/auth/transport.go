package auth

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// Transport wraps an HTTP RoundTripper to automatically add authentication
type Transport struct {
	base              http.RoundTripper // underlying HTTP transport
	manager           *Manager          // handles token lifecycle
	skipAuthEndpoints []string          // endpoints that don't need auth (e.g., /auth)
}

// NewTransport creates a new authenticated transport
func NewTransport(base http.RoundTripper, manager *Manager, skip []string) *Transport {
	if base == nil {
		base = http.DefaultTransport
	}

	if skip == nil {
		skip = []string{
			"/api/v0/auth",
			"/api/v0/auth_extended",
			"/api/v0/authok",
		}
	}

	return &Transport{
		base:              base,
		manager:           manager,
		skipAuthEndpoints: skip,
	}
}

// RoundTrip implements http.RoundTripper
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Skip authentication for certain endpoints
	if t.shouldSkipAuth(req) {
		slog.Debug("skip auth", "url", req.URL.Path)
		return t.base.RoundTrip(req)
	}

	slog.Debug("add auth", "method", req.Method, "url", req.URL.String())

	// Clone the request to avoid modifying the original
	reqWithAuth := req.Clone(req.Context())

	// Get authentication token
	token, err := t.manager.GetToken(req.Context())
	if err != nil {
		slog.Error("Failed to get authentication token", "error", err)
		return nil, err
	}

	// Add authorization header
	reqWithAuth.Header.Set("Authorization", "Bearer "+token)

	slog.Debug("Adding authentication to request",
		"method", req.Method,
		"path", req.URL.Path,
		"has_token", token != "",
	)

	// Execute the request
	res, err := t.base.RoundTrip(reqWithAuth)
	if err != nil {
		return nil, err
	}

	// Handle authentication failures
	if res.StatusCode == http.StatusUnauthorized {
		slog.Warn("Received 401 Unauthorized, invalidating token and retrying")

		// Close the current response body
		_ = drainAndClose(res.Body)

		// Invalidate the current token
		t.manager.InvalidateToken()

		// Get a fresh token
		newToken, err := t.manager.ForceRefresh(req.Context())
		if err != nil {
			slog.Error("Failed to refresh token after 401", "error", err)
			return nil, err
		}

		// Retry with the new token
		retryReq := req.Clone(req.Context())
		retryReq.Header.Set("Authorization", "Bearer "+newToken)

		slog.Debug("Retrying request with fresh token")
		return t.base.RoundTrip(retryReq)
	}

	return res, nil
}

// shouldSkipAuth determines if authentication should be skipped for a request
func (t *Transport) shouldSkipAuth(req *http.Request) bool {
	path := req.URL.Path

	for _, skipPath := range t.skipAuthEndpoints {
		// exact match and suffix match for flexibility
		if path == skipPath || strings.HasSuffix(path, skipPath) {
			slog.Debug("shouldSkipAuth MATCH", "path", path, "skipPath", skipPath)
			return true
		}
	}

	slog.Debug("shouldSkipAuth NO MATCH, will add auth", "path", path)
	return false
}

// drainAndClose drains and closes a response body
func drainAndClose(r io.ReadCloser) error {
	if r == nil {
		return nil
	}
	_, _ = io.Copy(io.Discard, r)
	return r.Close()
}

// Middleware creates an API middleware from the auth transport
// This is useful when you want to use the auth logic with the middleware system
// func (t *Transport) Middleware() func(http.RoundTripper) http.RoundTripper {
// 	return func(base http.RoundTripper) http.RoundTripper {
// 		return &Transport{
// 			base:              base,
// 			manager:           t.manager,
// 			skipAuthEndpoints: t.skipAuthEndpoints,
// 		}
// 	}
// }
//
// // Stats returns authentication statistics from the underlying manager
// func (t *Transport) Stats() Stats {
// 	if t.manager == nil {
// 		return Stats{}
// 	}
// 	return t.manager.Stats()
// }
