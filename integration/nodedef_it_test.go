//go:build integration

package integration

import "testing"

func TestIntegration_NodeDefinition(t *testing.T) {
	cfg := LoadConfigFromEnv()
	c := newClient(t, cfg)
	requireReady(t, c, cfg)

	ctx, cancel := testContext(t, cfg)
	defer cancel()

	defs, err := c.NodeDefinition.NodeDefinitions(ctx)
	if err != nil {
		t.Fatalf("NodeDefinition.NodeDefinitions: %v", err)
	}
	if len(defs) == 0 {
		t.Skip("no node definitions returned")
	}
}
