package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermission_Constants(t *testing.T) {
	assert.Equal(t, Permission("lab_admin"), PermissionAdmin)
	assert.Equal(t, Permission("lab_edit"), PermissionEdit)
	assert.Equal(t, Permission("lab_exec"), PermissionExec)
	assert.Equal(t, Permission("lab_view"), PermissionView)
}

func TestPermissions_JSON(t *testing.T) {
	perms := Permissions{PermissionView, PermissionEdit, PermissionAdmin}

	// Test marshaling
	data, err := json.Marshal(perms)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled Permissions
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, perms, unmarshaled)
	assert.Len(t, unmarshaled, 3)
	assert.Contains(t, unmarshaled, PermissionView)
	assert.Contains(t, unmarshaled, PermissionEdit)
	assert.Contains(t, unmarshaled, PermissionAdmin)
}

func TestAssociation_JSON(t *testing.T) {
	assoc := Association{
		ID: "user123",
		Permissions: Permissions{
			PermissionView,
			PermissionEdit,
		},
	}

	// Test marshaling
	data, err := json.Marshal(assoc)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled Association
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, assoc.ID, unmarshaled.ID)
	assert.Equal(t, assoc.Permissions, unmarshaled.Permissions)
}

func TestAssociationList_JSON(t *testing.T) {
	associations := AssociationList{
		{
			ID:          "user1",
			Permissions: Permissions{PermissionView},
		},
		{
			ID:          "user2",
			Permissions: Permissions{PermissionEdit, PermissionExec},
		},
	}

	// Test marshaling
	data, err := json.Marshal(associations)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled AssociationList
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Len(t, unmarshaled, 2)
	assert.Equal(t, "user1", string(unmarshaled[0].ID))
	assert.Equal(t, "user2", string(unmarshaled[1].ID))
	assert.Contains(t, unmarshaled[0].Permissions, PermissionView)
	assert.Contains(t, unmarshaled[1].Permissions, PermissionEdit)
	assert.Contains(t, unmarshaled[1].Permissions, PermissionExec)
}

func TestAssociationUsersGroups_JSON(t *testing.T) {
	assoc := AssociationUsersGroups{
		Users: AssociationList{
			{
				ID:          "user1",
				Permissions: Permissions{PermissionView},
			},
		},
		Groups: AssociationList{
			{
				ID:          "group1",
				Permissions: Permissions{PermissionEdit},
			},
		},
	}

	// Test marshaling
	data, err := json.Marshal(assoc)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled AssociationUsersGroups
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Len(t, unmarshaled.Users, 1)
	assert.Len(t, unmarshaled.Groups, 1)
	assert.Equal(t, "user1", string(unmarshaled.Users[0].ID))
	assert.Equal(t, "group1", string(unmarshaled.Groups[0].ID))
}

func TestAssociationUsersGroups_JSON_Empty(t *testing.T) {
	assoc := AssociationUsersGroups{}

	// Test marshaling
	data, err := json.Marshal(assoc)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled AssociationUsersGroups
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Len(t, unmarshaled.Users, 0)
	assert.Len(t, unmarshaled.Groups, 0)
}
