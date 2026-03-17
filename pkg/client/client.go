// Package client provides a Cisco Modeling Labs API client
package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/auth"
	"github.com/rschmied/gocmlclient/internal/httputil"
	"github.com/rschmied/gocmlclient/internal/logging"
	"github.com/rschmied/gocmlclient/internal/services"
	"github.com/rschmied/gocmlclient/internal/version"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/google/uuid"
)

// Client is the main CML API client that provides access to all services.
type Client struct {
	config    *Config
	apiClient *api.Client

	// services
	Interface       *services.InterfaceService
	Lab             *services.LabService
	Link            *services.LinkService
	Node            *services.NodeService
	System          *services.SystemService
	Group           *services.GroupService
	User            *services.UserService
	ImageDefinition *services.ImageDefinitionService
	NodeDefinition  *services.NodeDefinitionService
	ExtConn         *services.ExtConnService
	Annotation      *services.AnnotationService
	SmartAnnotation *services.SmartAnnotationService
}

// New creates a new CML client with the given options.
func New(baseURL string, opts ...Option) (*Client, error) {
	// start with defaults
	cfg := &Config{
		logLevel:     slog.LevelWarn,
		baseURL:      baseURL,
		namedConfigs: true, // make this the default!
	}

	// apply all options
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.logger == nil {
		cfg.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: cfg.logLevel}))
	}
	logging.SetDefault(cfg.logger)

	apiClient, err := newAPIClient(cfg)
	if err != nil {
		return nil, err
	}

	groupService := services.NewGroupService(apiClient)
	userService := services.NewUserService(apiClient, groupService)
	nodeService := services.NewNodeService(apiClient, cfg.namedConfigs)
	interfaceService := services.NewInterfaceService(apiClient)
	linkService := services.NewLinkService(apiClient)
	linkService.Interface = interfaceService
	linkService.Node = nodeService

	imageDefinitionService := services.NewImageDefinitionService(apiClient)
	nodeDefinitionService := services.NewNodeDefinitionService(apiClient)
	extConnService := services.NewExtConnService(apiClient)
	annotationService := services.NewAnnotationService(apiClient)
	smartAnnotationService := services.NewSmartAnnotationService(apiClient)

	c := &Client{
		config:          cfg,
		apiClient:       apiClient,
		Lab:             services.NewLabService(apiClient, interfaceService, linkService, userService, nodeService),
		Interface:       interfaceService,
		Link:            linkService,
		Node:            nodeService,
		Group:           groupService,
		User:            userService,
		System:          services.NewSystemService(apiClient),
		ImageDefinition: imageDefinitionService,
		NodeDefinition:  nodeDefinitionService,
		ExtConn:         extConnService,
		Annotation:      annotationService,
		SmartAnnotation: smartAnnotationService,
	}

	// If configured, force deterministic node configuration query behavior across
	// CML versions.
	if cfg.nodeExcludeConfigurations != nil {
		nodeService.SetExcludeConfigurations(cfg.nodeExcludeConfigurations)
	}

	// Perform system readiness check unless explicitly skipped
	if !cfg.skipReadyCheck {
		if err := c.System.Ready(context.Background()); err != nil {
			return nil, err
		}
	}

	// If configured, force deterministic node configuration query behavior across
	// CML versions.
	if cfg.nodeExcludeConfigurations != nil {
		nodeService.SetExcludeConfigurations(cfg.nodeExcludeConfigurations)
	}

	return c, nil
}

func newAPIClient(c *Config) (*api.Client, error) {
	clientUUID := uuid.NewString()
	clientVersion := version.Effective()

	// 1. create or use provided HTTP client
	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	// 2. handle TLS configuration if needed
	if c.insecureSkipVerify || len(c.caCertPEM) > 0 {
		transport, ok := c.httpClient.Transport.(*http.Transport)
		if !ok || transport == nil {
			transport = http.DefaultTransport.(*http.Transport).Clone()
		} else {
			transport = transport.Clone()
		}

		tlsCfg := transport.TLSClientConfig
		if tlsCfg == nil {
			tlsCfg = &tls.Config{}
		} else {
			tlsCfg = tlsCfg.Clone()
		}

		if len(c.caCertPEM) > 0 {
			pool, err := x509.SystemCertPool()
			if err != nil || pool == nil {
				pool = x509.NewCertPool()
			}
			if ok := pool.AppendCertsFromPEM(c.caCertPEM); !ok {
				return nil, fmt.Errorf("append CA certs: invalid PEM")
			}
			tlsCfg.RootCAs = pool
		}

		if c.insecureSkipVerify {
			tlsCfg.InsecureSkipVerify = true
		}

		transport.TLSClientConfig = tlsCfg
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
		ClientID:    httputil.ClientID,
		ClientUUID:  clientUUID,
		Version:     clientVersion,
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
	apiClient := api.New(c.baseURL,
		api.WithHTTPClient(c.httpClient),
		api.WithStats(),
		api.WithMiddlewares(
			api.UserAgentMiddleware("gocmlclient"),
			// api.LoggingMiddleware(c.logger),
			// api.LogRequestBodyMiddleware(c.logger),
			api.RetryMiddleware(api.DefaultRetryPolicy()),
		),
	)
	apiClient.SetClientInfo(httputil.ClientID, clientUUID, clientVersion)

	return apiClient, nil
}

func (c *Client) Stats() *models.Stats {
	return c.apiClient.Stats()
}
