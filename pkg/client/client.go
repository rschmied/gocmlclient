// Package client provides a Cisco Modeling Labs API client
package client

import (
	"crypto/tls"
	"log/slog"
	"net/http"
	"time"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/auth"
	"github.com/rschmied/gocmlclient/internal/services"
)

type Client struct {
	config    *Config
	apiClient *api.Client

	// services
	Labs   *services.LabService
	System *services.SystemService
	// Nodes  *NodeClient
	// System *SystemClient
	// Users  *UserClient
}

func New(baseURL string, opts ...Option) (*Client, error) {
	// start with defaults
	cfg := &Config{
		// httpClient: http.DefaultClient,
		logger:  slog.Default(), // whatever you use
		baseURL: baseURL,
	}

	// apply all options
	for _, opt := range opts {
		opt(cfg)
	}

	apiClient := createAPIclient(cfg)

	// build the client using cfg
	c := &Client{
		config:    cfg,
		apiClient: apiClient,
		Labs:      services.NewLabService(apiClient),
		System:    services.NewSystemService(apiClient),
	}

	return c, nil
}

func createAPIclient(c *Config) *api.Client {
	// 1. Create token provider
	provider := auth.NewAuthProvider(auth.AuthConfig{
		BaseURL:            c.baseURL,
		Username:           c.username,
		Password:           c.password,
		PresetToken:        c.token,
		InsecureSkipVerify: c.insecureSkipVerify,
	})

	// 2. create the auth manager
	manager := auth.NewManager(provider, auth.Config{
		RefreshBuffer: 30 * time.Second,
	})

	// 3. create a sane base transport
	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: c.insecureSkipVerify,
	}
	tr.Proxy = http.ProxyFromEnvironment

	// 4. create authenticated transport
	authTransport := auth.NewTransport(tr, manager)

	// 5. create a default HTTP client, if needed
	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	// 6. attach the auth transport
	c.httpClient.Transport = authTransport

	// 7. create API client with middlewares
	apiClient := api.New(c.baseURL, api.Options{
		HTTPClient: c.httpClient,
		Middlewares: []api.Middleware{
			api.LoggingMiddleware(c.logger),
			api.RetryMiddleware(api.DefaultRetryPolicy()),
		},
	})

	return apiClient
}
