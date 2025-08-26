// Package models provides the models for Cisco Modeling Labs
// here: user related types
package models

type User struct {
	ID           UUID            `json:"id,omitempty"`
	Created      string          `json:"created,omitempty"`
	Modified     string          `json:"modified,omitempty"`
	Username     string          `json:"username"`
	Fullname     string          `json:"fullname"`
	Password     string          `json:"password"`
	Description  string          `json:"description"`
	Email        string          `json:"email"`
	IsAdmin      bool            `json:"admin"`
	DirectoryDN  string          `json:"directory_dn,omitempty"`
	Groups       []UUID          `json:"groups,omitempty"` // deprecated?
	Associations AssociationList `json:"associations"`     // with 2.9.0
	Labs         []UUID          `json:"labs,omitempty"`
	OptIn        bool            `json:"opt_in"` // with 2.5.0
	TourVersion  string          `json:"tour_version"`
	PubkeyInfo   string          `json:"pubkey_info"`
	ResourcePool []UUID          `json:"resource_pool,omitempty"` // with 2.5.0
}

type UserList []*User
