//go:build integration

package integration

import (
	"testing"

	"github.com/rschmied/gocmlclient/pkg/models"
)

func TestIntegration_NodeInterfaceLink(t *testing.T) {
	cfg := LoadConfigFromEnv()
	c := newClient(t, cfg)
	wireClientServices(c)
	requireReady(t, c, cfg)

	ctx, cancel := testContext(t, cfg)
	defer cancel()
	if !cfg.AllowMutations {
		t.Skip("set CML_IT_ALLOW_MUTATIONS=1 to run")
	}

	lab := createTempLab(t, c, cfg, "it-nil")

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

	// Get node list / get by id
	_, err = c.Node.GetNodesForLab(ctx, lab.ID)
	if err != nil {
		t.Fatalf("Node.GetNodesForLab: %v", err)
	}
	_, err = c.Node.GetByID(ctx, lab.ID, n1.ID)
	if err != nil {
		t.Fatalf("Node.GetByID: %v", err)
	}

	// Interfaces for node + create explicit interface
	ifaces, err := c.Interface.GetInterfacesForNode(ctx, lab.ID, n1.ID)
	if err != nil {
		t.Fatalf("Interface.GetInterfacesForNode: %v", err)
	}
	if len(ifaces) == 0 {
		// try creating first interface
		_, err = c.Interface.Create(ctx, lab.ID, n1.ID, -1)
		if err != nil {
			t.Fatalf("Interface.Create: %v", err)
		}
		ifaces, err = c.Interface.GetInterfacesForNode(ctx, lab.ID, n1.ID)
		if err != nil {
			t.Fatalf("Interface.GetInterfacesForNode (2): %v", err)
		}
	}
	if len(ifaces) > 0 {
		_, err = c.Interface.GetByID(ctx, lab.ID, ifaces[0].ID)
		if err != nil {
			t.Fatalf("Interface.GetByID: %v", err)
		}
	}

	// Link create between nodes (slot 0)
	link, err := c.Link.Create(ctx, models.Link{LabID: lab.ID, SrcNode: n1.ID, DstNode: n2.ID, SrcSlot: 0, DstSlot: 0})
	if err != nil {
		requireNoErrorOrSkipStatus(t, err, 400, 403, 404)
	}

	// Link get/list
	_, err = c.Link.GetLinksForLab(ctx, lab.ID)
	if err != nil {
		t.Fatalf("Link.GetLinksForLab: %v", err)
	}
	_, err = c.Link.GetByID(ctx, lab.ID, link.ID)
	if err != nil {
		t.Fatalf("Link.GetByID: %v", err)
	}

	// Link conditioning (best effort)
	_, _ = c.Link.GetCondition(ctx, lab.ID, link.ID)
	_, _ = c.Link.SetCondition(ctx, lab.ID, link.ID, &models.LinkConditionConfiguration{Enabled: false})
	_ = c.Link.DeleteCondition(ctx, lab.ID, link.ID)

	// Update node (small label change)
	n1.Label = "n1-upd"
	_, _ = c.Node.Update(ctx, n1)

	// Node start/stop/wipe are best-effort (may fail on permissions)
	_ = c.Node.Start(ctx, lab.ID, n1.ID)
	_ = c.Node.Stop(ctx, lab.ID, n1.ID)
	_ = c.Node.Wipe(ctx, lab.ID, n1.ID)

	// Delete: link -> nodes
	_ = c.Link.Delete(ctx, lab.ID, link.ID)
	_ = c.Node.Delete(ctx, lab.ID, n1.ID)
	_ = c.Node.Delete(ctx, lab.ID, n2.ID)
}
