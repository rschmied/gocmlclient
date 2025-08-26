package models

import "encoding/json"

type ErrorResponse struct {
	Code        int             `json:"code"`
	Description json.RawMessage `json:"description"`
}

type UUID string

type Permissions []string

type (
	Association struct {
		ID          UUID        `json:"id"`
		Permissions Permissions `json:"permissions"`
	}
)

type AssociationList []*Association

type AssociationUsersGroups struct {
	Groups AssociationList `json:"groups"`
	Users  AssociationList `json:"users"`
}
