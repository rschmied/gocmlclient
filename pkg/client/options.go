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
	insecureSkipVerify bool
	namedConfigs       bool
	httpClient         *http.Client
	logger             *slog.Logger
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
