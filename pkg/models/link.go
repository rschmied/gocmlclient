// Package models provides the models for Cisco Modeling Labs
// here: link related types
package models

const (
	LinkStateDefined = "DEFINED_ON_CORE"
	LinkStateStopped = "STOPPED"
	LinkStateStarted = "STARTED"
)

// {
// 	"id": "4d76f475-2915-444e-bfd1-425a517120bc",
// 	"interface_a": "20681832-36e8-4ba9-9d8d-0588e0f7b517",
// 	"interface_b": "1959cc9f-361c-410e-a960-9d9a896482a0",
// 	"lab_id": "52d5c824-e10c-450a-b9c5-b700bd3bc17a",
// 	"label": "ext-conn-0-port<->unmanaged-switch-0-port0",
// 	"link_capture_key": "d827ce92-db2e-4933-bc0d-7a2c38e39ad5",
// 	"node_a": "9efb1503-7e2a-4d2a-959e-865209f1acc0",
// 	"node_b": "1cc0cbcd-6b4f-4bbe-9f69-2c3da5e3495a",
// 	"state": "STARTED"
// }

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
