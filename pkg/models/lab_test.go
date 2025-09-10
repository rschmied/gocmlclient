package models

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabState_Constants(t *testing.T) {
	assert.Equal(t, LabState("DEFINED_ON_CORE"), LabStateDefined)
	assert.Equal(t, LabState("STOPPED"), LabStateStopped)
	assert.Equal(t, LabState("STARTED"), LabStateStarted)
}

func TestLab_JSON(t *testing.T) {
	lab := Lab{
		ID:                   "lab123",
		State:                LabStateStarted,
		Created:              "2023-01-01T00:00:00Z",
		Modified:             "2023-01-02T00:00:00Z",
		Title:                "Test Lab",
		Description:          "A test laboratory",
		Notes:                "Some notes",
		Owner:                "owner123",
		OwnerUsername:        "testuser",
		NodeCount:            5,
		LinkCount:            4,
		EffectivePermissions: Permissions{PermissionView, PermissionEdit},
	}

	// Test marshaling
	data, err := json.Marshal(lab)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled Lab
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, lab.ID, unmarshaled.ID)
	assert.Equal(t, lab.State, unmarshaled.State)
	assert.Equal(t, lab.Title, unmarshaled.Title)
	assert.Equal(t, lab.Description, unmarshaled.Description)
	assert.Equal(t, lab.NodeCount, unmarshaled.NodeCount)
	assert.Equal(t, lab.LinkCount, unmarshaled.LinkCount)
}

func TestLab_CanBeWiped(t *testing.T) {
	tests := []struct {
		name     string
		lab      Lab
		expected bool
	}{
		{
			name: "empty lab with defined state",
			lab: Lab{
				State: LabStateDefined,
				Nodes: NodeMap{},
			},
			expected: false,
		},
		{
			name: "empty lab with stopped state",
			lab: Lab{
				State: LabStateStopped,
				Nodes: NodeMap{},
			},
			expected: true,
		},
		{
			name: "lab with defined nodes",
			lab: Lab{
				State: LabStateStopped,
				Nodes: NodeMap{
					"node1": &Node{State: NodeStateDefined},
					"node2": &Node{State: NodeStateDefined},
				},
			},
			expected: true,
		},
		{
			name: "lab with mixed node states",
			lab: Lab{
				State: LabStateStopped,
				Nodes: NodeMap{
					"node1": &Node{State: NodeStateDefined},
					"node2": &Node{State: NodeStateStarted},
				},
			},
			expected: false,
		},
		{
			name: "lab with started nodes",
			lab: Lab{
				State: LabStateStopped,
				Nodes: NodeMap{
					"node1": &Node{State: NodeStateStarted},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.lab.CanBeWiped()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLab_Running(t *testing.T) {
	tests := []struct {
		name     string
		lab      Lab
		expected bool
	}{
		{
			name: "empty lab",
			lab: Lab{
				Nodes: NodeMap{},
			},
			expected: false,
		},
		{
			name: "lab with defined nodes only",
			lab: Lab{
				Nodes: NodeMap{
					"node1": &Node{State: NodeStateDefined},
					"node2": &Node{State: NodeStateDefined},
				},
			},
			expected: false,
		},
		{
			name: "lab with started nodes",
			lab: Lab{
				Nodes: NodeMap{
					"node1": &Node{State: NodeStateDefined},
					"node2": &Node{State: NodeStateStarted},
				},
			},
			expected: true,
		},
		{
			name: "lab with booted nodes",
			lab: Lab{
				Nodes: NodeMap{
					"node1": &Node{State: NodeStateBooted},
				},
			},
			expected: true,
		},
		{
			name: "lab with stopped nodes",
			lab: Lab{
				Nodes: NodeMap{
					"node1": &Node{State: NodeStateStopped},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.lab.Running()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLab_Booted(t *testing.T) {
	tests := []struct {
		name     string
		lab      Lab
		expected bool
	}{
		{
			name: "empty lab",
			lab: Lab{
				Nodes: NodeMap{},
			},
			expected: true, // vacuously true
		},
		{
			name: "lab with all booted nodes",
			lab: Lab{
				Nodes: NodeMap{
					"node1": &Node{State: NodeStateBooted},
					"node2": &Node{State: NodeStateBooted},
				},
			},
			expected: true,
		},
		{
			name: "lab with mixed node states",
			lab: Lab{
				Nodes: NodeMap{
					"node1": &Node{State: NodeStateBooted},
					"node2": &Node{State: NodeStateStarted},
				},
			},
			expected: false,
		},
		{
			name: "lab with defined nodes",
			lab: Lab{
				Nodes: NodeMap{
					"node1": &Node{State: NodeStateDefined},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.lab.Booted()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLab_NodeByLabel(t *testing.T) {
	lab := Lab{
		Nodes: NodeMap{
			"node1": &Node{ID: "node1", Label: "router1"},
			"node2": &Node{ID: "node2", Label: "switch1"},
			"node3": &Node{ID: "node3", Label: "router2"},
		},
	}

	tests := []struct {
		name        string
		label       string
		expectedID  string
		expectError bool
	}{
		{
			name:       "existing node",
			label:      "router1",
			expectedID: "node1",
		},
		{
			name:       "another existing node",
			label:      "switch1",
			expectedID: "node2",
		},
		{
			name:        "non-existing node",
			label:       "nonexistent",
			expectError: true,
		},
		{
			name:        "empty label",
			label:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			node, err := lab.NodeByLabel(ctx, tt.label)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, node)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, node)
				assert.Equal(t, tt.expectedID, string(node.ID))
			}
		})
	}
}

func TestLabCreateRequest_JSON(t *testing.T) {
	request := LabCreateRequest{
		Title:       "New Lab",
		Description: "A new laboratory",
		Notes:       "Some notes",
		Owner:       "owner123",
		Associations: AssociationUsersGroups{
			Users: []Association{
				{ID: "user1", Permissions: Permissions{PermissionEdit}},
			},
			Groups: []Association{
				{ID: "group1", Permissions: Permissions{PermissionView}},
			},
		},
	}

	// Test marshaling
	data, err := json.Marshal(request)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled LabCreateRequest
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, request.Title, unmarshaled.Title)
	assert.Equal(t, request.Description, unmarshaled.Description)
	assert.Equal(t, request.Notes, unmarshaled.Notes)
	assert.Equal(t, request.Owner, unmarshaled.Owner)
	assert.Len(t, unmarshaled.Associations.Users, 1)
	assert.Len(t, unmarshaled.Associations.Groups, 1)
}

func TestLabImport_JSON(t *testing.T) {
	importData := LabImport{
		ID: "import123",
		Warnings: []string{
			"Warning 1",
			"Warning 2",
		},
	}

	// Test marshaling
	data, err := json.Marshal(importData)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled LabImport
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, importData.ID, unmarshaled.ID)
	assert.Equal(t, importData.Warnings, unmarshaled.Warnings)
}
