// Package testutil provides some common test functions
package testutil

import (
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/lmittmann/tint"
	cmlclient "github.com/rschmied/gocmlclient"
)

// ClientConfig holds configuration for test clients
type ClientConfig struct {
	BaseURL  string
	Username string
	Password string
	Token    string
}

// DefaultConfig returns default test configuration from environment
func DefaultConfig() ClientConfig {
	return ClientConfig{
		BaseURL:  GetEnvOrDefault("CML_HOST", "https://controller"),
		Username: os.Getenv("CML_USER"),
		Password: os.Getenv("CML_PASS"),
		Token:    os.Getenv("CML_TOKEN"),
	}
}

// NewAPIClient creates a test API client - either mock or live based on TEST_LIVE env var
func NewAPIClient(t *testing.T) (*cmlclient.Client, func()) {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			AddSource:  true,
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))
	return NewAPIClientWithConfig(t, DefaultConfig())
}

// NewAPIClientWithConfig creates a test API client with custom config
func NewAPIClientWithConfig(t *testing.T, config ClientConfig) (*cmlclient.Client, func()) {
	if IsLiveTesting() {
		return newLiveClient(t, config)
	}
	return newMockClient(t, config)
}

// IsLiveTesting returns true if running against live backend
func IsLiveTesting() bool {
	// return true
	return os.Getenv("TEST_LIVE") != ""
}

// RequireLive skips the test if not running in live mode
func RequireLive(t *testing.T) {
	if !IsLiveTesting() {
		t.Skip("Set TEST_LIVE=1 to run live integration tests")
	}
}

// SkipIfLive skips the test if running in live mode
func SkipIfLive(t *testing.T) {
	if IsLiveTesting() {
		t.Skip("Skipping mock-only test in live mode")
	}
}

// GetEnvOrDefault returns environment variable value or default
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func newLiveClient(t *testing.T, config ClientConfig) (*cmlclient.Client, func()) {
	if config.Token == "" && (config.Username == "" || config.Password == "") {
		t.Skip("Live tests require either CML_TOKEN or CML_USER+CML_PASS")
	}
	client := cmlclient.New(config.BaseURL, true)
	return client, func() {}
}

func newMockClient(t *testing.T, config ClientConfig) (*cmlclient.Client, func()) {
	_ = t
	_ = config
	httpmock.Activate()

	httpClient := &http.Client{}
	httpmock.ActivateNonDefault(httpClient)

	SetupCommonMocks()

	client := cmlclient.New("https://mock", true)
	client.SetHTTPClient(httpClient, true)
	return client, func() {
		httpmock.DeactivateAndReset()
	}
}

// SetupCommonMocks sets up HTTP mocks common across all services
func SetupCommonMocks() {
	httpmock.RegisterResponder("POST", "https://mock/api/v0/auth_extended",
		httpmock.NewStringResponder(401, `{"id":"user-123","username":"testuser","token":"mock-token-12345","admin":false}`))

	httpmock.RegisterResponder("GET", "https://mock/api/v0/system_information",
		httpmock.NewStringResponder(200, `{"ready":true,"version":"2.8.1","build":"123"}`))
	httpmock.RegisterResponder("GET", "https://mock/api/v0/authok",
		httpmock.NewStringResponder(200, `"OK"`))
}
