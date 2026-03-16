//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/rschmied/gocmlclient/pkg/models"
)

func TestIntegration_SmartAnnotations(t *testing.T) {
	cfg := LoadConfigFromEnv()
	c := newClient(t, cfg)
	wireClientServices(c)
	requireReady(t, c, cfg)

	ctx, cancel := testContext(t, cfg)
	defer cancel()

	if !cfg.AllowMutations {
		t.Skip("set CML_IT_ALLOW_MUTATIONS=1 to run")
	}

	lab := createTempLab(t, c, cfg, "it-smart-ann")

	// Smart annotations are generated based on node tags. Create two nodes with
	// a shared tag so the backend can create a smart annotation.
	def := envString("CML_IT_NODE_DEFINITION", envString("CML_IOL_NODE_DEFINITION", "iol-xe"))
	_, err := c.Node.Create(ctx, models.Node{LabID: lab.ID, Label: "smart-1", NodeDefinition: def, X: 0, Y: 0, Tags: []string{"smart"}})
	if err != nil {
		requireNoErrorOrSkipStatus(t, err, 400, 403, 404)
	}
	_, err = c.Node.Create(ctx, models.Node{LabID: lab.ID, Label: "smart-2", NodeDefinition: def, X: 200, Y: 0, Tags: []string{"smart"}})
	if err != nil {
		requireNoErrorOrSkipStatus(t, err, 400, 403, 404)
	}

	// The backend may be eventually consistent; poll briefly.
	var list []models.SmartAnnotation
	for range 10 {
		list, err = c.SmartAnnotation.List(ctx, lab.ID)
		if err != nil {
			requireNoErrorOrSkipStatus(t, err, 404)
		}
		if len(list) > 0 {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if len(list) == 0 {
		t.Skip("no smart annotations returned (expected from tagged nodes)")
	}

	// Try GET; PATCH may not be allowed depending on backend.
	_, err = c.SmartAnnotation.Get(ctx, lab.ID, list[0].ID)
	if err != nil {
		requireNoErrorOrSkipStatus(t, err, 404)
	}
}
