package api

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/rschmied/gocmlclient/internal/httputil"
)

// mockLogger captures log entries for testing
type mockLogger struct {
	entries []logEntry
}

type logEntry struct {
	level   slog.Level
	message string
	attrs   map[string]any
}

func (m *mockLogger) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (m *mockLogger) Handle(ctx context.Context, r slog.Record) error {
	entry := logEntry{
		level:   r.Level,
		message: r.Message,
		attrs:   make(map[string]any),
	}

	r.Attrs(func(a slog.Attr) bool {
		entry.attrs[a.Key] = a.Value.Any()
		return true
	})

	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return m
}

func (m *mockLogger) WithGroup(name string) slog.Handler {
	return m
}

func newMockLogger() *mockLogger {
	return &mockLogger{entries: make([]logEntry, 0)}
}

func (m *mockLogger) findEntry(level slog.Level, message string) *logEntry {
	for _, entry := range m.entries {
		if entry.level == level && entry.message == message {
			return &entry
		}
	}
	return nil
}

func (m *mockLogger) hasEntry(level slog.Level, message string) bool {
	return m.findEntry(level, message) != nil
}

// func (m *mockLogger) countEntries(level slog.Level, message string) int {
// 	count := 0
// 	for _, entry := range m.entries {
// 		if entry.level == level && entry.message == message {
// 			count++
// 		}
// 	}
// 	return count
// }

// TestUserAgentMiddleware tests the UserAgentMiddleware
func TestUserAgentMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		userAgent  string
		expectedUA string
	}{
		{
			name:       "standard user agent",
			userAgent:  "test-agent/1.0",
			expectedUA: "test-agent/1.0",
		},
		{
			name:       "empty user agent",
			userAgent:  "",
			expectedUA: "",
		},
		{
			name:       "complex user agent",
			userAgent:  "MyApp/1.2.3 (Linux; x86_64) Go/1.21",
			expectedUA: "MyApp/1.2.3 (Linux; x86_64) Go/1.21",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock next function that captures the request
			var capturedReq *http.Request
			next := func(req *http.Request) (*http.Response, error) {
				capturedReq = req
				return &http.Response{StatusCode: 200}, nil
			}

			// Create the middleware
			middleware := UserAgentMiddleware(tt.userAgent)
			wrappedNext := middleware(next)

			// Create a test request
			req, _ := http.NewRequest("GET", "http://example.com", nil)

			// Execute the middleware
			_, err := wrappedNext(req)
			if err != nil {
				t.Fatalf("middleware execution failed: %v", err)
			}

			// Verify the User-Agent header was set
			actualUA := capturedReq.Header.Get("User-Agent")
			if actualUA != tt.expectedUA {
				t.Errorf("expected User-Agent %q, got %q", tt.expectedUA, actualUA)
			}
		})
	}
}

// TestLoggingMiddleware tests the LoggingMiddleware
func TestLoggingMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		responseStatus int
		expectError    bool
		expectedLogs   []struct {
			level    slog.Level
			message  string
			hasAttrs bool
		}
	}{
		{
			name:           "successful request",
			method:         "GET",
			url:            "http://example.com/api/test",
			responseStatus: 200,
			expectError:    false,
			expectedLogs: []struct {
				level    slog.Level
				message  string
				hasAttrs bool
			}{
				{slog.LevelInfo, "HTTP request", true},
				{slog.LevelInfo, "HTTP response", true},
			},
		},
		{
			name:           "failed request",
			method:         "POST",
			url:            "http://example.com/api/error",
			responseStatus: 500,
			expectError:    true,
			expectedLogs: []struct {
				level    slog.Level
				message  string
				hasAttrs bool
			}{
				{slog.LevelInfo, "HTTP request", true},
				{slog.LevelError, "HTTP request failed", true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock logger
			mockLog := newMockLogger()
			logger := slog.New(mockLog)

			// Create a mock next function
			next := func(req *http.Request) (*http.Response, error) {
				if tt.expectError {
					return nil, errors.New("mock error")
				}
				return &http.Response{
					StatusCode: tt.responseStatus,
					Header:     make(http.Header),
				}, nil
			}

			// Create the middleware
			middleware := LoggingMiddleware(logger)
			wrappedNext := middleware(next)

			// Create a test request
			req, _ := http.NewRequest(tt.method, tt.url, nil)
			req.Header.Set("Content-Type", httputil.ContentTypeJSON)

			// Execute the middleware
			_, err := wrappedNext(req)

			// Verify error expectation
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify log entries
			for _, expectedLog := range tt.expectedLogs {
				if !mockLog.hasEntry(expectedLog.level, expectedLog.message) {
					t.Errorf("expected log entry: level=%s, message=%s", expectedLog.level, expectedLog.message)
				}
			}

			// Verify request log contains expected attributes
			if reqLog := mockLog.findEntry(slog.LevelInfo, "HTTP request"); reqLog != nil {
				if reqLog.attrs["method"] != tt.method {
					t.Errorf("expected method %s in request log, got %v", tt.method, reqLog.attrs["method"])
				}
				if reqLog.attrs["url"] != tt.url {
					t.Errorf("expected url %s in request log, got %v", tt.url, reqLog.attrs["url"])
				}
			}
		})
	}
}

// TestLogRequestBodyMiddleware tests the LogRequestBodyMiddleware
func TestLogRequestBodyMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		requestBody  string
		expectLog    bool
		expectedBody string
		expectError  bool
	}{
		{
			name:         "with request body",
			requestBody:  `{"key": "value"}`,
			expectLog:    true,
			expectedBody: `{"key": "value"}`,
			expectError:  false,
		},
		{
			name:         "without request body",
			expectLog:    false,
			expectedBody: "",
			expectError:  false,
		},
		{
			name:         "empty request body",
			requestBody:  "",
			expectLog:    false, // Empty body might not be logged
			expectedBody: "",
			expectError:  false,
		},
		{
			name:         "error reading body",
			requestBody:  "error", // Special case to trigger error
			expectLog:    false,
			expectedBody: "",
			expectError:  false, // Error is logged but not propagated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock logger
			mockLog := newMockLogger()
			logger := slog.New(mockLog)

			// Create a mock next function that captures the request
			var capturedReq *http.Request
			next := func(req *http.Request) (*http.Response, error) {
				capturedReq = req
				return &http.Response{StatusCode: 200}, nil
			}

			// Create the middleware
			middleware := LogRequestBodyMiddleware(logger)
			wrappedNext := middleware(next)

			// Create a test request
			var body io.Reader
			if tt.requestBody != "" {
				if tt.requestBody == "error" {
					// Create a reader that will return an error
					body = &errorReader{}
				} else {
					body = strings.NewReader(tt.requestBody)
				}
			}
			req, _ := http.NewRequest("POST", "http://example.com", body)

			// Execute the middleware
			_, err := wrappedNext(req)
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				// Verify error was logged
				if !mockLog.hasEntry(slog.LevelError, "Failed to read request body") {
					t.Error("expected error to be logged")
				}
				return
			}

			if err != nil {
				t.Fatalf("middleware execution failed: %v", err)
			}

			// For error reading body case, verify error was logged
			if tt.requestBody == "error" {
				if !mockLog.hasEntry(slog.LevelError, "Failed to read request body") {
					t.Error("expected error to be logged when reading body fails")
				}
			}

			// Verify request body was preserved (skip for error case)
			if tt.requestBody != "error" && capturedReq.Body != nil {
				bodyBytes, err := io.ReadAll(capturedReq.Body)
				if err != nil {
					t.Fatalf("failed to read captured body: %v", err)
				}
				if string(bodyBytes) != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, string(bodyBytes))
				}
			} else if tt.expectedBody != "" && tt.requestBody != "error" {
				t.Error("expected request body to be preserved")
			}

			// Verify logging
			if tt.expectLog {
				if !mockLog.hasEntry(slog.LevelInfo, "Request body") {
					t.Error("expected request body to be logged")
				}
				if bodyLog := mockLog.findEntry(slog.LevelDebug, "Request body"); bodyLog != nil {
					if bodyLog.attrs["body"] != tt.requestBody {
						t.Errorf("expected logged body %q, got %v", tt.requestBody, bodyLog.attrs["body"])
					}
				}
			} else {
				if mockLog.hasEntry(slog.LevelDebug, "Request body") {
					t.Error("did not expect request body to be logged")
				}
			}
		})
	}
}

// errorReader is an io.Reader that always returns an error
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("mock read error")
}

// TestRetryMiddleware tests the RetryMiddleware
func TestRetryMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		policy         RetryPolicy
		responseStatus int
		returnError    error
		expectedCalls  int
		expectError    bool
	}{
		{
			name: "successful request",
			policy: RetryPolicy{
				MaxRetries:    3,
				InitialDelay:  1 * time.Millisecond,
				MaxDelay:      100 * time.Millisecond,
				BackoffFactor: 2.0,
			},
			responseStatus: 200,
			expectedCalls:  1,
			expectError:    false,
		},
		{
			name: "retryable status code",
			policy: RetryPolicy{
				MaxRetries:    2,
				InitialDelay:  1 * time.Millisecond,
				MaxDelay:      100 * time.Millisecond,
				BackoffFactor: 2.0,
			},
			responseStatus: 500,
			expectedCalls:  3, // initial + 2 retries
			expectError:    true,
		},
		{
			name: "non-retryable status code",
			policy: RetryPolicy{
				MaxRetries:    3,
				InitialDelay:  1 * time.Millisecond,
				MaxDelay:      100 * time.Millisecond,
				BackoffFactor: 2.0,
			},
			responseStatus: 400,
			expectedCalls:  1,
			expectError:    false, // Should return response, not error
		},
		{
			name: "network error",
			policy: RetryPolicy{
				MaxRetries:    2,
				InitialDelay:  1 * time.Millisecond,
				MaxDelay:      100 * time.Millisecond,
				BackoffFactor: 2.0,
			},
			returnError:   errors.New("network error"),
			expectedCalls: 3, // initial + 2 retries
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0

			// Create a mock next function
			next := func(req *http.Request) (*http.Response, error) {
				callCount++
				if tt.returnError != nil {
					return nil, tt.returnError
				}
				return &http.Response{
					StatusCode: tt.responseStatus,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}

			// Create the middleware
			middleware := RetryMiddleware(tt.policy)
			wrappedNext := middleware(next)

			// Create a test request
			req, _ := http.NewRequest("GET", "http://example.com", nil)

			// Execute the middleware
			_, err := wrappedNext(req)

			// Verify call count
			if callCount != tt.expectedCalls {
				t.Errorf("expected %d calls, got %d", tt.expectedCalls, callCount)
			}

			// Verify error expectation
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestRetryMiddlewareWithBody tests retry middleware with request bodies
func TestRetryMiddlewareWithBody(t *testing.T) {
	callCount := 0
	originalBody := `{"test": "data"}`

	// Create a mock next function that verifies body is preserved
	next := func(req *http.Request) (*http.Response, error) {
		callCount++

		// Read the body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		// Verify body content
		if string(body) != originalBody {
			t.Errorf("body mismatch on call %d: expected %q, got %q", callCount, originalBody, string(body))
		}

		// Return retryable error for first two calls
		if callCount <= 2 {
			return &http.Response{
				StatusCode: 500,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}

		// Success on third call
		return &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}

	// Create retry policy
	policy := RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  1 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	// Create the middleware
	middleware := RetryMiddleware(policy)
	wrappedNext := middleware(next)

	// Create a test request with body
	req, _ := http.NewRequest("POST", "http://example.com", strings.NewReader(originalBody))

	// Execute the middleware
	_, err := wrappedNext(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all calls were made
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

// TestRetryMiddlewareBodyError tests retry middleware with body read errors
func TestRetryMiddlewareBodyError(t *testing.T) {
	callCount := 0

	// Create a mock next function that returns success on final attempt
	next := func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount < 3 { // First two attempts fail
			return &http.Response{
				StatusCode: 500,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}
		// Success on third attempt
		return &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}

	// Create retry policy
	policy := RetryPolicy{
		MaxRetries:    2,
		InitialDelay:  1 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	// Create the middleware
	middleware := RetryMiddleware(policy)
	wrappedNext := middleware(next)

	// Create a test request with error-prone body
	req, _ := http.NewRequest("POST", "http://example.com", &errorReader{})

	// Execute the middleware
	_, err := wrappedNext(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still make retry attempts despite body read error
	if callCount != 3 { // initial + 2 retries
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

// TestIsRetryableStatus tests the isRetryableStatus function
func TestIsRetryableStatus(t *testing.T) {
	tests := []struct {
		statusCode int
		retryable  bool
	}{
		{200, false},
		{400, false},
		{401, false},
		{404, false},
		{429, true}, // Too Many Requests
		{500, true}, // Internal Server Error
		{502, true}, // Bad Gateway
		{503, true}, // Service Unavailable
		{504, true}, // Gateway Timeout
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.statusCode), func(t *testing.T) {
			result := isRetryableStatus(tt.statusCode)
			if result != tt.retryable {
				t.Errorf("status %d: expected retryable=%v, got %v", tt.statusCode, tt.retryable, result)
			}
		})
	}
}

// TestIsRetryableError tests the isRetryableError function
func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"nil error", nil, false},
		{"retryable HTTP error", &HTTPError{StatusCode: 500}, true},
		{"non-retryable HTTP error", &HTTPError{StatusCode: 400}, false},
		{"generic error", errors.New("network timeout"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			if result != tt.retryable {
				t.Errorf("expected retryable=%v, got %v", tt.retryable, result)
			}
		})
	}
}

// TestDefaultRetryPolicy tests the DefaultRetryPolicy function
func TestDefaultRetryPolicy(t *testing.T) {
	policy := DefaultRetryPolicy()

	expected := RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
	}

	if policy != expected {
		t.Errorf("expected %+v, got %+v", expected, policy)
	}
}

// TestHTTPError tests the HTTPError type
func TestHTTPError(t *testing.T) {
	tests := []struct {
		statusCode  int
		expectedMsg string
	}{
		{400, "Bad Request"},
		{404, "Not Found"},
		{500, "Internal Server Error"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedMsg, func(t *testing.T) {
			err := &HTTPError{StatusCode: tt.statusCode}
			if err.Error() != tt.expectedMsg {
				t.Errorf("expected %q, got %q", tt.expectedMsg, err.Error())
			}
		})
	}
}

// TestMiddlewareIntegration tests multiple middlewares working together
func TestMiddlewareIntegration(t *testing.T) {
	// Create mock logger
	mockLog := newMockLogger()
	logger := slog.New(mockLog)

	// Create a chain of middlewares
	next := func(req *http.Request) (*http.Response, error) {
		// Verify User-Agent was set
		if req.Header.Get("User-Agent") != "test-agent" {
			t.Errorf("expected User-Agent 'test-agent', got %q", req.Header.Get("User-Agent"))
		}
		return &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
		}, nil
	}

	// Apply middlewares in reverse order (as they would be applied)
	middlewares := []Middleware{
		UserAgentMiddleware("test-agent"),
		LoggingMiddleware(logger),
		LogRequestBodyMiddleware(logger),
	}

	// Apply middlewares
	for i := len(middlewares) - 1; i >= 0; i-- {
		next = middlewares[i](next)
	}

	// Create test request
	req, _ := http.NewRequest("GET", "http://example.com", strings.NewReader(`{"test": true}`))

	// Execute the middleware chain
	_, err := next(req)
	if err != nil {
		t.Fatalf("middleware chain execution failed: %v", err)
	}

	// Verify logging occurred
	if !mockLog.hasEntry(slog.LevelInfo, "HTTP request") {
		t.Error("expected HTTP request to be logged")
	}
	if !mockLog.hasEntry(slog.LevelInfo, "HTTP response") {
		t.Error("expected HTTP response to be logged")
	}
	if !mockLog.hasEntry(slog.LevelInfo, "Request body") {
		t.Error("expected request body to be logged")
	}
}

// BenchmarkUserAgentMiddleware benchmarks the UserAgentMiddleware
func BenchmarkUserAgentMiddleware(b *testing.B) {
	next := func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200}, nil
	}

	middleware := UserAgentMiddleware("benchmark-agent")
	wrappedNext := middleware(next)

	req, _ := http.NewRequest("GET", "http://example.com", nil)

	for b.Loop() {
		wrappedNext(req)
	}
}

// BenchmarkLoggingMiddleware benchmarks the LoggingMiddleware
func BenchmarkLoggingMiddleware(b *testing.B) {
	mockLog := newMockLogger()
	logger := slog.New(mockLog)

	next := func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header)}, nil
	}

	middleware := LoggingMiddleware(logger)
	wrappedNext := middleware(next)

	req, _ := http.NewRequest("GET", "http://example.com", nil)

	for b.Loop() {
		wrappedNext(req)
	}
}

// BenchmarkLogRequestBodyMiddleware benchmarks the LogRequestBodyMiddleware
func BenchmarkLogRequestBodyMiddleware(b *testing.B) {
	mockLog := newMockLogger()
	logger := slog.New(mockLog)

	next := func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200}, nil
	}

	middleware := LogRequestBodyMiddleware(logger)
	wrappedNext := middleware(next)

	body := strings.NewReader(`{"benchmark": "data", "size": 100}`)
	req, _ := http.NewRequest("POST", "http://example.com", body)

	for b.Loop() {
		// Reset body for each iteration
		req.Body = io.NopCloser(strings.NewReader(`{"benchmark": "data", "size": 100}`))
		wrappedNext(req)
	}
}
