// Package gocmlclient provides a client for Cisco Modeling Labs
package gocmlclient

import (
	"github.com/rschmied/gocmlclient/pkg/client"
	"github.com/rschmied/gocmlclient/pkg/models"
)

// New creates a new CML client - convenience constructor
func New(baseURL string, opts ...client.Option) (*client.Client, error) {
	return client.New(baseURL, opts...)
}

// Re-export common types for convenience
type (
	Lab  = models.Lab
	Node = models.Node
)

// Re-export common options for convenience
var (
	SkipReadyCheck       = client.SkipReadyCheck
	WithUsernamePassword = client.WithUsernamePassword
	WithToken            = client.WithToken
	WithInsecureTLS      = client.WithInsecureTLS
)
