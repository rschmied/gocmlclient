package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/pkg/models"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		opts     []Option
		wantErr  bool
		validate func(t *testing.T, client *Client)
	}{
		{
			name:    "basic client creation",
			baseURL: "https://api.example.com",
			opts:    []Option{SkipReadyCheck()},
			wantErr: false,
			validate: func(t *testing.T, client *Client) {
				assert.NotNil(t, client)
				assert.NotNil(t, client.config)
				assert.NotNil(t, client.apiClient)
				assert.NotNil(t, client.Lab)
				assert.NotNil(t, client.Interface)
				assert.NotNil(t, client.Link)
				assert.NotNil(t, client.Node)
				assert.NotNil(t, client.Group)
				assert.NotNil(t, client.User)
				assert.NotNil(t, client.System)
			},
		},
		{
			name:    "with username password",
			baseURL: "https://api.example.com",
			opts: []Option{
				WithUsernamePassword("user", "pass"),
				SkipReadyCheck(),
			},
			wantErr: false,
			validate: func(t *testing.T, client *Client) {
				assert.Equal(t, "user", client.config.username)
				assert.Equal(t, "pass", client.config.password)
			},
		},
		{
			name:    "with token",
			baseURL: "https://api.example.com",
			opts: []Option{
				WithToken("test-token"),
				SkipReadyCheck(),
			},
			wantErr: false,
			validate: func(t *testing.T, client *Client) {
				assert.Equal(t, "test-token", client.config.token)
			},
		},
		{
			name:    "with insecure TLS",
			baseURL: "https://api.example.com",
			opts: []Option{
				WithInsecureTLS(),
				SkipReadyCheck(),
			},
			wantErr: false,
			validate: func(t *testing.T, client *Client) {
				assert.True(t, client.config.insecureSkipVerify)
			},
		},
		{
			name:    "with custom HTTP client",
			baseURL: "https://api.example.com",
			opts: []Option{
				WithHTTPClient(&http.Client{Timeout: 5 * time.Second}),
				SkipReadyCheck(),
			},
			wantErr: false,
			validate: func(t *testing.T, client *Client) {
				assert.Equal(t, 5*time.Second, client.config.httpClient.Timeout)
			},
		},
		{
			name:    "without named configs",
			baseURL: "https://api.example.com",
			opts: []Option{
				WithoutNamedConfigs(),
				SkipReadyCheck(),
			},
			wantErr: false,
			validate: func(t *testing.T, client *Client) {
				assert.False(t, client.config.namedConfigs)
			},
		},
		{
			name:    "with skip ready check",
			baseURL: "https://api.example.com",
			opts: []Option{
				SkipReadyCheck(),
			},
			wantErr: false,
			validate: func(t *testing.T, client *Client) {
				assert.True(t, client.config.skipReadyCheck)
			},
		},
		{
			name:    "error new api client",
			baseURL: "https://api.example.com",
			opts: []Option{
				WithTokenStorageFile("/nonexistent/path/token.json"),
			},
			wantErr: true,
			validate: func(t *testing.T, client *Client) {
				// Should not reach here
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.baseURL, tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
				return
			}
			assert.NoError(t, err)
			tt.validate(t, client)
		})
	}
}

func TestReadyCheckIntegration(t *testing.T) {
	// Test that Ready() check works with a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v0/system_information":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"version": "2.5.0", "ready": true}`)) //nolint:errcheck
		case "/api/v0/auth_extended":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"user-123","username":"testuser","token":"mock-token-12345","admin":false}`)) //nolint:errcheck
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// This should work without SkipReadyCheck since we have a mock server
	client, err := New(server.URL, WithToken("test-token"))
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "2.5.0", client.System.Version())
}

func TestNewAPIClient(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		wantErr  bool
		validate func(t *testing.T, apiClient *api.Client)
	}{
		{
			name: "basic API client",
			config: &Config{
				baseURL:        "https://api.example.com",
				skipReadyCheck: true, // Skip ready check for tests
			},
			wantErr: false,
			validate: func(t *testing.T, apiClient *api.Client) {
				assert.NotNil(t, apiClient)
			},
		},
		{
			name: "with insecure TLS",
			config: &Config{
				baseURL:            "https://api.example.com",
				insecureSkipVerify: true,
				skipReadyCheck:     true, // Skip ready check for tests
			},
			wantErr: false,
			validate: func(t *testing.T, apiClient *api.Client) {
				assert.NotNil(t, apiClient)
			},
		},
		{
			name: "with custom HTTP client",
			config: &Config{
				baseURL:        "https://api.example.com",
				httpClient:     &http.Client{Timeout: 10 * time.Second},
				skipReadyCheck: true, // Skip ready check for tests
			},
			wantErr: false,
			validate: func(t *testing.T, apiClient *api.Client) {
				assert.NotNil(t, apiClient)
			},
		},
		{
			name: "with token storage file",
			config: &Config{
				baseURL:          "https://api.example.com",
				tokenStorageFile: "/tmp/test_token.json",
				skipReadyCheck:   true,
			},
			wantErr: false,
			validate: func(t *testing.T, apiClient *api.Client) {
				assert.NotNil(t, apiClient)
			},
		},
		{
			name: "with invalid token storage file",
			config: &Config{
				baseURL:          "https://api.example.com",
				tokenStorageFile: "/nonexistent/path/token.json",
				skipReadyCheck:   true,
			},
			wantErr: true,
			validate: func(t *testing.T, apiClient *api.Client) {
				// Should not reach here
			},
		},
		{
			name: "with username password",
			config: &Config{
				baseURL:        "https://api.example.com",
				username:       "testuser",
				password:       "testpass",
				skipReadyCheck: true,
			},
			wantErr: false,
			validate: func(t *testing.T, apiClient *api.Client) {
				assert.NotNil(t, apiClient)
			},
		},
		{
			name: "with preset token",
			config: &Config{
				baseURL:        "https://api.example.com",
				token:          "test-token",
				skipReadyCheck: true,
			},
			wantErr: false,
			validate: func(t *testing.T, apiClient *api.Client) {
				assert.NotNil(t, apiClient)
			},
		},
		{
			name: "with CA cert PEM",
			config: &Config{
				baseURL:        "https://api.example.com",
				caCertPEM:      []byte("-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n"),
				skipReadyCheck: true,
			},
			wantErr: true,
			validate: func(t *testing.T, apiClient *api.Client) {
				// Should not reach here
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiClient, err := newAPIClient(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			tt.validate(t, apiClient)
		})
	}
}

func TestClient_Stats(t *testing.T) {
	// Create a test server that handles authentication
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v0/auth_extended":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"user-123","username":"testuser","token":"mock-token-12345","admin":false}`)) //nolint:errcheck
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Create client with authentication and stats enabled
	client, err := New(server.URL, WithToken("test-token"), SkipReadyCheck())
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Get stats - should return public Stats type
	stats := client.Stats()

	// Verify it's the correct type (public Stats from pkg/client)
	assert.IsType(t, &models.Stats{}, stats)

	// Verify EndpointGroups field is accessible
	assert.NotNil(t, stats.EndpointGroups)

	// Test computed getter methods
	assert.NotNil(t, stats.CallsByMethod())
	assert.NotNil(t, stats.CallsByEndpoint())
	assert.NotNil(t, stats.StatusCounts())
	assert.Equal(t, 0, stats.TotalCalls()) // Should be 0 since no calls made yet
}
