// Package auth provides usage examples for the enhanced authentication system
package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// ExampleMemoryStorage demonstrates using memory storage (default)
func ExampleMemoryStorage() {
	// Create auth provider
	provider := &AuthProvider{
		baseURL:  "https://cml-controller.example.com",
		username: "admin",
		password: "password",
		client:   &http.Client{Timeout: 30 * time.Second},
	}

	// Create manager with default memory storage
	manager := NewManager(provider, DefaultConfig())

	// Use the manager
	ctx := context.Background()
	token, err := manager.GetToken(ctx)
	if err != nil {
		log.Printf("Failed to get token: %v", err)
		return
	}

	fmt.Printf("Got token: %s\n", token)
	fmt.Printf("Storage type: %s\n", manager.storage.Type())
}

// ExampleFileStorage demonstrates using file storage for token persistence
func ExampleFileStorage() {
	// Create file storage
	storage, err := NewFileStorage("/tmp/cml_tokens.json")
	if err != nil {
		log.Printf("Failed to create file storage: %v", err)
		return
	}

	// Create auth provider
	provider := &AuthProvider{
		baseURL:  "https://cml-controller.example.com",
		username: "admin",
		password: "password",
		client:   &http.Client{Timeout: 30 * time.Second},
	}

	// Create manager with file storage
	config := DefaultConfig()
	config.Storage = storage
	manager := NewManager(provider, config)

	// Use the manager
	ctx := context.Background()
	token, err := manager.GetToken(ctx)
	if err != nil {
		log.Printf("Failed to get token: %v", err)
		return
	}

	fmt.Printf("Got token: %s\n", token)
	fmt.Printf("Storage type: %s\n", manager.storage.Type())

	// Token will be persisted to file and restored on next run
}

// CustomProvider demonstrates creating a custom token provider
type CustomProvider struct {
	apiKey string
}

func (p *CustomProvider) FetchToken(ctx context.Context) (string, time.Time, error) {
	// Custom token fetching logic
	token := "custom-" + p.apiKey
	expiry := time.Now().Add(1 * time.Hour)
	return token, expiry, nil
}

func (p *CustomProvider) Type() string {
	return "custom"
}

func ExampleCustomProvider() {
	// Create custom provider
	provider := &CustomProvider{apiKey: "my-api-key"}

	// Create manager with custom provider
	manager := NewManager(provider, DefaultConfig())

	// Use the manager
	ctx := context.Background()
	token, err := manager.GetToken(ctx)
	if err != nil {
		log.Printf("Failed to get token: %v", err)
		return
	}

	fmt.Printf("Got token: %s\n", token)
	fmt.Printf("Provider type: %s\n", provider.Type())
}
