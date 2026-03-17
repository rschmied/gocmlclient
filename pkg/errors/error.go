// Package errors provides error related functions and types
package errors

import (
	"crypto/x509"
	"errors"
	"fmt"
	"strings"
)

// Sentinel Errors
var (
	// System Errors
	ErrSystemNotReady       = errors.New("system not ready")
	ErrElementNotFound      = errors.New("element not found")
	ErrNoNamedConfigSupport = errors.New("backend does not support named configs")
	ErrTLSCertificate       = errors.New("TLS certificate validation failed")

	// API Errors
	ErrAPIRequestFailed = errors.New("API request failed")
	ErrAPIUnauthorized  = errors.New("API authentication failed")
	ErrAPINotFound      = errors.New("API resource not found")
	ErrAPIConflict      = errors.New("API resource conflict")
	ErrAPIServerError   = errors.New("API server error")

	// Validation Errors
	ErrValidationFailed = errors.New("validation failed")
	ErrInvalidInput     = errors.New("invalid input")
	ErrMissingRequired  = errors.New("missing required field")

	// Authentication Errors
	ErrAuthFailed   = errors.New("authentication failed")
	ErrTokenExpired = errors.New("authentication token expired")
	ErrTokenInvalid = errors.New("authentication token invalid")

	// Network/Connection Errors
	ErrConnectionFailed = errors.New("connection failed")
	ErrTimeout          = errors.New("operation timeout")
)

// APIError represents a structured API error response
type APIError struct {
	Operation  string `json:"operation"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Cause      error  `json:"-"`
}

func (e *APIError) Error() string {
	if e.Operation != "" && e.StatusCode > 0 {
		if e.Message != "" {
			return fmt.Sprintf("API %s failed (HTTP %d): %s", e.Operation, e.StatusCode, e.Message)
		}
		return fmt.Sprintf("API %s failed (HTTP %d)", e.Operation, e.StatusCode)
	}
	if e.StatusCode > 0 {
		if e.Message != "" {
			return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
		}
		return fmt.Sprintf("HTTP %d", e.StatusCode)
	}
	if e.Message != "" {
		return e.Message
	}
	return "API error"
}

func (e *APIError) Unwrap() error {
	return e.Cause
}

// ValidationError represents a structured validation error
type ValidationError struct {
	Field  string `json:"field"`
	Value  any    `json:"value,omitempty"`
	Reason string `json:"reason"`
	Cause  error  `json:"-"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field %q: %s", e.Field, e.Reason)
}

func (e *ValidationError) Unwrap() error {
	return e.Cause
}

// WrapTLSError wraps a TLS error with a clear sentinel value
func WrapTLSError(err error) error {
	return err
}

// IsTLSCertificateError checks if an error is a TLS/certificate validation error
func IsTLSCertificateError(err error) bool {
	if err == nil {
		return false
	}

	// Import needed: "crypto/x509"
	var (
		unknownAuthorityErr *x509.UnknownAuthorityError
		hostnameErr         *x509.HostnameError
		certInvalidErr      *x509.CertificateInvalidError
	)

	if errors.As(err, &unknownAuthorityErr) ||
		errors.As(err, &hostnameErr) ||
		errors.As(err, &certInvalidErr) {
		return true
	}

	// Also check for string-based matching as fallback
	errStr := err.Error()
	return strings.Contains(errStr, "x509:") ||
		strings.Contains(errStr, "certificate") ||
		strings.Contains(errStr, "TLS handshake")
}

// Error Constructor Functions

// Wrap creates a standardized wrapped error
func Wrap(err error, operation string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}

// Wrapf creates a formatted wrapped error
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", msg, err)
}

// NewAPIError creates a structured API error
func NewAPIError(operation string, statusCode int, message string, cause error) *APIError {
	return &APIError{
		Operation:  operation,
		StatusCode: statusCode,
		Message:    message,
		Cause:      cause,
	}
}

// NewValidationError creates a structured validation error
func NewValidationError(field string, value any, reason string, cause error) *ValidationError {
	return &ValidationError{
		Field:  field,
		Value:  value,
		Reason: reason,
		Cause:  cause,
	}
}

// Error Checking Helpers

// IsNotFound checks if error indicates resource not found
func IsNotFound(err error) bool {
	if errors.Is(err, ErrAPINotFound) || errors.Is(err, ErrElementNotFound) {
		return true
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}

	return false
}

// IsUnauthorized checks if error indicates authentication failure
func IsUnauthorized(err error) bool {
	if errors.Is(err, ErrAPIUnauthorized) || errors.Is(err, ErrAuthFailed) {
		return true
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401 || apiErr.StatusCode == 403
	}

	return false
}

// IsTemporary checks if error is temporary and operation can be retried
func IsTemporary(err error) bool {
	if errors.Is(err, ErrTimeout) || errors.Is(err, ErrConnectionFailed) {
		return true
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 500
	}

	return false
}
