package cmlclient

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type TokenProvider interface {
	FetchToken(ctx context.Context) (string, time.Time, error)
	// RefreshToken(ctx context.Context) (string, time.Time, error)
}

type AuthManager struct {
	provider TokenProvider
	mu       sync.RWMutex
	token    string
	// refreshToken string
	expiry time.Time
}

func NewAuthManager(provider TokenProvider) *AuthManager {
	return &AuthManager{provider: provider}
}

func (a *AuthManager) GetToken(ctx context.Context) (string, error) {
	a.mu.RLock()
	if a.token != "" && time.Now().Before(a.expiry) {
		a.mu.RUnlock()
		return a.token, nil
	}
	a.mu.RUnlock()
	return a.ForceRefresh(ctx)
}

func (a *AuthManager) ForceRefresh(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	tok, exp, err := a.provider.FetchToken(ctx)
	if err != nil {
		slog.Error("auth failed", "err", err)
		return "", err
	}
	a.token, a.expiry = tok, exp
	return tok, nil
}

type AuthTransport struct {
	Base http.RoundTripper
	Auth *AuthManager
}

func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	// if strings.HasSuffix(req.URL.Path, authAPI) {
	// 	return base.RoundTrip(req)
	// }

	ctx := req.Context()
	r2 := req.Clone(ctx)

	tok, err := t.Auth.GetToken(ctx)
	if err != nil {
		return nil, err
	}
	r2.Header.Set("Authorization", "Bearer "+tok)

	res, err := base.RoundTrip(r2)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusUnauthorized {
		_ = drainAndClose(res.Body)
		newTok, err := t.Auth.ForceRefresh(ctx)
		if err != nil {
			return nil, err
		}
		r2 = req.Clone(ctx)
		r2.Header.Set("Authorization", "Bearer "+newTok)
		return base.RoundTrip(r2)
	}
	return res, nil
}

func drainAndClose(r io.ReadCloser) error {
	_, _ = io.Copy(io.Discard, r)
	return r.Close()
}
