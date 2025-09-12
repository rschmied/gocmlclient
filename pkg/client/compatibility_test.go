package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rschmied/gocmlclient/pkg/models"
	"github.com/stretchr/testify/assert"
)

// TestCompatibilityCoreFunctionality tests the essential compatibility layer behaviors
func TestCompatibilityCoreFunctionality(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v0/auth_extended":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"user-123","username":"testuser","token":"mock-token-12345","admin":false}`))
		case "/api/v0/system_information":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"version": "2.5.0", "ready": true}`))
		case "/api/v0/labs/test-lab-id":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "test-lab-id",
				"lab_title": "Test Lab",
				"lab_description": "A test lab",
				"created": "2025-01-01T00:00:00Z",
				"modified": "2025-01-01T00:00:00Z"
			}`))
		case "/api/v0/labs/test-lab-id/state/start":
			if r.Method == "PUT" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{}`))
			}
		case "/api/v0/users":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{"id": "user1", "username": "user1", "admin": false},
				{"id": "user2", "username": "user2", "admin": true}
			]`))
		case "/api/v0/labs/nonexistent-lab":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Lab not found"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := New(server.URL, SkipReadyCheck())
	assert.NoError(t, err)

	ctx := context.Background()

	t.Run("UUID conversion works", func(t *testing.T) {
		// Test that string IDs are properly converted to UUIDs
		lab, err := client.LabGet(ctx, "test-lab-id", false)
		assert.NoError(t, err)
		assert.Equal(t, models.UUID("test-lab-id"), lab.ID)
		assert.Equal(t, "Test Lab", lab.Title)
	})

	t.Run("Error propagation works", func(t *testing.T) {
		// Test that service errors are properly propagated
		_, err := client.LabGet(ctx, "nonexistent-lab", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})

	t.Run("System methods work", func(t *testing.T) {
		// Test system-level compatibility methods
		err := client.Ready(ctx)
		assert.NoError(t, err)

		version := client.Version()
		assert.Equal(t, "2.5.0", version)
	})

	t.Run("Collection methods work", func(t *testing.T) {
		// Test methods that return collections
		users, err := client.Users(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		// Verify the conversion from UserList to []*User works
		assert.Equal(t, "user2", users[0].Username) // sorted by ID desc
		assert.Equal(t, "user1", users[1].Username)
	})

	t.Run("Action methods work", func(t *testing.T) {
		// Test action methods (start/stop/etc)
		err := client.LabStart(ctx, "test-lab-id")
		assert.NoError(t, err)
	})
}

// TestCompatibilityEdgeCases tests basic error conditions
func TestCompatibilityEdgeCases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v0/auth_extended":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"user-123","username":"testuser","token":"mock-token-12345","admin":false}`))
		case "/api/v0/labs/empty-id":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": "", "lab_title": "Empty Lab"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := New(server.URL, SkipReadyCheck())
	assert.NoError(t, err)

	ctx := context.Background()

	t.Run("Empty ID handling", func(t *testing.T) {
		// Test that empty string IDs are handled (though they'll likely fail at API level)
		_, err := client.LabGet(ctx, "", false)
		// This will likely fail, but importantly shouldn't panic
		assert.Error(t, err) // Expected to fail with empty ID
	})
}
