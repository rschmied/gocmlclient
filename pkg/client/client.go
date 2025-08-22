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
		// HTTPClient: http.DefaultClient,
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
	slog.Info("=== Full Integration Example ===")

	// 1. Create token provider
	provider := auth.NewUsernamePasswordProvider(auth.UsernamePasswordConfig{
		BaseURL:            c.baseURL,
		Username:           c.username,
		Password:           c.password,
		PresetToken:        c.token,
		InsecureSkipVerify: c.insecureSkipVerify,
	})

	// 2. Create auth manager
	manager := auth.NewManager(provider, auth.Config{
		RefreshBuffer: 30 * time.Second,
	})

	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: c.insecureSkipVerify,
	}
	tr.Proxy = http.ProxyFromEnvironment

	// // 4. Create authenticated transport
	authTransport := auth.NewTransport(auth.TransportConfig{
		// Base:    api.NewTransport(c.insecureSkipVerify),
		Base:    tr,
		Manager: manager,
		// SkipAuthEndpoints: []string{
		// 	"/api/v0/auth_extended",
		// 	"/health", // additional endpoints to skip
		// },
	})

	// 5. Create HTTP client with auth transport
	httpClient := &http.Client{
		Transport: authTransport,
		Timeout:   15 * time.Second,
	}

	// 6. Create API client with middlewares
	middlewares := []api.Middleware{
		api.LoggingMiddleware(slog.Default()),
		api.RetryMiddleware(api.DefaultRetryPolicy()),
	}

	apiClient := api.New(c.baseURL, api.Options{
		HTTPClient:  httpClient,
		Middlewares: middlewares,
	})

	return apiClient
}

// type LabClient struct {
// 	// Embed or compose internal services
// }
//
// func (c *LabClient) Get(ctx context.Context, id string, deep bool) (*models.Lab, error)
// func (c *LabClient) Create(ctx context.Context, lab *models.Lab) (*models.Lab, error)
