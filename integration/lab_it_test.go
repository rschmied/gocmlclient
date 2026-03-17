//go:build integration

package integration

import (
	stdErrors "errors"
	"testing"
	"time"

	pkgerrors "github.com/rschmied/gocmlclient/pkg/errors"
	"github.com/rschmied/gocmlclient/pkg/models"
)

func TestIntegration_Lab(t *testing.T) {
	cfg := LoadConfigFromEnv()
	c := newClient(t, cfg)
	wireClientServices(c)
	requireReady(t, c, cfg)

	ctx, cancel := testContext(t, cfg)
	defer cancel()

	_, err := c.Lab.Labs(ctx, false)
	if err != nil {
		t.Fatalf("Lab.Labs(false): %v", err)
	}
	_, err = c.Lab.Labs(ctx, true)
	if err != nil {
		t.Fatalf("Lab.Labs(true): %v", err)
	}
	_, err = c.Lab.LabsWithData(ctx)
	if err != nil {
		t.Fatalf("Lab.LabsWithData: %v", err)
	}

	if !cfg.AllowMutations {
		t.Skip("set CML_IT_ALLOW_MUTATIONS=1 to run")
	}
	lab := createTempLab(t, c, cfg, "it-lab")

	// GetByID deep/shallow
	_, err = c.Lab.GetByID(ctx, lab.ID, false)
	if err != nil {
		t.Fatalf("Lab.GetByID(false): %v", err)
	}
	_, err = c.Lab.GetByID(ctx, lab.ID, true)
	if err != nil {
		t.Fatalf("Lab.GetByID(true): %v", err)
	}

	// Update
	_, err = c.Lab.Update(ctx, lab.ID, models.LabUpdateRequest{Title: "it-lab-updated-" + randSuffix(4)})
	if err != nil {
		t.Fatalf("Lab.Update: %v", err)
	}

	// Node staging update (best effort; requires CML 2.10+)
	_, err = c.Lab.Update(ctx, lab.ID, models.LabUpdateRequest{NodeStaging: &models.NodeStaging{Enabled: false, StartRemaining: true, AbortOnFailure: false}})
	if err != nil {
		requireNoErrorOrSkipStatus(t, err, 400, 403, 404)
	}

	// GetByTitle
	// This endpoint uses /populate_lab_tiles under the hood and can be eventually consistent.
	var lastErr error
	for range 10 {
		_, lastErr = c.Lab.GetByTitle(ctx, lab.Title, false)
		if lastErr == nil {
			break
		}
		if stdErrors.Is(lastErr, pkgerrors.ErrElementNotFound) {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		break
	}
	if lastErr != nil {
		if stdErrors.Is(lastErr, pkgerrors.ErrElementNotFound) {
			t.Logf("Lab.GetByTitle: not found yet; skipping strict assertion: %v", lastErr)
		} else {
			// some backends may not return tiles / title lookup works differently
			requireNoErrorOrSkipStatus(t, lastErr, 404)
		}
	}

	// Start/Stop/Wipe once each.
	if err := c.Lab.Start(ctx, lab.ID); err != nil {
		requireNoErrorOrSkipStatus(t, err, 400, 403)
	}
	_ = c.Lab.Stop(ctx, lab.ID)
	_ = c.Lab.Wipe(ctx, lab.ID)

	// Converged check (may be false or error depending on backend state)
	_, _ = c.Lab.HasConverged(ctx, lab.ID)
}
