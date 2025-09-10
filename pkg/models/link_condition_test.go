package models

import (
	"encoding/json"
	"testing"
)

func TestConditionResponse_JSON(t *testing.T) {
	config := LinkConditionConfiguration{
		Bandwidth: 1000,
		Latency:   10,
		Enabled:   true,
	}

	operational := LinkConditionStricted{
		Bandwidth: 1000,
		Latency:   10,
	}

	response := ConditionResponse{
		LinkConditionConfiguration: config,
		Operational:                &operational,
	}

	// Test marshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal ConditionResponse: %v", err)
	}

	// Test unmarshaling
	var unmarshaled ConditionResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ConditionResponse: %v", err)
	}

	if unmarshaled.Bandwidth != response.Bandwidth {
		t.Errorf("Bandwidth mismatch: got %d, want %d", unmarshaled.Bandwidth, response.Bandwidth)
	}
	if unmarshaled.Operational == nil || unmarshaled.Operational.Bandwidth != operational.Bandwidth {
		t.Errorf("Operational bandwidth mismatch")
	}
}
