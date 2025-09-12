// Package testutil provides some common test functions
package testutil

import (
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/lmittmann/tint"
	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/auth"
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
		BaseURL:  GetEnvOrDefault("CML_BASE_URL", "https://localhost:8443"),
		Username: os.Getenv("CML_USER"),
		Password: os.Getenv("CML_PASS"),
		Token:    os.Getenv("CML_TOKEN"),
	}
}

// NewAPIClient creates a test API client - either mock or live based on TEST_LIVE env var
func NewAPIClient(t *testing.T) (*api.Client, func()) {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			AddSource:  true,
			Level:      slog.LevelInfo,
			TimeFormat: time.Kitchen,
		}),
	))
	return NewAPIClientWithConfig(t, DefaultConfig())
}

// NewAPIClientWithConfig creates a test API client with custom config
func NewAPIClientWithConfig(t *testing.T, config ClientConfig) (*api.Client, func()) {
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

func newLiveClient(t *testing.T, config ClientConfig) (*api.Client, func()) {
	if config.Token == "" && (config.Username == "" || config.Password == "") {
		t.Skip("Live tests require either CML_TOKEN or CML_USER+CML_PASS")
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	if config.Token != "" || (config.Username != "" && config.Password != "") {
		provider := auth.NewAuthProvider(auth.AuthConfig{
			BaseURL:     config.BaseURL,
			Username:    config.Username,
			Password:    config.Password,
			PresetToken: config.Token,
			Client:      httpClient,
		})

		manager := auth.NewManager(provider, auth.DefaultConfig())
		authTransport := auth.NewTransport(httpClient.Transport, manager, nil)
		httpClient.Transport = authTransport
	}

	apiClient := api.New(config.BaseURL,
		api.WithHTTPClient(httpClient),
		api.WithMiddlewares(
			api.UserAgentMiddleware("gocmlclient"),
			// api.LoggingMiddleware(slog.Default()),
			// api.LogRequestBodyMiddleware(slog.Default()),
			// api.RetryMiddleware(api.DefaultRetryPolicy()),
		),
	)

	return apiClient, func() {}
}

func newMockClient(t *testing.T, config ClientConfig) (*api.Client, func()) {
	_ = t
	_ = config
	httpmock.Activate()

	client := &http.Client{}
	httpmock.ActivateNonDefault(client)

	SetupCommonMocks()

	apiClient := api.New("https://mock",
		api.WithHTTPClient(client),
		api.WithMiddlewares(api.UserAgentMiddleware("gocmlclient")),
	)

	return apiClient, func() {
		httpmock.DeactivateAndReset()
	}
}

// SetupCommonMocks sets up HTTP mocks common across all services
func SetupCommonMocks() {
	httpmock.RegisterResponder("POST", "https://mock/api/v0/auth_extended",
		httpmock.NewStringResponder(200, `{"id":"user-123","username":"testuser","token":"mock-token-12345","admin":false}`))

	httpmock.RegisterResponder("GET", "https://mock/api/v0/system_information",
		httpmock.NewStringResponder(200, `{"ready":true,"version":"2.8.1","build":"123"}`))
}
