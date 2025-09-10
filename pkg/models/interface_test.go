package models

import (
	"encoding/json"
	"testing"
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
		"operational": {
			"device_name": "eth0",
			"mac_address": "00:11:22:33:44:55",
			"src_udp_port": 10000,
			"dst_udp_port": 10001
		}
	}`

	var iface Interface
	err := json.Unmarshal([]byte(testJSON), &iface)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify all fields are properly set
	if iface.ID != "90f84e38-a71c-4d57-8d90-00fa8a197385" {
		t.Errorf("Expected ID %s, got %s", "90f84e38-a71c-4d57-8d90-00fa8a197385", iface.ID)
	}

	if iface.Label != "GigabitEthernet0/0" {
		t.Errorf("Expected Label %s, got %s", "GigabitEthernet0/0", iface.Label)
	}

	if iface.Slot == nil || *iface.Slot != 0 {
		t.Errorf("Expected Slot %d, got %v", 0, iface.Slot)
	}

	if !iface.IsConnected {
		t.Errorf("Expected IsConnected true, got %v", iface.IsConnected)
	}

	// Test marshaling back to JSON
	marshaled, err := json.Marshal(iface)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Verify the marshaled JSON contains the expected fields
	var unmarshaled map[string]any
	err = json.Unmarshal(marshaled, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal marshaled JSON: %v", err)
	}

	// Check that all expected fields are present
	expectedFields := []string{"id", "lab_id", "node", "label", "slot", "type", "state", "is_connected", "operational"}
	for _, field := range expectedFields {
		if _, exists := unmarshaled[field]; !exists {
			t.Errorf("Expected field %s in marshaled JSON", field)
		}
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
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON with null values: %v", err)
	}

	if iface.IsConnected != false {
		t.Errorf("Expected IsConnected to be false, got %v", iface.IsConnected)
	}
}
