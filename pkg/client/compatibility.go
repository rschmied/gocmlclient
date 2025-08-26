package client

import (
	"context"

	"github.com/rschmied/gocmlclient/pkg/models"
)

func (c *Client) LabGet(ctx context.Context, id string, deep bool) (*models.Lab, error) {
	return c.Lab.GetByID(ctx, models.UUID(id), deep)
}
