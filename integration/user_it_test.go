//go:build integration

package integration

import (
	"testing"

	"github.com/rschmied/gocmlclient/pkg/models"
)

func TestIntegration_User(t *testing.T) {
	cfg := LoadConfigFromEnv()
	c := newClient(t, cfg)
	requireReady(t, c, cfg)

	ctx, cancel := testContext(t, cfg)
	defer cancel()

	if cfg.AllowUserMgmt {
		username := "it-user-" + randSuffix(6)
		created, err := c.User.Create(ctx, models.NewUserCreateRequest(username, "integration-test-secret"))
		if err != nil {
			requireNoErrorOrSkipStatus(t, err, 403, 404)
		}
		if created.ID == "" {
			t.Fatalf("User.Create: expected id")
		}

		deleted := false
		t.Cleanup(func() {
			if deleted {
				return
			}
			cleanupCtx, cleanupCancel := testContext(t, cfg)
			defer cleanupCancel()
			_ = c.User.Delete(cleanupCtx, created.ID)
		})

		// Read back
		_, err = c.User.GetByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("User.GetByID(created): %v", err)
		}
		_, err = c.User.GetByName(ctx, created.Username)
		if err != nil {
			requireNoErrorOrSkipStatus(t, err, 404)
		}
		_, err = c.User.Groups(ctx, created.ID)
		if err != nil {
			requireNoErrorOrSkipStatus(t, err, 404)
		}

		// Update
		upd := models.UserUpdateRequest{UserBase: created.UserBase}
		upd.Description = "integration test (updated)"
		_, err = c.User.Update(ctx, created.ID, upd)
		if err != nil {
			requireNoErrorOrSkipStatus(t, err, 403, 404)
		}

		// Delete
		if err := c.User.Delete(ctx, created.ID); err != nil {
			requireNoErrorOrSkipStatus(t, err, 403, 404)
		}
		deleted = true

		// Optional: list once to exercise endpoint post-mutation
		_, _ = c.User.Users(ctx)
		return
	}

	// Read-only path
	users, err := c.User.Users(ctx)
	if err != nil {
		t.Fatalf("User.Users: %v", err)
	}
	if len(users) == 0 {
		t.Skip("no users returned")
	}
	first := users[0]
	if first.ID != "" {
		_, err := c.User.GetByID(ctx, first.ID)
		if err != nil {
			t.Fatalf("User.GetByID: %v", err)
		}
		_, err = c.User.Groups(ctx, first.ID)
		if err != nil {
			requireNoErrorOrSkipStatus(t, err, 404)
		}
	}
	if first.Username != "" {
		_, err := c.User.GetByName(ctx, first.Username)
		if err != nil {
			requireNoErrorOrSkipStatus(t, err, 404)
		}
	}
}
