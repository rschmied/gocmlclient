package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/testutil"
)

func TestSystemReady(t *testing.T) {
	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	service := NewSystemService(client)
	ctx := context.Background()

	err := service.Ready(ctx)
	assert.NoError(t, err)
}

func TestNewSystemService(t *testing.T) {
	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	service := NewSystemService(client)
	assert.NotNil(t, service)
	assert.Equal(t, client, service.apiClient)
	assert.Empty(t, service.version)
	assert.False(t, service.useNamedConfigs)
}

func TestVersion(t *testing.T) {
	service := NewSystemService(nil)

	// Test empty version
	assert.Empty(t, service.Version())

	// Test with version set
	service.version = "2.5.0"
	assert.Equal(t, "2.5.0", service.Version())
}

func TestUseNamedConfigs(t *testing.T) {
	service := NewSystemService(nil)
	assert.False(t, service.useNamedConfigs)

	service.UseNamedConfigs()
	assert.True(t, service.useNamedConfigs)
}

func TestVersionCheck(t *testing.T) {
	service := NewSystemService(nil)
	ctx := context.Background()

	t.Run("empty constraint", func(t *testing.T) {
		compatible, err := service.VersionCheck(ctx, "")
		assert.False(t, compatible)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "constraint string cannot be empty")
	})

	t.Run("unknown version", func(t *testing.T) {
		compatible, err := service.VersionCheck(ctx, ">=2.4.0")
		assert.False(t, compatible)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version unknown")
	})

	t.Run("valid version and constraint", func(t *testing.T) {
		service.version = "2.5.0"
		compatible, err := service.VersionCheck(ctx, ">=2.4.0")
		assert.NoError(t, err)
		assert.True(t, compatible)
	})

	t.Run("invalid constraint", func(t *testing.T) {
		service.version = "2.5.0"
		compatible, err := service.VersionCheck(ctx, "invalid-constraint")
		assert.False(t, compatible)
		assert.Error(t, err)
	})
}

func TestCheckVersionConstraint(t *testing.T) {
	service := NewSystemService(nil)

	tests := []struct {
		name        string
		version     string
		constraint  string
		expectError bool
		expected    bool
	}{
		{
			name:        "valid version satisfies constraint",
			version:     "2.5.0",
			constraint:  ">=2.4.0",
			expectError: false,
			expected:    true,
		},
		{
			name:        "valid version does not satisfy constraint",
			version:     "2.3.0",
			constraint:  ">=2.4.0",
			expectError: false,
			expected:    false,
		},
		{
			name:        "dev version",
			version:     "2.5.0-dev0+build.abc123",
			constraint:  ">=2.4.0",
			expectError: false,
			expected:    true,
		},
		{
			name:        "invalid version format",
			version:     "invalid",
			constraint:  ">=2.4.0",
			expectError: true,
			expected:    false,
		},
		{
			name:        "invalid constraint",
			version:     "2.5.0",
			constraint:  "invalid",
			expectError: true,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.checkVersionConstraint(tt.version, tt.constraint)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestVersionError(t *testing.T) {
	err := versionError("1.0.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server not compatible")
	assert.Contains(t, err.Error(), ">=2.4.0,<3.0.0")
	assert.Contains(t, err.Error(), "1.0.0")
}

func TestSystemReadyErrors(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		statusCode    int
		expectError   bool
		errorContains string
	}{
		{
			name:          "system not ready",
			responseBody:  `{"version": "2.5.0", "ready": false}`,
			statusCode:    http.StatusOK,
			expectError:   true,
			errorContains: "system not ready",
		},
		{
			name:          "incompatible version",
			responseBody:  `{"version": "1.0.0", "ready": true}`,
			statusCode:    http.StatusOK,
			expectError:   true,
			errorContains: "server not compatible",
		},
		{
			name:          "api error",
			responseBody:  "",
			statusCode:    http.StatusInternalServerError,
			expectError:   true,
			errorContains: "get system info",
		},
		{
			name:          "malformed json",
			responseBody:  `{"invalid": json}`,
			statusCode:    http.StatusOK,
			expectError:   true,
			errorContains: "get system info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody)) //nolint:errcheck
				}
			}))
			defer server.Close()

			client := api.New(server.URL, api.WithHTTPClient(&http.Client{}))
			service := NewSystemService(client)
			ctx := context.Background()

			err := service.Ready(ctx)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSystemServiceIntegration(t *testing.T) {
	// Test the full integration with a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the correct endpoint is called
		assert.Equal(t, "/api/v0/system_information", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"version": "2.5.0", "ready": true}`)) //nolint:errcheck
	}))
	defer server.Close()

	client := api.New(server.URL, api.WithHTTPClient(&http.Client{}))
	service := NewSystemService(client)
	ctx := context.Background()

	// Test Ready() method
	err := service.Ready(ctx)
	assert.NoError(t, err)

	// Verify version is set
	assert.Equal(t, "2.5.0", service.Version())

	// Test VersionCheck
	compatible, err := service.VersionCheck(ctx, ">=2.4.0")
	assert.NoError(t, err)
	assert.True(t, compatible)

	// Test UseNamedConfigs
	service.UseNamedConfigs()
	assert.True(t, service.useNamedConfigs)
}

// Test version constraint logic through public interface
func TestSystemService_VersionConstraints(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test focuses on version constraint logic")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	service := NewSystemService(client)

	// Set version manually for testing
	service.version = "2.5.0"

	// Test valid version constraints
	compatible, err := service.VersionCheck(context.Background(), ">=2.4.0")
	assert.NoError(t, err)
	assert.True(t, compatible)

	compatible, err = service.VersionCheck(context.Background(), ">=2.6.0")
	assert.NoError(t, err)
	assert.False(t, compatible)

	compatible, err = service.VersionCheck(context.Background(), ">=2.4.0,<3.0.0")
	assert.NoError(t, err)
	assert.True(t, compatible)

	// Test invalid constraint
	_, err = service.VersionCheck(context.Background(), "invalid")
	assert.Error(t, err)

	// Test empty constraint
	_, err = service.VersionCheck(context.Background(), "")
	assert.Error(t, err)
}

func TestSystemService_Ready_Success(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock successful system info response
	httpmock.RegisterResponder("GET", "https://mock/api/v0/system_information",
		httpmock.NewStringResponder(200, `{
			"version": "2.5.0",
			"ready": true
		}`))

	service := NewSystemService(client)
	ctx := context.Background()

	err := service.Ready(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "2.5.0", service.Version())
}

func TestSystemService_Ready_SystemNotReady(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock system not ready response
	httpmock.RegisterResponder("GET", "https://mock/api/v0/system_information",
		httpmock.NewStringResponder(200, `{
			"version": "2.5.0",
			"ready": false
		}`))

	service := NewSystemService(client)
	ctx := context.Background()

	err := service.Ready(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not ready")
}

func TestSystemService_Ready_IncompatibleVersion(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock incompatible version response
	httpmock.RegisterResponder("GET", "https://mock/api/v0/system_information",
		httpmock.NewStringResponder(200, `{
			"version": "1.5.0",
			"ready": true
		}`))

	service := NewSystemService(client)
	ctx := context.Background()

	err := service.Ready(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not compatible")
}

func TestSystemService_Ready_APIError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock API error response
	httpmock.RegisterResponder("GET", "https://mock/api/v0/system_information",
		httpmock.NewStringResponder(500, `{"error": "Internal server error"}`))

	service := NewSystemService(client)
	ctx := context.Background()

	err := service.Ready(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get system info")
}
