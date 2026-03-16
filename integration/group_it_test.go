//go:build integration

package integration

import (
	"fmt"
	"testing"

	"github.com/rschmied/gocmlclient/pkg/models"
)

func TestIntegration_Group(t *testing.T) {
	cfg := LoadConfigFromEnv()
	c := newClient(t, cfg)
	requireReady(t, c, cfg)

	ctx, cancel := testContext(t, cfg)
	defer cancel()

	if cfg.AllowMutations {
		// Create a couple labs to associate permissions with.
		lab1 := createTempLab(t, c, cfg, "it-group-lab")
		lab2 := createTempLab(t, c, cfg, "it-group-lab")

		// Create a few users (preferred) or fall back to existing users.
		members := make([]models.UUID, 0, 3)
		if cfg.AllowUserMgmt {
			for i := range 3 {
				username := fmt.Sprintf("it-group-user-%d-%s", i, randSuffix(4))
				u, err := c.User.Create(ctx, models.NewUserCreateRequest(username, "integration-test-secret"))
				if err != nil {
					requireNoErrorOrSkipStatus(t, err, 403, 404)
				}
				members = append(members, u.ID)
				t.Cleanup(func() {
					cleanupCtx, cleanupCancel := testContext(t, cfg)
					defer cleanupCancel()
					_ = c.User.Delete(cleanupCtx, u.ID)
				})
			}
		} else {
			users, err := c.User.Users(ctx)
			if err == nil {
				for i := 0; i < len(users) && len(members) < 3; i++ {
					if users[i].ID != "" {
						members = append(members, users[i].ID)
					}
				}
			}
		}
		if members == nil {
			members = []models.UUID{}
		}

		name := "it-group-" + randSuffix(6)
		createReq := models.Group{
			Name:        name,
			Description: "integration test",
			Members:     members,
			Associations: []models.Association{
				{ID: lab1.ID, Permissions: models.Permissions{models.PermissionView, models.PermissionEdit}},
				{ID: lab2.ID, Permissions: models.Permissions{models.PermissionView}},
			},
		}
		created, err := c.Group.Create(ctx, createReq)
		if err != nil {
			requireNoErrorOrSkipStatus(t, err, 403, 404)
		}
		if created.ID == "" {
			t.Fatalf("Group.Create: expected id")
		}

		deleted := false
		t.Cleanup(func() {
			if deleted {
				return
			}
			cleanupCtx, cleanupCancel := testContext(t, cfg)
			defer cleanupCancel()
			_ = c.Group.Delete(cleanupCtx, string(created.ID))
		})

		// Read back via ID
		got, err := c.Group.GetByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("Group.GetByID(created): %v", err)
		}
		if got.ID != created.ID {
			t.Fatalf("Group.GetByID(created): id mismatch: %s != %s", got.ID, created.ID)
		}

		// ByName should work for the group we just created.
		_, err = c.Group.ByName(ctx, created.Name)
		if err != nil {
			requireNoErrorOrSkipStatus(t, err, 404)
		}

		// Update
		updateReq := models.Group{
			ID:          created.ID,
			Name:        created.Name,
			Description: "integration test (updated)",
			Members:     created.Members,
			Associations: func() []models.Association {
				if created.Associations != nil {
					return created.Associations
				}
				return []models.Association{}
			}(),
		}
		updated, err := c.Group.Update(ctx, updateReq)
		if err != nil {
			requireNoErrorOrSkipStatus(t, err, 403, 404)
		}
		_ = updated

		// Delete
		if err := c.Group.Delete(ctx, string(created.ID)); err != nil {
			requireNoErrorOrSkipStatus(t, err, 403, 404)
		}
		deleted = true

		return
	}

	// Read-only path (no mutations): just exercise list + get + optional by-name.
	groups, err := c.Group.Groups(ctx)
	if err != nil {
		t.Fatalf("Group.Groups: %v", err)
	}
	if len(groups) == 0 {
		t.Skip("no groups returned")
	}

	first := groups[0]
	if first.ID != "" {
		_, err := c.Group.GetByID(ctx, first.ID)
		if err != nil {
			t.Fatalf("Group.GetByID: %v", err)
		}
	}
	if first.Name != "" {
		_, err := c.Group.ByName(ctx, first.Name)
		if err != nil {
			requireNoErrorOrSkipStatus(t, err, 404)
		}
	}
}
