package client

import (
	"log/slog"
	"net/http"
)

type Option func(*Config)

type Config struct {
	baseURL            string
	username           string
	password           string
	token              string
	tokenStorageFile   string
	insecureSkipVerify bool
	namedConfigs       bool
	httpClient         *http.Client
	logger             *slog.Logger
}

func Conditional(condition bool, option Option) Option {
	if condition {
		return option
	}
	return func(c *Config) {
		// No-op - condition is false, so don't apply the option
	}
}

func WithUsernamePassword(username, password string) Option {
	return func(c *Config) {
		c.username = username
		c.password = password
	}
}

func WithInsecureTLS() Option {
	return func(c *Config) {
		c.insecureSkipVerify = true
	}
}

func WithToken(token string) Option {
	return func(c *Config) {
		c.token = token
	}
}

func WithTokenStorageFile(filename string) Option {
	return func(c *Config) {
		c.tokenStorageFile = filename
	}
}

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Config) {
		c.httpClient = hc
	}
}

func WithLogger(l *slog.Logger) Option {
	return func(c *Config) {
		c.logger = l
	}
}

func WithNamedConfigs() Option {
	return func(c *Config) {
		c.namedConfigs = true
	}
}
