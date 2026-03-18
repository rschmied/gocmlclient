// Package gocmlclient provides a client for Cisco Modeling Labs
package gocmlclient

import (
	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/pkg/client"
	"github.com/rschmied/gocmlclient/pkg/models"
)

// New creates a new CML client - convenience constructor
func New(baseURL string, opts ...client.Option) (*client.Client, error) {
	return client.New(baseURL, opts...)
}

// Option is re-exported for convenience.
type Option = client.Option

// Re-export common types for convenience
type (
	// Lab is a CML lab.
	Lab = models.Lab
	// Node is a CML node.
	Node = models.Node
	// Stats represents API client statistics.
	Stats = api.Stats
)

// Re-export common options for convenience.
var (
	Conditional                   = client.Conditional
	SkipReadyCheck                = client.SkipReadyCheck
	WithCACertPEM                 = client.WithCACertPEM
	WithHTTPClient                = client.WithHTTPClient
	WithInsecureTLS               = client.WithInsecureTLS
	WithLogLevel                  = client.WithLogLevel
	WithLogger                    = client.WithLogger
	WithNodeExcludeConfigurations = client.WithNodeExcludeConfigurations
	WithStaticToken               = client.WithStaticToken
	WithToken                     = client.WithToken
	WithTokenStorageFile          = client.WithTokenStorageFile
	WithUsernamePassword          = client.WithUsernamePassword
	WithoutNamedConfigs           = client.WithoutNamedConfigs
)
