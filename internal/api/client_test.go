package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/rschmied/gocmlclient/internal/httputil"
	gocmlerrors "github.com/rschmied/gocmlclient/pkg/errors"
)

func TestNew(t *testing.T) {
	client := &http.Client{Timeout: 30 * time.Second}

	apiClient := New("https://api.example.com",
		WithHTTPClient(client),
		WithMiddlewares(UserAgentMiddleware("test-agent")),
	)

	if apiClient.baseURL != "https://api.example.com" {
		t.Errorf("expected baseURL 'https://api.example.com', got %s", apiClient.baseURL)
	}

	if apiClient.do == nil {
		t.Error("expected do function to be set")
	}
}

func TestRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected method POST, got %s", r.Method)
		}
		if r.URL.Path != "/test" {
			t.Errorf("expected path /test, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != httputil.ContentTypeJSON {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("X-CML-CLIENT") != "gocmlclient" {
			t.Errorf("expected X-CML-CLIENT header 'gocmlclient', got %q", r.Header.Get("X-CML-CLIENT"))
		}
		if r.Header.Get("X-Client-UUID") != "test-uuid" {
			t.Errorf("expected X-Client-UUID header 'test-uuid', got %q", r.Header.Get("X-Client-UUID"))
		}
		if r.Header.Get("X-CML-CLIENT-VERSION") != "test-version" {
			t.Errorf("expected X-CML-CLIENT-VERSION header 'test-version', got %q", r.Header.Get("X-CML-CLIENT-VERSION"))
		}

		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		expectedBody := `{"message":"test"}`
		if string(body) != expectedBody {
			t.Errorf("expected body %s, got %s", expectedBody, string(body))
		}

		// Send response
		w.Header().Set("Content-Type", httputil.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"success"}`))
	}))
	defer server.Close()

	// Create API client
	client := New(server.URL, WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))
	client.SetClientInfo("gocmlclient", "test-uuid", "test-version")

	// Make request
	ctx := context.Background()
	resp, err := client.Request(ctx, "POST", "/test", nil, map[string]string{"message": "test"})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	expectedResponse := `{"result":"success"}`
	if string(body) != expectedResponse {
		t.Errorf("expected response %s, got %s", expectedResponse, string(body))
	}
}

func TestDoJSON(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		endpoint       string
		requestBody    any
		responseBody   string
		statusCode     int
		expectError    bool
		expectedResult any
	}{
		{
			name:         "successful GET",
			method:       "GET",
			endpoint:     "/users",
			responseBody: `{"users":[{"id":1,"name":"John"}]}`,
			statusCode:   http.StatusOK,
			expectedResult: map[string]any{
				"users": []any{
					map[string]any{"id": float64(1), "name": "John"},
				},
			},
		},
		{
			name:         "successful POST",
			method:       "POST",
			endpoint:     "/users",
			requestBody:  map[string]string{"name": "Jane"},
			responseBody: `{"id":2,"name":"Jane"}`,
			statusCode:   http.StatusCreated,
			expectedResult: map[string]any{
				"id":   float64(2),
				"name": "Jane",
			},
		},
		{
			name:        "server error",
			method:      "GET",
			endpoint:    "/error",
			statusCode:  http.StatusInternalServerError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			client := New(server.URL, WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))

			ctx := context.Background()
			var result any
			err := client.doJSON(ctx, tt.method, tt.endpoint, nil, tt.requestBody, &result)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectedResult != nil {
				// Simple comparison for test purposes
				resultJSON, _ := json.Marshal(result)
				expectedJSON, _ := json.Marshal(tt.expectedResult)
				if string(resultJSON) != string(expectedJSON) {
					t.Errorf("expected result %s, got %s", string(expectedJSON), string(resultJSON))
				}
			}
		})
	}
}

func TestGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":"test"}`))
	}))
	defer server.Close()

	client := New(server.URL, WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))

	var result map[string]string
	err := client.GetJSON(context.Background(), "/test", nil, &result)
	if err != nil {
		t.Fatalf("GetJSON failed: %v", err)
	}

	expected := map[string]string{"data": "test"}
	if result["data"] != expected["data"] {
		t.Errorf("expected data 'test', got %s", result["data"])
	}
}

func TestPostJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		var requestData map[string]string
		if err := json.Unmarshal(body, &requestData); err != nil {
			t.Fatalf("failed to unmarshal request: %v", err)
		}

		if requestData["name"] != "test" {
			t.Errorf("expected name 'test', got %s", requestData["name"])
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":123,"name":"test"}`))
	}))
	defer server.Close()

	client := New(server.URL, WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))

	requestData := map[string]string{"name": "test"}
	var result map[string]any
	err := client.PostJSON(context.Background(), "/users", nil, requestData, &result)
	if err != nil {
		t.Fatalf("PostJSON failed: %v", err)
	}

	if result["id"] != float64(123) {
		t.Errorf("expected id 123, got %v", result["id"])
	}
}

func TestPutJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(server.URL, WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))

	err := client.PutJSON(context.Background(), "/users/123", map[string]string{"name": "updated"})
	if err != nil {
		t.Fatalf("PutJSON failed: %v", err)
	}
}

func TestPatchJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"updated":true}`))
	}))
	defer server.Close()

	client := New(server.URL, WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))

	var result map[string]bool
	err := client.PatchJSON(context.Background(), "/users/123", nil, map[string]string{"name": "patched"}, &result)
	if err != nil {
		t.Fatalf("PatchJSON failed: %v", err)
	}

	if !result["updated"] {
		t.Error("expected updated to be true")
	}
}

func TestDeleteJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(server.URL, WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))

	err := client.DeleteJSON(context.Background(), "/users/123", nil)
	if err != nil {
		t.Fatalf("DeleteJSON failed: %v", err)
	}
}

func TestHandleHTTPError(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		expectedMsg  string
	}{
		{
			name:         "structured JSON error",
			statusCode:   http.StatusBadRequest,
			responseBody: `{"message":"Invalid request","details":{"field":"name"}}`,
			expectedMsg:  "API request failed (HTTP 400): Invalid request",
		},
		{
			name:         "plain text error",
			statusCode:   http.StatusNotFound,
			responseBody: "Resource not found",
			expectedMsg:  "API request failed (HTTP 404): Resource not found",
		},
		{
			name:         "empty response",
			statusCode:   http.StatusInternalServerError,
			responseBody: "",
			expectedMsg:  "API request failed (HTTP 500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			client := New(server.URL, WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))

			var result any
			err := client.GetJSON(context.Background(), "/error", nil, &result)
			if err == nil {
				t.Fatal("expected error but got none")
			}

			if !strings.Contains(err.Error(), tt.expectedMsg) {
				t.Errorf("expected error message to contain %q, got %q", tt.expectedMsg, err.Error())
			}

			// Check if it's our APIError type
			if apiErr, ok := err.(*gocmlerrors.APIError); ok {
				if apiErr.StatusCode != tt.statusCode {
					t.Errorf("expected status code %d, got %d", tt.statusCode, apiErr.StatusCode)
				}
			}
		})
	}
}

func TestWrapConnectionError(t *testing.T) {
	client := New("https://invalid-url", WithHTTPClient(&http.Client{Timeout: 10 * time.Nanosecond}))

	_, err := client.Request(context.Background(), "GET", "/test", nil, nil)
	if err == nil {
		t.Fatal("expected connection error but got none")
	}

	// Should be wrapped as our domain error
	if err.Error() != "system not ready" {
		t.Errorf("expected 'system not ready' error, got %q", err.Error())
	}
}

// TestWrapConnectionErrorSpecific tests specific error types
func TestWrapConnectionErrorSpecific(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "connection refused",
			err:      syscall.ECONNREFUSED,
			expected: "system not ready",
		},
		{
			name:     "host unreachable",
			err:      syscall.EHOSTUNREACH,
			expected: "system not ready",
		},
		{
			name:     "network unreachable",
			err:      syscall.ENETUNREACH,
			expected: "system not ready",
		},
		{
			name:     "generic error",
			err:      errors.New("generic error"),
			expected: "generic error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.wrapConnectionError(tt.err)
			if result.Error() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.Error())
			}
		})
	}
}

// TestRequestError tests the Request function error path
func TestRequestError(t *testing.T) {
	client := New("https://api.example.com", WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))

	// Test with invalid body that should cause httputil.BuildRequest to fail
	// We'll use a channel as body which should cause an error
	ctx := context.Background()
	_, err := client.Request(ctx, "POST", "/test", nil, make(chan int))
	if err == nil {
		t.Error("expected error when using invalid body type, got nil")
	}
}

// TestDoJSONNetworkError tests doJSON with network errors
func TestDoJSONNetworkError(t *testing.T) {
	// Create a client with an invalid URL that will cause network errors
	client := New("https://invalid-url-that-does-not-exist-12345.com",
		WithHTTPClient(&http.Client{Timeout: 1 * time.Millisecond}), // Very short timeout
	)

	ctx := context.Background()
	var result any
	err := client.doJSON(ctx, "GET", "/test", nil, nil, &result)
	if err == nil {
		t.Error("expected network error, got nil")
	}
}

// TestDoJSONMalformedResponse tests doJSON with malformed JSON response
func TestDoJSONMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Send malformed JSON that will fail to decode
		w.Write([]byte(`{"invalid": json}`))
	}))
	defer server.Close()

	client := New(server.URL, WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))

	ctx := context.Background()
	var result map[string]string
	err := client.doJSON(ctx, "GET", "/test", nil, nil, &result)
	if err == nil {
		t.Error("expected JSON decode error, got nil")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Errorf("expected decode error message, got: %v", err)
	}
}

// TestAPIErrorEmptyFields tests APIError.Error() with empty Message
func TestAPIErrorEmptyFields(t *testing.T) {
	apiErr := &gocmlerrors.APIError{
		StatusCode: 500,
		Message:    "", // Empty message
	}

	expected := "HTTP 500"
	if apiErr.Error() != expected {
		t.Errorf("expected %q, got %q", expected, apiErr.Error())
	}
}

// TestAPIErrorWithRawBody tests APIError.Error() with no Message
func TestAPIErrorWithRawBody(t *testing.T) {
	apiErr := &gocmlerrors.APIError{
		StatusCode: 404,
		Message:    "", // Empty message
	}

	expected := "HTTP 404"
	if apiErr.Error() != expected {
		t.Errorf("expected %q, got %q", expected, apiErr.Error())
	}
}

// TestHandleHTTPErrorBodyReadError tests handleHTTPError with body read error
func TestHandleHTTPErrorBodyReadError(t *testing.T) {
	// Create a mock response with a body reader that returns an error
	resp := &http.Response{
		StatusCode: 500,
		Header:     make(http.Header),
		Body:       &errorReader{}, // This will return an error when Read is called
	}

	client := &Client{}
	err := client.handleHTTPError(resp)

	if err == nil {
		t.Error("expected error when reading response body fails, got nil")
	}

	// Check that it's an APIError with the expected message
	if apiErr, ok := err.(*gocmlerrors.APIError); ok {
		if apiErr.StatusCode != 500 {
			t.Errorf("expected status code 500, got %d", apiErr.StatusCode)
		}
		if !strings.Contains(apiErr.Message, "failed to read error response") {
			t.Errorf("expected error message about reading response, got %q", apiErr.Message)
		}
	} else {
		t.Errorf("expected *APIError, got %T", err)
	}
}

func BenchmarkRequest(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := New(server.URL, WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))

	ctx := context.Background()

	for b.Loop() {
		resp, err := client.Request(ctx, "GET", "/bench", nil, nil)
		if err != nil {
			b.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()
	}
}

func BenchmarkGetJSON(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":"benchmark"}`))
	}))
	defer server.Close()

	client := New(server.URL, WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))

	ctx := context.Background()
	var result map[string]string

	for b.Loop() {
		err := client.GetJSON(ctx, "/bench", nil, &result)
		if err != nil {
			b.Fatalf("GetJSON failed: %v", err)
		}
	}
}

func TestStats(t *testing.T) {
	// Test server that returns different status codes
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch r.URL.Path {
		case "/api/v0/ok":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		case "/api/v0/notfound":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"not found"}`))
		case "/api/v0/error":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"server error"}`))
		default:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"default"}`))
		}
	}))
	defer server.Close()

	// Create client with stats enabled
	client := New(server.URL,
		WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
		WithStats(),
	)

	ctx := context.Background()

	// Make some requests
	client.GetJSON(ctx, "/ok", nil, nil)
	client.GetJSON(ctx, "/notfound", nil, nil)
	client.GetJSON(ctx, "/ok", nil, nil)
	client.GetJSON(ctx, "/error", nil, nil)

	// Get stats
	stats := client.Stats()

	// Verify stats
	if stats.TotalCalls() != 4 {
		t.Errorf("expected 4 total calls, got %d", stats.TotalCalls())
	}

	if stats.CallsByMethod()["GET"] != 4 {
		t.Errorf("expected 4 GET calls, got %d", stats.CallsByMethod()["GET"])
	}

	if stats.StatusCounts()[200] != 2 {
		t.Errorf("expected 2 status 200 calls, got %d", stats.StatusCounts()[200])
	}

	if stats.StatusCounts()[404] != 1 {
		t.Errorf("expected 1 status 404 call, got %d", stats.StatusCounts()[404])
	}

	if stats.StatusCounts()[500] != 1 {
		t.Errorf("expected 1 status 500 call, got %d", stats.StatusCounts()[500])
	}

	if len(stats.EndpointGroups) != 3 {
		t.Errorf("expected 3 endpoint groups, got %d", len(stats.EndpointGroups))
	}

	// Test with stats disabled
	clientNoStats := New(server.URL, WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))
	clientNoStats.GetJSON(ctx, "/ok", nil, nil)
	statsEmpty := clientNoStats.Stats()

	if statsEmpty.TotalCalls() != 0 {
		t.Errorf("expected 0 calls when stats disabled, got %d", statsEmpty.TotalCalls())
	}
}

func TestNewWithOptions(t *testing.T) {
	// Test functional options
	client := New("https://api.example.com",
		WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
		WithStats(),
		WithMiddlewares(UserAgentMiddleware("test-agent")),
	)

	if client.baseURL != "https://api.example.com" {
		t.Errorf("expected baseURL 'https://api.example.com', got %s", client.baseURL)
	}

	if client.stats == nil {
		t.Error("expected stats to be enabled")
	}

	// Test that stats work with middleware
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "test-agent" {
			t.Errorf("expected User-Agent header 'test-agent', got %s", r.Header.Get("User-Agent"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	clientWithServer := New(server.URL,
		WithStats(),
		WithMiddlewares(UserAgentMiddleware("test-agent")),
	)
	ctx := context.Background()
	clientWithServer.GetJSON(ctx, "/test", nil, nil)

	stats := clientWithServer.Stats()
	if stats.TotalCalls() != 1 {
		t.Errorf("expected 1 call, got %d", stats.TotalCalls())
	}
}
