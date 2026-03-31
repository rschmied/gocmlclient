package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/testutil"
	"github.com/rschmied/gocmlclient/pkg/models"
)

// mockGroupService implements GroupServiceInterface for testing
type mockGroupService struct {
	client *api.Client
}

func (m *mockGroupService) GetByID(ctx context.Context, id models.UUID) (models.Group, error) {
	group := models.Group{}
	api := fmt.Sprintf("groups/%s", id)
	err := m.client.GetJSON(ctx, api, nil, &group)
	return group, err
}

func initTest(t *testing.T) (*api.Client, func()) {
	client, cleanup := testutil.NewAPIClient(t)
	if !testutil.IsLiveTesting() {
		user1 := `{
			"id": "admin_id",
			"created": "2025-09-10T07:21:49+00:00",
			"modified": "2025-09-10T07:21:49+00:00",
			"username": "admin",
			"fullname": "The Super User",
			"email": "",
			"description": "",
			"admin": true,
			"directory_dn": "",
			"groups": [],
			"labs": [],
			"associations": [],
			"opt_in": null,
			"resource_pool": null,
			"tour_version": "",
			"pubkey_info": ""
		}`
		user2 := `{
			"id": "user_id",
			"created": "2025-09-10T07:21:49+00:00",
			"modified": "2025-09-10T07:21:49+00:00",
			"username": "bla",
			"fullname": "",
			"email": "",
			"description": "",
			"admin": false,
			"directory_dn": "",
			"groups": [],
			"labs": [],
			"associations": [],
			"opt_in": null,
			"resource_pool": null,
			"tour_version": "",
			"pubkey_info": ""
		}`
		userList := httpmock.NewStringResponder(200, fmt.Sprintf("[%s,%s]", user1, user2))
		userResponse := httpmock.NewStringResponder(200, user2)
		userIDResponse := httpmock.NewStringResponder(200, `"user_id"`)
		httpmock.RegisterResponder("GET", "https://mock/api/v0/users/bla/id", userIDResponse)
		httpmock.RegisterResponder("GET", "https://mock/api/v0/users", userList)
		httpmock.RegisterResponder("POST", "https://mock/api/v0/users", userResponse)
		httpmock.RegisterResponder("GET", "https://mock/api/v0/users/user_id", userResponse)
		httpmock.RegisterResponder("DELETE", "https://mock/api/v0/users/user_id",
			httpmock.NewJsonResponderOrPanic(204, nil))
		httpmock.RegisterResponder("PATCH", "https://mock/api/v0/users/user_id", userResponse)

		// Group mocks for Groups method
		group1 := `{
			"id": "group1_id",
			"created": "2025-09-10T07:21:49+00:00",
			"modified": "2025-09-10T07:21:49+00:00",
			"name": "test_group_1",
			"description": "Test group 1",
			"members": ["user_id"],
			"directory_dn": "",
			"directory_exists": false
		}`
		group2 := `{
			"id": "group2_id",
			"created": "2025-09-10T07:21:49+00:00",
			"modified": "2025-09-10T07:21:49+00:00",
			"name": "test_group_2",
			"description": "Test group 2",
			"members": ["user_id"],
			"directory_dn": "",
			"directory_exists": false
		}`
		groupsList := httpmock.NewStringResponder(200, `["group1_id", "group2_id"]`)
		group1Response := httpmock.NewStringResponder(200, group1)
		group2Response := httpmock.NewStringResponder(200, group2)
		httpmock.RegisterResponder("GET", "https://mock/api/v0/users/user_id/groups", groupsList)
		httpmock.RegisterResponder("GET", "https://mock/api/v0/groups/group1_id", group1Response)
		httpmock.RegisterResponder("GET", "https://mock/api/v0/groups/group2_id", group2Response)
	}
	return client, cleanup
}

func TestUserCRUD(t *testing.T) {
	client, cleanup := initTest(t)
	defer cleanup()

	service := NewUserService(client, nil)
	ctx := context.Background()

	// create new user
	request := models.NewUserCreateRequest("bla", "süpersücret")
	user, err := service.Create(ctx, request)
	if err != nil {
		_ = testutil.PrettyPrintError(err)
	}
	assert.NoError(t, err)

	// get user by name
	namedUser, err := service.GetByName(ctx, "bla")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, namedUser.ID)

	// list all users
	users, err := service.Users(ctx)
	userCount := len(users)
	assert.NoError(t, err)
	assert.Greater(t, userCount, 1)

	// list all users
	updateRequest := models.UserUpdateRequest{
		UserBase: user.UserBase,
		Password: &models.UpdatePassword{
			Old: "süpersücret",
			New: "extremelysücret",
		},
	}
	updatedUser, err := service.Update(ctx, user.ID, updateRequest)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, updatedUser.ID)

	// delete the user
	err = service.Delete(ctx, user.ID)
	assert.NoError(t, err)
}

func TestUserGetByID_NotFound(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - UUID validation differs")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 404 response for non-existent user
	httpmock.RegisterResponder("GET", "https://mock/api/v0/users/nonexistent",
		httpmock.NewStringResponder(404, `{"message": "User not found"}`))

	service := NewUserService(client, nil)
	ctx := context.Background()

	_, err := service.GetByID(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "User not found")
}

func TestUserGetByName_NotFound(t *testing.T) {
	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 404 response for non-existent user name lookup
	httpmock.RegisterResponder("GET", "https://mock/api/v0/users/nonexistent/id",
		httpmock.NewStringResponder(404, `{"description": "User does not exist: nonexistent.","code":404}`))

	service := NewUserService(client, nil)
	ctx := context.Background()

	_, err := service.GetByName(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "User does not exist")
}

func TestUserCreate_ValidationError(t *testing.T) {
	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 400 validation error for invalid user creation
	httpmock.RegisterResponder("POST", "https://mock/api/v0/users",
		httpmock.NewStringResponder(400, `{"description": "{\"Input validation failed\": [{\"location\": [\"body\", \"username\"], \"type\": \"string_too_short\", \"message\": \"String should have at least 1 character\"}]}","code":400}`))

	service := NewUserService(client, nil)
	ctx := context.Background()

	// Create request with missing required fields
	invalidRequest := models.UserCreateRequest{}
	_, err := service.Create(ctx, invalidRequest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "Input validation failed")
}

func TestUserUpdate_NotFound(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - UUID validation differs")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 404 response for updating non-existent user
	httpmock.RegisterResponder("PATCH", "https://mock/api/v0/users/nonexistent",
		httpmock.NewStringResponder(404, `{"message": "User not found"}`))

	service := NewUserService(client, nil)
	ctx := context.Background()

	updateRequest := models.UserUpdateRequest{
		UserBase: models.UserBase{
			Username: "test",
		},
	}
	_, err := service.Update(ctx, "nonexistent", updateRequest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "User not found")
}

func TestUserDelete_NotFound(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - UUID validation differs")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 404 response for deleting non-existent user
	httpmock.RegisterResponder("DELETE", "https://mock/api/v0/users/nonexistent",
		httpmock.NewStringResponder(404, `{"message": "User not found"}`))

	service := NewUserService(client, nil)
	ctx := context.Background()

	err := service.Delete(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "User not found")
}

func TestUserServerError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - can't force errors")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 500 internal server error
	httpmock.RegisterResponder("GET", "https://mock/api/v0/users",
		httpmock.NewStringResponder(500, `{"message": "Internal server error"}`))

	service := NewUserService(client, nil)
	ctx := context.Background()

	_, err := service.Users(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "Internal server error")
}

func TestUserAuthError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - authentication succeeds")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 401 unauthorized error
	httpmock.RegisterResponder("GET", "https://mock/api/v0/users",
		httpmock.NewStringResponder(401, `{"message": "Unauthorized"}`))

	service := NewUserService(client, nil)
	ctx := context.Background()

	_, err := service.Users(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
	assert.Contains(t, err.Error(), "Unauthorized")
}

func TestUserPermissionError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - permissions are sufficient")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 403 forbidden error
	httpmock.RegisterResponder("GET", "https://mock/api/v0/users",
		httpmock.NewStringResponder(403, `{"message": "Forbidden"}`))

	service := NewUserService(client, nil)
	ctx := context.Background()

	_, err := service.Users(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "403")
	assert.Contains(t, err.Error(), "Forbidden")
}

func TestUserMalformedJSON(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - returns valid JSON")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock response with malformed JSON
	httpmock.RegisterResponder("GET", "https://mock/api/v0/users",
		httpmock.NewStringResponder(200, `[{"id": "invalid-json"`)) // Missing closing brace

	service := NewUserService(client, nil)
	ctx := context.Background()

	_, err := service.Users(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode response")
}

func TestUserConnectionError(t *testing.T) {
	// This test requires special setup to simulate connection errors
	// Typically handled by the API client's connection error wrapping
	t.Skip("Connection error testing requires special network setup")
}

func TestUserGroups(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - requires valid UUID user ID")
	}

	client, cleanup := initTest(t)
	defer cleanup()

	service := NewUserService(client, &mockGroupService{client: client})
	ctx := context.Background()

	groups, err := service.Groups(ctx, "user_id")
	assert.NoError(t, err)
	assert.Len(t, groups, 2)
	// Sorted by ID descending: group2_id > group1_id
	assert.Equal(t, "test_group_2", groups[0].Name)
	assert.Equal(t, "test_group_1", groups[1].Name)
}

func TestUserGroups_GroupNotFound(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - can't mock group not found")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock groups list with a non-existent group ID
	httpmock.RegisterResponder("GET", "https://mock/api/v0/users/user_id/groups",
		httpmock.NewStringResponder(200, `["group1_id", "nonexistent_group"]`))
	// Mock existing group
	httpmock.RegisterResponder("GET", "https://mock/api/v0/groups/group1_id",
		httpmock.NewStringResponder(200, `{
			"id": "group1_id",
			"created": "2025-09-10T07:21:49+00:00",
			"modified": "2025-09-10T07:21:49+00:00",
			"name": "test_group_1",
			"description": "Test group 1",
			"members": ["user_id"],
			"directory_dn": "",
			"directory_exists": false
		}`))
	// Mock 404 for non-existent group
	httpmock.RegisterResponder("GET", "https://mock/api/v0/groups/nonexistent_group",
		httpmock.NewStringResponder(404, `{"message": "Group not found"}`))

	service := NewUserService(client, &mockGroupService{client: client})
	ctx := context.Background()

	_, err := service.Groups(ctx, "user_id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "Group not found")
}

func TestUserGroups_Empty(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - requires valid UUID user ID")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock empty groups list
	httpmock.RegisterResponder("GET", "https://mock/api/v0/users/user_id/groups",
		httpmock.NewStringResponder(200, `[]`))

	service := NewUserService(client, &mockGroupService{client: client})
	ctx := context.Background()

	groups, err := service.Groups(ctx, "user_id")
	assert.NoError(t, err)
	assert.Len(t, groups, 0)
}

func TestUserGroups_FetchError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - can't force groups list error")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock 500 error for groups list
	httpmock.RegisterResponder("GET", "https://mock/api/v0/users/user_id/groups",
		httpmock.NewStringResponder(500, `{"message": "Internal server error"}`))

	service := NewUserService(client, &mockGroupService{client: client})
	ctx := context.Background()

	_, err := service.Groups(ctx, "user_id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "Internal server error")
}

func TestUserGroups_EndpointNotFound(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - can't force endpoint mismatch")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("GET", "https://mock/api/v0/users/user_id/groups",
		httpmock.NewStringResponder(404, `{"description": "Not Found","code":404}`))

	service := NewUserService(client, &mockGroupService{client: client})
	ctx := context.Background()

	_, err := service.Groups(ctx, "user_id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestUserGetByID_Success(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - requires valid UUID user ID")
	}

	client, cleanup := initTest(t)
	defer cleanup()

	service := NewUserService(client, nil)
	ctx := context.Background()

	user, err := service.GetByID(ctx, "user_id")
	assert.NoError(t, err)
	assert.Equal(t, "bla", user.Username)
	assert.Equal(t, models.UUID("user_id"), user.ID)
}

func TestNewUserService(t *testing.T) {
	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	service := NewUserService(client, nil)
	assert.NotNil(t, service)
	assert.Equal(t, client, service.apiClient)
	assert.Nil(t, service.Group)
}
