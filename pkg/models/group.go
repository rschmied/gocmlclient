// Package models provides the models for Cisco Modeling Labs
// here: group related types
package models

type Group struct {
	ID           UUID            `json:"id,omitempty"`
	Description  string          `json:"description"`
	Members      []UUID          `json:"members"`
	Name         string          `json:"name"`
	Asocciations AssociationList `json:"asocciations"`
}

type GroupList []*Group
