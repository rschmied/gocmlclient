// Package models provides the models for Cisco Modeling Labs
// here: link related types
package models

const (
	LinkStateDefined = "DEFINED_ON_CORE"
	LinkStateStopped = "STOPPED"
	LinkStateStarted = "STARTED"
)

// Link defines the data structure for a CML link between nodes.
type Link struct {
	ID      UUID   `json:"id"`
	LabID   UUID   `json:"lab_id"`
	State   string `json:"state"`
	Label   string `json:"label"`
	PCAPkey UUID   `json:"link_capture_key"`
	SrcID   UUID   `json:"interface_a"`
	DstID   UUID   `json:"interface_b"`
	SrcNode UUID   `json:"node_a"`
	DstNode UUID   `json:"node_b"`
	SrcSlot int    `json:"slot_a"`
	DstSlot int    `json:"slot_b"`
}
