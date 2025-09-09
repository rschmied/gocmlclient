package api

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/rschmied/gocmlclient/pkg/errors"
)

// LoggingMiddleware logs HTTP requests and responses
func LoggingMiddleware(logger *slog.Logger) Middleware {
	return func(next DoFunc) DoFunc {
		return func(req *http.Request) (*http.Response, error) {
			start := time.Now()

			logger.Info("HTTP request",
				"method", req.Method,
				"url", req.URL.String(),
				"headers", req.Header,
			)

			res, err := next(req)
			duration := time.Since(start)

			if err != nil {
				logger.Error("HTTP request failed",
					"method", req.Method,
					"url", req.URL.String(),
					"duration", duration,
					"error", err,
				)
				return res, err
			}

			logger.Info("HTTP response",
				"method", req.Method,
				"url", req.URL.String(),
				"status", res.StatusCode,
				"duration", duration,
				"headers", res.Header,
			)

			return res, err
		}
	}
}

// LogRequestBodyMiddleware logs request bodies for debugging
func LogRequestBodyMiddleware(logger *slog.Logger) Middleware {
	return func(next DoFunc) DoFunc {
		return func(req *http.Request) (*http.Response, error) {
			if req.Body != nil {
				body, err := io.ReadAll(req.Body)
				if err != nil {
					logger.Error("Failed to read request body", "error", err)
				} else {
					logger.Debug("Request body",
						"method", req.Method,
						"url", req.URL.String(),
						"body", string(body),
						"length", len(body),
					)
					// Restore the body
					req.Body = io.NopCloser(bytes.NewReader(body))
				}
			}

			return next(req)
		}
	}
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// DefaultRetryPolicy returns a sensible default retry policy
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
	}
}

// RetryMiddleware implements retry logic with exponential backoff
func RetryMiddleware(policy RetryPolicy) Middleware {
	return func(next DoFunc) DoFunc {
		return func(req *http.Request) (*http.Response, error) {
			var lastErr error
			delay := policy.InitialDelay

			for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
				// Clone the request for retry attempts
				reqClone := req.Clone(req.Context())

				// If there's a body, we need to restore it for each attempt
				if req.Body != nil {
					body, err := io.ReadAll(req.Body)
					if err == nil {
						req.Body = io.NopCloser(bytes.NewReader(body))
						reqClone.Body = io.NopCloser(bytes.NewReader(body))
					}
				}

				res, err := next(reqClone)
				if err == nil {
					// Check if the response indicates a retryable error
					if !isRetryableStatus(res.StatusCode) {
						return res, nil
					}
					// Close the response body before retrying
					res.Body.Close()
					lastErr = &HTTPError{StatusCode: res.StatusCode}
				} else {
					lastErr = err
				}

				// Don't wait after the last attempt
				if attempt < policy.MaxRetries {
					if isRetryableError(lastErr) {
						time.Sleep(delay)
						delay = min(time.Duration(float64(delay)*policy.BackoffFactor), policy.MaxDelay)
					} else {
						// Non-retryable error, return immediately
						return res, lastErr
					}
				}
			}

			return nil, lastErr
		}
	}
}

// HTTPError represents an HTTP error
type HTTPError struct {
	StatusCode int
}

func (e *HTTPError) Error() string {
	return http.StatusText(e.StatusCode)
}

// isRetryableStatus checks if an HTTP status code is retryable
func isRetryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for HTTP errors
	if httpErr, ok := err.(*HTTPError); ok {
		return isRetryableStatus(httpErr.StatusCode)
	}

	// Check for TLS/certificate validation errors - these should NOT be retried
	if errors.IsTLSCertificateError(err) {
		return false
	}

	// Check for network/timeout errors
	// You might want to add more specific error type checks here
	// based on your error handling needs

	return true // Default to retrying for unknown errors
}

// UserAgentMiddleware adds a User-Agent header to requests
func UserAgentMiddleware(userAgent string) Middleware {
	return func(next DoFunc) DoFunc {
		return func(req *http.Request) (*http.Response, error) {
			req.Header.Set("User-Agent", userAgent)
			return next(req)
		}
	}
}
