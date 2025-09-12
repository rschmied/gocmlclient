package services

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/testutil"
	"github.com/rschmied/gocmlclient/pkg/models"
	"github.com/stretchr/testify/assert"
)

func addExtConnGetResponders() {
	// Mock responder for getting a single external connector
	httpmock.RegisterResponder("GET", "https://mock/api/v0/system/external_connectors/extconn-123",
		httpmock.NewStringResponder(200, `{
			"id": "extconn-123",
			"device_name": "eth0",
			"label": "External Connection 1",
			"protected": false,
			"snooped": true,
			"tags": ["tag1", "tag2"],
			"operational": {
				"forwarding": "bridged",
				"label": "eth0",
				"mtu": 1500,
				"exists": true,
				"enabled": true,
				"protected": false,
				"snooped": true,
				"stp": false,
				"ip_networks": ["192.168.1.0/24"]
			}
		}`))

	// Mock responder for listing all external connectors
	httpmock.RegisterResponder("GET", "https://mock/api/v0/system/external_connectors",
		httpmock.NewStringResponder(200, `[
			{
				"id": "extconn-123",
				"device_name": "eth0",
				"label": "External Connection 1",
				"protected": false,
				"snooped": true,
				"tags": ["tag1", "tag2"],
				"operational": {
					"forwarding": "bridged",
					"label": "eth0",
					"mtu": 1500,
					"exists": true,
					"enabled": true,
					"protected": false,
					"snooped": true,
					"stp": false,
					"ip_networks": ["192.168.1.0/24"]
				}
			},
			{
				"id": "extconn-456",
				"device_name": "eth1",
				"label": "External Connection 2",
				"protected": true,
				"snooped": false,
				"tags": ["tag3"],
				"operational": {
					"forwarding": "routed",
					"label": "eth1",
					"mtu": 9000,
					"exists": true,
					"enabled": false,
					"protected": true,
					"snooped": false,
					"stp": true,
					"ip_networks": ["10.0.0.0/8", "172.16.0.0/12"]
				}
			}
		]`))
}

func addExtConnErrorResponders() {
	// Mock responder for 404 error
	httpmock.RegisterResponder("GET", "https://mock/api/v0/system/external_connectors/nonexistent",
		httpmock.NewStringResponder(404, `{"error": "External connector not found"}`))

	// Mock responder for 500 error
	httpmock.RegisterResponder("GET", "https://mock/api/v0/system/external_connectors/error",
		httpmock.NewStringResponder(500, `{"error": "Internal server error"}`))
}

func initExtConnTest(t *testing.T, responders func()) (*api.Client, func()) {
	client, cleanup := testutil.NewAPIClient(t)
	if !testutil.IsLiveTesting() {
		responders()
	}
	return client, cleanup
}

func TestExtConnService_Get(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initExtConnTest(t, addExtConnGetResponders)
	defer cleanup()

	service := NewExtConnService(client)
	ctx := context.Background()

	extConn, err := service.Get(ctx, "extconn-123")

	assert.NoError(t, err)
	assert.Equal(t, models.UUID("extconn-123"), extConn.ID)
	assert.Equal(t, "eth0", extConn.DeviceName)
	assert.Equal(t, "External Connection 1", extConn.Label)
	assert.False(t, extConn.Protected)
	assert.True(t, extConn.Snooped)
	assert.Equal(t, []string{"tag1", "tag2"}, extConn.Tags)

	// Test operational data
	assert.Equal(t, "bridged", extConn.Operational.Forwarding)
	assert.Equal(t, "eth0", extConn.Operational.Label)
	assert.Equal(t, 1500, extConn.Operational.MTU)
	assert.True(t, extConn.Operational.Exists)
	assert.True(t, extConn.Operational.Enabled)
	assert.False(t, extConn.Operational.Protected)
	assert.True(t, extConn.Operational.Snooped)
	assert.False(t, extConn.Operational.STP)
	assert.Equal(t, []string{"192.168.1.0/24"}, extConn.Operational.IPNetworks)
}

func TestExtConnService_Get_Error(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initExtConnTest(t, addExtConnErrorResponders)
	defer cleanup()

	service := NewExtConnService(client)
	ctx := context.Background()

	// Test 404 error
	_, err := service.Get(ctx, "nonexistent")
	assert.Error(t, err)

	// Test 500 error
	_, err = service.Get(ctx, "error")
	assert.Error(t, err)
}

func TestExtConnService_List(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initExtConnTest(t, addExtConnGetResponders)
	defer cleanup()

	service := NewExtConnService(client)
	ctx := context.Background()

	extConns, err := service.List(ctx)

	assert.NoError(t, err)
	assert.Len(t, extConns, 2)

	// Test first external connector
	assert.Equal(t, models.UUID("extconn-123"), extConns[0].ID)
	assert.Equal(t, "eth0", extConns[0].DeviceName)
	assert.Equal(t, "External Connection 1", extConns[0].Label)
	assert.False(t, extConns[0].Protected)
	assert.True(t, extConns[0].Snooped)
	assert.Equal(t, []string{"tag1", "tag2"}, extConns[0].Tags)

	// Test second external connector
	assert.Equal(t, models.UUID("extconn-456"), extConns[1].ID)
	assert.Equal(t, "eth1", extConns[1].DeviceName)
	assert.Equal(t, "External Connection 2", extConns[1].Label)
	assert.True(t, extConns[1].Protected)
	assert.False(t, extConns[1].Snooped)
	assert.Equal(t, []string{"tag3"}, extConns[1].Tags)
	assert.Equal(t, 9000, extConns[1].Operational.MTU)
	assert.Equal(t, []string{"10.0.0.0/8", "172.16.0.0/12"}, extConns[1].Operational.IPNetworks)
}

func TestExtConnService_List_Error(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Register error responder
	httpmock.RegisterResponder("GET", "https://mock/api/v0/system/external_connectors",
		httpmock.NewStringResponder(500, `{"error": "Internal server error"}`))

	service := NewExtConnService(client)
	ctx := context.Background()

	_, err := service.List(ctx)
	assert.Error(t, err)
}

func TestExtConnService_NewExtConnService(t *testing.T) {
	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	service := NewExtConnService(client)
	assert.NotNil(t, service)
	assert.Equal(t, client, service.apiClient)
}
