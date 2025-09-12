package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/stretchr/testify/assert"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.baseURL, tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			tt.validate(t, client)
		})
	}
}

func TestLabGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v0/auth_extended":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"user-123","username":"testuser","token":"mock-token-12345","admin":false}`))
		case "/api/v0/labs/test-id":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "test-id",
				"lab_title": "Test Lab",
				"lab_description": "A test lab",
				"created": "2025-01-01T00:00:00Z",
				"modified": "2025-01-01T00:00:00Z"
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := New(server.URL, SkipReadyCheck())
	assert.NoError(t, err)
	assert.NotNil(t, client.Lab)

	ctx := context.Background()
	lab, err := client.LabGet(ctx, "test-id", false)
	assert.NoError(t, err)
	assert.NotNil(t, lab)
	assert.Equal(t, "test-id", string(lab.ID))
	assert.Equal(t, "Test Lab", lab.Title)
}

func TestReadyCheckIntegration(t *testing.T) {
	// Test that Ready() check works with a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v0/system_information":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"version": "2.5.0", "ready": true}`))
		case "/api/v0/auth_extended":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"user-123","username":"testuser","token":"mock-token-12345","admin":false}`))
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
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with stats enabled
	client, err := New(server.URL, SkipReadyCheck())
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Get stats - should return public Stats type
	stats := client.Stats()
	
	// Verify it's the correct type (public Stats from pkg/client)
	assert.IsType(t, Stats{}, stats)
	
	// Verify fields are accessible (this would fail if it was internal api.Stats)
	assert.NotNil(t, stats.CallsByMethod)
	assert.NotNil(t, stats.CallsByEndpoint)
	assert.NotNil(t, stats.StatusCounts)
	assert.NotNil(t, stats.ResponseTimes)
}
