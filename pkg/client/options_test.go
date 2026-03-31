package client

import (
	"log/slog"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConditional(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		option    Option
		validate  func(t *testing.T, config *Config)
	}{
		{
			name:      "condition true",
			condition: true,
			option:    WithToken("test-token"),
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "test-token", config.token)
			},
		},
		{
			name:      "condition false",
			condition: false,
			option:    WithToken("test-token"),
			validate: func(t *testing.T, config *Config) {
				assert.Empty(t, config.token)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{}
			option := Conditional(tt.condition, tt.option)
			option(config)
			tt.validate(t, config)
		})
	}
}

func TestWithUsernamePassword(t *testing.T) {
	c := &Config{}
	opt := WithUsernamePassword("testuser", "testpass")
	opt(c)
	assert.Equal(t, "testuser", c.username)
	assert.Equal(t, "testpass", c.password)
}

func TestWithInsecureTLS(t *testing.T) {
	c := &Config{}
	opt := WithInsecureTLS()
	opt(c)
	assert.True(t, c.insecureSkipVerify)
}

func TestWithToken(t *testing.T) {
	c := &Config{}
	opt := WithToken("testtoken")
	opt(c)
	assert.Equal(t, "testtoken", c.token)
}

func TestWithStaticToken(t *testing.T) {
	c := &Config{}
	opt := WithStaticToken("statictoken")
	opt(c)
	assert.Equal(t, "statictoken", c.staticToken)
}

func TestWithTokenStorageFile(t *testing.T) {
	c := &Config{}
	opt := WithTokenStorageFile("/path/to/file")
	opt(c)
	assert.Equal(t, "/path/to/file", c.tokenStorageFile)
}

func TestWithHTTPClient(t *testing.T) {
	c := &Config{}
	hc := &http.Client{}
	opt := WithHTTPClient(hc)
	opt(c)
	assert.Equal(t, hc, c.httpClient)
}

func TestWithLogger(t *testing.T) {
	c := &Config{}
	logger := slog.New(slog.NewTextHandler(nil, nil))
	opt := WithLogger(logger)
	opt(c)
	assert.Equal(t, logger, c.logger)
}

func TestWithLogLevel(t *testing.T) {
	c := &Config{}
	opt := WithLogLevel(slog.LevelInfo)
	opt(c)
	assert.Equal(t, slog.LevelInfo, c.logLevel)
}

func TestWithoutNamedConfigs(t *testing.T) {
	c := &Config{namedConfigs: true} // default might be true, but test setting to false
	opt := WithoutNamedConfigs()
	opt(c)
	assert.False(t, c.namedConfigs)
}

func TestWithNodeExcludeConfigurations(t *testing.T) {
	c := &Config{}
	opt := WithNodeExcludeConfigurations(false)
	opt(c)
	if c.nodeExcludeConfigurations == nil {
		t.Fatalf("expected nodeExcludeConfigurations to be set")
	}
	assert.False(t, *c.nodeExcludeConfigurations)
}

func TestSkipReadyCheck(t *testing.T) {
	c := &Config{}
	opt := SkipReadyCheck()
	opt(c)
	assert.True(t, c.skipReadyCheck)
}

func TestWithCACertPEM(t *testing.T) {
	c := &Config{}
	opt := WithCACertPEM([]byte("-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n"))
	opt(c)
	assert.NotEmpty(t, c.caCertPEM)
}

func TestWithRequestHeader(t *testing.T) {
	c := &Config{}
	opt := WithRequestHeader("X-Proxy-Token", "proxy-secret")
	opt(c)
	assert.Equal(t, map[string]string{"X-Proxy-Token": "proxy-secret"}, c.requestHeaders)
}

func TestWithRequestHeaders(t *testing.T) {
	c := &Config{}
	headers := map[string]string{
		"X-Proxy-Token": "proxy-secret",
		"X-Trace-ID":    "trace-123",
	}
	opt := WithRequestHeaders(headers)
	opt(c)

	headers["X-Trace-ID"] = "mutated"

	assert.Equal(t, map[string]string{
		"X-Proxy-Token": "proxy-secret",
		"X-Trace-ID":    "trace-123",
	}, c.requestHeaders)
}
