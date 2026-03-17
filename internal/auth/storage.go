// Package auth provides authentication services with pluggable storage
package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TokenStorage defines the interface for token storage backends
type TokenStorage interface {
	// Store saves a token with its expiry time
	Store(token string, expiry time.Time) error

	// Retrieve gets the stored token and expiry time
	Retrieve() (token string, expiry time.Time, err error)

	// Clear removes the stored token
	Clear() error

	// Type returns the storage type identifier
	Type() string
}

// MemoryStorage implements in-memory token storage (default)
type MemoryStorage struct {
	mu     sync.RWMutex
	token  string
	expiry time.Time
}

// NewMemoryStorage creates a new in-memory token storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

// Store implements TokenStorage
func (s *MemoryStorage) Store(token string, expiry time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.token = token
	s.expiry = expiry
	return nil
}

// Retrieve implements TokenStorage
func (s *MemoryStorage) Retrieve() (string, time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.token == "" {
		return "", time.Time{}, fmt.Errorf("no token stored")
	}

	return s.token, s.expiry, nil
}

// Clear implements TokenStorage
func (s *MemoryStorage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.token = ""
	s.expiry = time.Time{}
	return nil
}

// Type implements TokenStorage
func (s *MemoryStorage) Type() string {
	return "memory"
}

// FileStorage implements file-based token storage
type FileStorage struct {
	filePath string
	mu       sync.RWMutex
}

// tokenData represents the structure stored in the file
type tokenData struct {
	Token  string    `json:"token"`
	Expiry time.Time `json:"expiry"`
}

// NewFileStorage creates a new file-based token storage
func NewFileStorage(filePath string) (*FileStorage, error) {
	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}

	return &FileStorage{
		filePath: filePath,
	}, nil
}

// Store implements TokenStorage
func (s *FileStorage) Store(token string, expiry time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data := tokenData{
		Token:  token,
		Expiry: expiry,
	}

	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("open token file: %w", err)
	}
	defer file.Close() //nolint:errcheck

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("encode token data: %w", err)
	}

	return nil
}

// Retrieve implements TokenStorage
func (s *FileStorage) Retrieve() (string, time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", time.Time{}, fmt.Errorf("no token stored")
		}
		return "", time.Time{}, fmt.Errorf("open token file: %w", err)
	}
	defer file.Close() //nolint:errcheck

	var data tokenData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return "", time.Time{}, fmt.Errorf("decode token data: %w", err)
	}

	if data.Token == "" {
		return "", time.Time{}, fmt.Errorf("no token stored")
	}

	return data.Token, data.Expiry, nil
}

// Clear implements TokenStorage
func (s *FileStorage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.Remove(s.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove token file: %w", err)
	}

	return nil
}

// Type implements TokenStorage
func (s *FileStorage) Type() string {
	return "file"
}
