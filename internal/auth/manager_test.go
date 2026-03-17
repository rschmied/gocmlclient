package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestGetTokenCached(t *testing.T) {
	provider := &mockProvider{
		token:  "cached-token",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	// First call to populate cache
	token1, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("initial GetToken failed: %v", err)
	}

	// Second call should return cached token without calling provider again
	token2, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("cached GetToken failed: %v", err)
	}

	if token1 != "cached-token" {
		t.Errorf("expected token 'cached-token', got %s", token1)
	}

	if token2 != token1 {
		t.Errorf("expected same cached token, got %s vs %s", token1, token2)
	}

	if provider.getCallCount() != 1 {
		t.Errorf("expected provider called 1 time (cached), got %d", provider.getCallCount())
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

	for range numGoroutines {
		wg.Go(func() {
			for range numCalls {
				_, err := manager.GetToken(context.Background())
				if err != nil {
					errChan <- err
				}
			}
		})
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
		manager.refreshToken(context.Background()) //nolint:errcheck
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

func TestConcurrentRefreshRace(t *testing.T) {
	provider := &mockProvider{
		token:  "race-test-token",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	// Expire the token to force refresh
	manager.mu.Lock()
	manager.token = "expired-token"
	manager.expiry = time.Now().Add(-time.Hour)
	manager.mu.Unlock()

	const numGoroutines = 20
	var wg sync.WaitGroup
	results := make(chan string, numGoroutines)
	errors := make(chan error, numGoroutines)

	for range numGoroutines {
		wg.Go(func() {
			token, err := manager.GetToken(context.Background())
			if err != nil {
				errors <- err
				return
			}
			results <- token
		})
	}

	wg.Wait()
	close(results)
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Errorf("concurrent refresh error: %v", err)
		errorCount++
	}
	if errorCount > 0 {
		t.Fatalf("got %d errors from concurrent refreshes", errorCount)
	}

	// All results should be the same token
	var firstToken string
	for token := range results {
		if firstToken == "" {
			firstToken = token
		} else if token != firstToken {
			t.Errorf("inconsistent tokens: expected %s, got %s", firstToken, token)
		}
	}

	// Provider should only be called once
	if provider.getCallCount() != 1 {
		t.Errorf("expected provider called 1 time, got %d", provider.getCallCount())
	}
}

func TestConcurrent401RefreshRace(t *testing.T) {
	provider := &mockProvider{
		token:  "401-race-token",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	// Simulate 401: invalidate token once
	manager.InvalidateToken()

	const numGoroutines = 10
	var wg sync.WaitGroup
	results := make(chan string, numGoroutines)
	errors := make(chan error, numGoroutines)

	for range numGoroutines {
		wg.Go(func() {
			// Simulate concurrent requests trying to refresh after 401
			token, err := manager.GetToken(context.Background())
			if err != nil {
				errors <- err
				return
			}
			results <- token
		})
	}

	wg.Wait()
	close(results)
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Errorf("concurrent 401 refresh error: %v", err)
		errorCount++
	}
	if errorCount > 0 {
		t.Fatalf("got %d errors from concurrent 401 refreshes", errorCount)
	}

	// All results should be the same token
	var firstToken string
	for token := range results {
		if firstToken == "" {
			firstToken = token
		} else if token != firstToken {
			t.Errorf("inconsistent tokens: expected %s, got %s", firstToken, token)
		}
	}

	// Provider should only be called once despite multiple concurrent GetToken calls
	if provider.getCallCount() != 1 {
		t.Errorf("expected provider called 1 time, got %d", provider.getCallCount())
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

	for b.Loop() {
		_, err := manager.GetToken(context.Background())
		if err != nil {
			b.Fatalf("GetToken failed: %v", err)
		}
	}
}

// mockStorage implements TokenStorage for testing error scenarios
type mockStorage struct {
	storeErr    error
	retrieveErr error
	clearErr    error
	token       string
	expiry      time.Time
}

func (m *mockStorage) Store(token string, expiry time.Time) error {
	if m.storeErr != nil {
		return m.storeErr
	}
	m.token = token
	m.expiry = expiry
	return nil
}

func (m *mockStorage) Retrieve() (string, time.Time, error) {
	if m.retrieveErr != nil {
		return "", time.Time{}, m.retrieveErr
	}
	return m.token, m.expiry, nil
}

func (m *mockStorage) Clear() error {
	if m.clearErr != nil {
		return m.clearErr
	}
	m.token = ""
	m.expiry = time.Time{}
	return nil
}

func (m *mockStorage) Type() string {
	return "mock"
}

func TestGetTokenStorageStoreFailure(t *testing.T) {
	provider := &mockProvider{
		token:  "test-token",
		expiry: time.Now().Add(time.Hour),
	}

	storage := &mockStorage{
		storeErr: errors.New("storage store failed"),
	}

	manager := NewManager(provider, Config{
		Storage: storage,
	})

	// First call should succeed (token fetched and cached in memory)
	token, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "test-token" {
		t.Errorf("expected token 'test-token', got %s", token)
	}

	// Second call should also succeed (cached in memory, not persisted)
	token2, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("cached GetToken failed: %v", err)
	}

	if token2 != "test-token" {
		t.Errorf("expected cached token 'test-token', got %s", token2)
	}
}

func TestGetTokenProviderExpiredToken(t *testing.T) {
	now := time.Now()
	provider := &mockProvider{
		token:  "expired-from-provider",
		expiry: now.Add(-time.Hour), // Provider returns already expired token
	}

	manager := NewManager(provider, DefaultConfig())

	token, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "expired-from-provider" {
		t.Errorf("expected token 'expired-from-provider', got %s", token)
	}

	// Token should be cached but marked as invalid
	if manager.HasValidToken() {
		t.Error("expected expired token to be invalid")
	}

	// But it should still be in memory
	if manager.token != "expired-from-provider" {
		t.Error("expected expired token to be cached in memory")
	}
}

func TestGetTokenProviderEmptyToken(t *testing.T) {
	provider := &mockProvider{
		token:  "", // Empty token
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	_, err := manager.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error for empty token")
	}

	expectedError := "provider returned empty token"
	if err.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, err.Error())
	}
}

func TestGetTokenProviderZeroExpiry(t *testing.T) {
	provider := &mockProvider{
		token:  "zero-expiry-token",
		expiry: time.Time{}, // Zero expiry
	}

	manager := NewManager(provider, DefaultConfig())

	token, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "zero-expiry-token" {
		t.Errorf("expected token 'zero-expiry-token', got %s", token)
	}

	// Token should be considered expired immediately
	if manager.HasValidToken() {
		t.Error("expected token with zero expiry to be invalid")
	}
}

func TestInvalidateTokenStorageFailure(t *testing.T) {
	provider := &mockProvider{
		token:  "test-token",
		expiry: time.Now().Add(time.Hour),
	}

	storage := &mockStorage{
		clearErr: errors.New("storage clear failed"),
	}

	manager := NewManager(provider, Config{
		Storage: storage,
	})

	// Populate cache
	_, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("initial GetToken failed: %v", err)
	}

	// Invalidate should succeed even if storage clear fails
	manager.InvalidateToken()

	if manager.HasValidToken() {
		t.Error("expected token to be invalidated")
	}
}

func TestStatsWithZeroValues(t *testing.T) {
	provider := &mockProvider{
		token:  "stats-test",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, Config{
		RefreshBuffer: 10 * time.Minute,
		Storage:       NewMemoryStorage(),
	})

	// Test stats before any token
	stats := manager.Stats()
	if stats.HasToken {
		t.Error("expected no token initially")
	}
	if stats.TokenExpiry != (time.Time{}) {
		t.Error("expected zero expiry initially")
	}
	if stats.TimeUntilRefresh != 0 {
		t.Errorf("expected zero time until refresh, got %v", stats.TimeUntilRefresh)
	}

	// Test stats after token
	_, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	stats = manager.Stats()
	if !stats.HasToken {
		t.Error("expected to have token")
	}
	if stats.TimeUntilRefresh <= 0 {
		t.Errorf("expected positive time until refresh, got %v", stats.TimeUntilRefresh)
	}
}

func TestNewManagerInvalidRefreshBuffer(t *testing.T) {
	provider := &mockProvider{
		token:  "test-token",
		expiry: time.Now().Add(time.Hour),
	}

	// Test negative refresh buffer
	config := Config{
		RefreshBuffer: -1 * time.Hour, // Invalid negative value
		Storage:       NewMemoryStorage(),
	}

	manager := NewManager(provider, config)

	// Should use default refresh buffer
	if manager.refreshBuffer != 30*time.Second {
		t.Errorf("expected default refresh buffer 30s, got %v", manager.refreshBuffer)
	}
}

func TestNewManagerZeroRefreshBuffer(t *testing.T) {
	provider := &mockProvider{
		token:  "test-token",
		expiry: time.Now().Add(time.Hour),
	}

	// Test zero refresh buffer
	config := Config{
		RefreshBuffer: 0, // Zero value
		Storage:       NewMemoryStorage(),
	}

	manager := NewManager(provider, config)

	// Should use default refresh buffer
	if manager.refreshBuffer != 30*time.Second {
		t.Errorf("expected default refresh buffer 30s, got %v", manager.refreshBuffer)
	}
}

func TestNewManagerNilStorage(t *testing.T) {
	provider := &mockProvider{
		token:  "test-token",
		expiry: time.Now().Add(time.Hour),
	}

	// Test nil storage
	config := Config{
		RefreshBuffer: 10 * time.Second,
		Storage:       nil, // Nil storage
	}

	manager := NewManager(provider, config)

	// Should create default memory storage
	if manager.storage.Type() != "memory" {
		t.Errorf("expected memory storage, got %s", manager.storage.Type())
	}
}

func TestNewManagerWithStoredToken(t *testing.T) {
	provider := &mockProvider{
		token:  "provider-token",
		expiry: time.Now().Add(time.Hour),
	}

	// Pre-populate storage
	storage := NewMemoryStorage()
	storedToken := "stored-token"
	storedExpiry := time.Now().Add(2 * time.Hour)
	err := storage.Store(storedToken, storedExpiry)
	if err != nil {
		t.Fatalf("failed to pre-populate storage: %v", err)
	}

	config := Config{
		RefreshBuffer: 10 * time.Second,
		Storage:       storage,
	}

	manager := NewManager(provider, config)

	// Should load token from storage
	if !manager.HasValidToken() {
		t.Error("expected token to be loaded from storage")
	}

	token, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != storedToken {
		t.Errorf("expected stored token %s, got %s", storedToken, token)
	}

	// Provider should not be called since token was loaded from storage
	if provider.getCallCount() != 0 {
		t.Errorf("expected provider not called, got %d calls", provider.getCallCount())
	}
}

func TestNewManagerWithExpiredStoredToken(t *testing.T) {
	provider := &mockProvider{
		token:  "fresh-token",
		expiry: time.Now().Add(time.Hour),
	}

	// Pre-populate storage with expired token
	storage := NewMemoryStorage()
	expiredToken := "expired-token"
	expiredExpiry := time.Now().Add(-time.Hour) // Already expired
	err := storage.Store(expiredToken, expiredExpiry)
	if err != nil {
		t.Fatalf("failed to pre-populate storage: %v", err)
	}

	config := Config{
		RefreshBuffer: 10 * time.Second,
		Storage:       storage,
	}

	manager := NewManager(provider, config)

	// Should not have valid token initially (expired)
	if manager.HasValidToken() {
		t.Error("expected stored token to be expired")
	}

	// GetToken should fetch from provider
	token, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "fresh-token" {
		t.Errorf("expected fresh token from provider, got %s", token)
	}

	if provider.getCallCount() != 1 {
		t.Errorf("expected provider called 1 time, got %d", provider.getCallCount())
	}
}

func TestGetTokenBoundaryExpiry(t *testing.T) {
	now := time.Now()
	provider := &mockProvider{
		token:  "boundary-token",
		expiry: now.Add(30 * time.Second), // Exactly at refresh boundary
	}

	manager := NewManager(provider, Config{
		RefreshBuffer: 30 * time.Second, // Same as token lifetime
		Storage:       NewMemoryStorage(),
	})

	// Token should be considered expired (at boundary)
	if manager.HasValidToken() {
		t.Error("expected token at boundary to be invalid")
	}

	// GetToken should refresh
	token, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "boundary-token" {
		t.Errorf("expected boundary token, got %s", token)
	}

	if provider.getCallCount() != 1 {
		t.Errorf("expected provider called 1 time, got %d", provider.getCallCount())
	}
}

func TestGetTokenRapidRequests(t *testing.T) {
	provider := &mockProvider{
		token:  "rapid-token",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	// Make multiple rapid requests
	const numRequests = 100
	tokens := make([]string, numRequests)

	for i := range numRequests {
		token, err := manager.GetToken(context.Background())
		if err != nil {
			t.Fatalf("GetToken %d failed: %v", i, err)
		}
		tokens[i] = token
	}

	// All tokens should be the same (cached)
	for i := 1; i < numRequests; i++ {
		if tokens[i] != tokens[0] {
			t.Errorf("inconsistent tokens: %s vs %s", tokens[0], tokens[i])
		}
	}

	// Provider should only be called once
	if provider.getCallCount() != 1 {
		t.Errorf("expected provider called 1 time, got %d", provider.getCallCount())
	}
}

func TestInvalidateDuringRefresh(t *testing.T) {
	provider := &mockProvider{
		token:  "invalidate-during-refresh",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	// Populate cache
	_, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("initial GetToken failed: %v", err)
	}

	// Start a refresh in background
	go func() {
		time.Sleep(10 * time.Millisecond) // Small delay
		manager.InvalidateToken()
	}()

	// This request might happen during invalidation
	token, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken during invalidation failed: %v", err)
	}

	if token != "invalidate-during-refresh" {
		t.Errorf("expected token, got %s", token)
	}

	// Should eventually get a fresh token
	time.Sleep(50 * time.Millisecond)
	token2, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("second GetToken failed: %v", err)
	}

	if token2 != "invalidate-during-refresh" {
		t.Errorf("expected fresh token, got %s", token2)
	}
}

func TestTokenExpiryBoundaryConditions(t *testing.T) {
	testCases := []struct {
		name          string
		tokenExpiry   time.Time
		buffer        time.Duration
		shouldBeValid bool
	}{
		{"fresh token", time.Now().Add(time.Hour), 30 * time.Second, true},
		{"token at buffer boundary", time.Now().Add(30 * time.Second), 30 * time.Second, false},
		{"token just inside buffer", time.Now().Add(31 * time.Second), 30 * time.Second, true},
		{"token just outside buffer", time.Now().Add(29 * time.Second), 30 * time.Second, false},
		{"expired token", time.Now().Add(-time.Hour), 30 * time.Second, false},
		{"zero buffer", time.Now().Add(time.Hour), 0, true},
		{"large buffer", time.Now().Add(10 * time.Second), time.Hour, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := &mockProvider{
				token:  "boundary-test",
				expiry: tc.tokenExpiry,
			}

			manager := NewManager(provider, Config{
				RefreshBuffer: tc.buffer,
				Storage:       NewMemoryStorage(),
			})

			// Populate cache
			_, err := manager.GetToken(context.Background())
			if err != nil {
				t.Fatalf("GetToken failed: %v", err)
			}

			isValid := manager.HasValidToken()
			if isValid != tc.shouldBeValid {
				t.Errorf("expected valid=%v, got valid=%v", tc.shouldBeValid, isValid)
			}
		})
	}
}

func TestConcurrentInvalidateAndGet(t *testing.T) {
	provider := &mockProvider{
		token:  "concurrent-invalidate",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	// Populate cache
	_, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("initial GetToken failed: %v", err)
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines*2)

	// Start goroutines that alternate between GetToken and InvalidateToken
	for range numGoroutines {
		wg.Go(func() {
			for j := range 10 {
				// Alternate between operations
				if j%2 == 0 {
					_, err := manager.GetToken(context.Background())
					if err != nil {
						errChan <- err
					}
				} else {
					manager.InvalidateToken()
				}
			}
		})
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	errorCount := 0
	for err := range errChan {
		t.Errorf("concurrent operation error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("got %d errors from concurrent invalidate/get operations", errorCount)
	}
}

// Integration test: Full auth flow with memory storage
func TestIntegrationMemoryStorage(t *testing.T) {
	provider := &mockProvider{
		token:  "integration-memory",
		expiry: time.Now().Add(30 * time.Minute),
	}

	config := Config{
		RefreshBuffer: 5 * time.Minute,
		Storage:       NewMemoryStorage(),
	}

	manager := NewManager(provider, config)

	// Test initial token fetch
	token1, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("initial GetToken failed: %v", err)
	}

	if token1 != "integration-memory" {
		t.Errorf("expected token 'integration-memory', got %s", token1)
	}

	// Test caching
	token2, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("cached GetToken failed: %v", err)
	}

	if token1 != token2 {
		t.Errorf("cached token mismatch: %s vs %s", token1, token2)
	}

	if provider.getCallCount() != 1 {
		t.Errorf("expected provider called 1 time, got %d", provider.getCallCount())
	}

	// Test stats
	stats := manager.Stats()
	if !stats.HasToken {
		t.Error("expected to have token")
	}
	if !stats.IsValid {
		t.Error("expected token to be valid")
	}
	if stats.StorageType != "memory" {
		t.Errorf("expected storage type 'memory', got %s", stats.StorageType)
	}
}

// Integration test: Full auth flow with file storage
func TestIntegrationFileStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "auth_integration_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck

	filePath := filepath.Join(tempDir, "integration_token.json")
	fileStorage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}

	provider := &mockProvider{
		token:  "integration-file",
		expiry: time.Now().Add(45 * time.Minute),
	}

	config := Config{
		RefreshBuffer: 10 * time.Minute,
		Storage:       fileStorage,
	}

	manager := NewManager(provider, config)

	// Test initial token fetch and persistence
	token1, err := manager.GetToken(context.Background())
	if err != nil {
		t.Fatalf("initial GetToken failed: %v", err)
	}

	if token1 != "integration-file" {
		t.Errorf("expected token 'integration-file', got %s", token1)
	}

	// Verify token was persisted
	retrievedToken, _, err := fileStorage.Retrieve()
	if err != nil {
		t.Fatalf("failed to retrieve persisted token: %v", err)
	}

	if retrievedToken != token1 {
		t.Errorf("persisted token mismatch: %s vs %s", token1, retrievedToken)
	}

	// Create new manager instance to test loading from storage
	provider2 := &mockProvider{
		token:  "should-not-be-called",
		expiry: time.Now().Add(time.Hour),
	}

	manager2 := NewManager(provider2, config)

	// Should load token from storage without calling provider
	token2, err := manager2.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken from storage failed: %v", err)
	}

	if token2 != token1 {
		t.Errorf("loaded token mismatch: %s vs %s", token1, token2)
	}

	if provider2.getCallCount() != 0 {
		t.Errorf("expected provider not called when loading from storage, got %d calls", provider2.getCallCount())
	}
}

// Integration test: Auth transport with real HTTP server
func TestIntegrationTransportWithServer(t *testing.T) {
	// Create a test server that simulates a protected API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for authorization header
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Missing authorization")) //nolint:errcheck
			return
		}

		if auth != "Bearer test-token" {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Invalid token")) //nolint:errcheck
			return
		}

		// Return success for authorized requests
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":"protected content","user":"test"}`)) //nolint:errcheck
	}))
	defer server.Close()

	provider := &mockProvider{
		token:  "test-token",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())
	transport := NewTransport(http.DefaultTransport, manager, nil)

	client := &http.Client{Transport: transport}

	// Test authorized request (should succeed after fetching token)
	resp, err := client.Get(server.URL + "/api/data")
	if err != nil {
		t.Fatalf("authorized request failed: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for authorized request, got %d", resp.StatusCode)
	}

	// Verify response content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	expectedContent := `{"data":"protected content","user":"test"}`
	if string(body) != expectedContent {
		t.Errorf("expected response %q, got %q", expectedContent, string(body))
	}

	// Verify provider was called once
	if provider.getCallCount() != 1 {
		t.Errorf("expected provider called 1 time, got %d", provider.getCallCount())
	}
}

// Integration test: 401 retry mechanism
func TestIntegration401Retry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		auth := r.Header.Get("Authorization")

		if callCount == 1 {
			// First call: return 401 to trigger retry
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized")) //nolint:errcheck
			return
		}

		// Second call: should have auth header
		if auth != "Bearer retry-token" {
			t.Errorf("expected Bearer retry-token on retry, got %s", auth)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"success"}`)) //nolint:errcheck
	}))
	defer server.Close()

	provider := &mockProvider{
		token:  "retry-token",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())
	transport := NewTransport(http.DefaultTransport, manager, nil)

	client := &http.Client{Transport: transport}

	// Make request that will trigger 401 and retry
	resp, err := client.Get(server.URL + "/api/data")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 after retry, got %d", resp.StatusCode)
	}

	if callCount != 2 {
		t.Errorf("expected 2 server calls (initial + retry), got %d", callCount)
	}

	// Provider should be called twice: once for initial request, once after 401 invalidation
	// This is correct behavior - 401 should trigger token refresh
	if provider.getCallCount() != 2 {
		t.Errorf("expected provider called 2 times (initial + after 401), got %d", provider.getCallCount())
	}
}

// Stress test: High concurrency with token operations
func TestStressConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	provider := &mockProvider{
		token:  "stress-token",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	const numGoroutines = 50
	const numOperations = 100

	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines*numOperations)

	// Start multiple goroutines performing various operations
	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range numOperations {
				// Mix of different operations
				switch j % 4 {
				case 0:
					// Get token
					_, err := manager.GetToken(context.Background())
					if err != nil {
						errChan <- fmt.Errorf("GetToken error: %w", err)
					}
				case 1:
					// Check validity
					_ = manager.HasValidToken()
				case 2:
					// Get stats
					_ = manager.Stats()
				case 3:
					// Invalidate (less frequently)
					if j%10 == 0 {
						manager.InvalidateToken()
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Check for any errors
	errorCount := 0
	for err := range errChan {
		t.Errorf("stress test error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("got %d errors in stress test", errorCount)
	}

	// Verify final state
	if !manager.HasValidToken() {
		t.Error("expected token to be valid after stress test")
	}
}

// Stress test: Memory leak detection
func TestStressMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory stress test in short mode")
	}

	provider := &mockProvider{
		token:  "memory-test-token",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	// Perform many operations to check for memory issues
	for i := range 10000 {
		_, err := manager.GetToken(context.Background())
		if err != nil {
			t.Fatalf("GetToken failed at iteration %d: %v", i, err)
		}

		if i%100 == 0 {
			manager.InvalidateToken()
		}
	}

	// Final check
	if !manager.HasValidToken() {
		t.Error("expected token to be valid after memory stress test")
	}
}

// Performance regression test: Compare cached vs refresh performance
func BenchmarkGetTokenCached(b *testing.B) {
	provider := &mockProvider{
		token:  "bench-cached-token",
		expiry: time.Now().Add(time.Hour),
	}

	manager := NewManager(provider, DefaultConfig())

	// Warm up cache
	_, err := manager.GetToken(context.Background())
	if err != nil {
		b.Fatalf("warmup failed: %v", err)
	}

	b.ResetTimer()
	for b.Loop() {
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

	for b.Loop() {
		// Invalidate token to force refresh
		manager.InvalidateToken()

		_, err := manager.GetToken(context.Background())
		if err != nil {
			b.Fatalf("GetToken failed: %v", err)
		}
	}
}
