package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkTopology_JSON(t *testing.T) {
	link := LinkTopology{
		ID:    "link123",
		I1:    "interface1",
		I2:    "interface2",
		N1:    "node1",
		N2:    "node2",
		Label: "Test Link",
		Conditioning: &LinkConditionConfiguration{
			Bandwidth: 1000,
			Latency:   10,
			Enabled:   true,
		},
	}

	// Test marshaling
	data, err := json.Marshal(link)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled LinkTopology
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, link.ID, unmarshaled.ID)
	assert.Equal(t, link.I1, unmarshaled.I1)
	assert.Equal(t, link.I2, unmarshaled.I2)
	assert.Equal(t, link.N1, unmarshaled.N1)
	assert.Equal(t, link.N2, unmarshaled.N2)
	assert.Equal(t, link.Label, unmarshaled.Label)
	assert.NotNil(t, unmarshaled.Conditioning)
	assert.Equal(t, link.Conditioning.Bandwidth, unmarshaled.Conditioning.Bandwidth)
	assert.Equal(t, link.Conditioning.Latency, unmarshaled.Conditioning.Latency)
	assert.Equal(t, link.Conditioning.Enabled, unmarshaled.Conditioning.Enabled)
}

func TestLinkTopology_JSON_Minimal(t *testing.T) {
	// Test with minimal required fields
	link := LinkTopology{
		ID: "link123",
		I1: "interface1",
		I2: "interface2",
		N1: "node1",
		N2: "node2",
	}

	// Test marshaling
	data, err := json.Marshal(link)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled LinkTopology
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, link.ID, unmarshaled.ID)
	assert.Equal(t, link.I1, unmarshaled.I1)
	assert.Equal(t, link.I2, unmarshaled.I2)
	assert.Equal(t, link.N1, unmarshaled.N1)
	assert.Equal(t, link.N2, unmarshaled.N2)
	assert.Empty(t, unmarshaled.Label)
	assert.Nil(t, unmarshaled.Conditioning)
}

func TestLinkTopology_JSON_WithNullConditioning(t *testing.T) {
	jsonStr := `{
		"id": "link123",
		"i1": "interface1",
		"i2": "interface2",
		"n1": "node1",
		"n2": "node2",
		"conditioning": null
	}`

	var link LinkTopology
	err := json.Unmarshal([]byte(jsonStr), &link)
	assert.NoError(t, err)

	assert.Equal(t, "link123", link.ID)
	assert.Nil(t, link.Conditioning)
}
