// Package models provides the models for Cisco Modeling Labs
// here: general types
package models

import "encoding/json"

// ErrorResponse represents an error response from the CML API.
type ErrorResponse struct {
	Code        int             `json:"code"`
	Description json.RawMessage `json:"description"`
}

// UUID represents a universally unique identifier as a string.
type UUID string
