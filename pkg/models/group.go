// Package models provides the models for Cisco Modeling Labs
// here: group related types
package models

// Group represents a user group in CML with associated permissions and members.
type Group struct {
	ID              UUID          `json:"id,omitempty"`
	Description     string        `json:"description,omitempty"`
	Members         []UUID        `json:"members,omitempty"`
	Name            string        `json:"name"`
	Associations    []Association `json:"associations,omitempty"`
	Created         string        `json:"created,omitempty"`
	Modified        string        `json:"modified,omitempty"`
	DirectoryDN     string        `json:"directory_dn,omitempty"`
	DirectoryExists *bool         `json:"directory_exists,omitempty"`
}

// GroupList is a slice of Group pointers.
type GroupList []*Group
