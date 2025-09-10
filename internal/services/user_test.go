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

	service := NewUserService(client, nil)
	ctx := context.Background()

	// create new user
	request := models.NewUserCreateRequest("bla", "suepersuecret")
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
			Old: "suepersuecret",
			New: "extremlysuecret",
		},
	}
	updatedUser, err := service.Update(ctx, user.ID, updateRequest)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, updatedUser.ID)

	// delete the user
	err = service.Delete(ctx, user.ID)
	assert.NoError(t, err)
}
