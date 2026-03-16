// Package client provides configuration options for the CML client.
package client

import (
	"log/slog"
	"net/http"
)

// Option is a functional option for configuring the client.
type Option func(*Config)

// Config holds the configuration for the CML client.
type Config struct {
	baseURL            string
	username           string
	password           string
	token              string
	tokenStorageFile   string
	insecureSkipVerify bool
	caCertPEM          []byte
	namedConfigs       bool
	// nodeExcludeConfigurations, when set, forces Node GET/LIST query behavior by
	// explicitly sending exclude_configurations=true/false.
	//
	// This is independent from namedConfigs. It exists because older CML versions
	// may default the configuration field to different shapes when the parameter
	// is omitted (e.g. string vs named-config list).
	nodeExcludeConfigurations *bool
	httpClient                *http.Client
	logger                    *slog.Logger
	skipReadyCheck            bool
}

// Conditional applies an option only if the condition is true.
func Conditional(condition bool, option Option) Option {
	if condition {
		return option
	}
	return func(c *Config) {
		// No-op - condition is false, so don't apply the option
	}
}

// WithUsernamePassword sets the username and password for authentication.
func WithUsernamePassword(username, password string) Option {
	return func(c *Config) {
		c.username = username
		c.password = password
	}
}

// WithInsecureTLS skips TLS certificate verification.
func WithInsecureTLS() Option {
	return func(c *Config) {
		c.insecureSkipVerify = true
	}
}

// WithToken sets the authentication token.
func WithToken(token string) Option {
	return func(c *Config) {
		c.token = token
	}
}

// WithCACertPEM sets a custom CA certificate bundle (PEM). The certificates
// are added to the system cert pool.
func WithCACertPEM(certPEM []byte) Option {
	return func(c *Config) {
		c.caCertPEM = certPEM
	}
}

// WithTokenStorageFile sets the file to store the authentication token. If no
// stoken storage file is provided, memory storage is being used. There's a
// tradeoff with memory storage as it requires more authentication API calls
// which is costly especially when a lot of clients are instantiated during
// e.g. Terraform runs. The file storage saves the token in the file system but
// this has a security implication as the token might be retrieved. It's up to
// the user to remove the configured file with a potentially still valid token!
func WithTokenStorageFile(filename string) Option {
	return func(c *Config) {
		c.tokenStorageFile = filename
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Config) {
		c.httpClient = hc
	}
}

// WithLogger sets the logger for the client.
func WithLogger(l *slog.Logger) Option {
	return func(c *Config) {
		c.logger = l
	}
}

// WithoutNamedConfigs enables support for named configurations. Named
// configurations were introduced with CML 2.7 and should be enabled.
func WithoutNamedConfigs() Option {
	return func(c *Config) {
		c.namedConfigs = false
	}
}

// WithNodeExcludeConfigurations forces the node GET/LIST query parameter
// exclude_configurations.
//
// Passing true omits configuration payloads, passing false includes configuration
// payloads. This is useful to keep node configuration shape stable across CML
// versions.
func WithNodeExcludeConfigurations(v bool) Option {
	return func(c *Config) {
		c.nodeExcludeConfigurations = &v
	}
}

// SkipReadyCheck disables the automatic system readiness check during client
// initialization. By default, the client will call Ready() to verify the CML
// server is compatible and ready before returning. This check can be skipped
// for performance reasons or when working with servers that don't support
// the system_information endpoint.
func SkipReadyCheck() Option {
	return func(c *Config) {
		c.skipReadyCheck = true
	}
}
