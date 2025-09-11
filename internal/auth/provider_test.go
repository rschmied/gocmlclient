package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rschmied/gocmlclient/internal/httputil"
)

func TestNewAuthProvider(t *testing.T) {
	config := AuthConfig{
		BaseURL:  "https://api.example.com",
		Username: "testuser",
		Password: "testpass",
		Client:   &http.Client{Timeout: 30 * time.Second},
	}

	provider := NewAuthProvider(config)

	if provider.baseURL != "https://api.example.com" {
		t.Errorf("expected baseURL 'https://api.example.com', got %s", provider.baseURL)
	}

	if provider.username != "testuser" {
		t.Errorf("expected username 'testuser', got %s", provider.username)
	}

	if provider.password != "testpass" {
		t.Errorf("expected password 'testpass', got %s", provider.password)
	}

	if provider.client.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", provider.client.Timeout)
	}
}

func TestNewAuthProviderDefaultTimeout(t *testing.T) {
	config := AuthConfig{
		BaseURL:  "https://api.example.com",
		Username: "testuser",
		Password: "testpass",
		Client:   &http.Client{}, // No timeout set
	}

	provider := NewAuthProvider(config)

	if provider.client.Timeout != 10*time.Second {
		t.Errorf("expected default timeout 10s, got %v", provider.client.Timeout)
	}
}

func TestFetchTokenWithPreset(t *testing.T) {
	config := AuthConfig{
		BaseURL:     "https://api.example.com",
		Username:    "testuser",
		Password:    "testpass",
		PresetToken: "preset-token-123",
		Client:      &http.Client{Timeout: 10 * time.Second},
	}

	provider := NewAuthProvider(config)

	// First call should return preset token
	token, expiry, err := provider.FetchToken(context.Background())
	if err != nil {
		t.Fatalf("FetchToken failed: %v", err)
	}

	if token != "preset-token-123" {
		t.Errorf("expected preset token 'preset-token-123', got %s", token)
	}

	// Expiry should be set to default 8 hours from now
	expectedExpiry := time.Now().Add(8 * time.Hour)
	if expiry.Before(expectedExpiry.Add(-time.Minute)) || expiry.After(expectedExpiry.Add(time.Minute)) {
		t.Errorf("expected expiry around %v, got %v", expectedExpiry, expiry)
	}

	// Second call should authenticate normally (but we'll mock the server)
	// For this test, we'll just verify the preset token is cleared
	if provider.presetToken != "" {
		t.Error("expected preset token to be cleared after first use")
	}
}

func TestFetchTokenAuthentication(t *testing.T) {
	// Create a test server that simulates authentication
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		if r.URL.Path != "/api/v0/auth_extended" {
			t.Errorf("expected path /api/v0/auth_extended, got %s", r.URL.Path)
		}

		// Verify request body
		var req authRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Username != "testuser" {
			t.Errorf("expected username 'testuser', got %s", req.Username)
		}

		if req.Password != "testpass" {
			t.Errorf("expected password 'testpass', got %s", req.Password)
		}

		// Return successful authentication response
		response := authResponse{
			ID:       "user-123",
			Username: "testuser",
			Token:    "auth-token-456",
			Admin:    true,
		}

		w.Header().Set("Content-Type", httputil.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := AuthConfig{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
		Client:   &http.Client{Timeout: 10 * time.Second},
	}

	provider := NewAuthProvider(config)

	token, expiry, err := provider.FetchToken(context.Background())
	if err != nil {
		t.Fatalf("FetchToken failed: %v", err)
	}

	if token != "auth-token-456" {
		t.Errorf("expected token 'auth-token-456', got %s", token)
	}

	// Expiry should be set to default 8 hours from now
	expectedExpiry := time.Now().Add(8 * time.Hour)
	if expiry.Before(expectedExpiry.Add(-time.Minute)) || expiry.After(expectedExpiry.Add(time.Minute)) {
		t.Errorf("expected expiry around %v, got %v", expectedExpiry, expiry)
	}
}

func TestFetchTokenAuthenticationFailure(t *testing.T) {
	// Create a test server that returns authentication failure
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid credentials"))
	}))
	defer server.Close()

	config := AuthConfig{
		BaseURL:  server.URL,
		Username: "wronguser",
		Password: "wrongpass",
		Client:   &http.Client{Timeout: 10 * time.Second},
	}

	provider := NewAuthProvider(config)

	_, _, err := provider.FetchToken(context.Background())
	if err == nil {
		t.Fatal("expected authentication error")
	}

	expectedError := "authentication failed: 401 Unauthorized"
	if err.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, err.Error())
	}
}

func TestFetchTokenServerError(t *testing.T) {
	// Create a test server that returns server error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	config := AuthConfig{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
		Client:   &http.Client{Timeout: 10 * time.Second},
	}

	provider := NewAuthProvider(config)

	_, _, err := provider.FetchToken(context.Background())
	if err == nil {
		t.Fatal("expected server error")
	}

	expectedError := "authentication failed: 500 Internal Server Error"
	if err.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, err.Error())
	}
}

func TestFetchTokenInvalidResponse(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", httputil.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	config := AuthConfig{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
		Client:   &http.Client{Timeout: 10 * time.Second},
	}

	provider := NewAuthProvider(config)

	_, _, err := provider.FetchToken(context.Background())
	if err == nil {
		t.Fatal("expected JSON parsing error")
	}

	if !strings.Contains(err.Error(), "decode auth response") {
		t.Errorf("expected decode error, got %q", err.Error())
	}
}

func TestFetchTokenEmptyToken(t *testing.T) {
	// Create a test server that returns empty token
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := authResponse{
			ID:       "user-123",
			Username: "testuser",
			Token:    "", // Empty token
			Admin:    false,
		}

		w.Header().Set("Content-Type", httputil.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := AuthConfig{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
		Client:   &http.Client{Timeout: 10 * time.Second},
	}

	provider := NewAuthProvider(config)

	_, _, err := provider.FetchToken(context.Background())
	if err == nil {
		t.Fatal("expected empty token error")
	}

	expectedError := "empty token in auth response"
	if err.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, err.Error())
	}
}

func TestFetchTokenNetworkError(t *testing.T) {
	config := AuthConfig{
		BaseURL:  "https://invalid-host-that-does-not-exist.com",
		Username: "testuser",
		Password: "testpass",
		Client:   &http.Client{Timeout: 1 * time.Millisecond}, // Very short timeout
	}

	provider := NewAuthProvider(config)

	_, _, err := provider.FetchToken(context.Background())
	if err == nil {
		t.Fatal("expected network error")
	}

	// Should contain some network-related error
	if !strings.Contains(err.Error(), "auth request failed") {
		t.Errorf("expected auth request error, got %q", err.Error())
	}
}

func TestSetPresetToken(t *testing.T) {
	config := AuthConfig{
		BaseURL:  "https://api.example.com",
		Username: "testuser",
		Password: "testpass",
		Client:   &http.Client{Timeout: 10 * time.Second},
	}

	provider := NewAuthProvider(config)

	// Initially no preset token
	if provider.presetToken != "" {
		t.Error("expected no initial preset token")
	}

	// Set preset token
	provider.SetPresetToken("new-preset-token")

	if provider.presetToken != "new-preset-token" {
		t.Errorf("expected preset token 'new-preset-token', got %s", provider.presetToken)
	}
}

func TestUpdateCredentials(t *testing.T) {
	config := AuthConfig{
		BaseURL:  "https://api.example.com",
		Username: "olduser",
		Password: "oldpass",
		Client:   &http.Client{Timeout: 10 * time.Second},
	}

	provider := NewAuthProvider(config)

	if provider.username != "olduser" {
		t.Errorf("expected initial username 'olduser', got %s", provider.username)
	}

	if provider.password != "oldpass" {
		t.Errorf("expected initial password 'oldpass', got %s", provider.password)
	}

	// Update credentials
	provider.UpdateCredentials("newuser", "newpass")

	if provider.username != "newuser" {
		t.Errorf("expected updated username 'newuser', got %s", provider.username)
	}

	if provider.password != "newpass" {
		t.Errorf("expected updated password 'newpass', got %s", provider.password)
	}
}

func TestAuthProviderType(t *testing.T) {
	config := AuthConfig{
		BaseURL:  "https://api.example.com",
		Username: "testuser",
		Password: "testpass",
		Client:   &http.Client{Timeout: 10 * time.Second},
	}

	provider := NewAuthProvider(config)

	if provider.Type() != "password" {
		t.Errorf("expected provider type 'password', got %s", provider.Type())
	}
}

func BenchmarkFetchToken(b *testing.B) {
	// Create a fast test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := authResponse{
			ID:       "user-123",
			Username: "testuser",
			Token:    "bench-token",
			Admin:    false,
		}

		w.Header().Set("Content-Type", httputil.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := AuthConfig{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
		Client:   &http.Client{Timeout: 10 * time.Second},
	}

	provider := NewAuthProvider(config)

	for b.Loop() {
		_, _, err := provider.FetchToken(context.Background())
		if err != nil {
			b.Fatalf("FetchToken failed: %v", err)
		}
	}
}

func TestFetchTokenNetworkTimeout(t *testing.T) {
	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Delay longer than client timeout
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"token":"delayed-token"}`))
	}))
	defer server.Close()

	config := AuthConfig{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
		Client:   &http.Client{Timeout: 100 * time.Millisecond}, // Short timeout
	}

	provider := NewAuthProvider(config)

	_, _, err := provider.FetchToken(context.Background())
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !strings.Contains(err.Error(), "auth request failed") {
		t.Errorf("expected auth request error, got %q", err.Error())
	}
}

func BenchmarkFetchTokenWithPreset(b *testing.B) {
	config := AuthConfig{
		BaseURL:     "https://api.example.com",
		Username:    "testuser",
		Password:    "testpass",
		PresetToken: "preset-bench-token",
		Client:      &http.Client{Timeout: 10 * time.Second},
	}

	provider := NewAuthProvider(config)

	for b.Loop() {
		// Reset preset token for each iteration
		provider.SetPresetToken("preset-bench-token")

		_, _, err := provider.FetchToken(context.Background())
		if err != nil {
			b.Fatalf("FetchToken failed: %v", err)
		}
	}
}
