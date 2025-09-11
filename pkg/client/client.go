// Package client provides a Cisco Modeling Labs API client
package client

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"time"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/auth"
	"github.com/rschmied/gocmlclient/internal/services"
)

// Client is the main CML API client that provides access to all services.
type Client struct {
	config    *Config
	apiClient *api.Client

	// services
	Interface *services.InterfaceService
	Lab       *services.LabService
	Link      *services.LinkService
	Node      *services.NodeService
	System    *services.SystemService
	Group     *services.GroupService
	User      *services.UserService
}

// New creates a new CML client with the given options.
func New(baseURL string, opts ...Option) (*Client, error) {
	// start with defaults
	cfg := &Config{
		logger:       slog.Default(),
		baseURL:      baseURL,
		namedConfigs: true, // make this the default!
	}

	// apply all options
	for _, opt := range opts {
		opt(cfg)
	}

	apiClient, err := newAPIClient(cfg)
	if err != nil {
		return nil, err
	}

	groupService := services.NewGroupService(apiClient)
	userService := services.NewUserService(apiClient)
	nodeService := services.NewNodeService(apiClient, cfg.namedConfigs)
	interfaceService := services.NewInterfaceService(apiClient)
	linkService := services.NewLinkService(apiClient)

	c := &Client{
		config:    cfg,
		apiClient: apiClient,
		Lab:       services.NewLabService(apiClient, interfaceService, linkService, userService, nodeService),
		Interface: interfaceService,
		Link:      linkService,
		Node:      nodeService,
		Group:     groupService,
		User:      userService,
		System:    services.NewSystemService(apiClient),
	}

	// Perform system readiness check unless explicitly skipped
	if !cfg.skipReadyCheck {
		if err := c.System.Ready(context.Background()); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func newAPIClient(c *Config) (*api.Client, error) {
	// 1. create or use provided HTTP client
	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	// 2. handle TLS configuration if needed
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

	// 3. get the base transport (before adding auth)
	baseTransport := c.httpClient.Transport
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}

	// Create file storage
	var storage auth.TokenStorage
	if len(c.tokenStorageFile) > 0 {
		var err error
		storage, err = auth.NewFileStorage(c.tokenStorageFile)
		if err != nil {
			return nil, err
		}
	}

	// 4. create token provider - it will use the SAME http client but the auth
	//    transport will skip auth endpoints
	provider := auth.NewAuthProvider(auth.AuthConfig{
		BaseURL:     c.baseURL,
		Username:    c.username,
		Password:    c.password,
		PresetToken: c.token,
		Client:      c.httpClient,
	})

	// 5. create manager with token storage (default is memory storage)
	config := auth.DefaultConfig()
	if storage != nil {
		config.Storage = storage
	}

	// 6. create the auth manager
	manager := auth.NewManager(provider, config)

	// 7. create authenticated transport that wraps the base transport
	authTransport := auth.NewTransport(baseTransport, manager, nil)

	// 8. set the auth transport on the client
	c.httpClient.Transport = authTransport

	// 9. create API client with middlewares
	apiClient := api.New(c.baseURL, api.Options{
		HTTPClient: c.httpClient,
		Middlewares: []api.Middleware{
			api.UserAgentMiddleware("gocmlclient"),
			// api.LoggingMiddleware(c.logger),
			api.LogRequestBodyMiddleware(c.logger),
			api.RetryMiddleware(api.DefaultRetryPolicy()),
		},
	})

	return apiClient, nil
}
