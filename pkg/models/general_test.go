package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorResponse_JSON(t *testing.T) {
	tests := []struct {
		name        string
		code        int
		description json.RawMessage
		jsonStr     string
	}{
		{
			name:        "basic error response",
			code:        400,
			description: json.RawMessage(`"Bad Request"`),
			jsonStr:     `{"code":400,"description":"Bad Request"}`,
		},
		{
			name:        "error with object description",
			code:        500,
			description: json.RawMessage(`{"message":"Internal Server Error","details":"Something went wrong"}`),
			jsonStr:     `{"code":500,"description":{"message":"Internal Server Error","details":"Something went wrong"}}`,
		},
		{
			name:        "error with array description",
			code:        422,
			description: json.RawMessage(`["Field is required","Invalid format"]`),
			jsonStr:     `{"code":422,"description":["Field is required","Invalid format"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test unmarshaling
			var resp ErrorResponse
			err := json.Unmarshal([]byte(tt.jsonStr), &resp)
			assert.NoError(t, err)
			assert.Equal(t, tt.code, resp.Code)
			assert.Equal(t, tt.description, resp.Description)

			// Test marshaling
			data, err := json.Marshal(resp)
			assert.NoError(t, err)

			// Verify it matches original JSON
			var unmarshaled map[string]any
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, float64(tt.code), unmarshaled["code"])
		})
	}
}

func TestUUID_Type(t *testing.T) {
	// Test that UUID is a string type
	var uuid UUID = "123e4567-e89b-12d3-a456-426614174000"
	assert.IsType(t, "", string(uuid))
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", string(uuid))

	// Test empty UUID
	var emptyUUID UUID = ""
	assert.Equal(t, "", string(emptyUUID))
}
