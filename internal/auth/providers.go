// Package auth provides authentication providers
package auth

import (
	"context"
	"time"
)

// TokenProvider defines the interface for token acquisition
type TokenProvider interface {
	// FetchToken retrieves a new authentication token
	FetchToken(ctx context.Context) (token string, expiry time.Time, err error)

	// Type returns the provider type identifier
	Type() string
}

// Ensure AuthProvider implements TokenProvider
var _ TokenProvider = (*AuthProvider)(nil)

// Type implements TokenProvider for AuthProvider
func (p *AuthProvider) Type() string {
	return "password"
}
