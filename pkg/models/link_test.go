package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkState_Constants(t *testing.T) {
	assert.Equal(t, "DEFINED_ON_CORE", LinkStateDefined)
	assert.Equal(t, "STOPPED", LinkStateStopped)
	assert.Equal(t, "STARTED", LinkStateStarted)
}

func TestLink_JSON(t *testing.T) {
	link := Link{
		ID:      "link123",
		LabID:   "lab123",
		State:   LinkStateStarted,
		Label:   "Test Link",
		PCAPkey: "pcap123",
		SrcID:   "interface1",
		DstID:   "interface2",
		SrcNode: "node1",
		DstNode: "node2",
		SrcSlot: 0,
		DstSlot: 1,
	}

	// Test marshaling
	data, err := json.Marshal(link)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled Link
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, link.ID, unmarshaled.ID)
	assert.Equal(t, link.LabID, unmarshaled.LabID)
	assert.Equal(t, link.State, unmarshaled.State)
	assert.Equal(t, link.Label, unmarshaled.Label)
	assert.Equal(t, link.PCAPkey, unmarshaled.PCAPkey)
	assert.Equal(t, link.SrcID, unmarshaled.SrcID)
	assert.Equal(t, link.DstID, unmarshaled.DstID)
	assert.Equal(t, link.SrcNode, unmarshaled.SrcNode)
	assert.Equal(t, link.DstNode, unmarshaled.DstNode)
	assert.Equal(t, link.SrcSlot, unmarshaled.SrcSlot)
	assert.Equal(t, link.DstSlot, unmarshaled.DstSlot)
}

func TestLink_JSON_Minimal(t *testing.T) {
	// Test with minimal required fields
	link := Link{
		ID:    "link123",
		State: LinkStateDefined,
	}

	// Test marshaling
	data, err := json.Marshal(link)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled Link
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, link.ID, unmarshaled.ID)
	assert.Equal(t, link.State, unmarshaled.State)
	assert.Empty(t, unmarshaled.Label)
	assert.Empty(t, unmarshaled.LabID)
}

func TestLinkList_JSON(t *testing.T) {
	links := []*Link{
		{
			ID:    "link1",
			State: LinkStateStarted,
			Label: "Link 1",
		},
		{
			ID:    "link2",
			State: LinkStateStopped,
			Label: "Link 2",
		},
	}

	// Test marshaling
	data, err := json.Marshal(links)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled []*Link
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Len(t, unmarshaled, 2)
	assert.Equal(t, "link1", string(unmarshaled[0].ID))
	assert.Equal(t, "Link 1", unmarshaled[0].Label)
	assert.Equal(t, "link2", string(unmarshaled[1].ID))
	assert.Equal(t, "Link 2", unmarshaled[1].Label)
}
