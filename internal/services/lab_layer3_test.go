package services

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/testutil"
)

func addL3InfoResponders() {
	// Mock responder for successful L3 info retrieval
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-123/layer3_addresses",
		httpmock.NewStringResponder(200, `{
			"node1": {
				"name": "node1",
				"interfaces": {
					"eth0": {
						"id": "eth0",
						"label": "Ethernet 0",
						"ip4": ["192.168.1.10/24", "10.0.0.1/8"],
						"ip6": ["2001:db8::1/32", "fe80::1/64"]
					},
					"eth1": {
						"id": "eth1",
						"label": "Ethernet 1",
						"ip4": ["172.16.0.1/16"],
						"ip6": []
					}
				}
			},
			"node2": {
				"name": "node2",
				"interfaces": {
					"eth0": {
						"id": "eth0",
						"label": "Ethernet 0",
						"ip4": [],
						"ip6": ["2001:db8:1::1/48"]
					}
				}
			}
		}`))

	// Mock responder for empty L3 info
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/empty-lab/layer3_addresses",
		httpmock.NewStringResponder(200, `{}`))

	// Mock responder for single node with minimal interfaces
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/single-node-lab/layer3_addresses",
		httpmock.NewStringResponder(200, `{
			"router1": {
				"name": "router1",
				"interfaces": {
					"GigabitEthernet0/0": {
						"id": "GigabitEthernet0/0",
						"label": "GE0/0",
						"ip4": ["203.0.113.1/24"],
						"ip6": ["2001:db8:2::1/64"]
					}
				}
			}
		}`))
}

func addL3InfoErrorResponders() {
	// Mock responder for 404 error
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/nonexistent-lab/layer3_addresses",
		httpmock.NewStringResponder(404, `{"error": "Lab not found"}`))

	// Mock responder for 500 error
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/error-lab/layer3_addresses",
		httpmock.NewStringResponder(500, `{"error": "Internal server error"}`))

	// Mock responder for malformed JSON
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/malformed-lab/layer3_addresses",
		httpmock.NewStringResponder(200, `{"invalid": json}`))
}

func initL3Test(t *testing.T, responders func()) (*api.Client, func()) {
	client, cleanup := testutil.NewAPIClient(t)
	if !testutil.IsLiveTesting() {
		responders()
	}
	return client, cleanup
}

func TestLabService_getL3Info(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initL3Test(t, addL3InfoResponders)
	defer cleanup()

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	nodes, err := service.getL3Info(ctx, "lab-123")

	assert.NoError(t, err)
	assert.NotNil(t, nodes)
	assert.Len(t, *nodes, 2)

	// Test node1
	node1, exists := (*nodes)["node1"]
	assert.True(t, exists)
	assert.Equal(t, "node1", node1.Name)
	assert.Len(t, node1.Interfaces, 2)

	// Test node1 eth0 interface
	eth0, exists := node1.Interfaces["eth0"]
	assert.True(t, exists)
	assert.Equal(t, "eth0", eth0.ID)
	assert.Equal(t, "Ethernet 0", eth0.Label)
	assert.Equal(t, []string{"192.168.1.10/24", "10.0.0.1/8"}, eth0.IP4)
	assert.Equal(t, []string{"2001:db8::1/32", "fe80::1/64"}, eth0.IP6)

	// Test node1 eth1 interface
	eth1, exists := node1.Interfaces["eth1"]
	assert.True(t, exists)
	assert.Equal(t, "eth1", eth1.ID)
	assert.Equal(t, "Ethernet 1", eth1.Label)
	assert.Equal(t, []string{"172.16.0.1/16"}, eth1.IP4)
	assert.Equal(t, []string{}, eth1.IP6)

	// Test node2
	node2, exists := (*nodes)["node2"]
	assert.True(t, exists)
	assert.Equal(t, "node2", node2.Name)
	assert.Len(t, node2.Interfaces, 1)

	// Test node2 eth0 interface
	node2eth0, exists := node2.Interfaces["eth0"]
	assert.True(t, exists)
	assert.Equal(t, "eth0", node2eth0.ID)
	assert.Equal(t, "Ethernet 0", node2eth0.Label)
	assert.Equal(t, []string{}, node2eth0.IP4)
	assert.Equal(t, []string{"2001:db8:1::1/48"}, node2eth0.IP6)
}

func TestLabService_getL3Info_EmptyResponse(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initL3Test(t, addL3InfoResponders)
	defer cleanup()

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	nodes, err := service.getL3Info(ctx, "empty-lab")

	assert.NoError(t, err)
	assert.NotNil(t, nodes)
	assert.Len(t, *nodes, 0)
}

func TestLabService_getL3Info_SingleNode(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initL3Test(t, addL3InfoResponders)
	defer cleanup()

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	nodes, err := service.getL3Info(ctx, "single-node-lab")

	assert.NoError(t, err)
	assert.NotNil(t, nodes)
	assert.Len(t, *nodes, 1)

	// Test router1
	router1, exists := (*nodes)["router1"]
	assert.True(t, exists)
	assert.Equal(t, "router1", router1.Name)
	assert.Len(t, router1.Interfaces, 1)

	// Test GigabitEthernet0/0 interface
	ge00, exists := router1.Interfaces["GigabitEthernet0/0"]
	assert.True(t, exists)
	assert.Equal(t, "GigabitEthernet0/0", ge00.ID)
	assert.Equal(t, "GE0/0", ge00.Label)
	assert.Equal(t, []string{"203.0.113.1/24"}, ge00.IP4)
	assert.Equal(t, []string{"2001:db8:2::1/64"}, ge00.IP6)
}

func TestLabService_getL3Info_Error_404(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initL3Test(t, addL3InfoErrorResponders)
	defer cleanup()

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	nodes, err := service.getL3Info(ctx, "nonexistent-lab")

	assert.Error(t, err)
	assert.Nil(t, nodes)
}

func TestLabService_getL3Info_Error_500(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initL3Test(t, addL3InfoErrorResponders)
	defer cleanup()

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	nodes, err := service.getL3Info(ctx, "error-lab")

	assert.Error(t, err)
	assert.Nil(t, nodes)
}

func TestLabService_getL3Info_Error_MalformedJSON(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initL3Test(t, addL3InfoErrorResponders)
	defer cleanup()

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	nodes, err := service.getL3Info(ctx, "malformed-lab")

	assert.Error(t, err)
	assert.Nil(t, nodes)
}

func TestLabService_getL3Info_NoInterfaces(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Register response for node with no interfaces
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/no-interfaces-lab/layer3_addresses",
		httpmock.NewStringResponder(200, `{
			"node1": {
				"name": "node1",
				"interfaces": {}
			}
		}`))

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	nodes, err := service.getL3Info(ctx, "no-interfaces-lab")

	assert.NoError(t, err)
	assert.NotNil(t, nodes)
	assert.Len(t, *nodes, 1)

	node1, exists := (*nodes)["node1"]
	assert.True(t, exists)
	assert.Equal(t, "node1", node1.Name)
	assert.Len(t, node1.Interfaces, 0)
}

func TestLabService_getL3Info_InterfaceWithoutIPs(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Register response for interface with no IP addresses
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/no-ips-lab/layer3_addresses",
		httpmock.NewStringResponder(200, `{
			"node1": {
				"name": "node1",
				"interfaces": {
					"eth0": {
						"id": "eth0",
						"label": "Ethernet 0",
						"ip4": [],
						"ip6": []
					}
				}
			}
		}`))

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	nodes, err := service.getL3Info(ctx, "no-ips-lab")

	assert.NoError(t, err)
	assert.NotNil(t, nodes)
	assert.Len(t, *nodes, 1)

	node1, exists := (*nodes)["node1"]
	assert.True(t, exists)
	assert.Equal(t, "node1", node1.Name)
	assert.Len(t, node1.Interfaces, 1)

	eth0, exists := node1.Interfaces["eth0"]
	assert.True(t, exists)
	assert.Equal(t, "eth0", eth0.ID)
	assert.Equal(t, "Ethernet 0", eth0.Label)
	assert.Equal(t, []string{}, eth0.IP4)
	assert.Equal(t, []string{}, eth0.IP6)
}
