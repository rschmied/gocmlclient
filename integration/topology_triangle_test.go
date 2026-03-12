//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rschmied/gocmlclient/pkg/models"
)

func TestIntegration_TriangleWithExtConnAndUnmanagedSwitch(t *testing.T) {
	// t.Skip("for now")
	cfg := LoadConfigFromEnv()
	if cfg.Timeout < 3*time.Minute {
		cfg.Timeout = 3 * time.Minute
	}

	// Node definition names are installation-dependent; allow override.
	iolDef := envString("CML_IOL_NODE_DEFINITION", "iol-xe")
	switchDef := envString("CML_UNMANAGED_SWITCH_NODE_DEFINITION", "unmanaged_switch")
	extDef := envString("CML_EXT_CONNECTOR_NODE_DEFINITION", "external_connector")
	extLabel := envString("CML_EXT_CONNECTOR_LABEL", "")
	extParamKey := envString("CML_EXT_CONNECTOR_PARAM_KEY", "external_connector_id")

	c := newClient(t, cfg)
	wireClientServices(c)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := c.System.Ready(ctx); err != nil {
		t.Fatalf("System.Ready: %v", err)
	}

	// Resolve external connector ID from label (or take first).
	extList, err := c.ExtConn.List(ctx)
	if err != nil {
		t.Fatalf("ExtConn.List: %v", err)
	}
	if len(extList) == 0 {
		t.Fatalf("no external connectors returned")
	}

	var ext *models.ExtConn
	if extLabel != "" {
		for _, e := range extList {
			if e != nil && e.Label == extLabel {
				ext = e
				break
			}
		}
		if ext == nil {
			t.Fatalf("external connector with label %q not found", extLabel)
		}
	} else {
		ext = extList[0]
	}

	labTitle := fmt.Sprintf("it-triangle-%d", time.Now().UnixNano())
	lab, err := c.Lab.Create(ctx, models.LabCreateRequest{Title: labTitle, Description: "integration test: triangle topology"})
	if err != nil {
		t.Fatalf("Lab.Create: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cleanupCancel()
		err = c.Lab.Stop(cleanupCtx, lab.ID)
		if err != nil {
			t.Logf("Lab.Stop(): %v", err)
		}
		converge(cleanupCtx, t, c, lab.ID)
		err = c.Lab.Wipe(cleanupCtx, lab.ID)
		if err != nil {
			t.Logf("Lab.Wipe(): %v", err)
		}
		err = c.Lab.Delete(cleanupCtx, lab.ID)
		if err != nil {
			t.Logf("Lab.Delete(): %v", err)
		}
	})

	// Create nodes.
	extNode, err := c.Node.Create(ctx, models.Node{LabID: lab.ID, Label: "ext", NodeDefinition: extDef, X: -400, Y: 0})
	if err != nil {
		t.Fatalf("Node.Create(ext): %v", err)
	}
	// Bind to a real external connector.
	extNode.Parameters = map[string]any{extParamKey: string(ext.ID)}
	if _, err := c.Node.Update(ctx, extNode); err != nil {
		t.Fatalf("Node.Update(ext parameters): %v", err)
	}

	swNode, err := c.Node.Create(ctx, models.Node{LabID: lab.ID, Label: "sw", NodeDefinition: switchDef, X: -150, Y: 0})
	if err != nil {
		t.Fatalf("Node.Create(sw): %v", err)
	}

	r1, err := c.Node.Create(ctx, models.Node{LabID: lab.ID, Label: "r1", NodeDefinition: iolDef, CPUs: 1, X: 150, Y: -150})
	if err != nil {
		t.Fatalf("Node.Create(r1): %v", err)
	}
	r2, err := c.Node.Create(ctx, models.Node{LabID: lab.ID, Label: "r2", NodeDefinition: iolDef, CPUs: 1, X: 350, Y: 0})
	if err != nil {
		t.Fatalf("Node.Create(r2): %v", err)
	}
	r3, err := c.Node.Create(ctx, models.Node{LabID: lab.ID, Label: "r3", NodeDefinition: iolDef, CPUs: 1, X: 150, Y: 150})
	if err != nil {
		t.Fatalf("Node.Create(r3): %v", err)
	}

	// Create links.
	linkSvc := c.Link
	linkSvc.Interface = c.Interface
	linkSvc.Node = c.Node

	// Each IOL -> unmanaged switch on first port (slot 0).
	_, err = linkSvc.Create(ctx, models.Link{LabID: lab.ID, SrcNode: r1.ID, DstNode: swNode.ID, SrcSlot: 0, DstSlot: -1})
	if err != nil {
		t.Fatalf("Link.Create(r1-sw): %v", err)
	}
	_, err = linkSvc.Create(ctx, models.Link{LabID: lab.ID, SrcNode: r2.ID, DstNode: swNode.ID, SrcSlot: 0, DstSlot: -1})
	if err != nil {
		t.Fatalf("Link.Create(r2-sw): %v", err)
	}
	_, err = linkSvc.Create(ctx, models.Link{LabID: lab.ID, SrcNode: r3.ID, DstNode: swNode.ID, SrcSlot: 0, DstSlot: -1})
	if err != nil {
		t.Fatalf("Link.Create(r3-sw): %v", err)
	}

	// Unmanaged switch <-> external connector (use next free on both).
	_, err = linkSvc.Create(ctx, models.Link{LabID: lab.ID, SrcNode: swNode.ID, DstNode: extNode.ID, SrcSlot: -1, DstSlot: -1})
	if err != nil {
		t.Fatalf("Link.Create(sw-ext): %v", err)
	}

	// Triangle links: r1<->r2 (slot 1), r2<->r3 (slot 1), r3<->r1 (slot 1).
	_, err = linkSvc.Create(ctx, models.Link{LabID: lab.ID, SrcNode: r1.ID, DstNode: r2.ID, SrcSlot: 1, DstSlot: 1})
	if err != nil {
		t.Fatalf("Link.Create(r1-r2): %v", err)
	}
	_, err = linkSvc.Create(ctx, models.Link{LabID: lab.ID, SrcNode: r2.ID, DstNode: r3.ID, SrcSlot: 2, DstSlot: 1})
	if err != nil {
		t.Fatalf("Link.Create(r2-r3): %v", err)
	}
	_, err = linkSvc.Create(ctx, models.Link{LabID: lab.ID, SrcNode: r3.ID, DstNode: r1.ID, SrcSlot: 2, DstSlot: 2})
	if err != nil {
		t.Fatalf("Link.Create(r3-r1): %v", err)
	}

	// Verify counts.
	loaded, err := c.Lab.GetByID(ctx, lab.ID, true)
	if err != nil {
		t.Fatalf("Lab.GetByID: %v", err)
	}
	if len(loaded.Nodes) != 5 {
		t.Fatalf("expected 5 nodes, got %d", len(loaded.Nodes))
	}
	if len(loaded.Links) != 7 {
		t.Fatalf("expected 7 links (3 iol->sw + sw->ext + iol triangle), got %d", len(loaded.Links))
	}

	// start the lab
	err = c.Lab.Start(ctx, lab.ID)
	if err != nil {
		t.Fatalf("Lab.Start(): %v", err)
	}
	// wait before done
	converge(ctx, t, c, lab.ID)
}
