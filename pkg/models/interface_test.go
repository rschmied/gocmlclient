package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterfaceJSONMarshalUnmarshal(t *testing.T) {
	// Test data matching OpenAPI InterfaceResponse schema
	testJSON := `{
		"id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
		"lab_id": "90f84e38-a71c-4d57-8d90-00fa8a197386",
		"node": "90f84e38-a71c-4d57-8d90-00fa8a197387",
		"label": "GigabitEthernet0/0",
		"slot": 0,
		"type": "physical",
		"state": "STARTED",
		"is_connected": true,
		"mac_address": null,
		"operational": {
			"device_name": "eth0",
			"mac_address": "00:11:22:33:44:55",
			"src_udp_port": 10000,
			"dst_udp_port": 10001
		}
	}`

	var iface Interface
	err := json.Unmarshal([]byte(testJSON), &iface)
	assert.NoError(t, err)

	// Verify all fields are properly set
	assert.Equal(t, "90f84e38-a71c-4d57-8d90-00fa8a197385", string(iface.ID))
	assert.Equal(t, "GigabitEthernet0/0", iface.Label)
	assert.NotNil(t, iface.Slot)
	assert.Equal(t, 0, *iface.Slot)
	assert.True(t, iface.IsConnected)
	assert.Nil(t, iface.MACAddress)

	// Test marshaling back to JSON
	marshaledData, err := json.Marshal(iface)
	assert.NoError(t, err)

	// Verify the marshaled JSON contains the expected fields
	var unmarshaled map[string]any
	err = json.Unmarshal(marshaledData, &unmarshaled)
	assert.NoError(t, err)

	// Check that all expected fields are present
	expectedFields := []string{"id", "lab_id", "node", "label", "slot", "type", "state", "is_connected", "operational"}
	for _, field := range expectedFields {
		assert.Contains(t, unmarshaled, field, "Expected field %s in marshaled JSON", field)
	}
}

func TestInterfaceWithNullValues(t *testing.T) {
	// Test JSON with zero values for required fields
	testJSON := `{
		"id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
		"label": "Loopback0",
		"is_connected": false,
		"operational": {
			"device_name": null,
			"mac_address": null,
			"src_udp_port": null,
			"dst_udp_port": null
		}
	}`

	var iface Interface
	err := json.Unmarshal([]byte(testJSON), &iface)
	assert.NoError(t, err)
	assert.False(t, iface.IsConnected)
}

func TestIfaceState_Constants(t *testing.T) {
	assert.Equal(t, IfaceState("DEFINED_ON_CORE"), IfaceStateDefined)
	assert.Equal(t, IfaceState("STOPPED"), IfaceStateStopped)
	assert.Equal(t, IfaceState("STARTED"), IfaceStateStarted)
}

func TestIfaceType_Constants(t *testing.T) {
	assert.Equal(t, IfaceType("physical"), IfaceTypePhysical)
	assert.Equal(t, IfaceType("loopback"), IfaceTypeLoopback)
}

func TestOperational_JSON(t *testing.T) {
	tests := []struct {
		name        string
		operational Operational
		jsonStr     string
	}{
		{
			name: "complete operational",
			operational: Operational{
				DeviceName: stringPtr("eth0"),
				MACaddress: stringPtr("00:11:22:33:44:55"),
				SrcUDPport: intPtr(10000),
				DstUDPport: intPtr(10001),
			},
			jsonStr: `{
				"device_name": "eth0",
				"mac_address": "00:11:22:33:44:55",
				"src_udp_port": 10000,
				"dst_udp_port": 10001
			}`,
		},
		{
			name: "operational with nulls",
			operational: Operational{
				DeviceName: nil,
				MACaddress: nil,
				SrcUDPport: nil,
				DstUDPport: nil,
			},
			jsonStr: `{
				"device_name": null,
				"mac_address": null,
				"src_udp_port": null,
				"dst_udp_port": null
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test unmarshaling
			var unmarshaled Operational
			err := json.Unmarshal([]byte(tt.jsonStr), &unmarshaled)
			assert.NoError(t, err)

			assert.Equal(t, tt.operational.DeviceName, unmarshaled.DeviceName)
			assert.Equal(t, tt.operational.MACaddress, unmarshaled.MACaddress)
			assert.Equal(t, tt.operational.SrcUDPport, unmarshaled.SrcUDPport)
			assert.Equal(t, tt.operational.DstUDPport, unmarshaled.DstUDPport)
		})
	}
}

func TestInterface_Exists(t *testing.T) {
	tests := []struct {
		name     string
		iface    Interface
		expected bool
	}{
		{
			name: "defined state",
			iface: Interface{
				State: IfaceStateDefined,
			},
			expected: false,
		},
		{
			name: "stopped state",
			iface: Interface{
				State: IfaceStateStopped,
			},
			expected: true,
		},
		{
			name: "started state",
			iface: Interface{
				State: IfaceStateStarted,
			},
			expected: true,
		},
		{
			name: "empty state",
			iface: Interface{
				State: "",
			},
			expected: true, // empty string != "DEFINED_ON_CORE"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.iface.Exists()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInterface_Runs(t *testing.T) {
	tests := []struct {
		name     string
		iface    Interface
		expected bool
	}{
		{
			name: "defined state",
			iface: Interface{
				State: IfaceStateDefined,
			},
			expected: false,
		},
		{
			name: "stopped state",
			iface: Interface{
				State: IfaceStateStopped,
			},
			expected: false,
		},
		{
			name: "started state",
			iface: Interface{
				State: IfaceStateStarted,
			},
			expected: true,
		},
		{
			name: "empty state",
			iface: Interface{
				State: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.iface.Runs()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInterface_IsPhysical(t *testing.T) {
	tests := []struct {
		name     string
		iface    Interface
		expected bool
	}{
		{
			name: "physical type",
			iface: Interface{
				Type: IfaceTypePhysical,
			},
			expected: true,
		},
		{
			name: "loopback type",
			iface: Interface{
				Type: IfaceTypeLoopback,
			},
			expected: false,
		},
		{
			name: "empty type",
			iface: Interface{
				Type: "",
			},
			expected: false,
		},
		{
			name: "other type",
			iface: Interface{
				Type: IfaceType("other"),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.iface.IsPhysical()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInterfaceList_JSON(t *testing.T) {
	interfaces := InterfaceList{
		{
			ID:          "iface1",
			Label:       "eth0",
			State:       IfaceStateStarted,
			Type:        IfaceTypePhysical,
			IsConnected: true,
		},
		{
			ID:          "iface2",
			Label:       "lo0",
			State:       IfaceStateDefined,
			Type:        IfaceTypeLoopback,
			IsConnected: false,
		},
	}

	// Test marshaling
	data, err := json.Marshal(interfaces)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled InterfaceList
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Len(t, unmarshaled, 2)
	assert.Equal(t, "iface1", string(unmarshaled[0].ID))
	assert.Equal(t, "eth0", unmarshaled[0].Label)
	assert.Equal(t, IfaceStateStarted, unmarshaled[0].State)
	assert.Equal(t, IfaceTypePhysical, unmarshaled[0].Type)
	assert.True(t, unmarshaled[0].IsConnected)

	assert.Equal(t, "iface2", string(unmarshaled[1].ID))
	assert.Equal(t, "lo0", unmarshaled[1].Label)
	assert.Equal(t, IfaceStateDefined, unmarshaled[1].State)
	assert.Equal(t, IfaceTypeLoopback, unmarshaled[1].Type)
	assert.False(t, unmarshaled[1].IsConnected)
}

func TestInterface_JSON_WithExtraFields(t *testing.T) {
	iface := Interface{
		ID:          "iface1",
		Label:       "eth0",
		State:       IfaceStateStarted,
		Type:        IfaceTypePhysical,
		IsConnected: true,
		IP4:         []string{"192.168.1.1/24", "10.0.0.1/8"},
		IP6:         []string{"2001:db8::1/32"},
	}

	// Test marshaling
	data, err := json.Marshal(iface)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled Interface
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, iface.ID, unmarshaled.ID)
	assert.Equal(t, iface.Label, unmarshaled.Label)
	assert.Equal(t, iface.State, unmarshaled.State)
	assert.Equal(t, iface.Type, unmarshaled.Type)
	assert.Equal(t, iface.IsConnected, unmarshaled.IsConnected)
	assert.Equal(t, iface.IP4, unmarshaled.IP4)
	assert.Equal(t, iface.IP6, unmarshaled.IP6)
}

func TestInterface_JSON_Minimal(t *testing.T) {
	iface := Interface{
		ID:          "iface1",
		Label:       "eth0",
		IsConnected: true,
	}

	// Test marshaling
	data, err := json.Marshal(iface)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled Interface
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, iface.ID, unmarshaled.ID)
	assert.Equal(t, iface.Label, unmarshaled.Label)
	assert.Equal(t, iface.IsConnected, unmarshaled.IsConnected)
	assert.Empty(t, unmarshaled.State)
	assert.Empty(t, unmarshaled.Type)
	assert.Nil(t, unmarshaled.Slot)
	assert.Nil(t, unmarshaled.Operational)
	assert.Empty(t, unmarshaled.IP4)
	assert.Empty(t, unmarshaled.IP6)
}
