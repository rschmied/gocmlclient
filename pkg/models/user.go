// Package models provides the models for Cisco Modeling Labs
// here: user related types
package models

// UserBase contains common user fields.
type UserBase struct {
	Username     string          `json:"username"`
	Fullname     string          `json:"fullname"`
	Description  string          `json:"description"`
	Email        string          `json:"email"`
	IsAdmin      bool            `json:"admin"`
	Groups       []UUID          `json:"groups,omitempty"`
	Associations AssociationList `json:"associations,omitempty"`  // with 2.9.0
	ResourcePool *UUID           `json:"resource_pool,omitempty"` // with 2.5.0
	OptIn        *bool           `json:"opt_in,omitempty"`        // with 2.5.0
	TourVersion  string          `json:"tour_version"`
	PubkeyInfo   string          `json:"pubkey_info,omitempty"`
}

// UserCreateRequest represents the data required to create a new user.
type UserCreateRequest struct {
	UserBase
	Password string `json:"password"`
}

// NewUserCreateRequest creates a new UserCreateRequest with the given username and password.
func NewUserCreateRequest(username, password string) UserCreateRequest {
	return UserCreateRequest{
		UserBase: UserBase{
			Username: username,
		},
		Password: password,
	}
}

// UpdatePassword represents a password update request.
type UpdatePassword struct {
	Old string `json:"old_password"`
	New string `json:"new_password"`
}

// UserUpdateRequest represents the data for updating a user.
type UserUpdateRequest struct {
	UserBase
	Password *UpdatePassword `json:"password,omitempty"`
}

// User represents a CML user with additional metadata.
type User struct {
	UserBase
	ID          UUID   `json:"id,omitempty"`
	Created     string `json:"created,omitempty"`
	Modified    string `json:"modified,omitempty"`
	DirectoryDN string `json:"directory_dn,omitempty"`
	Labs        []UUID `json:"labs,omitempty"`
	Password    string `json:"password,omitempty"` // For backward compatibility
}

// UserList is a slice of User.
type UserList []User
