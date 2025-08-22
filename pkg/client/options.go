package client

import (
	"log/slog"
)

type Option func(*Config)

type Config struct {
	baseURL            string
	username           string
	password           string
	token              string
	insecureSkipVerify bool
	logger             *slog.Logger

	// Token      string
	// HTTPClient *http.Client
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

// func WithHTTPClient(hc *http.Client) Option {
// 	return func(c *Config) {
// 		c.HTTPClient = hc
// 	}
// }

func WithLogger(l *slog.Logger) Option {
	return func(c *Config) {
		c.logger = l
	}
}
