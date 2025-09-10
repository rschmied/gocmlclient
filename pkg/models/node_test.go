package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeState_Constants(t *testing.T) {
	assert.Equal(t, NodeState("DEFINED_ON_CORE"), NodeStateDefined)
	assert.Equal(t, NodeState("STOPPED"), NodeStateStopped)
	assert.Equal(t, NodeState("STARTED"), NodeStateStarted)
	assert.Equal(t, NodeState("QUEUED"), NodeStateQueued)
	assert.Equal(t, NodeState("BOOTED"), NodeStateBooted)
	assert.Equal(t, NodeState("DISCONNECTED"), NodeStateDisconnected)
}

func TestBootProgress_Constants(t *testing.T) {
	assert.Equal(t, BootProgress("Not running"), BootProgressNotRunning)
	assert.Equal(t, BootProgress("Booting"), BootProgressBooting)
	assert.Equal(t, BootProgress("Booted"), BootProgressBooted)
}

func TestNode_JSON(t *testing.T) {
	node := Node{
		ID:              "node123",
		LabID:           "lab123",
		Label:           "Test Node",
		X:               100,
		Y:               200,
		NodeDefinition:  "iosv",
		CPUs:            2,
		ImageDefinition: stringPtr("vios-adventerprisek9-m"),
		RAM:             intPtr(512),
		State:           NodeStateStarted,
		Tags:            []string{"test", "iosv"},
	}

	// Test marshaling
	data, err := json.Marshal(node)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled Node
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, node.ID, unmarshaled.ID)
	assert.Equal(t, node.LabID, unmarshaled.LabID)
	assert.Equal(t, node.Label, unmarshaled.Label)
	assert.Equal(t, node.X, unmarshaled.X)
	assert.Equal(t, node.Y, unmarshaled.Y)
	assert.Equal(t, node.NodeDefinition, unmarshaled.NodeDefinition)
	assert.Equal(t, node.CPUs, unmarshaled.CPUs)
	assert.Equal(t, node.ImageDefinition, unmarshaled.ImageDefinition)
	assert.Equal(t, node.RAM, unmarshaled.RAM)
	assert.Equal(t, node.State, unmarshaled.State)
	assert.Equal(t, node.Tags, unmarshaled.Tags)
}

func TestNode_JSON_WithConfigurations(t *testing.T) {
	// Test the complex configuration marshaling by setting up the node properly
	node := Node{
		ID:    "node123",
		Label: "Test Node",
	}

	// Manually set Configurations (this would normally be done by UnmarshalJSON)
	node.Configurations = []NodeConfig{
		{Name: "startup", Content: "hostname test"},
		{Name: "running", Content: "interface GigabitEthernet0/0"},
	}

	// Test marshaling with named configurations
	data, err := json.Marshal(&node)
	assert.NoError(t, err)

	// Should marshal as array of configurations
	var jsonMap map[string]any
	err = json.Unmarshal(data, &jsonMap)
	assert.NoError(t, err)

	configs, exists := jsonMap["configuration"]
	assert.True(t, exists, "configuration field should exist")

	// The MarshalJSON method creates a struct with NamedConfig field
	// So we need to check if it's an array
	if configArray, ok := configs.([]any); ok {
		assert.Len(t, configArray, 2, "should have 2 configurations")

		// Check first config
		config0 := configArray[0].(map[string]any)
		assert.Equal(t, "startup", config0["name"])
		assert.Equal(t, "hostname test", config0["content"])

		// Check second config
		config1 := configArray[1].(map[string]any)
		assert.Equal(t, "running", config1["name"])
		assert.Equal(t, "interface GigabitEthernet0/0", config1["content"])
	} else {
		t.Errorf("Expected configuration to be []any, got %T", configs)
	}
}

func TestNode_JSON_WithStringConfiguration(t *testing.T) {
	node := Node{
		ID:            "node123",
		Label:         "Test Node",
		Configuration: "hostname test",
	}

	// Test marshaling with string configuration
	data, err := json.Marshal(node)
	assert.NoError(t, err)

	// Should marshal as string
	var jsonMap map[string]any
	err = json.Unmarshal(data, &jsonMap)
	assert.NoError(t, err)

	config, exists := jsonMap["configuration"]
	assert.True(t, exists)
	assert.IsType(t, "", config)
	assert.Equal(t, "hostname test", config)
}

func TestNode_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name            string
		jsonStr         string
		expectedConfig  any
		expectedConfigs []NodeConfig
	}{
		{
			name:           "string configuration",
			jsonStr:        `{"id":"node1","label":"test","configuration":"hostname test"}`,
			expectedConfig: func() any { s := "hostname test"; return &s }(),
		},
		{
			name:    "array configuration",
			jsonStr: `{"id":"node1","label":"test","configuration":[{"name":"startup","content":"hostname test"}]}`,
			expectedConfigs: []NodeConfig{
				{Name: "startup", Content: "hostname test"},
			},
		},
		{
			name:           "null configuration",
			jsonStr:        `{"id":"node1","label":"test","configuration":null}`,
			expectedConfig: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node Node
			err := json.Unmarshal([]byte(tt.jsonStr), &node)
			assert.NoError(t, err)

			if tt.expectedConfig != nil {
				assert.Equal(t, tt.expectedConfig, node.Configuration)
			} else {
				assert.Nil(t, node.Configuration)
			}

			if len(tt.expectedConfigs) > 0 {
				assert.Equal(t, tt.expectedConfigs, node.Configurations)
			}
		})
	}
}

func TestNode_UnmarshalJSON_InvalidConfiguration(t *testing.T) {
	// Test with invalid configuration type
	jsonStr := `{"id":"node1","label":"test","configuration":{"invalid":"type"}}`

	var node Node
	err := json.Unmarshal([]byte(jsonStr), &node)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected type")
}

func TestNode_UnmarshalJSON_Null(t *testing.T) {
	// Test unmarshaling null
	var node Node
	err := json.Unmarshal([]byte("null"), &node)
	assert.NoError(t, err)
	assert.Empty(t, node.ID)
}

func TestNode_UnmarshalJSON_EmptyString(t *testing.T) {
	// Test unmarshaling empty string
	var node Node
	err := json.Unmarshal([]byte(`""`), &node)
	assert.NoError(t, err)
	assert.Empty(t, node.ID)
}

func TestNodeMap_MarshalJSON(t *testing.T) {
	nodeMap := NodeMap{
		"node2": &Node{ID: "node2", Label: "Node B"},
		"node1": &Node{ID: "node1", Label: "Node A"},
		"node3": &Node{ID: "node3", Label: "Node C"},
	}

	data, err := json.Marshal(nodeMap)
	assert.NoError(t, err)

	// Should be sorted by UUID
	var nodes []*Node
	err = json.Unmarshal(data, &nodes)
	assert.NoError(t, err)

	assert.Len(t, nodes, 3)
	assert.Equal(t, "node1", string(nodes[0].ID))
	assert.Equal(t, "node2", string(nodes[1].ID))
	assert.Equal(t, "node3", string(nodes[2].ID))
}

func TestNode_SameConfig(t *testing.T) {
	tests := []struct {
		name     string
		node1    Node
		node2    Node
		expected bool
	}{
		{
			name: "both string configs equal",
			node1: Node{
				Configuration: "config1",
			},
			node2: Node{
				Configuration: "config1",
			},
			expected: true,
		},
		{
			name: "both string configs different",
			node1: Node{
				Configuration: "config1",
			},
			node2: Node{
				Configuration: "config2",
			},
			expected: false,
		},
		{
			name: "string vs non-string",
			node1: Node{
				Configuration: "config1",
			},
			node2: Node{
				Configuration: 123,
			},
			expected: false,
		},
		{
			name: "both non-string equal",
			node1: Node{
				Configuration: 123,
			},
			node2: Node{
				Configuration: 123,
			},
			expected: true,
		},
		{
			name: "both nil",
			node1: Node{
				Configuration: nil,
			},
			node2: Node{
				Configuration: nil,
			},
			expected: true,
		},
		{
			name: "one nil one not",
			node1: Node{
				Configuration: nil,
			},
			node2: Node{
				Configuration: "config",
			},
			expected: false,
		},
		{
			name: "named configs equal",
			node1: Node{
				Configuration: nil,
				Configurations: []NodeConfig{
					{Name: "config1", Content: "content1"},
					{Name: "config2", Content: "content2"},
				},
			},
			node2: Node{
				Configuration: nil,
				Configurations: []NodeConfig{
					{Name: "config1", Content: "content1"},
					{Name: "config2", Content: "content2"},
				},
			},
			expected: true,
		},
		{
			name: "named configs different length",
			node1: Node{
				Configurations: []NodeConfig{
					{Name: "config1", Content: "content1"},
				},
			},
			node2: Node{
				Configurations: []NodeConfig{
					{Name: "config1", Content: "content1"},
					{Name: "config2", Content: "content2"},
				},
			},
			expected: false,
		},
		{
			name: "named configs different content",
			node1: Node{
				Configurations: []NodeConfig{
					{Name: "config1", Content: "content1"},
				},
			},
			node2: Node{
				Configurations: []NodeConfig{
					{Name: "config1", Content: "different"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.node1.SameConfig(&tt.node2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNodeConfig_JSON(t *testing.T) {
	config := NodeConfig{
		Name:    "startup",
		Content: "hostname test\ninterface GigabitEthernet0/0",
	}

	data, err := json.Marshal(config)
	assert.NoError(t, err)

	var unmarshaled NodeConfig
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, config.Name, unmarshaled.Name)
	assert.Equal(t, config.Content, unmarshaled.Content)
}

func TestSerialDevice_JSON(t *testing.T) {
	device := SerialDevice{
		ConsoleKey:   "console123",
		DeviceNumber: 1,
	}

	data, err := json.Marshal(device)
	assert.NoError(t, err)

	var unmarshaled SerialDevice
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, device.ConsoleKey, unmarshaled.ConsoleKey)
	assert.Equal(t, device.DeviceNumber, unmarshaled.DeviceNumber)
}

func TestSerialConsole_JSON(t *testing.T) {
	console := SerialConsole{
		ConsoleKey:   "console123",
		DeviceNumber: 1,
		Label:        "Console 1",
	}

	data, err := json.Marshal(console)
	assert.NoError(t, err)

	var unmarshaled SerialConsole
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, console.ConsoleKey, unmarshaled.ConsoleKey)
	assert.Equal(t, console.DeviceNumber, unmarshaled.DeviceNumber)
	assert.Equal(t, console.Label, unmarshaled.Label)
}

func TestNodeOperational_JSON(t *testing.T) {
	operational := NodeOperational{
		BootDiskSize:    intPtr(1000),
		CPUlimit:        intPtr(50),
		CPUs:            intPtr(2),
		DataVolume:      intPtr(500),
		RAM:             intPtr(1024),
		ComputeID:       uuidPtr("compute123"),
		ImageDefinition: stringPtr("vios-adventerprisek9-m"),
		VNCkey:          uuidPtr("vnc123"),
		ResourcePool:    stringPtr("pool1"),
		IOLAppID:        intPtr(123),
		SerialConsoles: []SerialConsole{
			{ConsoleKey: "console1", DeviceNumber: 0},
		},
	}

	data, err := json.Marshal(operational)
	assert.NoError(t, err)

	var unmarshaled NodeOperational
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, operational.BootDiskSize, unmarshaled.BootDiskSize)
	assert.Equal(t, operational.CPUlimit, unmarshaled.CPUlimit)
	assert.Equal(t, operational.CPUs, unmarshaled.CPUs)
	assert.Equal(t, operational.DataVolume, unmarshaled.DataVolume)
	assert.Equal(t, operational.RAM, unmarshaled.RAM)
	assert.Equal(t, operational.ComputeID, unmarshaled.ComputeID)
	assert.Equal(t, operational.ImageDefinition, unmarshaled.ImageDefinition)
	assert.Equal(t, operational.VNCkey, unmarshaled.VNCkey)
	assert.Equal(t, operational.ResourcePool, unmarshaled.ResourcePool)
	assert.Equal(t, operational.IOLAppID, unmarshaled.IOLAppID)
	assert.Len(t, unmarshaled.SerialConsoles, 1)
}

// Helper functions for creating pointers
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
