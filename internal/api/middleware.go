package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	cmlerrors "github.com/rschmied/gocmlclient/pkg/errors"
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
					logger.Info("Request body",
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
			ctx := req.Context()

			getBody, err := makeGetBody(req)
			if err != nil {
				return nil, err
			}

			var lastErr error
			delay := policy.InitialDelay

			for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
				reqClone := req.Clone(ctx)
				if getBody != nil {
					body, err := getBody()
					if err != nil {
						return nil, err
					}
					reqClone.Body = body
				}

				res, err := next(reqClone)
				if err == nil {
					// Check if the response indicates a retryable error
					if !isRetryableStatus(res.StatusCode) {
						return res, nil
					}
					// Drain + close the response body before retrying (helps connection reuse)
					_ = drainAndClose(res.Body)
					lastErr = &HTTPError{StatusCode: res.StatusCode}
				} else {
					lastErr = err
				}

				// Don't wait after the last attempt
				if attempt < policy.MaxRetries {
					if !isRetryableError(lastErr) {
						// Non-retryable error, return immediately
						return res, lastErr
					}
					if err := sleepWithContext(ctx, delay); err != nil {
						return nil, err
					}
					delay = min(time.Duration(float64(delay)*policy.BackoffFactor), policy.MaxDelay)
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
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for HTTP errors
	httpErr := &HTTPError{}
	if errors.As(err, &httpErr) {
		return isRetryableStatus(httpErr.StatusCode)
	}

	// Check for TLS/certificate validation errors - these should NOT be retried
	if cmlerrors.IsTLSCertificateError(err) {
		return false
	}

	// Check for network/timeout errors
	// You might want to add more specific error type checks here
	// based on your error handling needs

	return true // Default to retrying for unknown errors
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}

	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func drainAndClose(r io.ReadCloser) error {
	if r == nil {
		return nil
	}
	io.Copy(io.Discard, r) //nolint:errcheck
	return r.Close()
}

func makeGetBody(req *http.Request) (func() (io.ReadCloser, error), error) {
	if req == nil {
		return nil, nil
	}
	if req.Body == nil {
		return nil, nil
	}
	if req.GetBody != nil {
		return req.GetBody, nil
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body.Close() //nolint:errcheck

	// Restore the body so other middleware/callers aren't surprised.
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	return func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(bodyBytes)), nil
	}, nil
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

// StatsMiddleware collects API call statistics
func StatsMiddleware(stats *Stats) Middleware {
	return func(next DoFunc) DoFunc {
		return func(req *http.Request) (*http.Response, error) {
			start := time.Now()

			res, err := next(req)
			duration := time.Since(start)

			// Extract endpoint from URL (remove base path)
			endpoint := strings.TrimPrefix(req.URL.Path, APIBasePath)

			if err != nil {
				// Record failed call
				stats.RecordCall(req.Method, endpoint, 0, duration)
				return res, err
			}

			// Record successful call
			stats.RecordCall(req.Method, endpoint, res.StatusCode, duration)

			return res, err
		}
	}
}
