package auth

import (
	"io"
	"log/slog"
	"net/http"
)

// Transport wraps an HTTP RoundTripper to automatically add authentication
type Transport struct {
	Base    http.RoundTripper // underlying HTTP transport
	Manager *Manager          // handles token lifecycle
}

// NewTransport creates a new authenticated transport
func NewTransport(base http.RoundTripper, manager *Manager) *Transport {
	if base == nil {
		base = http.DefaultTransport
	}

	return &Transport{
		Base:    base,
		Manager: manager,
	}
}

// RoundTrip implements http.RoundTripper
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	reqWithAuth := req.Clone(req.Context())

	// Get authentication token
	token, err := t.Manager.GetToken(req.Context())
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
	res, err := t.Base.RoundTrip(reqWithAuth)
	if err != nil {
		return nil, err
	}

	// Handle authentication failures
	if res.StatusCode == http.StatusUnauthorized {
		slog.Warn("Received 401 Unauthorized, invalidating token and retrying")

		// Close the current response body
		_ = drainAndClose(res.Body)

		// Invalidate the current token
		t.Manager.InvalidateToken()

		// Get a fresh token
		newToken, err := t.Manager.ForceRefresh(req.Context())
		if err != nil {
			slog.Error("Failed to refresh token after 401", "error", err)
			return nil, err
		}

		// Retry with the new token
		retryReq := req.Clone(req.Context())
		retryReq.Header.Set("Authorization", "Bearer "+newToken)

		slog.Debug("Retrying request with fresh token")
		return t.Base.RoundTrip(retryReq)
	}

	return res, nil
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
func (t *Transport) Middleware() func(http.RoundTripper) http.RoundTripper {
	return func(base http.RoundTripper) http.RoundTripper {
		return &Transport{
			Base:    base,
			Manager: t.Manager,
		}
	}
}

// Stats returns authentication statistics from the underlying manager
func (t *Transport) Stats() Stats {
	if t.Manager == nil {
		return Stats{}
	}
	return t.Manager.Stats()
}
