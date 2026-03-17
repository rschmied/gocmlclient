package httputil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildRequest(t *testing.T) {
	baseURL := "https://api.example.com"

	tests := []struct {
		name        string
		method      string
		endpoint    string
		query       map[string]string
		body        any
		expectURL   string
		expectBody  string
		expectError bool
	}{
		{
			name:      "GET request without query",
			method:    "GET",
			endpoint:  "/users",
			expectURL: "https://api.example.com/users",
		},
		{
			name:     "GET request with query parameters",
			method:   "GET",
			endpoint: "/users",
			query: map[string]string{
				"limit":  "10",
				"offset": "20",
			},
			expectURL: "https://api.example.com/users?limit=10&offset=20",
		},
		{
			name:     "POST request with JSON body",
			method:   "POST",
			endpoint: "/users",
			body: map[string]string{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			expectURL:  "https://api.example.com/users",
			expectBody: `{"email":"john@example.com","name":"John Doe"}`,
		},
		{
			name:       "POST request with string body",
			method:     "POST",
			endpoint:   "/logs",
			body:       "log message",
			expectURL:  "https://api.example.com/logs",
			expectBody: "log message",
		},
		{
			name:       "PUT request with byte slice body",
			method:     "PUT",
			endpoint:   "/files",
			body:       []byte("file content"),
			expectURL:  "https://api.example.com/files",
			expectBody: "file content",
		},
		{
			name:       "PATCH request with buffer body",
			method:     "PATCH",
			endpoint:   "/data",
			body:       bytes.NewBufferString("buffered data"),
			expectURL:  "https://api.example.com/data",
			expectBody: "buffered data",
		},
		{
			name:      "DELETE request",
			method:    "DELETE",
			endpoint:  "/users/123",
			expectURL: "https://api.example.com/users/123",
		},
		{
			name:      "HEAD request",
			method:    "HEAD",
			endpoint:  "/health",
			expectURL: "https://api.example.com/health",
		},
		{
			name:      "endpoint with leading slash",
			method:    "GET",
			endpoint:  "/api/v1/status",
			expectURL: "https://api.example.com/api/v1/status",
		},
		{
			name:      "endpoint without leading slash",
			method:    "GET",
			endpoint:  "health",
			expectURL: "https://api.example.com/health",
		},
		{
			name:     "complex query parameters",
			method:   "GET",
			endpoint: "/search",
			query: map[string]string{
				"q":      "golang testing",
				"sort":   "relevance",
				"filter": "active",
			},
			expectURL: "https://api.example.com/search?filter=active&q=golang+testing&sort=relevance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testURL := baseURL
			if tt.expectError {
				testURL = "invalid-url"
			}

			req, err := BuildRequest(ctx, testURL, tt.method, tt.endpoint, tt.query, tt.body)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check method
			if req.Method != tt.method {
				t.Errorf("expected method %s, got %s", tt.method, req.Method)
			}

			// Check URL
			if req.URL.String() != tt.expectURL {
				t.Errorf("expected URL %s, got %s", tt.expectURL, req.URL.String())
			}

			// Check body
			if tt.expectBody != "" {
				bodyBytes, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("failed to read request body: %v", err)
				}
				if string(bodyBytes) != tt.expectBody {
					t.Errorf("expected body %q, got %q", tt.expectBody, string(bodyBytes))
				}
			}

			// Check Content-Type header for requests with body
			if tt.body != nil {
				contentType := req.Header.Get("Content-Type")
				if contentType != ContentTypeJSON {
					t.Errorf("expected Content-Type application/json, got %s", contentType)
				}
			}
		})
	}
}

func TestBuildRequestWithReaderBody(t *testing.T) {
	ctx := context.Background()
	baseURL := "https://api.example.com"

	reader := strings.NewReader("reader content")
	req, err := BuildRequest(ctx, baseURL, "POST", "/upload", nil, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read request body: %v", err)
	}

	if string(bodyBytes) != "reader content" {
		t.Errorf("expected body %q, got %q", "reader content", string(bodyBytes))
	}
}

func TestBuildRequestContext(t *testing.T) {
	type myKey string
	const testKey myKey = "test-key"
	ctx := context.WithValue(context.Background(), testKey, "test-value")
	baseURL := "https://api.example.com"

	req, err := BuildRequest(ctx, baseURL, "GET", "/test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check if context value is preserved (contexts may not be identical objects)
	if req.Context().Value(testKey) != "test-value" {
		t.Error("context value not preserved")
	}

	// Also verify the context has the expected value
	if req.Context() == nil {
		t.Error("request context is nil")
	}
}

func TestBuildRequestNilBody(t *testing.T) {
	ctx := context.Background()
	baseURL := "https://api.example.com"

	req, err := BuildRequest(ctx, baseURL, "GET", "/test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Body != nil {
		t.Error("expected nil body for nil input")
	}

	contentType := req.Header.Get("Content-Type")
	if contentType != "" {
		t.Errorf("expected no Content-Type header for nil body, got %s", contentType)
	}
}

func TestBuildRequestEmptyQuery(t *testing.T) {
	ctx := context.Background()
	baseURL := "https://api.example.com"

	req, err := BuildRequest(ctx, baseURL, "GET", "/test", map[string]string{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.URL.RawQuery != "" {
		t.Errorf("expected no query string for empty query map, got %s", req.URL.RawQuery)
	}
}

func TestBuildRequestURLConstruction(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		endpoint string
		expected string
	}{
		{
			name:     "base URL with path",
			baseURL:  "https://api.example.com/v1",
			endpoint: "/users",
			expected: "https://api.example.com/v1/users",
		},
		{
			name:     "base URL without trailing slash",
			baseURL:  "https://api.example.com",
			endpoint: "/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "endpoint without leading slash",
			baseURL:  "https://api.example.com",
			endpoint: "users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "both with slashes",
			baseURL:  "https://api.example.com/",
			endpoint: "/users",
			expected: "https://api.example.com/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req, err := BuildRequest(ctx, tt.baseURL, "GET", tt.endpoint, nil, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if req.URL.String() != tt.expected {
				t.Errorf("expected URL %s, got %s", tt.expected, req.URL.String())
			}
		})
	}
}

func TestBuildRequestQueryParameterEncoding(t *testing.T) {
	ctx := context.Background()
	baseURL := "https://api.example.com"

	query := map[string]string{
		"search": "golang & rust",
		"tags":   "web,api",
		"active": "true",
	}

	req, err := BuildRequest(ctx, baseURL, "GET", "/search", query, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsedQuery, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	if parsedQuery.Get("search") != "golang & rust" {
		t.Errorf("expected search param 'golang & rust', got %s", parsedQuery.Get("search"))
	}
	if parsedQuery.Get("tags") != "web,api" {
		t.Errorf("expected tags param 'web,api', got %s", parsedQuery.Get("tags"))
	}
	if parsedQuery.Get("active") != "true" {
		t.Errorf("expected active param 'true', got %s", parsedQuery.Get("active"))
	}
}

func TestBuildRequestComplexJSONBody(t *testing.T) {
	ctx := context.Background()
	baseURL := "https://api.example.com"

	type User struct {
		Name    string   `json:"name"`
		Age     int      `json:"age"`
		Active  bool     `json:"active"`
		Hobbies []string `json:"hobbies"`
	}

	user := User{
		Name:    "Alice",
		Age:     30,
		Active:  true,
		Hobbies: []string{"reading", "coding"},
	}

	req, err := BuildRequest(ctx, baseURL, "POST", "/users", nil, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read request body: %v", err)
	}

	var parsedUser User
	if err := json.Unmarshal(bodyBytes, &parsedUser); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if !reflect.DeepEqual(parsedUser, user) {
		t.Errorf("expected user %+v, got %+v", user, parsedUser)
	}
}

func TestBuildRequestNilQuery(t *testing.T) {
	ctx := context.Background()
	baseURL := "https://api.example.com"

	req, err := BuildRequest(ctx, baseURL, "GET", "/test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.URL.RawQuery != "" {
		t.Errorf("expected no query string for nil query map, got %s", req.URL.RawQuery)
	}
}

func BenchmarkBuildRequest(b *testing.B) {
	ctx := context.Background()
	baseURL := "https://api.example.com"

	body := map[string]string{
		"key":   "value",
		"test":  "data",
		"count": "42",
	}

	query := map[string]string{
		"limit": "10",
		"page":  "1",
	}

	for b.Loop() {
		_, err := BuildRequest(ctx, baseURL, "POST", "/benchmark", query, body)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkBuildRequestSimple(b *testing.B) {
	ctx := context.Background()
	baseURL := "https://api.example.com"

	for b.Loop() {
		_, err := BuildRequest(ctx, baseURL, "GET", "/simple", nil, nil)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestBuildRequestErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		baseURL     string
		method      string
		endpoint    string
		query       map[string]string
		body        any
		expectError bool
	}{
		{
			name:        "invalid base URL",
			ctx:         context.Background(),
			baseURL:     "://invalid-url",
			method:      "GET",
			endpoint:    "/test",
			expectError: true,
		},
		{
			name:        "invalid body marshaling",
			ctx:         context.Background(),
			baseURL:     "https://api.example.com",
			method:      "POST",
			endpoint:    "/test",
			body:        make(chan int), // unmarshalable
			expectError: true,
		},
		{
			name:        "invalid method",
			ctx:         context.Background(),
			baseURL:     "https://api.example.com",
			method:      "INVALID METHOD",
			endpoint:    "/test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := BuildRequest(tt.ctx, tt.baseURL, tt.method, tt.endpoint, tt.query, tt.body)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, req)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
			}
		})
	}
}
