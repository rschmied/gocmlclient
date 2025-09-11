// Package errors provides error related functions and types
package errors

import (
	"crypto/x509"
	"errors"
	"strings"
)

var (
	// ErrSystemNotReady is returned when the CML system is not ready.
	ErrSystemNotReady = errors.New("system not ready")
	// ErrElementNotFound is returned when a requested element is not found.
	ErrElementNotFound = errors.New("element not found")
	// ErrNoNamedConfigSupport is returned when the backend does not support named configs.
	ErrNoNamedConfigSupport = errors.New("backend does not support named configs")
	// ErrTLSCertificate is returned when TLS certificate validation fails.
	ErrTLSCertificate = errors.New("TLS certificate validation failed")
)

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
