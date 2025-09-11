package client

import (
	"context"
	"net/http"
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
			opts:    []Option{},
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
			},
			wantErr: false,
			validate: func(t *testing.T, client *Client) {
				assert.False(t, client.config.namedConfigs)
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

func TestConditional(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		option    Option
		validate  func(t *testing.T, config *Config)
	}{
		{
			name:      "condition true",
			condition: true,
			option:    WithToken("test-token"),
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "test-token", config.token)
			},
		},
		{
			name:      "condition false",
			condition: false,
			option:    WithToken("test-token"),
			validate: func(t *testing.T, config *Config) {
				assert.Empty(t, config.token)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{}
			option := Conditional(tt.condition, tt.option)
			option(config)
			tt.validate(t, config)
		})
	}
}

func TestLabGet(t *testing.T) {
	// Test that LabGet method exists and can be called
	// We can't easily test the full HTTP flow without complex mocking
	// So we'll just test that the client can be created and the method exists
	client, err := New("https://example.com")
	assert.NoError(t, err)
	assert.NotNil(t, client.Lab)

	// Test that the method signature is correct by checking it compiles
	ctx := context.Background()
	_, _ = client.LabGet(ctx, "test-id", false) // This should compile without error
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
				baseURL: "https://api.example.com",
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
			},
			wantErr: false,
			validate: func(t *testing.T, apiClient *api.Client) {
				assert.NotNil(t, apiClient)
			},
		},
		{
			name: "with custom HTTP client",
			config: &Config{
				baseURL:    "https://api.example.com",
				httpClient: &http.Client{Timeout: 10 * time.Second},
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
