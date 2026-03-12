//go:build integration

package integration

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	stdErrors "errors"
	"testing"
	"time"

	"github.com/rschmied/gocmlclient/pkg/client"
	pkgerrors "github.com/rschmied/gocmlclient/pkg/errors"
	"github.com/rschmied/gocmlclient/pkg/models"
)

func wireClientServices(c *client.Client) {
	if c == nil {
		return
	}
	if c.Link != nil {
		c.Link.Interface = c.Interface
		c.Link.Node = c.Node
	}
}

func requireReady(t *testing.T, c *client.Client, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err := c.System.Ready(ctx); err != nil {
		t.Fatalf("System.Ready: %v", err)
	}
}

func testContext(t *testing.T, cfg Config) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), cfg.Timeout)
}

func randSuffix(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func createTempLab(t *testing.T, c *client.Client, cfg Config, titlePrefix string) models.Lab {
	t.Helper()
	if !cfg.AllowMutations {
		t.Skip("set CML_IT_ALLOW_MUTATIONS=1 to run")
	}
	ctx, cancel := testContext(t, cfg)
	defer cancel()

	title := titlePrefix + "-" + randSuffix(6)
	lab, err := c.Lab.Create(ctx, models.LabCreateRequest{Title: title, Description: "integration test"})
	if err != nil {
		t.Fatalf("Lab.Create: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cleanupCancel()
		_ = c.Lab.Stop(cleanupCtx, lab.ID)
		_ = c.Lab.Wipe(cleanupCtx, lab.ID)
		_ = c.Lab.Delete(cleanupCtx, lab.ID)
	})

	return lab
}

func requireNoErrorOrSkipStatus(t *testing.T, err error, skipStatusCodes ...int) {
	t.Helper()
	if err == nil {
		return
	}
	var apiErr *pkgerrors.APIError
	if stdErrors.As(err, &apiErr) {
		for _, code := range skipStatusCodes {
			if apiErr.StatusCode == code {
				t.Skipf("skipping due to HTTP %d: %v", code, err)
			}
		}
	}
	t.Fatalf("unexpected error: %v", err)
}
