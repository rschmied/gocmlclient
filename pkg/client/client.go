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
	Interface *services.InterfaceService
	Lab       *services.LabService
	Node      *services.NodeService
	System    *services.SystemService
	Group     *services.GroupService
	User      *services.UserService
}

func New(baseURL string, opts ...Option) (*Client, error) {
	// start with defaults
	cfg := &Config{
		logger:  slog.Default(),
		baseURL: baseURL,
	}

	// apply all options
	for _, opt := range opts {
		opt(cfg)
	}

	apiClient := newAPIClient(cfg)

	c := &Client{
		config:    cfg,
		apiClient: apiClient,
		Interface: services.NewInterfaceService(apiClient),
		Lab:       services.NewLabService(apiClient),
		Node:      services.NewNodeService(apiClient, true),
		System:    services.NewSystemService(apiClient),
		Group:     services.NewGroupService(apiClient),
		User:      services.NewUserService(apiClient),
	}

	// inject dependencies
	c.Lab.Interface = c.Interface
	c.Lab.User = c.User
	c.Lab.Node = c.Node
	c.User.Group = c.Group

	// check version
	// c.System.Ready(context.Background())

	return c, nil
}

func newAPIClient(c *Config) *api.Client {
	// 1. Create or use provided HTTP client
	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	// 2. Handle TLS configuration if needed
	if c.insecureSkipVerify {
		transport, ok := c.httpClient.Transport.(*http.Transport)
		if !ok || transport == nil {
			transport = http.DefaultTransport.(*http.Transport).Clone()
		} else {
			transport = transport.Clone()
		}
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		c.httpClient.Transport = transport
	}

	// 3. Get the base transport (before adding auth)
	baseTransport := c.httpClient.Transport
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}

	// 4. Create token provider - it will use the SAME http client
	// but the auth transport will skip auth endpoints
	provider := auth.NewAuthProvider(auth.AuthConfig{
		BaseURL:     c.baseURL,
		Username:    c.username,
		Password:    c.password,
		PresetToken: c.token,
		HTTPclient:  c.httpClient, // Same client!
	})

	// 5. Create the auth manager
	manager := auth.NewManager(provider, auth.DefaultConfig())

	// 6. Create authenticated transport that wraps the base transport
	authTransport := auth.NewTransport(baseTransport, manager, nil)

	// 7. Set the auth transport on the client
	c.httpClient.Transport = authTransport

	// 8. Create API client
	apiClient := api.New(c.baseURL, api.Options{
		HTTPClient: c.httpClient,
		Middlewares: []api.Middleware{
			// api.LoggingMiddleware(c.logger),
			api.RetryMiddleware(api.DefaultRetryPolicy()),
		},
	})

	return apiClient
}
