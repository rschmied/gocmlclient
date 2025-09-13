package errors

import (
	"crypto/x509"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTLSCertificateError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name: "unknown authority error",
			err: &x509.UnknownAuthorityError{
				Cert: &x509.Certificate{},
			},
			expected: true,
		},
		{
			name: "hostname error",
			err: &x509.HostnameError{
				Certificate: &x509.Certificate{},
				Host:        "example.com",
			},
			expected: true,
		},
		{
			name: "certificate invalid error",
			err: &x509.CertificateInvalidError{
				Cert:   &x509.Certificate{},
				Reason: x509.Expired,
			},
			expected: true,
		},
		{
			name:     "string containing x509",
			err:      errors.New("x509: certificate signed by unknown authority"),
			expected: true,
		},
		{
			name:     "string containing certificate",
			err:      errors.New("certificate validation failed"),
			expected: true,
		},
		{
			name:     "string containing TLS handshake",
			err:      errors.New("TLS handshake failed"),
			expected: true,
		},
		{
			name:     "wrapped x509 error",
			err:      errors.New("connection failed: x509: certificate has expired"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTLSCertificateError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWrapTLSError(t *testing.T) {
	// WrapTLSError is currently a no-op, so it should return the same error
	originalErr := errors.New("test TLS error")
	result := WrapTLSError(originalErr)
	assert.Equal(t, originalErr, result)
}

func TestErrorVariables(t *testing.T) {
	// Test that error variables are properly defined and not nil
	assert.NotNil(t, ErrSystemNotReady)
	assert.NotNil(t, ErrElementNotFound)
	assert.NotNil(t, ErrNoNamedConfigSupport)
	assert.NotNil(t, ErrTLSCertificate)

	// Test error messages
	assert.Equal(t, "system not ready", ErrSystemNotReady.Error())
	assert.Equal(t, "element not found", ErrElementNotFound.Error())
	assert.Equal(t, "backend does not support named configs", ErrNoNamedConfigSupport.Error())
	assert.Equal(t, "TLS certificate validation failed", ErrTLSCertificate.Error())
}

func TestErrorUniqueness(t *testing.T) {
	// Test that all error variables are unique
	errors := []error{
		ErrSystemNotReady,
		ErrElementNotFound,
		ErrNoNamedConfigSupport,
		ErrTLSCertificate,
		ErrAPIRequestFailed,
		ErrAPIUnauthorized,
		ErrAPINotFound,
		ErrAPIConflict,
		ErrAPIServerError,
		ErrValidationFailed,
		ErrInvalidInput,
		ErrMissingRequired,
		ErrAuthFailed,
		ErrTokenExpired,
		ErrTokenInvalid,
		ErrConnectionFailed,
		ErrTimeout,
	}

	for i, err1 := range errors {
		for j, err2 := range errors {
			if i != j {
				assert.NotEqual(t, err1, err2, "error %d and %d should be different", i, j)
			}
		}
	}
}

func TestAPIError(t *testing.T) {
	tests := []struct {
		name     string
		apiErr   *APIError
		expected string
	}{
		{
			name: "full error with operation",
			apiErr: &APIError{
				Operation:  "create",
				StatusCode: 400,
				Message:    "invalid input",
				RawBody:    "detailed error",
			},
			expected: "API create failed (HTTP 400): invalid input",
		},
		{
			name: "error with status code only",
			apiErr: &APIError{
				StatusCode: 500,
			},
			expected: "HTTP 500",
		},
		{
			name: "error with message only",
			apiErr: &APIError{
				Message: "server error",
			},
			expected: "server error",
		},
		{
			name: "error with raw body only",
			apiErr: &APIError{
				RawBody: "internal server error",
			},
			expected: "internal server error",
		},
		{
			name:     "empty error",
			apiErr:   &APIError{},
			expected: "API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.apiErr.Error())
		})
	}
}

func TestAPIErrorUnwrap(t *testing.T) {
	cause := errors.New("original error")
	apiErr := &APIError{
		Operation:  "test",
		StatusCode: 500,
		Message:    "test error",
		Cause:      cause,
	}

	assert.Equal(t, cause, apiErr.Unwrap())
	assert.True(t, errors.Is(apiErr.Unwrap(), cause))
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		valErr   *ValidationError
		expected string
	}{
		{
			name: "full validation error",
			valErr: &ValidationError{
				Field:  "username",
				Value:  "",
				Reason: "cannot be empty",
			},
			expected: "validation failed for field \"username\": cannot be empty",
		},
		{
			name: "validation error without value",
			valErr: &ValidationError{
				Field:  "email",
				Reason: "invalid format",
			},
			expected: "validation failed for field \"email\": invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.valErr.Error())
		})
	}
}

func TestValidationErrorUnwrap(t *testing.T) {
	cause := errors.New("original validation error")
	valErr := &ValidationError{
		Field:  "password",
		Reason: "too short",
		Cause:  cause,
	}

	assert.Equal(t, cause, valErr.Unwrap())
	assert.True(t, errors.Is(valErr.Unwrap(), cause))
}

func TestNewAPIError(t *testing.T) {
	cause := errors.New("connection failed")
	apiErr := NewAPIError("login", 401, "invalid credentials", cause)

	assert.Equal(t, "login", apiErr.Operation)
	assert.Equal(t, 401, apiErr.StatusCode)
	assert.Equal(t, "invalid credentials", apiErr.Message)
	assert.Equal(t, cause, apiErr.Cause)
	assert.Equal(t, "API login failed (HTTP 401): invalid credentials", apiErr.Error())
}

func TestNewValidationError(t *testing.T) {
	cause := errors.New("value too long")
	valErr := NewValidationError("description", "very long text", "exceeds maximum length", cause)

	assert.Equal(t, "description", valErr.Field)
	assert.Equal(t, "very long text", valErr.Value)
	assert.Equal(t, "exceeds maximum length", valErr.Reason)
	assert.Equal(t, cause, valErr.Cause)
	assert.Equal(t, "validation failed for field \"description\": exceeds maximum length", valErr.Error())
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")

	t.Run("wrap with error", func(t *testing.T) {
		wrapped := Wrap(originalErr, "operation")
		assert.Equal(t, "operation: original error", wrapped.Error())
		assert.True(t, errors.Is(wrapped, originalErr))
	})

	t.Run("wrap with nil error", func(t *testing.T) {
		wrapped := Wrap(nil, "operation")
		assert.Nil(t, wrapped)
	})
}

func TestWrapf(t *testing.T) {
	originalErr := errors.New("original error")

	t.Run("wrapf with error", func(t *testing.T) {
		wrapped := Wrapf(originalErr, "failed to %s %s", "process", "request")
		assert.Equal(t, "failed to process request: original error", wrapped.Error())
		assert.True(t, errors.Is(wrapped, originalErr))
	})

	t.Run("wrapf with nil error", func(t *testing.T) {
		wrapped := Wrapf(nil, "failed to %s", "process")
		assert.Nil(t, wrapped)
	})
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrAPINotFound",
			err:      ErrAPINotFound,
			expected: true,
		},
		{
			name:     "ErrElementNotFound",
			err:      ErrElementNotFound,
			expected: true,
		},
		{
			name: "APIError with 404",
			err: &APIError{
				StatusCode: 404,
				Message:    "not found",
			},
			expected: true,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name: "APIError with 500",
			err: &APIError{
				StatusCode: 500,
				Message:    "server error",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsNotFound(tt.err))
		})
	}
}

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrAPIUnauthorized",
			err:      ErrAPIUnauthorized,
			expected: true,
		},
		{
			name:     "ErrAuthFailed",
			err:      ErrAuthFailed,
			expected: true,
		},
		{
			name: "APIError with 401",
			err: &APIError{
				StatusCode: 401,
				Message:    "unauthorized",
			},
			expected: true,
		},
		{
			name: "APIError with 403",
			err: &APIError{
				StatusCode: 403,
				Message:    "forbidden",
			},
			expected: true,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name: "APIError with 404",
			err: &APIError{
				StatusCode: 404,
				Message:    "not found",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsUnauthorized(tt.err))
		})
	}
}

func TestIsTemporary(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrTimeout",
			err:      ErrTimeout,
			expected: true,
		},
		{
			name:     "ErrConnectionFailed",
			err:      ErrConnectionFailed,
			expected: true,
		},
		{
			name: "APIError with 500",
			err: &APIError{
				StatusCode: 500,
				Message:    "server error",
			},
			expected: true,
		},
		{
			name: "APIError with 502",
			err: &APIError{
				StatusCode: 502,
				Message:    "bad gateway",
			},
			expected: true,
		},
		{
			name: "APIError with 503",
			err: &APIError{
				StatusCode: 503,
				Message:    "service unavailable",
			},
			expected: true,
		},
		{
			name: "APIError with 504",
			err: &APIError{
				StatusCode: 504,
				Message:    "gateway timeout",
			},
			expected: true,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name: "APIError with 404",
			err: &APIError{
				StatusCode: 404,
				Message:    "not found",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsTemporary(tt.err))
		})
	}
}
