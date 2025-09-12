package services

import (
	"context"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/testutil"
	"github.com/rschmied/gocmlclient/pkg/models"
	"github.com/stretchr/testify/assert"
)

func addLabsGetResponders() {
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs",
		httpmock.NewStringResponder(200, `["uuid-1", "uuid-2", "uuid-3"]`))
}

func addLabCreateResponders() {
	httpmock.RegisterResponder("POST", "https://mock/api/v0/labs",
		httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"DEFINED_ON_CORE","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
	httpmock.RegisterResponder("PATCH", "https://mock/api/v0/labs/lab_uuid",
		httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"DEFINED_ON_CORE","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid",
		httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"DEFINED_ON_CORE","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
	httpmock.RegisterResponder("PUT", "https://mock/api/v0/labs/lab_uuid/state/start",
		httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"QUEUED","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
	httpmock.RegisterResponder("PUT", "https://mock/api/v0/labs/lab_uuid/state/stop",
		httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"STOPPED","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
	httpmock.RegisterResponder("PUT", "https://mock/api/v0/labs/lab_uuid/state/wipe",
		httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"DEFINED_ON_CORE","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
	httpmock.RegisterResponder("DELETE", "https://mock/api/v0/labs/lab_uuid",
		httpmock.NewJsonResponderOrPanic(204, nil))
}

func initLabTest(t *testing.T, responders func()) (*api.Client, func()) {
	client, cleanup := testutil.NewAPIClient(t)
	if !testutil.IsLiveTesting() {
		responders()
	}
	return client, cleanup
}

func TestLabs(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initLabTest(t, addLabsGetResponders)
	defer cleanup()

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	labs, err := service.Labs(ctx, true)
	if err != nil {
		testutil.PrettyPrintError(err)
	}
	assert.NoError(t, err)
	assert.Len(t, labs, 3)
}

func TestLabCreate(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - requires lab lifecycle management permissions")
	}

	client, cleanup := initLabTest(t, addLabCreateResponders)
	defer cleanup()

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	lab := models.LabCreateRequest{Title: "this"}
	newLab, err := service.Create(ctx, lab)
	if err != nil {
		testutil.PrettyPrintError(err)
	}
	assert.NoError(t, err)

	newLab, err = service.GetByID(ctx, newLab.ID, false)
	assert.NoError(t, err)

	err = service.Start(ctx, newLab.ID)
	assert.NoError(t, err)

	err = service.Stop(ctx, newLab.ID)
	assert.NoError(t, err)

	err = service.Wipe(ctx, newLab.ID)
	assert.NoError(t, err)

	err = service.Delete(ctx, newLab.ID)
	assert.NoError(t, err)
}

func TestGetByIDDeep(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initLabTest(t, func() {
		// Mock responder for basic lab data
		httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid",
			httpmock.NewStringResponder(200, `{
				"id":"lab_uuid",
				"state":"DEFINED_ON_CORE",
				"created":"2025-08-26T09:41:36+00:00",
				"modified":"2025-08-26T09:41:36+00:00",
				"lab_title":"test lab",
				"owner":"00000000-0000-4000-a000-000000000000",
				"owner_username":"admin",
				"effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]
			}`))

		// Mock responders for deep fetch data
		httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid/nodes",
			httpmock.NewStringResponder(200, `[
				{
					"id": "node1",
					"label": "test-node",
					"node_definition": "test-def",
					"state": "DEFINED_ON_CORE"
				}
			]`))

		// Mock responder for user data
		httpmock.RegisterResponder("GET", "https://mock/api/v0/users/00000000-0000-4000-a000-000000000000",
			httpmock.NewStringResponder(200, `{
				"id": "00000000-0000-4000-a000-000000000000",
				"username": "admin",
				"fullname": "Administrator",
				"admin": true
			}`))

		// Mock responder for L3 info
		httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid/state/layer3_addresses",
			httpmock.NewStringResponder(200, `{
				"node1": {
					"name": "R1",
					"interfaces": {
						"52:54:00:b3:0b:ed": {
							"id": "eth0",
							"label": "Ethernet 0",
							"ip4": ["10.0.10.1"],
							"ip6": []
						}
					}
				}
			}`))

		httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid/links",
			httpmock.NewStringResponder(200, `[
				{
					"id": "link1",
					"src": "node1",
					"dst": "node2",
					"src_int": "eth0",
					"dst_int": "eth0"
				}
			]`))

		httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid/nodes/node1/interfaces",
			httpmock.NewStringResponder(200, `[
				{
					"id": "eth0",
					"label": "Ethernet 0",
					"mac_address": "aa:bb:cc:dd:ee:ff",
					"is_connected": true
				}
			]`))
	})
	defer cleanup()

	// Create mock services
	nodeService := NewNodeService(client, false)
	interfaceService := NewInterfaceService(client)
	linkService := NewLinkService(client)
	linkService.Interface = interfaceService
	linkService.Node = nodeService
	groupService := NewGroupService(client)
	userService := NewUserService(client, groupService)

	service := NewLabService(client, interfaceService, linkService, userService, nodeService)
	ctx := context.Background()

	// Test deep=false (should not fetch additional data)
	lab, err := service.GetByID(ctx, "lab_uuid", false)
	assert.NoError(t, err)
	assert.Equal(t, models.UUID("lab_uuid"), lab.ID)
	assert.Empty(t, lab.Nodes) // Should be empty without deep fetch
	assert.Empty(t, lab.Links) // Should be empty without deep fetch

	// Test deep=true (should fetch additional data)
	lab, err = service.GetByID(ctx, "lab_uuid", true)
	assert.NoError(t, err)
	assert.Equal(t, models.UUID("lab_uuid"), lab.ID)
	assert.NotEmpty(t, lab.Nodes) // Should have nodes with deep fetch
	assert.NotEmpty(t, lab.Links) // Should have links with deep fetch

	// Verify node has interfaces populated
	if len(lab.Nodes) > 0 {
		node := lab.Nodes["node1"]
		assert.NotEmpty(t, node.Interfaces) // Should have interfaces populated
	}
}

func TestGetByIDDeepErrorHandling(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initLabTest(t, func() {
		// Mock responder for basic lab data
		httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/error_lab",
			httpmock.NewStringResponder(200, `{
				"id":"error_lab",
				"state":"DEFINED_ON_CORE",
				"lab_title":"error test lab"
			}`))

		// Mock error responder for nodes
		httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/error_lab/nodes",
			httpmock.NewStringResponder(500, `{"error": "internal server error"}`))
	})
	defer cleanup()

	nodeService := NewNodeService(client, false)
	interfaceService := NewInterfaceService(client)
	linkService := NewLinkService(client)

	service := NewLabService(client, interfaceService, linkService, nil, nodeService)
	ctx := context.Background()

	// Test that errors in deep fetch are properly handled
	_, err := service.GetByID(ctx, "error_lab", true)
	assert.Error(t, err)
	// The error could come from nodes or links fetch - both are acceptable
	assert.True(t, strings.Contains(err.Error(), "failed to get nodes") ||
		strings.Contains(err.Error(), "failed to get links"))
}
