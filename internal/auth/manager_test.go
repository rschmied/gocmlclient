package auth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type mockProvider struct {
	token     string
	expiry    time.Time
	err       error
	callCount int
	mu        sync.Mutex
}

func (m *mockProvider) FetchToken(ctx context.Context) (string, time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	return m.token, m.expiry, m.err
}

func (m *mockProvider) Type() string {
	return "mock"
}

func (m *mockProvider) getCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func TestNewManager(t *testing.T) {
	provider := &mockProvider{
		token:  "test-token",
		expiry: time.Now().Add(time.Hour),
	}

	// Test with default config
	manager := NewManager(provider, DefaultConfig())

	if manager.provider != provider {
		t.Error("provider not set correctly")
	}

	if manager.refreshBuffer != 30*time.Second {
		t.Errorf("expected refresh buffer 30s, got %v", manager.refreshBuffer)
	}

	if manager.storage.Type() != "memory" {
		t.Errorf("expected memory storage, got %s", manager.storage.Type())
	}
}

func TestNewManagerWithStorage(t *testing.T) {
	provider := &mockProvider{
		token:  "test-token",
		expiry: time.Now().Add(time.Hour),
	}

	storage := NewMemoryStorage()
	config := Config{
		RefreshBuffer: 45 * time.Second,
		Storage:       storage,
	}

	manager := NewManager(provider, config)

	if manager.refreshBuffer != 45*time.Second {
		t.Errorf("expected refresh buffer 45s, got %v", manager.refreshBuffer)
	}

	if manager.storage != storage {
		t.Error("storage not set correctly")
	}
}

func TestGetToken(t *testing.T) {
	now := time.Now()
	provider := &mockProvider{
		token:  "fresh-token",
		expiry: now.Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	// First call should fetch from provider
	token, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "fresh-token" {
		t.Errorf("expected token 'fresh-token', got %s", token)
	}

	if provider.getCallCount() != 1 {
		t.Errorf("expected provider called 1 time, got %d", provider.getCallCount())
	}

	// Second call should return cached token
	token2, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token2 != "fresh-token" {
		t.Errorf("expected cached token 'fresh-token', got %s", token2)
	}

	if provider.getCallCount() != 1 {
		t.Errorf("expected provider called only 1 time (cached), got %d", provider.getCallCount())
	}
}

func TestGetTokenExpired(t *testing.T) {
	now := time.Now()
	provider := &mockProvider{
		token:  "expired-token",
		expiry: now.Add(-time.Hour), // Already expired
	}

	manager := NewManager(provider, Config{
		RefreshBuffer: 1 * time.Nanosecond, // Very short buffer
		Storage:       NewMemoryStorage(),
	})

	// Should refresh immediately due to expiry
	token, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "expired-token" {
		t.Errorf("expected token 'expired-token', got %s", token)
	}

	if provider.getCallCount() != 1 {
		t.Errorf("expected provider called 1 time, got %d", provider.getCallCount())
	}
}

func TestGetTokenNearExpiry(t *testing.T) {
	now := time.Now()
	provider := &mockProvider{
		token:  "near-expiry-token",
		expiry: now.Add(10 * time.Second), // Expires soon
	}

	manager := NewManager(provider, Config{
		RefreshBuffer: 30 * time.Second, // Buffer is larger than remaining time
		Storage:       NewMemoryStorage(),
	})

	// Should refresh because token is within refresh buffer
	token, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "near-expiry-token" {
		t.Errorf("expected token 'near-expiry-token', got %s", token)
	}

	if provider.getCallCount() != 1 {
		t.Errorf("expected provider called 1 time, got %d", provider.getCallCount())
	}
}

func TestGetTokenProviderError(t *testing.T) {
	provider := &mockProvider{
		err: errors.New("provider error"),
	}

	manager := NewManager(provider, DefaultConfig())

	_, err := manager.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error from provider")
	}

	if err.Error() != "fetch token: provider error" {
		t.Errorf("expected 'fetch token: provider error', got %q", err.Error())
	}
}

func TestForceRefresh(t *testing.T) {
	provider := &mockProvider{
		token:  "force-refresh-token",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	// First call to populate cache
	_, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("initial GetToken failed: %v", err)
	}

	// Verify first call happened
	if provider.getCallCount() != 1 {
		t.Errorf("expected provider called 1 time after initial GetToken, got %d", provider.getCallCount())
	}

	// Force refresh should call provider again
	token, err := manager.ForceRefresh(context.Background())
	if err != nil {
		t.Fatalf("ForceRefresh failed: %v", err)
	}

	if token != "force-refresh-token" {
		t.Errorf("expected token 'force-refresh-token', got %s", token)
	}

	if provider.getCallCount() != 2 {
		t.Errorf("expected provider called 2 times, got %d", provider.getCallCount())
	}
}

func TestInvalidateToken(t *testing.T) {
	provider := &mockProvider{
		token:  "invalidate-test",
		expiry: time.Now().Add(time.Hour),
	}

	storage := NewMemoryStorage()
	manager := NewManager(provider, Config{
		Storage: storage,
	})

	// Populate cache and storage
	_, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	// Verify token is cached
	if !manager.HasValidToken() {
		t.Error("expected token to be valid")
	}

	// Invalidate
	manager.InvalidateToken()

	// Verify token is invalidated
	if manager.HasValidToken() {
		t.Error("expected token to be invalid after invalidation")
	}

	// Next call should fetch from provider again
	_, err = manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken after invalidation failed: %v", err)
	}

	if provider.getCallCount() != 2 {
		t.Errorf("expected provider called 2 times, got %d", provider.getCallCount())
	}
}

func TestHasValidToken(t *testing.T) {
	now := time.Now()
	provider := &mockProvider{
		token:  "valid-token",
		expiry: now.Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	// No token initially
	if manager.HasValidToken() {
		t.Error("expected no valid token initially")
	}

	// After fetching
	_, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if !manager.HasValidToken() {
		t.Error("expected valid token after fetching")
	}

	// After invalidation
	manager.InvalidateToken()
	if manager.HasValidToken() {
		t.Error("expected no valid token after invalidation")
	}
}

func TestTokenExpiry(t *testing.T) {
	expiry := time.Now().Add(time.Hour)
	provider := &mockProvider{
		token:  "expiry-test",
		expiry: expiry,
	}

	manager := NewManager(provider, DefaultConfig())

	// No expiry initially
	if !manager.TokenExpiry().IsZero() {
		t.Error("expected zero expiry initially")
	}

	// After fetching
	_, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if !manager.TokenExpiry().Equal(expiry) {
		t.Errorf("expected expiry %v, got %v", expiry, manager.TokenExpiry())
	}
}

func TestStats(t *testing.T) {
	now := time.Now()
	expiry := now.Add(time.Hour)
	provider := &mockProvider{
		token:  "stats-test",
		expiry: expiry,
	}

	storage := NewMemoryStorage()
	manager := NewManager(provider, Config{
		RefreshBuffer: 10 * time.Minute,
		Storage:       storage,
	})

	// Get stats before token
	stats := manager.Stats()
	if stats.HasToken {
		t.Error("expected no token initially")
	}
	if stats.IsValid {
		t.Error("expected token not valid initially")
	}
	if stats.StorageType != "memory" {
		t.Errorf("expected storage type 'memory', got %s", stats.StorageType)
	}

	// Get stats after token
	_, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	stats = manager.Stats()
	if !stats.HasToken {
		t.Error("expected to have token")
	}
	if !stats.IsValid {
		t.Error("expected token to be valid")
	}
	if !stats.TokenExpiry.Equal(expiry) {
		t.Errorf("expected expiry %v, got %v", expiry, stats.TokenExpiry)
	}
	if stats.TimeUntilRefresh <= 0 {
		t.Errorf("expected positive time until refresh, got %v", stats.TimeUntilRefresh)
	}
}

func TestConcurrentAccess(t *testing.T) {
	provider := &mockProvider{
		token:  "concurrent-test",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	const numGoroutines = 10
	const numCalls = 100

	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines*numCalls)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numCalls; j++ {
				_, err := manager.GetToken(context.Background())
				if err != nil {
					errChan <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
		t.Errorf("concurrent access error: %v", err)
	}

	// Provider should only be called once due to caching
	if provider.getCallCount() != 1 {
		t.Errorf("expected provider called 1 time, got %d", provider.getCallCount())
	}
}

func TestRefreshTokenDoubleCheck(t *testing.T) {
	provider := &mockProvider{
		token:  "double-check-test",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	// Simulate concurrent refresh scenario
	manager.mu.Lock()
	manager.token = "old-token"
	manager.expiry = time.Now().Add(-time.Hour) // Expired
	manager.mu.Unlock()

	// First call starts refresh
	go func() {
		manager.refreshToken(context.Background(), false)
	}()

	// Small delay to ensure first refresh starts
	time.Sleep(10 * time.Millisecond)

	// Second call should see that refresh is already in progress
	// and wait for the result
	token, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken during refresh failed: %v", err)
	}

	if token != "double-check-test" {
		t.Errorf("expected token 'double-check-test', got %s", token)
	}
}

func BenchmarkGetToken(b *testing.B) {
	provider := &mockProvider{
		token:  "bench-token",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	// Warm up cache
	_, err := manager.GetToken(context.Background())
	if err != nil {
		b.Fatalf("warmup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetToken(context.Background())
		if err != nil {
			b.Fatalf("GetToken failed: %v", err)
		}
	}
}

func BenchmarkGetTokenWithRefresh(b *testing.B) {
	provider := &mockProvider{
		token:  "bench-refresh-token",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Invalidate token to force refresh
		manager.InvalidateToken()

		_, err := manager.GetToken(context.Background())
		if err != nil {
			b.Fatalf("GetToken failed: %v", err)
		}
	}
}
