//go:build integration

package integration

import "testing"

func TestIntegration_ImageDefinition(t *testing.T) {
	cfg := LoadConfigFromEnv()
	c := newClient(t, cfg)
	requireReady(t, c, cfg)

	ctx, cancel := testContext(t, cfg)
	defer cancel()

	defs, err := c.ImageDefinition.ImageDefinitions(ctx)
	if err != nil {
		t.Fatalf("ImageDefinition.ImageDefinitions: %v", err)
	}
	if len(defs) == 0 {
		t.Skip("no image definitions returned")
	}
}
