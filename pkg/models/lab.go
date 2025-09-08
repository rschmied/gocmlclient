// Package models provides the models for Cisco Modeling Labs
// here: lab related types
package models

import (
	"context"

	cmlerror "github.com/rschmied/gocmlclient/pkg/errors"
)

// {
// 	"state": "STOPPED",
// 	"created": "2023-02-08T10:02:43+00:00",
// 	"modified": "2023-02-08T11:59:45+00:00",
// 	"lab_title": "sample",
// 	"lab_description": "",
// 	"lab_notes": "",
// 	"owner": "00000000-0000-4000-a000-000000000000",
// 	"owner_username": "admin",
// 	"node_count": 5,
// 	"link_count": 4,
// 	"id": "a7d20917-5e57-407f-80ea-63596c53f198",
// 	"groups": [
// 	  {
// 		"id": "bc9b796e-48bc-4369-b131-231dfa057d41",
// 		"name": "students",
// 		"permission": "read_only"
// 	  }
// 	]
// }

type LabState string

const (
	LabStateDefined LabState = "DEFINED_ON_CORE"
	LabStateStopped LabState = "STOPPED"
	LabStateStarted LabState = "STARTED"
	LabStateBooted  LabState = "BOOTED"
)

type (
	NodeMap  map[UUID]*Node
	LinkList []*Link
)

type LabImport struct {
	ID       string   `json:"id"`
	Warnings []string `json:"warnings"`
}

type Lab struct {
	ID                   UUID        `json:"id"`
	State                LabState    `json:"state,omitempty"`
	Created              string      `json:"created,omitempty"`
	Modified             string      `json:"modified,omitempty"`
	Title                string      `json:"lab_title,omitempty"`
	Description          string      `json:"lab_description,omitempty"`
	Notes                string      `json:"lab_notes,omitempty"`
	Owner                UUID        `json:"owner,omitempty"`
	OwnerUsername        string      `json:"owner_username,omitempty"`
	OwnerFullname        string      `json:"owner_fullname,omitempty"`
	NodeCount            int         `json:"node_count,omitempty"`
	LinkCount            int         `json:"link_count,omitempty"`
	EffectivePermissions Permissions `json:"effective_permissions,omitempty"`

	// non-schema helpers
	Nodes NodeMap  `json:"nodes,omitempty"`
	Links LinkList `json:"links,omitempty"`
}

// CanBeWiped returns `true` when all nodes in the lab are wiped.
func (l *Lab) CanBeWiped() bool {
	if len(l.Nodes) == 0 {
		return l.State != LabStateDefined
	}
	for _, node := range l.Nodes {
		if node.State != NodeStateDefined {
			return false
		}
	}
	return true
}

// Running returns `true` if at least one node is running (started or booted).
func (l *Lab) Running() bool {
	for _, node := range l.Nodes {
		if node.State != NodeStateDefined && node.State != NodeStateStopped {
			return true
		}
	}
	return false
}

// Booted returns `true` if all nodes in the lab are in booted state.
func (l *Lab) Booted() bool {
	for _, node := range l.Nodes {
		if node.State != NodeStateBooted {
			return false
		}
	}
	return true
}

// NodeByLabel returns the node of a lab identified by its `label“ or an error
// if not found.
func (l *Lab) NodeByLabel(ctx context.Context, label string) (*Node, error) {
	for _, node := range l.Nodes {
		if node.Label == label {
			return node, nil
		}
	}
	return nil, cmlerror.ErrElementNotFound
}
