package services

import (
	"context"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/testutil"
	"github.com/rschmied/gocmlclient/pkg/models"
	"github.com/stretchr/testify/assert"
)

func initNodeTest(t *testing.T) (*api.Client, func()) {
	client, cleanup := testutil.NewAPIClient(t)
	if !testutil.IsLiveTesting() {
		// Mock responses will be registered in individual tests
		addLabCreateResponders()
	}
	return client, cleanup
}

func TestNodeCRUD(t *testing.T) {
	client, cleanup := initNodeTest(t)
	defer cleanup()

	if !testutil.IsLiveTesting() {
		// provide different responses when calling the GET multiple times
		var createCounter int

		// Mock responses for CRUD operations
		createResponse := `{"id": "test-node-id"}`
		nodeResponse := `{
			"boot_disk_size": 4096,
			"configuration": "# test configuration",
			"cpu_limit": 20,
			"cpus": 1,
			"data_volume": 4096,
			"hide_links": true,
			"id": "test-node-id",
			"image_definition": null,
			"lab_id": "lab_uuid",
			"label": "ubuntu-0",
			"node_definition": "ubuntu",
			"parameters": {
				"smbios.bios.vendor": "Lenovo"
			},
			"pinned_compute_id": null,
			"ram": 1,
			"tags": ["test"],
			"x": -15000,
			"y": -15000,
			"state": "DEFINED_ON_CORE",
			"boot_progress": "Not running"
		}`
		namedConfigResponse := `{
			"boot_disk_size": null,
			"configuration": [
				{
					"name": "user-data",
					"content": "#cloud-config\\nhostname: inserthostname-here\\nmanage_etc_hosts: True\\nsystem_info:\\n  default_user:\\n    name: cisco\\npassword: cisco\\nchpasswd: { expire: False }\\nssh_pwauth: True\\nssh_authorized_keys:\\n  - your-ssh-pubkey-line-goes-here\\n"
				},
				{
					"name": "network-config",
					"content": "#network-config\\nnetwork:\\n  version: 2\\n  ethernets:\\n    ens2:\\n      dhcp4: true\\n"
				}
			],
			"cpu_limit": null,
			"cpus": null,
			"data_volume": null,
			"hide_links": false,
			"id": "test-node-id",
			"image_definition": null,
			"lab_id": "lab_uuid",
			"label": "ubuntu-0",
			"node_definition": "ubuntu",
			"parameters": {},
			"pinned_compute_id": null,
			"ram": null,
			"tags": [],
			"x": -240,
			"y": -40,
			"state": "DEFINED_ON_CORE",
			"boot_progress": "Not running"
		}`

		_ = createCounter
		_ = namedConfigResponse

		getResponder := func(req *http.Request) (*http.Response, error) {
			_ = req
			createCounter++
			if createCounter == 1 {
				return httpmock.NewStringResponse(200, nodeResponse), nil
			}
			return httpmock.NewStringResponse(200, namedConfigResponse), nil
		}

		httpmock.RegisterResponder("POST", "https://mock/api/v0/labs/lab_uuid/nodes",
			httpmock.NewStringResponder(200, createResponse))
		httpmock.RegisterResponder("PATCH", "https://mock/api/v0/labs/lab_uuid/nodes/test-node-id",
			httpmock.NewStringResponder(200, `"test-node-id"`))
		httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid/nodes/test-node-id", getResponder)
		httpmock.RegisterResponder("DELETE", "https://mock/api/v0/labs/lab_uuid/nodes/test-node-id",
			httpmock.NewJsonResponderOrPanic(204, nil))

	}

	ctx := context.Background()
	labService := NewLabService(client, nil, nil, nil)
	nodeService := NewNodeService(client, false)

	lab := models.LabCreateRequest{Title: "this"}
	newLab, err := labService.Create(ctx, lab)
	if err != nil {
		testutil.PrettyPrintError(err)
	}
	labID := newLab.ID

	assert.NoError(t, err)
	// Create test
	node := &models.Node{
		LabID:          labID,
		Label:          "ubuntu-0",
		NodeDefinition: "ubuntu",
		CPUs:           1,
		X:              100,
		Y:              200,
	}
	created, err := nodeService.Create(ctx, node)
	assert.NoError(t, err)

	// Get by ID test
	fetched, err := nodeService.GetByID(ctx, labID, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "ubuntu-0", fetched.Label)

	// Update test
	fetched.Label = "updated-node"
	updated, err := nodeService.Update(ctx, fetched)
	assert.NoError(t, err)
	assert.Equal(t, "updated-node", fetched.Label)

	// Delete test
	err = nodeService.Delete(ctx, labID, updated.ID)
	assert.NoError(t, err)

	// use named configs, create one more node
	nodeService.useNamedConfigs = true
	created, err = nodeService.Create(ctx, node)
	assert.NoError(t, err)
	node, err = nodeService.GetByID(ctx, labID, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "ubuntu-0", node.Label)
	assert.Equal(t, "ubuntu", node.NodeDefinition)
	assert.Equal(t, models.NodeStateDefined, node.State)
	assert.Equal(t, models.BootProgressNotRunning, node.BootProgress)

	// Verify named configurations are parsed
	// assert.NotNil(t, node.Configuration)
	// configs, ok := node.Configuration.([]models.NodeConfig)
	// assert.True(t, ok)
	assert.Len(t, node.Configurations, 2)
	assert.Equal(t, "user-data", node.Configurations[0].Name)
	assert.Equal(t, "network-config", node.Configurations[1].Name)
	assert.Contains(t, node.Configurations[0].Content, "#cloud-config")
	assert.Contains(t, node.Configurations[1].Content, "#network-config")

	err = nodeService.Delete(ctx, labID, created.ID)
	assert.NoError(t, err)

	err = labService.Delete(ctx, labID)
	assert.NoError(t, err)
}

func TestNodeGetByID_NotFound(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - UUID validation differs")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-123/nodes/nonexistent",
		httpmock.NewStringResponder(404, `{"message": "Node not found"}`))

	service := NewNodeService(client, false)
	ctx := context.Background()

	_, err := service.GetByID(ctx, "lab-123", "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "Node not found")
}

func TestNodeCreate_ValidationError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - UUID validation differs")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("POST", "https://mock/api/v0/labs/lab-123/nodes?populate_interfaces=true",
		httpmock.NewStringResponder(400, `{"description": "Input validation failed","code":400}`))

	service := NewNodeService(client, false)
	ctx := context.Background()

	invalidNode := &models.Node{
		LabID: "lab-123",
		// Missing required fields: Label, NodeDefinition
	}
	_, err := service.Create(ctx, invalidNode)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "Input validation failed")
}

func TestNodeUpdate_NotFound(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - UUID validation differs")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("PATCH", "https://mock/api/v0/labs/lab-123/nodes/nonexistent",
		httpmock.NewStringResponder(404, `{"message": "Node not found"}`))

	service := NewNodeService(client, false)
	ctx := context.Background()

	node := &models.Node{
		ID:    "nonexistent",
		LabID: "lab-123",
		Label: "test",
	}
	_, err := service.Update(ctx, node)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "Node not found")
}

func TestNodeDelete_NotFound(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - UUID validation differs")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("DELETE", "https://mock/api/v0/labs/lab-123/nodes/nonexistent",
		httpmock.NewStringResponder(404, `{"message": "Node not found"}`))

	service := NewNodeService(client, false)
	ctx := context.Background()

	err := service.Delete(ctx, "lab-123", "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "Node not found")
}

func TestNodeServerError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - can't force errors")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-123/nodes/node-123",
		httpmock.NewStringResponder(500, `{"message": "Internal server error"}`))

	service := NewNodeService(client, false)
	ctx := context.Background()

	_, err := service.GetByID(ctx, "lab-123", "node-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "Internal server error")
}

func TestNodeAuthError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - authentication succeeds")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-123/nodes/node-123",
		httpmock.NewStringResponder(401, `{"message": "Unauthorized"}`))

	service := NewNodeService(client, false)
	ctx := context.Background()

	_, err := service.GetByID(ctx, "lab-123", "node-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
	assert.Contains(t, err.Error(), "Unauthorized")
}

func TestNodePermissionError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - permissions are sufficient")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-123/nodes/node-123",
		httpmock.NewStringResponder(403, `{"message": "Forbidden"}`))

	service := NewNodeService(client, false)
	ctx := context.Background()

	_, err := service.GetByID(ctx, "lab-123", "node-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "403")
	assert.Contains(t, err.Error(), "Forbidden")
}

func TestNodeMalformedJSON(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - returns valid JSON")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-123/nodes/node-123",
		httpmock.NewStringResponder(200, `{"id": "node-123"`)) // Missing closing brace

	service := NewNodeService(client, false)
	ctx := context.Background()

	_, err := service.GetByID(ctx, "lab-123", "node-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode response")
}

func TestNodeStateOperations(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - UUID validation differs")
	}

	client, cleanup := initNodeTest(t)
	defer cleanup()

	if !testutil.IsLiveTesting() {
		httpmock.RegisterResponder("PUT", "https://mock/api/v0/labs/lab-123/nodes/node-123/state/start",
			httpmock.NewJsonResponderOrPanic(200, nil))
		httpmock.RegisterResponder("PUT", "https://mock/api/v0/labs/lab-123/nodes/node-123/state/stop",
			httpmock.NewJsonResponderOrPanic(200, nil))
		httpmock.RegisterResponder("PUT", "https://mock/api/v0/labs/lab-123/nodes/node-123/wipe_disks",
			httpmock.NewJsonResponderOrPanic(200, nil))
	}

	service := NewNodeService(client, false)
	ctx := context.Background()

	// Start test
	err := service.Start(ctx, "lab-123", "node-123")
	assert.NoError(t, err)

	// Stop test
	err = service.Stop(ctx, "lab-123", "node-123")
	assert.NoError(t, err)

	// Wipe test
	err = service.Wipe(ctx, "lab-123", "node-123")
	assert.NoError(t, err)
}
