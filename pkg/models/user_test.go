package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUserCreateRequest(t *testing.T) {
	username := "testuser"
	password := "testpass"

	request := NewUserCreateRequest(username, password)

	assert.Equal(t, username, request.Username)
	assert.Equal(t, password, request.Password)
	assert.Equal(t, UserBase{Username: username}, request.UserBase)
}

// uuidPtr creates a pointer to a UUID from a string
func uuidPtr(id string) *UUID {
	u := UUID(id)
	return &u
}

func TestUserBase_JSON(t *testing.T) {
	userBase := UserBase{
		Username:     "testuser",
		Fullname:     "Test User",
		Description:  "A test user",
		Email:        "test@example.com",
		IsAdmin:      true,
		Groups:       []UUID{"group1", "group2"},
		ResourcePool: uuidPtr("pool1"),
		OptIn:        boolPtr(true),
		TourVersion:  "1.0.0",
		PubkeyInfo:   "ssh-rsa AAAAB3NzaC1yc...",
	}

	// Test marshaling
	data, err := json.Marshal(userBase)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled UserBase
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, userBase.Username, unmarshaled.Username)
	assert.Equal(t, userBase.Fullname, unmarshaled.Fullname)
	assert.Equal(t, userBase.Description, unmarshaled.Description)
	assert.Equal(t, userBase.Email, unmarshaled.Email)
	assert.Equal(t, userBase.IsAdmin, unmarshaled.IsAdmin)
	assert.Equal(t, userBase.Groups, unmarshaled.Groups)
	assert.Equal(t, userBase.ResourcePool, unmarshaled.ResourcePool)
	assert.Equal(t, userBase.OptIn, unmarshaled.OptIn)
	assert.Equal(t, userBase.TourVersion, unmarshaled.TourVersion)
	assert.Equal(t, userBase.PubkeyInfo, unmarshaled.PubkeyInfo)
}

func TestUserCreateRequest_JSON(t *testing.T) {
	request := UserCreateRequest{
		UserBase: UserBase{
			Username: "newuser",
			Fullname: "New User",
			Email:    "new@example.com",
		},
		Password: "secret123",
	}

	// Test marshaling
	data, err := json.Marshal(request)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled UserCreateRequest
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, request.Username, unmarshaled.Username)
	assert.Equal(t, request.Fullname, unmarshaled.Fullname)
	assert.Equal(t, request.Email, unmarshaled.Email)
	assert.Equal(t, request.Password, unmarshaled.Password)
}

func TestUpdatePassword_JSON(t *testing.T) {
	update := UpdatePassword{
		Old: "oldpass",
		New: "newpass",
	}

	// Test marshaling
	data, err := json.Marshal(update)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled UpdatePassword
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, update.Old, unmarshaled.Old)
	assert.Equal(t, update.New, unmarshaled.New)
}

func TestUserUpdateRequest_JSON(t *testing.T) {
	request := UserUpdateRequest{
		UserBase: UserBase{
			Username: "testuser",
			Fullname: "Updated User",
		},
		Password: &UpdatePassword{
			Old: "oldpass",
			New: "newpass",
		},
	}

	// Test marshaling
	data, err := json.Marshal(request)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled UserUpdateRequest
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, request.Username, unmarshaled.Username)
	assert.Equal(t, request.Fullname, unmarshaled.Fullname)
	assert.NotNil(t, unmarshaled.Password)
	assert.Equal(t, request.Password.Old, unmarshaled.Password.Old)
	assert.Equal(t, request.Password.New, unmarshaled.Password.New)
}

func TestUserUpdateRequest_JSON_WithoutPassword(t *testing.T) {
	request := UserUpdateRequest{
		UserBase: UserBase{
			Username: "testuser",
			Fullname: "Updated User",
		},
	}

	// Test marshaling
	data, err := json.Marshal(request)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled UserUpdateRequest
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, request.Username, unmarshaled.Username)
	assert.Equal(t, request.Fullname, unmarshaled.Fullname)
	assert.Nil(t, unmarshaled.Password)
}

func TestUser_JSON(t *testing.T) {
	user := User{
		UserBase: UserBase{
			Username: "testuser",
			Fullname: "Test User",
			Email:    "test@example.com",
		},
		ID:       "user123",
		Created:  "2023-01-01T00:00:00Z",
		Modified: "2023-01-02T00:00:00Z",
		Labs:     []UUID{"lab1", "lab2"},
	}

	// Test marshaling
	data, err := json.Marshal(user)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled User
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, user.ID, unmarshaled.ID)
	assert.Equal(t, user.Username, unmarshaled.Username)
	assert.Equal(t, user.Fullname, unmarshaled.Fullname)
	assert.Equal(t, user.Email, unmarshaled.Email)
	assert.Equal(t, user.Created, unmarshaled.Created)
	assert.Equal(t, user.Modified, unmarshaled.Modified)
	assert.Equal(t, user.Labs, unmarshaled.Labs)
}

func TestUserList_JSON(t *testing.T) {
	users := UserList{
		{
			UserBase: UserBase{Username: "user1", Fullname: "User One"},
			ID:       "user1",
		},
		{
			UserBase: UserBase{Username: "user2", Fullname: "User Two"},
			ID:       "user2",
		},
	}

	// Test marshaling
	data, err := json.Marshal(users)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled UserList
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Len(t, unmarshaled, 2)
	assert.Equal(t, "user1", string(unmarshaled[0].ID))
	assert.Equal(t, "user2", string(unmarshaled[1].ID))
	assert.Equal(t, "User One", unmarshaled[0].Fullname)
	assert.Equal(t, "User Two", unmarshaled[1].Fullname)
}

func TestUser_JSON_WithAssociations(t *testing.T) {
	user := User{
		UserBase: UserBase{
			Username: "testuser",
			Associations: AssociationList{
				{
					ID:          "lab1",
					Permissions: Permissions{PermissionView, PermissionEdit},
				},
			},
		},
		ID: "user123",
	}

	// Test marshaling
	data, err := json.Marshal(user)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled User
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Len(t, unmarshaled.Associations, 1)
	assert.Equal(t, "lab1", string(unmarshaled.Associations[0].ID))
	assert.Contains(t, unmarshaled.Associations[0].Permissions, PermissionView)
	assert.Contains(t, unmarshaled.Associations[0].Permissions, PermissionEdit)
}
