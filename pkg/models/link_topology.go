// Package models provides the models for Cisco Modeling Labs
// here: link topology related types
package models

// LinkTopology defines the data structure for a CML link in topology context.
// This is used for topology import/export operations, not for API operations.
type LinkTopology struct {
	ID           string                      `json:"id"`
	I1           string                      `json:"i1"`
	I2           string                      `json:"i2"`
	N1           string                      `json:"n1"`
	N2           string                      `json:"n2"`
	Label        string                      `json:"label,omitempty"`
	Conditioning *LinkConditionConfiguration `json:"conditioning,omitempty"`
}
