package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type mockProviderForTransport struct {
	token  string
	expiry time.Time
	err    error
	mu     sync.Mutex
}

func (m *mockProviderForTransport) FetchToken(ctx context.Context) (string, time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return "", time.Time{}, m.err
	}
	return m.token, m.expiry, nil
}

func (m *mockProviderForTransport) Type() string {
	return "mock"
}

func createMockManager(token string, expiry time.Time, err error) *Manager {
	provider := &mockProviderForTransport{
		token:  token,
		expiry: expiry,
		err:    err,
	}

	return NewManager(provider, Config{
		Storage: NewMemoryStorage(),
	})
}

func createFailingManager() *Manager {
	provider := &mockProviderForTransport{
		err: http.ErrHandlerTimeout,
	}

	return NewManager(provider, Config{
		Storage: NewMemoryStorage(),
	})
}

func TestNewTransport(t *testing.T) {
	baseTransport := http.DefaultTransport
	manager := createMockManager("test-token", time.Now().Add(time.Hour), nil)

	transport := NewTransport(baseTransport, manager, nil)

	if transport.base != baseTransport {
		t.Error("base transport not set correctly")
	}

	// Note: We can't directly compare managers, but we can test functionality

	// Check default skip endpoints
	expectedSkips := []string{
		"/api/v0/auth",
		"/api/v0/auth_extended",
		"/api/v0/authok",
	}

	for _, expected := range expectedSkips {
		found := false
		for _, skip := range transport.skipAuthEndpoints {
			if skip == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected skip endpoint %s not found", expected)
		}
	}
}

func TestNewTransportWithCustomSkips(t *testing.T) {
	baseTransport := http.DefaultTransport
	manager := createMockManager("test-token", time.Now().Add(time.Hour), nil)

	customSkips := []string{"/custom/skip"}
	transport := NewTransport(baseTransport, manager, customSkips)

	if len(transport.skipAuthEndpoints) != len(customSkips) {
		t.Errorf("expected %d skip endpoints, got %d", len(customSkips), len(transport.skipAuthEndpoints))
	}

	for i, expected := range customSkips {
		if transport.skipAuthEndpoints[i] != expected {
			t.Errorf("expected skip endpoint %s, got %s", expected, transport.skipAuthEndpoints[i])
		}
	}
}

func TestTransportRoundTripSkipAuth(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For auth endpoint, check that no Authorization header is present
		if r.URL.Path == "/api/v0/auth_extended" {
			if auth := r.Header.Get("Authorization"); auth != "" {
				t.Errorf("unexpected Authorization header on auth endpoint: %s", auth)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"token":"auth-token"}`))
		} else {
			// For regular endpoint, auth header should be present
			if auth := r.Header.Get("Authorization"); auth == "" {
				t.Errorf("missing Authorization header on regular endpoint")
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":"success"}`))
		}
	}))
	defer server.Close()

	manager := createMockManager("test-token", time.Now().Add(time.Hour), nil)
	transport := NewTransport(http.DefaultTransport, manager, nil)

	// Test auth endpoint (should skip auth)
	req, _ := http.NewRequest("POST", server.URL+"/api/v0/auth_extended", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("auth request failed: %v", err)
	}
	resp.Body.Close()

	// Test regular endpoint (should add auth)
	req, _ = http.NewRequest("GET", server.URL+"/api/data", nil)
	resp, err = transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("regular request failed: %v", err)
	}
	resp.Body.Close()
}

func TestTransportRoundTripAddAuth(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that Authorization header is present for non-auth endpoints
		auth := r.Header.Get("Authorization")
		if r.URL.Path != "/api/v0/auth_extended" {
			if auth != "Bearer test-token" {
				t.Errorf("expected Authorization 'Bearer test-token', got %s", auth)
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"ok"}`))
	}))
	defer server.Close()

	manager := createMockManager("test-token", time.Now().Add(time.Hour), nil)
	transport := NewTransport(http.DefaultTransport, manager, nil)

	// Test regular endpoint (should add auth)
	req, _ := http.NewRequest("GET", server.URL+"/api/data", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
}

func TestTransportRoundTrip401Retry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First call returns 401
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
		} else {
			// Second call succeeds
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"result":"ok"}`))
		}
	}))
	defer server.Close()

	manager := createMockManager("test-token", time.Now().Add(time.Hour), nil)
	transport := NewTransport(http.DefaultTransport, manager, nil)

	req, _ := http.NewRequest("GET", server.URL+"/api/data", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if callCount != 2 {
		t.Errorf("expected 2 calls (initial + retry), got %d", callCount)
	}
}

func TestTransportRoundTripManagerError(t *testing.T) {
	manager := createFailingManager()
	transport := NewTransport(http.DefaultTransport, manager, nil)

	req, _ := http.NewRequest("GET", "http://example.com/api/data", nil)
	_, err := transport.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error from manager")
	}

	if !strings.Contains(err.Error(), "fetch token") {
		t.Errorf("expected manager error, got %q", err.Error())
	}
}

func TestShouldSkipAuth(t *testing.T) {
	manager := createMockManager("test-token", time.Now().Add(time.Hour), nil)
	transport := NewTransport(http.DefaultTransport, manager, nil)

	tests := []struct {
		path     string
		expected bool
	}{
		{"/api/v0/auth", true},
		{"/api/v0/auth_extended", true},
		{"/api/v0/authok", true},
		{"/api/v0/auth_extended/extra", false}, // suffix match doesn't work as expected
		{"/api/v0/users", false},
		{"/api/v0/data", false},
		{"/health", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com"+tt.path, nil)
			result := transport.shouldSkipAuth(req)

			if result != tt.expected {
				t.Errorf("shouldSkipAuth(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestTransportRoundTripWithBody(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request has the correct body after potential retries
		body := make([]byte, 100)
		n, _ := r.Body.Read(body)
		body = body[:n]

		expectedBody := "test request body"
		if string(body) != expectedBody {
			t.Errorf("expected body %q, got %q", expectedBody, string(body))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"ok"}`))
	}))
	defer server.Close()

	manager := createMockManager("test-token", time.Now().Add(time.Hour), nil)
	transport := NewTransport(http.DefaultTransport, manager, nil)

	req, _ := http.NewRequest("POST", server.URL+"/api/data", strings.NewReader("test request body"))
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
}

func TestTransportRoundTripConcurrent(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"ok"}`))
	}))
	defer server.Close()

	manager := createMockManager("test-token", time.Now().Add(time.Hour), nil)
	transport := NewTransport(http.DefaultTransport, manager, nil)

	const numGoroutines = 10
	const numRequests = 50

	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines*numRequests)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numRequests; j++ {
				req, _ := http.NewRequest("GET", server.URL+"/api/data", nil)
				resp, err := transport.RoundTrip(req)
				if err != nil {
					errChan <- err
					continue
				}
				resp.Body.Close()
			}
		}()
	}

	wg.Wait()
	close(errChan)

	// Check for any errors
	errorCount := 0
	for err := range errChan {
		t.Errorf("concurrent request error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("got %d errors from concurrent requests", errorCount)
	}
}

func BenchmarkTransportRoundTrip(b *testing.B) {
	// Create a fast test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"ok"}`))
	}))
	defer server.Close()

	manager := createMockManager("bench-token", time.Now().Add(time.Hour), nil)
	transport := NewTransport(http.DefaultTransport, manager, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", server.URL+"/api/data", nil)
		resp, err := transport.RoundTrip(req)
		if err != nil {
			b.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()
	}
}

func BenchmarkTransportRoundTripSkipAuth(b *testing.B) {
	// Create a fast test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"token":"auth-token"}`))
	}))
	defer server.Close()

	manager := createMockManager("bench-token", time.Now().Add(time.Hour), nil)
	transport := NewTransport(http.DefaultTransport, manager, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", server.URL+"/api/v0/auth_extended", nil)
		resp, err := transport.RoundTrip(req)
		if err != nil {
			b.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()
	}
}
