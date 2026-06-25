//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/rschmied/gocmlclient/pkg/client"
	"github.com/rschmied/gocmlclient/pkg/models"
)

func waitForLinkState(t *testing.T, c *client.Client, ctx context.Context, labID, linkID models.UUID, expected string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	var last models.Link
	for time.Now().Before(deadline) {
		updated, err := c.Link.GetByID(ctx, labID, linkID)
		if err == nil {
			last = updated
			if updated.State == expected {
				return
			}
		}
		time.Sleep(2 * time.Second)
	}

	t.Fatalf("Link did not reach expected state %q within timeout (last state=%q)", expected, last.State)
}

func TestIntegration_LinkStartStop(t *testing.T) {
	cfg := LoadConfigFromEnv()
	c := newClient(t, cfg)
	wireClientServices(c)
	requireReady(t, c, cfg)

	ctx, cancel := testContext(t, cfg)
	defer cancel()
	if !cfg.AllowMutations {
		t.Skip("set CML_IT_ALLOW_MUTATIONS=1 to run")
	}

	lab := createTempLab(t, c, cfg, "it-link-start-stop")

	// Create two nodes (definition via env so users can adapt)
	def := envString("CML_IT_NODE_DEFINITION", envString("CML_IOL_NODE_DEFINITION", "iol-xe"))

	n1, err := c.Node.Create(ctx, models.Node{LabID: lab.ID, Label: "n1", NodeDefinition: def, X: 0, Y: 0})
	if err != nil {
		requireNoErrorOrSkipStatus(t, err, 400, 403, 404)
	}
	n2, err := c.Node.Create(ctx, models.Node{LabID: lab.ID, Label: "n2", NodeDefinition: def, X: 200, Y: 0})
	if err != nil {
		requireNoErrorOrSkipStatus(t, err, 400, 403, 404)
	}

	// Ensure at least one interface exists per node
	ifaces1, err := c.Interface.GetInterfacesForNode(ctx, lab.ID, n1.ID)
	if err != nil {
		t.Fatalf("Interface.GetInterfacesForNode(n1): %v", err)
	}
	if len(ifaces1) == 0 {
		_, err = c.Interface.Create(ctx, lab.ID, n1.ID, -1)
		if err != nil {
			t.Fatalf("Interface.Create(n1): %v", err)
		}
		ifaces1, err = c.Interface.GetInterfacesForNode(ctx, lab.ID, n1.ID)
		if err != nil {
			t.Fatalf("Interface.GetInterfacesForNode(n1) (2): %v", err)
		}
	}
	ifaces2, err := c.Interface.GetInterfacesForNode(ctx, lab.ID, n2.ID)
	if err != nil {
		t.Fatalf("Interface.GetInterfacesForNode(n2): %v", err)
	}
	if len(ifaces2) == 0 {
		_, err = c.Interface.Create(ctx, lab.ID, n2.ID, -1)
		if err != nil {
			t.Fatalf("Interface.Create(n2): %v", err)
		}
		ifaces2, err = c.Interface.GetInterfacesForNode(ctx, lab.ID, n2.ID)
		if err != nil {
			t.Fatalf("Interface.GetInterfacesForNode(n2) (2): %v", err)
		}
	}

	// Create link between nodes (slot 0 -> if slots unsupported, CML may auto-create)
	link, err := c.Link.Create(ctx, models.Link{LabID: lab.ID, SrcNode: n1.ID, DstNode: n2.ID, SrcSlot: 0, DstSlot: 0})
	if err != nil {
		requireNoErrorOrSkipStatus(t, err, 400, 403, 404)
	}

	// Start the lab now that all resources (nodes + link) exist.
	if err := c.Lab.Start(ctx, lab.ID); err != nil {
		t.Fatalf("Lab.Start: %v", err)
	}

	labDeadline := time.Now().Add(180 * time.Second)
	for time.Now().Before(labDeadline) {
		ok, err := c.Lab.HasConverged(ctx, lab.ID)
		if err != nil {
			t.Fatalf("Lab.HasConverged: %v", err)
		}
		if ok {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if ok, err := c.Lab.HasConverged(ctx, lab.ID); err != nil {
		t.Fatalf("Lab.HasConverged (final): %v", err)
	} else if !ok {
		t.Fatalf("lab did not converge within timeout")
	}

	// Start link
	if err := c.Link.Start(ctx, lab.ID, link.ID); err != nil {
		t.Fatalf("Link.Start: %v", err)
	}

	// Wait a bit and verify started state
	waitForLinkState(t, c, ctx, lab.ID, link.ID, string(models.LinkStateStarted), 30*time.Second)

	// Stop link
	if err := c.Link.Stop(ctx, lab.ID, link.ID); err != nil {
		t.Fatalf("Link.Stop: %v", err)
	}

	waitForLinkState(t, c, ctx, lab.ID, link.ID, string(models.LinkStateStopped), 30*time.Second)

	// Start link again
	if err := c.Link.Start(ctx, lab.ID, link.ID); err != nil {
		t.Fatalf("Link.Start (2nd time): %v", err)
	}

	waitForLinkState(t, c, ctx, lab.ID, link.ID, string(models.LinkStateStarted), 30*time.Second)

	// Best-effort delete/cleanup
	_ = c.Link.Delete(ctx, lab.ID, link.ID)
	_ = c.Node.Delete(ctx, lab.ID, n1.ID)
	_ = c.Node.Delete(ctx, lab.ID, n2.ID)
	_ = c.Lab.Delete(ctx, lab.ID)
}
