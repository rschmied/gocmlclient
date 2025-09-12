package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/testutil"
	"github.com/rschmied/gocmlclient/pkg/models"
	"github.com/stretchr/testify/assert"
)

func initGroupTest(t *testing.T) (*api.Client, func()) {
	client, cleanup := testutil.NewAPIClient(t)
	if !testutil.IsLiveTesting() {
		group1 := `{
			"id": "group-1",
			"name": "admin-group",
			"description": "Administrators group",
			"members": ["user-1", "user-2"],
			"asocciations": []
		}`
		group2 := `{
			"id": "group-2", 
			"name": "user-group",
			"description": "Regular users group",
			"members": ["user-3"],
			"asocciations": []
		}`
		groupList := httpmock.NewStringResponder(200, fmt.Sprintf("[%s,%s]", group1, group2))
		groupResponse := httpmock.NewStringResponder(200, group2)
		groupCreateResponse := httpmock.NewStringResponder(200, group2)
		groupIDResponse := httpmock.NewStringResponder(200, `"group-2"`)
		groupUpdateResponse := httpmock.NewStringResponder(200, `{
			"id": "group-2", 
			"name": "user-group",
			"description": "Updated regular users group",
			"members": ["user-3", "user-4"],
			"asocciations": []
		}`)
		httpmock.RegisterResponder("GET", "https://mock/api/v0/groups/user-group/id", groupIDResponse)
		httpmock.RegisterResponder("GET", "https://mock/api/v0/groups", groupList)
		httpmock.RegisterResponder("POST", "https://mock/api/v0/groups", groupCreateResponse)
		httpmock.RegisterResponder("GET", "https://mock/api/v0/groups/group-2", groupResponse)
		httpmock.RegisterResponder("DELETE", "https://mock/api/v0/groups/group-2",
			httpmock.NewJsonResponderOrPanic(204, nil))
		httpmock.RegisterResponder("PATCH", "https://mock/api/v0/groups/group-2", groupUpdateResponse)
	}
	return client, cleanup
}

func TestGroupCRUD(t *testing.T) {
	client, cleanup := initGroupTest(t)
	defer cleanup()

	service := NewGroupService(client)
	ctx := context.Background()

	// create new group
	group := models.Group{
		Name:        "user-group",
		Description: "Regular users group",
	}

	// Only add members for mock testing (live server requires valid UUIDs)
	if !testutil.IsLiveTesting() {
		group.Members = []models.UUID{"user-3"}
	}

	createdGroup, err := service.Create(ctx, group)
	if err != nil {
		testutil.PrettyPrintError(err)
	}
	assert.NoError(t, err)

	// get group by name
	namedGroup, err := service.ByName(ctx, "user-group")
	assert.NoError(t, err)
	assert.Equal(t, createdGroup.ID, namedGroup.ID)

	// list all groups
	groups, err := service.Groups(ctx)
	groupCount := len(groups)
	assert.NoError(t, err)
	assert.Greater(t, groupCount, 1)

	// update the group - for live tests, just change the description
	updateGroup := models.Group{
		ID:          createdGroup.ID,
		Name:        "user-group",
		Description: "Updated regular users group",
	}

	// Only update members for mock testing
	if !testutil.IsLiveTesting() {
		updateGroup.Members = []models.UUID{"user-3", "user-4"}
	}

	updatedGroup, err := service.Update(ctx, updateGroup)
	assert.NoError(t, err)
	assert.Equal(t, createdGroup.ID, updatedGroup.ID)
	assert.Equal(t, "Updated regular users group", updatedGroup.Description)

	// delete the group
	err = service.Delete(ctx, string(createdGroup.ID))
	assert.NoError(t, err)
}

func TestGroupGetByID_NotFound(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - UUID validation differs")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 404 response for non-existent group
	httpmock.RegisterResponder("GET", "https://mock/api/v0/groups/nonexistent",
		httpmock.NewStringResponder(404, `{"message": "Group not found"}`))

	service := NewGroupService(client)
	ctx := context.Background()

	_, err := service.GetByID(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "Group not found")
}

func TestGroupGetByName_NotFound(t *testing.T) {
	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 404 response for non-existent group name lookup
	httpmock.RegisterResponder("GET", "https://mock/api/v0/groups/nonexistent/id",
		httpmock.NewStringResponder(404, `{"description": "Group does not exist: nonexistent.","code":404}`))

	service := NewGroupService(client)
	ctx := context.Background()

	_, err := service.ByName(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "Group does not exist")
}

func TestGroupCreate_ValidationError(t *testing.T) {
	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 400 validation error for invalid group creation
	httpmock.RegisterResponder("POST", "https://mock/api/v0/groups",
		httpmock.NewStringResponder(400, `{"description": "{\"Input validation failed\": [{\"location\": [\"body\", \"name\"], \"type\": \"string_too_short\", \"message\": \"String should have at least 1 character\"}]}","code":400}`))

	service := NewGroupService(client)
	ctx := context.Background()

	// Create group with missing required fields
	invalidGroup := models.Group{}
	_, err := service.Create(ctx, invalidGroup)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "Input validation failed")
}

func TestGroupUpdate_NotFound(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - UUID validation differs")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 404 response for updating non-existent group
	httpmock.RegisterResponder("PATCH", "https://mock/api/v0/groups/nonexistent",
		httpmock.NewStringResponder(404, `{"message": "Group not found"}`))

	service := NewGroupService(client)
	ctx := context.Background()

	updateGroup := models.Group{
		ID:   "nonexistent",
		Name: "test-group",
	}
	_, err := service.Update(ctx, updateGroup)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "Group not found")
}

func TestGroupDelete_NotFound(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - UUID validation differs")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 404 response for deleting non-existent group
	httpmock.RegisterResponder("DELETE", "https://mock/api/v0/groups/nonexistent",
		httpmock.NewStringResponder(404, `{"message": "Group not found"}`))

	service := NewGroupService(client)
	ctx := context.Background()

	err := service.Delete(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "Group not found")
}

func TestGroupServerError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - can't force errors")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 500 internal server error
	httpmock.RegisterResponder("GET", "https://mock/api/v0/groups",
		httpmock.NewStringResponder(500, `{"message": "Internal server error"}`))

	service := NewGroupService(client)
	ctx := context.Background()

	_, err := service.Groups(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "Internal server error")
}

func TestGroupAuthError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - authentication succeeds")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 401 unauthorized error
	httpmock.RegisterResponder("GET", "https://mock/api/v0/groups",
		httpmock.NewStringResponder(401, `{"message": "Unauthorized"}`))

	service := NewGroupService(client)
	ctx := context.Background()

	_, err := service.Groups(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
	assert.Contains(t, err.Error(), "Unauthorized")
}

func TestGroupPermissionError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - permissions are sufficient")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 403 forbidden error
	httpmock.RegisterResponder("GET", "https://mock/api/v0/groups",
		httpmock.NewStringResponder(403, `{"message": "Forbidden"}`))

	service := NewGroupService(client)
	ctx := context.Background()

	_, err := service.Groups(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "403")
	assert.Contains(t, err.Error(), "Forbidden")
}

func TestGroupMalformedJSON(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - returns valid JSON")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock response with malformed JSON
	httpmock.RegisterResponder("GET", "https://mock/api/v0/groups",
		httpmock.NewStringResponder(200, `[{"id": "invalid-json"`)) // Missing closing brace

	service := NewGroupService(client)
	ctx := context.Background()

	_, err := service.Groups(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode response")
}

func TestGroupConnectionError(t *testing.T) {
	// This test requires special setup to simulate connection errors
	// Typically handled by the API client's connection error wrapping
	t.Skip("Connection error testing requires special network setup")
}
