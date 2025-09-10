// Package models provides the models for Cisco Modeling Labs
// here: user related types
package models

type UserBase struct {
	Username     string          `json:"username"`
	Fullname     string          `json:"fullname"`
	Description  string          `json:"description"`
	Email        string          `json:"email"`
	IsAdmin      bool            `json:"admin"`
	Groups       []UUID          `json:"groups,omitempty"`        // deprecated?
	Associations AssociationList `json:"associations,omitempty"`  // with 2.9.0
	ResourcePool *UUID           `json:"resource_pool,omitempty"` // with 2.5.0
	OptIn        *bool           `json:"opt_in,omitempty"`        // with 2.5.0
	TourVersion  string          `json:"tour_version"`
	PubkeyInfo   string          `json:"pubkey_info,omitempty"`
}

type UserCreateRequest struct {
	UserBase
	Password string `json:"password"`
}

func NewUserCreateRequest(username, password string) UserCreateRequest {
	return UserCreateRequest{
		UserBase: UserBase{
			Username: username,
		},
		Password: password,
	}
}

type UpdatePassword struct {
	Old string `json:"old_password"`
	New string `json:"new_password"`
}

type UserUpdateRequest struct {
	UserBase
	Password *UpdatePassword `json:"password,omitempty"`
}

type User struct {
	UserBase
	ID          UUID   `json:"id,omitempty"`
	Created     string `json:"created,omitempty"`
	Modified    string `json:"modified,omitempty"`
	DirectoryDN string `json:"directory_dn,omitempty"`
	Labs        []UUID `json:"labs,omitempty"`
}

type UserList []User
