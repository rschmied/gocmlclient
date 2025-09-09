package auth

import (
	"os"
	"path/filepath"
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
