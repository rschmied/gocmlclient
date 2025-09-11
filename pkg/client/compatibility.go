// Package client provides compatibility functions for the CML client.
package client

import (
	"context"

	"github.com/rschmied/gocmlclient/pkg/models"
)

// LabGet retrieves a lab by ID with optional deep loading.
func (c *Client) LabGet(ctx context.Context, id string, deep bool) (*models.Lab, error) {
	return c.Lab.GetByID(ctx, models.UUID(id), deep)
}
