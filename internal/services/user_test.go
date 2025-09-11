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
	}
	return client, cleanup
}

func TestUserCRUD(t *testing.T) {
	client, cleanup := initTest(t)
	defer cleanup()

	service := NewUserService(client)
	ctx := context.Background()

	// create new user
	request := models.NewUserCreateRequest("bla", "süpersücret")
	user, err := service.Create(ctx, request)
	if err != nil {
		testutil.PrettyPrintError(err)
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

	service := NewUserService(client)
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

	service := NewUserService(client)
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

	service := NewUserService(client)
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

	service := NewUserService(client)
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

	service := NewUserService(client)
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

	service := NewUserService(client)
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

	service := NewUserService(client)
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

	service := NewUserService(client)
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

	service := NewUserService(client)
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
