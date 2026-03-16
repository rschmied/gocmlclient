// Package models provides the models for Cisco Modeling Labs
// here: permission related types
package models

// Permission represents a permission level for CML resources.
type Permission string

const (
	// PermissionAdmin grants administrative access.
	PermissionAdmin Permission = "lab_admin"
	// PermissionEdit allows editing resources.
	PermissionEdit Permission = "lab_edit"
	// PermissionExec allows executing operations.
	PermissionExec Permission = "lab_exec"
	// PermissionView allows viewing resources.
	PermissionView Permission = "lab_view"
)

// Permissions is a slice of Permission.
type Permissions []Permission

type (
	// Association represents an association between a user/group and permissions.
	Association struct {
		ID          UUID        `json:"id"`
		Permissions Permissions `json:"permissions"`
	}
)

// AssociationList is a slice of Association.
type AssociationList []Association

// AssociationUsersGroups contains associations for users and groups.
type AssociationUsersGroups struct {
	Groups AssociationList `json:"groups,omitempty"`
	Users  AssociationList `json:"users,omitempty"`
}
