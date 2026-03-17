package services

import (
	"context"
	"sort"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/rschmied/gocmlclient/internal/testutil"
	"github.com/rschmied/gocmlclient/pkg/models"
)

func TestInterfaceGetInterfacesForNode(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - requires specific lab and node setup")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	if !testutil.IsLiveTesting() {
		httpmock.RegisterResponder(
			"GET",
			"=~/labs/lab_id_1/nodes/node_id_1/interfaces",
			httpmock.NewJsonResponderOrPanic(200, []any{
				map[string]any{
					"id":           "iface_id_1",
					"lab_id":       "lab_id_1",
					"node":         "node_id_1",
					"label":        "GigabitEthernet0/0",
					"slot":         0,
					"type":         "physical",
					"state":        "STARTED",
					"is_connected": true,
					"mac_address":  "00:11:22:33:44:55",
					"src_udp_port": 10000,
					"dst_udp_port": 10001,
					"device_name":  "eth0",
				},
				map[string]any{
					"id":           "iface_id_2",
					"lab_id":       "lab_id_1",
					"node":         "node_id_1",
					"label":        "GigabitEthernet0/1",
					"slot":         1,
					"type":         "physical",
					"state":        "STARTED",
					"is_connected": false,
					"mac_address":  "00:11:22:33:44:56",
					"src_udp_port": 10002,
					"dst_udp_port": 10003,
					"device_name":  "eth1",
				},
			}),
		)
	}

	service := NewInterfaceService(client)

	interfaces, err := service.GetInterfacesForNode(context.Background(), "lab_id_1", "node_id_1")
	assert.NoError(t, err)
	assert.Len(t, interfaces, 2)
	assert.Equal(t, models.UUID("iface_id_1"), interfaces[0].ID)
	assert.Equal(t, models.UUID("iface_id_2"), interfaces[1].ID)
	assert.Equal(t, 0, *interfaces[0].Slot)
	assert.Equal(t, 1, *interfaces[1].Slot)
}

func TestInterfaceGetByID(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - requires specific lab and interface setup")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	if !testutil.IsLiveTesting() {
		httpmock.RegisterResponder(
			"GET",
			"=~/labs/lab_id_1/interfaces/iface_id_1",
			httpmock.NewJsonResponderOrPanic(200, map[string]any{
				"id":           "iface_id_1",
				"lab_id":       "lab_id_1",
				"node":         "node_id_1",
				"label":        "GigabitEthernet0/0",
				"slot":         0,
				"type":         "physical",
				"state":        "STARTED",
				"is_connected": true,
				"mac_address":  "00:11:22:33:44:55",
				"src_udp_port": 10000,
				"dst_udp_port": 10001,
				"device_name":  "eth0",
			}),
		)
	}

	service := NewInterfaceService(client)

	iface, err := service.GetByID(context.Background(), "lab_id_1", "iface_id_1")
	assert.NoError(t, err)
	assert.Equal(t, models.UUID("iface_id_1"), iface.ID)
	assert.Equal(t, "GigabitEthernet0/0", iface.Label)
	assert.Equal(t, 0, *iface.Slot)
	assert.True(t, iface.IsConnected)
}

func TestInterfaceCreateWithSlot(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - requires specific lab and node setup")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	if !testutil.IsLiveTesting() {
		// Mock Create with slot specified (returns array)
		httpmock.RegisterResponder(
			"POST",
			"=~/labs/lab_id_1/interfaces",
			httpmock.NewJsonResponderOrPanic(200, []any{
				map[string]any{
					"id":     "iface_id_1",
					"lab_id": "lab_id_1",
					"node":   "node_id_1",
					"label":  "GigabitEthernet0/0",
					"slot":   0,
				},
			}),
		)
	}

	service := NewInterfaceService(client)

	// Test creating with slot specified (returns array)
	iface, err := service.Create(context.Background(), "lab_id_1", "node_id_1", 0)
	assert.NoError(t, err)
	assert.Equal(t, models.UUID("iface_id_1"), iface.ID)
	assert.Equal(t, 0, *iface.Slot)
}

func TestInterfaceCreateWithoutSlot(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - requires specific lab and node setup")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	if !testutil.IsLiveTesting() {
		// Mock Create without slot specified (returns single object)
		httpmock.RegisterResponder(
			"POST",
			"=~/labs/lab_id_2/interfaces",
			httpmock.NewJsonResponderOrPanic(200, map[string]any{
				"id":     "iface_id_2",
				"lab_id": "lab_id_2",
				"node":   "node_id_2",
				"label":  "GigabitEthernet0/1",
				"slot":   1,
			}),
		)
	}

	service := NewInterfaceService(client)

	// Test creating without slot specified (returns single object)
	iface, err := service.Create(context.Background(), "lab_id_2", "node_id_2", -1)
	assert.NoError(t, err)
	assert.Equal(t, models.UUID("iface_id_2"), iface.ID)
	assert.Equal(t, 1, *iface.Slot)
}

func TestInterfaceGetByID_NotFound(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - requires specific lab setup")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	if !testutil.IsLiveTesting() {
		httpmock.RegisterResponder(
			"GET",
			"=~/labs/lab_id_1/interfaces/nonexistent",
			httpmock.NewStringResponder(404, `{"detail":"Interface not found"}`),
		)
	}

	service := NewInterfaceService(client)

	_, err := service.GetByID(context.Background(), "lab_id_1", "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestInterfaceServerError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - requires specific lab and node setup")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	if !testutil.IsLiveTesting() {
		httpmock.RegisterResponder(
			"GET",
			"=~/labs/error_lab/nodes/error_node/interfaces",
			httpmock.NewStringResponder(500, `{"detail":"Internal server error"}`),
		)
	}

	service := NewInterfaceService(client)

	_, err := service.GetInterfacesForNode(context.Background(), "error_lab", "error_node")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server error")
}

func TestInterfaceAuthError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - requires specific lab setup")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	if !testutil.IsLiveTesting() {
		httpmock.RegisterResponder(
			"GET",
			"=~/labs/lab_id_1/interfaces/auth_error",
			httpmock.NewStringResponder(401, `{"detail":"Unauthorized"}`),
		)
	}

	service := NewInterfaceService(client)

	_, err := service.GetByID(context.Background(), "lab_id_1", "auth_error")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unauthorized")
}

func TestInterfacePermissionError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - requires specific lab setup")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	if !testutil.IsLiveTesting() {
		httpmock.RegisterResponder(
			"GET",
			"=~/labs/lab_id_1/interfaces/perm_error",
			httpmock.NewStringResponder(403, `{"detail":"Forbidden"}`),
		)
	}

	service := NewInterfaceService(client)

	_, err := service.GetByID(context.Background(), "lab_id_1", "perm_error")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Forbidden")
}

func TestInterfaceMalformedJSON(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - requires specific lab setup")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	if !testutil.IsLiveTesting() {
		httpmock.RegisterResponder(
			"GET",
			"=~/labs/lab_id_1/interfaces/malformed_json",
			httpmock.NewStringResponder(200, `{invalid json`),
		)
	}

	service := NewInterfaceService(client)

	_, err := service.GetByID(context.Background(), "lab_id_1", "malformed_json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character")
}

func TestInterfaceConnectionError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Connection error testing requires special network setup")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	service := NewInterfaceService(client)

	_, err := service.GetByID(context.Background(), "lab_id_1", "iface_id_1")
	assert.Error(t, err)
}

func TestSortInterfacesBySlot(t *testing.T) {
	// Create test interfaces with various slot configurations
	interfaces := models.InterfaceList{
		{Slot: intPtr(2)}, // slot 2
		{Slot: intPtr(1)}, // slot 1
		{Slot: nil},       // nil slot
		{Slot: intPtr(0)}, // slot 0
		{Slot: nil},       // another nil slot
		{Slot: intPtr(3)}, // slot 3
	}

	// Test the sort function directly
	result := sort.SliceIsSorted(interfaces, func(i, j int) bool {
		return sortInterfacesBySlot(i, j, interfaces)
	})
	assert.False(t, result, "Interfaces should not be sorted initially")

	// Sort the interfaces using our function
	sort.Slice(interfaces, func(i, j int) bool {
		return sortInterfacesBySlot(i, j, interfaces)
	})

	// Verify the sorting order: nil slots first, then sorted by slot number
	expectedOrder := []*int{nil, nil, intPtr(0), intPtr(1), intPtr(2), intPtr(3)}

	for i, iface := range interfaces {
		if expectedOrder[i] == nil {
			assert.Nil(t, iface.Slot, "Expected nil slot at position %d", i)
		} else {
			assert.NotNil(t, iface.Slot, "Expected non-nil slot at position %d", i)
			if iface.Slot != nil {
				assert.Equal(t, *expectedOrder[i], *iface.Slot, "Wrong slot value at position %d", i)
			}
		}
	}

	// Test specific edge cases
	testCases := []struct {
		name     string
		iface1   *models.Interface
		iface2   *models.Interface
		expected bool
	}{
		{
			name:     "both nil slots",
			iface1:   &models.Interface{Slot: nil},
			iface2:   &models.Interface{Slot: nil},
			expected: false,
		},
		{
			name:     "first nil, second not nil",
			iface1:   &models.Interface{Slot: nil},
			iface2:   &models.Interface{Slot: intPtr(1)},
			expected: true,
		},
		{
			name:     "first not nil, second nil",
			iface1:   &models.Interface{Slot: intPtr(1)},
			iface2:   &models.Interface{Slot: nil},
			expected: false,
		},
		{
			name:     "both not nil, first smaller",
			iface1:   &models.Interface{Slot: intPtr(1)},
			iface2:   &models.Interface{Slot: intPtr(2)},
			expected: true,
		},
		{
			name:     "both not nil, first larger",
			iface1:   &models.Interface{Slot: intPtr(2)},
			iface2:   &models.Interface{Slot: intPtr(1)},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interfaceList := models.InterfaceList{tc.iface1, tc.iface2}
			result := sortInterfacesBySlot(0, 1, interfaceList)
			assert.Equal(t, tc.expected, result, "Test case %s failed", tc.name)
		})
	}
}

// intPtr is a helper function to create a pointer to an int
func intPtr(i int) *int {
	return &i
}
