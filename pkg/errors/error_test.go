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
	}

	for i, err1 := range errors {
		for j, err2 := range errors {
			if i != j {
				assert.NotEqual(t, err1, err2, "error %d and %d should be different", i, j)
			}
		}
	}
}
