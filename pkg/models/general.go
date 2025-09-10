package models

import "encoding/json"

type ErrorResponse struct {
	Code        int             `json:"code"`
	Description json.RawMessage `json:"description"`
}

type UUID string
