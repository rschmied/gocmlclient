//go:build integration

package integration

import (
	"testing"
)

func TestIntegration_ExtConn(t *testing.T) {
	cfg := LoadConfigFromEnv()
	c := newClient(t, cfg)
	wireClientServices(c)
	requireReady(t, c, cfg)

	ctx, cancel := testContext(t, cfg)
	defer cancel()

	list, err := c.ExtConn.List(ctx)
	if err != nil {
		t.Fatalf("ExtConn.List: %v", err)
	}
	if len(list) == 0 {
		t.Fatalf("ExtConn.List: expected at least 1 connector")
	}

	first := list[0]
	if first == nil {
		t.Fatalf("ExtConn.List: got nil entry")
	}

	_, err = c.ExtConn.Get(ctx, first.ID)
	if err != nil {
		t.Fatalf("ExtConn.Get: %v", err)
	}
}
