//go:build integration

package integration

import (
	"context"
	"testing"
	"time"
)

func TestIntegration_Smoke(t *testing.T) {
	cfg := LoadConfigFromEnv()
	c := newClient(t, cfg)
	wireClientServices(c)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := c.System.Ready(ctx); err != nil {
		t.Fatalf("System.Ready: %v", err)
	}

	// Basic authenticated call
	_, err := c.Lab.Labs(ctx, true)
	if err != nil {
		t.Fatalf("Lab.Labs: %v", err)
	}
}

func TestIntegration_LabImportFromFiles(t *testing.T) {
	cfg := LoadConfigFromEnv()
	if len(cfg.LabTopologyFiles) == 0 {
		t.Skip("set CML_LAB_TOPOLOGY_FILES to run")
	}
	if cfg.Timeout < 2*time.Minute {
		cfg.Timeout = 2 * time.Minute
	}

	c := newClient(t, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	for _, file := range cfg.LabTopologyFiles {
		// file := file
		t.Run(file, func(t *testing.T) {
			topo, err := readFile(file)
			if err != nil {
				t.Fatalf("read topology: %v", err)
			}

			lab, err := c.Lab.Import(ctx, topo)
			if err != nil {
				t.Fatalf("Lab.Import: %v", err)
			}

			// Best-effort cleanup.
			t.Cleanup(func() {
				cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Minute)
				defer cleanupCancel()
				_ = c.Lab.Delete(cleanupCtx, lab.ID)
			})

			// Sanity check: imported lab must be retrievable.
			_, err = c.Lab.GetByID(ctx, lab.ID, true)
			if err != nil {
				t.Fatalf("Lab.GetByID: %v", err)
			}
		})
	}
}
