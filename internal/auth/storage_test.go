package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage()

	// Test empty storage
	_, _, err := storage.Retrieve()
	if err == nil {
		t.Error("expected error for empty storage")
	}

	// Test storing and retrieving
	token := "test-token"
	expiry := time.Now().Add(time.Hour)

	err = storage.Store(token, expiry)
	if err != nil {
		t.Fatalf("failed to store token: %v", err)
	}

	retrievedToken, retrievedExpiry, err := storage.Retrieve()
	if err != nil {
		t.Fatalf("failed to retrieve token: %v", err)
	}

	if retrievedToken != token {
		t.Errorf("expected token %s, got %s", token, retrievedToken)
	}

	if !retrievedExpiry.Equal(expiry) {
		t.Errorf("expected expiry %v, got %v", expiry, retrievedExpiry)
	}

	// Test clearing
	err = storage.Clear()
	if err != nil {
		t.Fatalf("failed to clear storage: %v", err)
	}

	_, _, err = storage.Retrieve()
	if err == nil {
		t.Error("expected error after clearing storage")
	}

	// Test type
	if storage.Type() != "memory" {
		t.Errorf("expected type 'memory', got %s", storage.Type())
	}
}

func TestFileStorage(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "auth_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "token.json")
	storage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}

	// Test empty storage
	_, _, err = storage.Retrieve()
	if err == nil {
		t.Error("expected error for empty storage")
	}

	// Test storing and retrieving
	token := "file-test-token"
	expiry := time.Now().Add(2 * time.Hour)

	err = storage.Store(token, expiry)
	if err != nil {
		t.Fatalf("failed to store token: %v", err)
	}

	retrievedToken, retrievedExpiry, err := storage.Retrieve()
	if err != nil {
		t.Fatalf("failed to retrieve token: %v", err)
	}

	if retrievedToken != token {
		t.Errorf("expected token %s, got %s", token, retrievedToken)
	}

	if !retrievedExpiry.Equal(expiry) {
		t.Errorf("expected expiry %v, got %v", expiry, retrievedExpiry)
	}

	// Test clearing
	err = storage.Clear()
	if err != nil {
		t.Fatalf("failed to clear storage: %v", err)
	}

	_, _, err = storage.Retrieve()
	if err == nil {
		t.Error("expected error after clearing storage")
	}

	// Test type
	if storage.Type() != "file" {
		t.Errorf("expected type 'file', got %s", storage.Type())
	}
}

func TestFileStoragePersistence(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "auth_persistence_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "persistent_token.json")

	// Create first storage instance and store token
	storage1, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("failed to create first storage: %v", err)
	}

	token := "persistent-token"
	expiry := time.Now().Add(3 * time.Hour)

	err = storage1.Store(token, expiry)
	if err != nil {
		t.Fatalf("failed to store token: %v", err)
	}

	// Create second storage instance and verify it can read the persisted token
	storage2, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("failed to create second storage: %v", err)
	}

	retrievedToken, retrievedExpiry, err := storage2.Retrieve()
	if err != nil {
		t.Fatalf("failed to retrieve persisted token: %v", err)
	}

	if retrievedToken != token {
		t.Errorf("expected persisted token %s, got %s", token, retrievedToken)
	}

	if !retrievedExpiry.Equal(expiry) {
		t.Errorf("expected persisted expiry %v, got %v", expiry, retrievedExpiry)
	}
}

func TestFileStorageInvalidFile(t *testing.T) {
	// Test with invalid file path
	_, err := NewFileStorage("/invalid/path/token.json")
	if err == nil {
		t.Error("expected error for invalid file path")
	}

	// Test with valid directory but invalid permissions
	tempDir, err := os.MkdirTemp("", "auth_invalid_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a file and make directory read-only
	filePath := filepath.Join(tempDir, "readonly", "token.json")
	err = os.MkdirAll(filepath.Dir(filePath), 0o400)
	if err != nil {
		t.Fatalf("failed to create readonly dir: %v", err)
	}

	storage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// This should fail due to permissions
	err = storage.Store("token", time.Now().Add(time.Hour))
	if err == nil {
		t.Error("expected error when storing to read-only directory")
	}
}

func BenchmarkMemoryStorage(b *testing.B) {
	storage := NewMemoryStorage()
	token := "benchmark-token"
	expiry := time.Now().Add(time.Hour)

	for b.Loop() {
		storage.Store(token, expiry)
		storage.Retrieve()
	}
}

func TestFileStoragePermissionDenied(t *testing.T) {
	// Create a read-only directory
	tempDir, err := os.MkdirTemp("", "auth_readonly_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a subdirectory and make it read-only
	readonlyDir := filepath.Join(tempDir, "readonly")
	err = os.MkdirAll(readonlyDir, 0o400) // Read-only for owner
	if err != nil {
		t.Fatalf("failed to create readonly dir: %v", err)
	}

	filePath := filepath.Join(readonlyDir, "token.json")

	storage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	token := "permission-test-token"
	expiry := time.Now().Add(time.Hour)

	// Store should fail due to permissions
	err = storage.Store(token, expiry)
	if err == nil {
		t.Error("expected permission denied error on store")
	}

	// Retrieve should also fail or return empty
	_, _, err = storage.Retrieve()
	if err == nil {
		t.Error("expected error on retrieve from read-only location")
	}
}

func TestFileStorageCorruptedJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "auth_corrupt_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "corrupt_token.json")

	// Create a file with invalid JSON
	err = os.WriteFile(filePath, []byte("invalid json content"), 0o600)
	if err != nil {
		t.Fatalf("failed to write corrupt file: %v", err)
	}

	storage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// Retrieve should fail due to corrupted JSON
	_, _, err = storage.Retrieve()
	if err == nil {
		t.Error("expected error when reading corrupted JSON")
	}
}

func TestFileStorageEmptyFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "auth_empty_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "empty_token.json")

	// Create an empty file
	err = os.WriteFile(filePath, []byte(""), 0o600)
	if err != nil {
		t.Fatalf("failed to write empty file: %v", err)
	}

	storage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// Retrieve should fail due to empty file
	_, _, err = storage.Retrieve()
	if err == nil {
		t.Error("expected error when reading empty file")
	}
}

func TestFileStorageInvalidJSONStructure(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "auth_invalid_json_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "invalid_structure.json")

	// Create a file with valid JSON but wrong structure
	invalidJSON := `{"wrong_field": "value"}`
	err = os.WriteFile(filePath, []byte(invalidJSON), 0o600)
	if err != nil {
		t.Fatalf("failed to write invalid JSON file: %v", err)
	}

	storage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// Retrieve should fail due to missing required fields
	_, _, err = storage.Retrieve()
	if err == nil {
		t.Error("expected error when reading JSON with missing fields")
	}
}

func TestFileStorageConcurrentAccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "auth_concurrent_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "concurrent_token.json")

	// Create storage instance
	storage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	const numGoroutines = 10
	const numOperations = 50

	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines*numOperations)

	// Start multiple goroutines performing concurrent operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				token := fmt.Sprintf("token-%d-%d", id, j)
				expiry := time.Now().Add(time.Duration(j) * time.Minute)

				// Alternate between store and retrieve operations
				if j%2 == 0 {
					err := storage.Store(token, expiry)
					if err != nil {
						errChan <- fmt.Errorf("store error: %v", err)
						return
					}
				} else {
					_, _, err := storage.Retrieve()
					if err != nil {
						errChan <- fmt.Errorf("retrieve error: %v", err)
						return
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
		t.Errorf("concurrent access error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("got %d errors from concurrent file access", errorCount)
	}
}

func TestFileStorageLargeToken(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "auth_large_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "large_token.json")
	storage, err := NewFileStorage(filePath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// Create a large token (10KB)
	largeToken := strings.Repeat("a", 10*1024)
	expiry := time.Now().Add(time.Hour)

	// Store large token
	err = storage.Store(largeToken, expiry)
	if err != nil {
		t.Fatalf("failed to store large token: %v", err)
	}

	// Retrieve and verify
	retrievedToken, retrievedExpiry, err := storage.Retrieve()
	if err != nil {
		t.Fatalf("failed to retrieve large token: %v", err)
	}

	if retrievedToken != largeToken {
		t.Errorf("expected large token, got different content (length: expected %d, got %d)",
			len(largeToken), len(retrievedToken))
	}

	if !retrievedExpiry.Equal(expiry) {
		t.Errorf("expected expiry %v, got %v", expiry, retrievedExpiry)
	}
}

func BenchmarkFileStorage(b *testing.B) {
	tempDir, _ := os.MkdirTemp("", "bench_file_test")
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "bench_token.json")
	storage, _ := NewFileStorage(filePath)

	token := "benchmark-token"
	expiry := time.Now().Add(time.Hour)

	for b.Loop() {
		storage.Store(token, expiry)
		storage.Retrieve()
	}
}
