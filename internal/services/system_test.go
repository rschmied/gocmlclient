package services

import (
	"context"
	"testing"

	"github.com/rschmied/gocmlclient/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestSystemReady(t *testing.T) {
	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	service := NewSystemService(client)
	ctx := context.Background()

	err := service.Ready(ctx)
	assert.NoError(t, err)
}
