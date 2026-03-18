package auth

import (
	"context"
	"time"
)

// StaticTokenProvider returns the same bearer token for every request.
//
// This is intended for token-only configurations where the token is managed
// outside of gocmlclient (e.g. Terraform provider config). It never attempts to
// authenticate via username/password.
type StaticTokenProvider struct {
	token string
}

// NewStaticTokenProvider returns a TokenProvider that always yields the given
// token.
func NewStaticTokenProvider(token string) *StaticTokenProvider {
	return &StaticTokenProvider{token: token}
}

// FetchToken implements TokenProvider.
func (p *StaticTokenProvider) FetchToken(ctx context.Context) (string, time.Time, error) {
	// Use a long horizon so the manager does not try to refresh proactively.
	return p.token, time.Now().Add(10 * 365 * 24 * time.Hour), nil
}

// Type implements TokenProvider.
func (p *StaticTokenProvider) Type() string {
	return "static"
}
