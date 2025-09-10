package models

type Permission string

const (
	PermissionAdmin Permission = "lab_admin"
	PermissionEdit  Permission = "lab_edit"
	PermissionExec  Permission = "lab_exec"
	PermissionView  Permission = "lab_view"
)

type Permissions []Permission

type (
	Association struct {
		ID          UUID        `json:"id"`
		Permissions Permissions `json:"permissions"`
	}
)

type AssociationList []Association

type AssociationUsersGroups struct {
	Groups AssociationList `json:"groups"`
	Users  AssociationList `json:"users"`
}
