// Package models provides the models for Cisco Modeling Labs
// here: user related types
package models

import (
	"encoding/json"
	"strings"
)

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
	OptIn        *OptInState     `json:"opt_in,omitempty"`        // with 2.5.0
	TourVersion  string          `json:"tour_version"`
	PubkeyInfo   string          `json:"pubkey_info,omitempty"`
}

type OptInState string

const (
	OptInAccepted OptInState = "accepted"
	OptInDeclined OptInState = "declined"
	OptInUnset    OptInState = "unset"
)

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

// UnmarshalJSON provides backwards-compatible handling for opt_in.
//
// CML versions have been observed returning opt_in as:
// - boolean (legacy): true/false
// - string (newer): "accepted" (and possibly other states)
func (u *User) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var wire struct {
		ID           UUID            `json:"id,omitempty"`
		Created      string          `json:"created,omitempty"`
		Modified     string          `json:"modified,omitempty"`
		DirectoryDN  string          `json:"directory_dn,omitempty"`
		Labs         []UUID          `json:"labs,omitempty"`
		Password     string          `json:"password,omitempty"`
		Username     string          `json:"username"`
		Fullname     string          `json:"fullname"`
		Description  string          `json:"description"`
		Email        string          `json:"email"`
		IsAdmin      bool            `json:"admin"`
		Groups       []UUID          `json:"groups,omitempty"`
		Associations AssociationList `json:"associations,omitempty"`
		ResourcePool *UUID           `json:"resource_pool,omitempty"`
		OptIn        json.RawMessage `json:"opt_in,omitempty"`
		TourVersion  string          `json:"tour_version"`
		PubkeyInfo   string          `json:"pubkey_info,omitempty"`
	}

	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	u.ID = wire.ID
	u.Created = wire.Created
	u.Modified = wire.Modified
	u.DirectoryDN = wire.DirectoryDN
	u.Labs = wire.Labs
	u.Password = wire.Password

	u.Username = wire.Username
	u.Fullname = wire.Fullname
	u.Description = wire.Description
	u.Email = wire.Email
	u.IsAdmin = wire.IsAdmin
	u.Groups = wire.Groups
	u.Associations = wire.Associations
	u.ResourcePool = wire.ResourcePool
	u.TourVersion = wire.TourVersion
	u.PubkeyInfo = wire.PubkeyInfo

	u.OptIn = nil
	if len(wire.OptIn) == 0 || string(wire.OptIn) == "null" {
		return nil
	}

	var b bool
	if err := json.Unmarshal(wire.OptIn, &b); err == nil {
		if b {
			v := OptInAccepted
			u.OptIn = &v
		} else {
			v := OptInDeclined
			u.OptIn = &v
		}
		return nil
	}

	var s string
	if err := json.Unmarshal(wire.OptIn, &s); err == nil {
		s = strings.ToLower(strings.TrimSpace(s))
		switch s {
		case string(OptInAccepted), "true", "yes", "y", "1":
			v := OptInAccepted
			u.OptIn = &v
		case string(OptInDeclined), "false", "no", "n", "0":
			v := OptInDeclined
			u.OptIn = &v
		case string(OptInUnset):
			v := OptInUnset
			u.OptIn = &v
		default:
			// Unknown/unsupported state; leave nil.
		}
		return nil
	}

	return nil
}

// UserList is a slice of User.
type UserList []User
