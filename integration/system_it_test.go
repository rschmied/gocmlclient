//go:build integration

package integration

import "testing"

func TestIntegration_System(t *testing.T) {
	cfg := LoadConfigFromEnv()
	c := newClient(t, cfg)

	ctx, cancel := testContext(t, cfg)
	defer cancel()

	if err := c.System.Ready(ctx); err != nil {
		t.Fatalf("System.Ready: %v", err)
	}

	_ = c.System.Version()
	_, _ = c.System.VersionCheck(ctx, ">=2.4.0")
}
