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

func TestGetByTitle(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initLabTest(t, func() {
		// Mock responder for populate_lab_tiles
		httpmock.RegisterResponder("GET", "https://mock/api/v0/populate_lab_tiles",
			httpmock.NewStringResponder(200, `{
				"lab_tiles": {
					"uuid-1": {
						"id": "uuid-1",
						"lab_title": "Lab One",
						"state": "STOPPED",
						"owner": "owner-uuid",
						"owner_username": "admin",
						"effective_permissions": ["lab_admin"]
					},
					"uuid-2": {
						"id": "uuid-2",
						"lab_title": "Lab Two",
						"state": "STARTED",
						"owner": "owner-uuid",
						"owner_username": "admin",
						"effective_permissions": ["lab_view"]
					}
				}
			}`))

		// Mock responder for individual lab details
		httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/uuid-1",
			httpmock.NewStringResponder(200, `{
				"id": "uuid-1",
				"lab_title": "Lab One",
				"state": "STOPPED",
				"owner": "owner-uuid",
				"owner_username": "admin",
				"effective_permissions": ["lab_admin"]
			}`))
	})
	defer cleanup()

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	// Test finding lab by title
	lab, err := service.GetByTitle(ctx, "Lab One", false)
	assert.NoError(t, err)
	assert.Equal(t, models.UUID("uuid-1"), lab.ID)
	assert.Equal(t, "Lab One", lab.Title)

	// Test finding non-existent lab
	_, err = service.GetByTitle(ctx, "Non-existent Lab", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lab with title \"Non-existent Lab\" not found")
}

func TestLabsWithData(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initLabTest(t, func() {
		// Mock responder for populate_lab_tiles
		httpmock.RegisterResponder("GET", "https://mock/api/v0/populate_lab_tiles",
			httpmock.NewStringResponder(200, `{
				"lab_tiles": {
					"lab1": {
						"id": "lab1",
						"lab_title": "Lab One",
						"state": "DEFINED_ON_CORE",
						"node_count": 5,
						"link_count": 4,
						"owner_username": "admin",
						"effective_permissions": ["lab_admin"]
					},
					"lab2": {
						"id": "lab2",
						"lab_title": "Lab Two",
						"state": "STOPPED",
						"node_count": 3,
						"link_count": 2,
						"owner_username": "user",
						"effective_permissions": ["lab_view"]
					}
				}
			}`))
	})
	defer cleanup()

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()
	labs, err := service.LabsWithData(ctx)
	assert.NoError(t, err)
	assert.Len(t, labs, 2)

	// Check first lab
	assert.Equal(t, "lab1", string(labs[0].ID))
	assert.Equal(t, "Lab One", labs[0].Title)
	assert.Equal(t, models.LabStateDefined, labs[0].State)
	assert.Equal(t, 5, labs[0].NodeCount)
	assert.Equal(t, 4, labs[0].LinkCount)
	assert.Equal(t, "admin", labs[0].OwnerUsername)
	assert.Equal(t, models.Permissions{models.PermissionAdmin}, labs[0].EffectivePermissions)

	// Check second lab
	assert.Equal(t, "lab2", string(labs[1].ID))
	assert.Equal(t, "Lab Two", labs[1].Title)
	assert.Equal(t, models.LabStateStopped, labs[1].State)
	assert.Equal(t, 3, labs[1].NodeCount)
	assert.Equal(t, 2, labs[1].LinkCount)
	assert.Equal(t, "user", labs[1].OwnerUsername)
	assert.Equal(t, models.Permissions{models.PermissionView}, labs[1].EffectivePermissions)
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
		httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid/layer3_addresses",
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

func TestLabUpdate(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Register mock responder for lab update
	httpmock.RegisterResponder("PATCH", "https://mock/api/v0/labs/lab-123",
		httpmock.NewStringResponder(200, `{
			"id": "lab-123",
			"lab_title": "Updated Lab Title",
			"lab_description": "Updated description",
			"lab_notes": "Updated notes",
			"state": "DEFINED_ON_CORE",
			"owner": "owner-uuid",
			"owner_username": "admin",
			"effective_permissions": ["lab_admin"],
			"nodes": {},
			"links": []
		}`))

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	updateData := models.LabUpdateRequest{
		Title:       "Updated Lab Title",
		Description: "Updated description",
		Notes:       "Updated notes",
	}

	updatedLab, err := service.Update(ctx, "lab-123", updateData)

	assert.NoError(t, err)
	assert.Equal(t, models.UUID("lab-123"), updatedLab.ID)
	assert.Equal(t, "Updated Lab Title", updatedLab.Title)
	assert.Equal(t, "Updated description", updatedLab.Description)
	assert.Equal(t, "Updated notes", updatedLab.Notes)
}

func TestLabUpdate_Error(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Register mock responder for error
	httpmock.RegisterResponder("PATCH", "https://mock/api/v0/labs/nonexistent",
		httpmock.NewStringResponder(404, `{"error": "Lab not found"}`))

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	updateData := models.LabUpdateRequest{
		Title: "Updated Title",
	}

	_, err := service.Update(ctx, "nonexistent", updateData)
	assert.Error(t, err)
}

func TestLabImport(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Mock the import response
	httpmock.RegisterResponder("POST", "https://mock/api/v0/import",
		httpmock.NewStringResponder(200, `{
			"id": "imported-lab-123",
			"warnings": ["Warning: interface eth0 not found"]
		}`))

	// Mock the user endpoint for fillLabData
	httpmock.RegisterResponder("GET", "https://mock/api/v0/users/owner-uuid",
		httpmock.NewStringResponder(200, `{
			"id": "owner-uuid",
			"username": "admin",
			"fullname": "Administrator",
			"admin": true
		}`))

	// Mock nodes endpoint for fillLabData
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/imported-lab-123/nodes?data=true",
		httpmock.NewStringResponder(200, `[]`))

	// Mock links endpoint for fillLabData
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/imported-lab-123/links?data=true",
		httpmock.NewStringResponder(200, `[]`))

	// Mock L3 info endpoint for fillLabData
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/imported-lab-123/layer3_addresses",
		httpmock.NewStringResponder(200, `{}`))

	// Mock the GetByID response for the imported lab
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/imported-lab-123",
		httpmock.NewStringResponder(200, `{
			"id": "imported-lab-123",
			"lab_title": "Imported Lab",
			"state": "DEFINED_ON_CORE",
			"owner": "owner-uuid",
			"owner_username": "admin",
			"effective_permissions": ["lab_admin"],
			"nodes": {},
			"links": []
		}`))

	// Create service dependencies to avoid nil pointer dereference
	groupService := NewGroupService(client)
	userService := NewUserService(client, groupService)
	nodeService := NewNodeService(client, false)
	interfaceService := NewInterfaceService(client)
	linkService := NewLinkService(client)

	service := NewLabService(client, interfaceService, linkService, userService, nodeService)
	ctx := context.Background()

	yamlTopology := `
lab:
  title: Imported Lab
  description: A test lab
nodes:
  - id: n0
    node_definition: iosv
    image_definition: iosv-159
`

	importedLab, err := service.Import(ctx, yamlTopology)

	assert.NoError(t, err)
	assert.Equal(t, models.UUID("imported-lab-123"), importedLab.ID)
	assert.Equal(t, "Imported Lab", importedLab.Title)
}

func TestLabImport_Error(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Register mock responder for import error
	httpmock.RegisterResponder("POST", "https://mock/api/v0/import",
		httpmock.NewStringResponder(400, `{"error": "Invalid YAML format"}`))

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	_, err := service.Import(ctx, "invalid yaml content")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "import lab")
}

func TestLabHasConverged(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Register mock responder for converged state
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-123/check_if_converged",
		httpmock.NewStringResponder(200, `true`))

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	converged, err := service.HasConverged(ctx, "lab-123")

	assert.NoError(t, err)
	assert.True(t, converged)
}

func TestLabHasConverged_False(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Register mock responder for not converged state
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-456/check_if_converged",
		httpmock.NewStringResponder(200, `false`))

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	converged, err := service.HasConverged(ctx, "lab-456")

	assert.NoError(t, err)
	assert.False(t, converged)
}

func TestLabHasConverged_Error(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Register mock responder for error
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/error-lab/check_if_converged",
		httpmock.NewStringResponder(500, `{"error": "Internal server error"}`))

	service := NewLabService(client, nil, nil, nil, nil)
	ctx := context.Background()

	_, err := service.HasConverged(ctx, "error-lab")
	assert.Error(t, err)
}
