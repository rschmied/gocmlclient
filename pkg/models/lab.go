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

// LabState represents the operational state of a CML lab.
type LabState string

const (
	// LabStateDefined indicates the lab is defined on core.
	LabStateDefined LabState = "DEFINED_ON_CORE"
	// LabStateStopped indicates the lab is stopped.
	LabStateStopped LabState = "STOPPED"
	// LabStateStarted indicates the lab is started.
	LabStateStarted LabState = "STARTED"
)

// LabList is a list of lab IDs
type LabList []UUID

// OldPermission represents the old permission system for lab groups
type OldPermission string

const (
	// OldPermissionReadOnly allows read-only access
	OldPermissionReadOnly OldPermission = "read_only"
	// OldPermissionReadWrite allows read-write access
	OldPermissionReadWrite OldPermission = "read_write"
)

// LabGroup represents a group with permissions for a lab
type LabGroup struct {
	ID         UUID          `json:"id"`
	Permission OldPermission `json:"permission"`
	Name       string        `json:"name,omitempty"`
}

// LabResponse represents the response from /labs/{lab_id} endpoint
type LabResponse struct {
	ID                   UUID        `json:"id"`
	Created              string      `json:"created,omitempty"`
	Modified             string      `json:"modified,omitempty"`
	Title                string      `json:"lab_title"`
	Description          string      `json:"lab_description,omitempty"`
	Notes                string      `json:"lab_notes,omitempty"`
	OwnerID              UUID        `json:"owner"`
	OwnerUsername        string      `json:"owner_username"`
	OwnerFullname        string      `json:"owner_fullname"`
	State                LabState    `json:"state"`
	NodeCount            int         `json:"node_count"`
	LinkCount            int         `json:"link_count"`
	Groups               []LabGroup  `json:"groups,omitempty"`
	EffectivePermissions Permissions `json:"effective_permissions"`
}

// LabImport represents the result of importing a lab, including any warnings.
type LabImport struct {
	ID       UUID     `json:"id"`
	Warnings []string `json:"warnings"`
}

// LabCreateRequest represents the data required to create a new lab.
// Only certain fields from the full Lab model are accepted during creation.
// This ensures API safety by preventing misuse of unsupported fields.
type LabCreateRequest struct {
	Title        string                 `json:"title,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Notes        string                 `json:"notes,omitempty"`
	Owner        UUID                   `json:"owner,omitempty"`
	Groups       []LabGroup             `json:"groups,omitzero"`
	Associations AssociationUsersGroups `json:"associations,omitzero"`
}

// LabUpdateRequest is identical to LabCreateRequest and, in fact,
// LabCreateRequest is also used in the OpenAPI spec. Using a new type makes it
// clearer.
type LabUpdateRequest LabCreateRequest

// Lab represents a CML lab with its nodes, links, and metadata.
type Lab struct {
	ID                   UUID        `json:"id"`
	State                LabState    `json:"state,omitempty"`
	Created              string      `json:"created,omitempty"`
	Modified             string      `json:"modified,omitempty"`
	Title                string      `json:"lab_title,omitempty"`
	Description          string      `json:"lab_description,omitempty"`
	Notes                string      `json:"lab_notes,omitempty"`
	OwnerID              UUID        `json:"owner,omitempty"`
	OwnerUsername        string      `json:"owner_username,omitempty"`
	OwnerFullname        string      `json:"owner_fullname,omitempty"`
	NodeCount            int         `json:"node_count,omitempty"`
	LinkCount            int         `json:"link_count,omitempty"`
	EffectivePermissions Permissions `json:"effective_permissions,omitempty"`
	Groups               []LabGroup  `json:"groups,omitempty"`

	// non-schema helpers
	Owner *User    `json:"-"` // Full user object (not serialized)
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

// LabTilesResponse represents the response from /populate_lab_tiles endpoint
type LabTilesResponse struct {
	LabTiles map[string]LabTile `json:"lab_tiles"`
}

// LabTile represents a lab tile with topology data
type LabTile struct {
	ID                   UUID        `json:"id"`
	State                LabState    `json:"state"`
	Created              string      `json:"created"`
	Modified             string      `json:"modified"`
	Title                string      `json:"lab_title"`
	Description          string      `json:"lab_description"`
	Notes                string      `json:"lab_notes"`
	OwnerID              UUID        `json:"owner"`
	OwnerUsername        string      `json:"owner_username"`
	OwnerFullname        string      `json:"owner_fullname"`
	NodeCount            int         `json:"node_count"`
	LinkCount            int         `json:"link_count"`
	Groups               []LabGroup  `json:"groups"`
	Topology             Topology    `json:"topology"`
	EffectivePermissions Permissions `json:"effective_permissions"`
}

// Topology represents the lab topology with nodes and links
type Topology struct {
	Nodes            []NodeTile       `json:"nodes"`
	Links            []LinkTile       `json:"links"`
	Annotations      []AnnotationTile `json:"annotations"`
	SmartAnnotations []any            `json:"smart_annotations"` // Keep as any for now
}

// NodeTile represents a node in the topology
type NodeTile struct {
	ID              UUID    `json:"id"`
	Label           string  `json:"label"`
	X               float64 `json:"x"`
	Y               float64 `json:"y"`
	NodeDefinition  string  `json:"node_definition"`
	ImageDefinition *string `json:"image_definition"`
	State           string  `json:"state"`
	CPUs            *int    `json:"cpus"`
	CPULimit        *int    `json:"cpu_limit"`
	RAM             *int    `json:"ram"`
	DataVolume      *int    `json:"data_volume"`
	BootDiskSize    *int    `json:"boot_disk_size"`
	Tags            []any   `json:"tags"` // Keep as any for now
}

// LinkTile represents a link between nodes
type LinkTile struct {
	ID    UUID   `json:"id"`
	NodeA UUID   `json:"node_a"`
	NodeB UUID   `json:"node_b"`
	State string `json:"state"`
}

// AnnotationTile represents an annotation in the topology
type AnnotationTile struct {
	ID           UUID    `json:"id"`
	BorderColor  string  `json:"border_color,omitempty"`
	BorderRadius float64 `json:"border_radius,omitempty"`
	BorderStyle  string  `json:"border_style,omitempty"`
	Color        string  `json:"color,omitempty"`
	Rotation     float64 `json:"rotation,omitempty"`
	TextBold     bool    `json:"text_bold,omitempty"`
	TextContent  string  `json:"text_content,omitempty"`
	TextFont     string  `json:"text_font,omitempty"`
	TextItalic   bool    `json:"text_italic,omitempty"`
	TextSize     float64 `json:"text_size,omitempty"`
	TextUnit     string  `json:"text_unit,omitempty"`
	Thickness    float64 `json:"thickness,omitempty"`
	Type         string  `json:"type"`
	X1           float64 `json:"x1,omitempty"`
	Y1           float64 `json:"y1,omitempty"`
	X2           float64 `json:"x2,omitempty"`
	Y2           float64 `json:"y2,omitempty"`
	ZIndex       float64 `json:"z_index,omitempty"`
}
