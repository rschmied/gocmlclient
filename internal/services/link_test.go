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

func initLinkTest(t *testing.T) (*api.Client, func()) {
	client, cleanup := testutil.NewAPIClient(t)
	if !testutil.IsLiveTesting() {
		// Mock responses will be registered in individual tests
		addLabResponders()
		addNodeResponders()
		addInterfaceResponders()
	}
	return client, cleanup
}

func addNodeResponders() {
	// Mock node creation
	httpmock.RegisterResponder("POST", "https://mock/api/v0/labs/lab_uuid/nodes",
		httpmock.NewStringResponder(200, `{"id": "node-a-uuid"}`))
	httpmock.RegisterResponder("POST", "https://mock/api/v0/labs/lab_uuid/nodes",
		httpmock.NewStringResponder(200, `{"id": "node-b-uuid"}`))
	// Mock node get
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid/nodes/node-a-uuid",
		httpmock.NewStringResponder(200, `{
			"id": "node-a-uuid",
			"lab_id": "lab_uuid",
			"label": "node-a",
			"node_definition": "ubuntu",
			"x": 100,
			"y": 100
		}`))
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid/nodes/node-b-uuid",
		httpmock.NewStringResponder(200, `{
			"id": "node-b-uuid",
			"lab_id": "lab_uuid",
			"label": "node-b",
			"node_definition": "ubuntu",
			"x": 200,
			"y": 200
		}`))
	// Mock node patch
	httpmock.RegisterResponder("PATCH", "https://mock/api/v0/labs/lab_uuid/nodes/node-a-uuid",
		httpmock.NewStringResponder(200, `{"id": "node-a-uuid"}`))
	httpmock.RegisterResponder("PATCH", "https://mock/api/v0/labs/lab_uuid/nodes/node-b-uuid",
		httpmock.NewStringResponder(200, `{"id": "node-b-uuid"}`))
	// Mock node delete
	httpmock.RegisterResponder("DELETE", "https://mock/api/v0/labs/lab_uuid/nodes/node-a-uuid",
		httpmock.NewJsonResponderOrPanic(204, nil))
	httpmock.RegisterResponder("DELETE", "https://mock/api/v0/labs/lab_uuid/nodes/node-b-uuid",
		httpmock.NewJsonResponderOrPanic(204, nil))
}

func addInterfaceResponders() {
	// Mock get interfaces for node
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid/nodes/node-a/interfaces?data=true",
		httpmock.NewJsonResponderOrPanic(200, []map[string]any{
			{
				"id":           "iface-a-uuid",
				"node":         "node-a-uuid",
				"slot":         0,
				"is_connected": false,
				"type":         "physical",
			},
		}))
	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid/nodes/node-b/interfaces?data=true",
		httpmock.NewJsonResponderOrPanic(200, []map[string]any{
			{
				"id":           "iface-b-uuid",
				"node":         "node-b-uuid",
				"slot":         0,
				"is_connected": false,
				"type":         "physical",
			},
		}))
	// Mock create interface
	httpmock.RegisterResponder("POST", "https://mock/api/v0/labs/lab_uuid/interfaces",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"id": "new-iface-uuid",
		}))
}

func TestLinkCRUD(t *testing.T) {
	client, cleanup := initLinkTest(t)
	defer cleanup()

	if !testutil.IsLiveTesting() {
		// Mock link responses
		createResponse := `{"id": "link-uuid"}`
		linkResponse := `{
			"id": "link-uuid",
			"lab_id": "lab_uuid",
			"interface_a": "iface-a-uuid",
			"interface_b": "iface-b-uuid",
			"node_a": "node-a-uuid",
			"node_b": "node-b-uuid",
			"state": "STARTED"
		}`
		linksResponse := `[` + linkResponse + `]`

		httpmock.RegisterResponder("POST", "https://mock/api/v0/labs/lab_uuid/links",
			httpmock.NewStringResponder(200, createResponse))
		httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid/links?data=true",
			httpmock.NewStringResponder(200, linksResponse))
		httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid/links/link-uuid",
			httpmock.NewStringResponder(200, linkResponse))
		httpmock.RegisterResponder("DELETE", "https://mock/api/v0/labs/lab_uuid/links/link-uuid",
			httpmock.NewJsonResponderOrPanic(204, nil))
	}

	ctx := context.Background()
	labService := NewLabService(client, nil, nil, nil, nil)
	nodeService := NewNodeService(client, false)
	interfaceService := NewInterfaceService(client)
	linkService := NewLinkService(client)
	linkService.Interface = interfaceService
	linkService.Node = nodeService

	// Create lab
	lab := models.LabCreateRequest{Title: "test lab"}
	newLab, err := labService.Create(ctx, lab)
	assert.NoError(t, err)
	labID := newLab.ID

	// Create nodes
	nodeA := &models.Node{
		LabID:          labID,
		Label:          "node-a",
		NodeDefinition: "ubuntu",
		X:              100,
		Y:              100,
	}
	createdNodeA, err := nodeService.Create(ctx, nodeA)
	assert.NoError(t, err)

	nodeB := &models.Node{
		LabID:          labID,
		Label:          "node-b",
		NodeDefinition: "ubuntu",
		X:              200,
		Y:              200,
	}
	createdNodeB, err := nodeService.Create(ctx, nodeB)
	assert.NoError(t, err)

	// Create link using node names and slots
	link := &models.Link{
		LabID:   labID,
		SrcNode: "node-a",
		DstNode: "node-b",
		SrcSlot: 0,
		DstSlot: 0,
	}
	createdLink, err := linkService.Create(ctx, link)
	assert.NoError(t, err)
	assert.NotEmpty(t, createdLink.ID)

	// Get links for lab
	links, err := linkService.GetLinksForLab(ctx, newLab)
	assert.NoError(t, err)
	assert.Len(t, links, 1)
	assert.Equal(t, createdLink.ID, links[0].ID)

	// Get link by ID
	fetchedLink, err := linkService.GetByID(ctx, labID, createdLink.ID)
	assert.NoError(t, err)
	assert.Equal(t, createdLink.ID, fetchedLink.ID)
	assert.Equal(t, labID, fetchedLink.LabID)

	// Delete link
	err = linkService.Delete(ctx, *createdLink)
	assert.NoError(t, err)

	// Cleanup
	err = nodeService.Delete(ctx, createdNodeA)
	assert.NoError(t, err)
	err = nodeService.Delete(ctx, createdNodeB)
	assert.NoError(t, err)
	err = labService.Delete(ctx, labID)
	assert.NoError(t, err)
}

func TestLinkCondition(t *testing.T) {
	client, cleanup := initLinkTest(t)
	defer cleanup()

	if !testutil.IsLiveTesting() {
		conditionResponse := `{
			"bandwidth": 1000,
			"latency": 10,
			"loss": 0.1,
			"enabled": true
		}`

		httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-uuid/links/link-uuid/condition",
			httpmock.NewStringResponder(200, conditionResponse))
		httpmock.RegisterResponder("PATCH", "https://mock/api/v0/labs/lab-uuid/links/link-uuid/condition",
			httpmock.NewStringResponder(200, conditionResponse))
		httpmock.RegisterResponder("DELETE", "https://mock/api/v0/labs/lab-uuid/links/link-uuid/condition",
			httpmock.NewJsonResponderOrPanic(204, nil))
	}

	linkService := NewLinkService(client)
	ctx := context.Background()

	// Get condition
	condition, err := linkService.GetCondition(ctx, "lab-uuid", "link-uuid")
	assert.NoError(t, err)
	assert.Equal(t, 1000, condition.Bandwidth)
	assert.Equal(t, 10, condition.Latency)
	assert.Equal(t, 0.1, condition.Loss)

	// Set condition
	config := &models.LinkConditionConfiguration{
		Bandwidth: 1000,
		Latency:   10,
		Loss:      0.1,
		Enabled:   true,
	}
	setCondition, err := linkService.SetCondition(ctx, "lab-uuid", "link-uuid", config)
	assert.NoError(t, err)
	assert.Equal(t, 1000, setCondition.Bandwidth)

	// Delete condition
	err = linkService.DeleteCondition(ctx, "lab-uuid", "link-uuid")
	assert.NoError(t, err)
}

func TestLinkGetByID_NotFound(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - UUID validation differs")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-123/links/nonexistent",
		httpmock.NewStringResponder(404, `{"message": "Link not found"}`))

	service := NewLinkService(client)
	ctx := context.Background()

	_, err := service.GetByID(ctx, "lab-123", "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "Link not found")
}

func TestLinkCreate_ValidationError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - UUID validation differs")
	}

	client, cleanup := initLinkTest(t)
	defer cleanup()

	httpmock.RegisterResponder("POST", "https://mock/api/v0/labs/lab-123/links",
		httpmock.NewStringResponder(400, `{"description": "Input validation failed","code":400}`))

	service := NewLinkService(client)
	ctx := context.Background()

	invalidLink := &models.Link{
		LabID: "lab-123",
		// Missing required fields
	}
	_, err := service.Create(ctx, invalidLink)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "Input validation failed")
}

func TestLinkDelete_NotFound(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - UUID validation differs")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("DELETE", "https://mock/api/v0/labs/lab-123/links/nonexistent",
		httpmock.NewStringResponder(404, `{"message": "Link not found"}`))

	service := NewLinkService(client)
	ctx := context.Background()

	link := &models.Link{
		ID:    "nonexistent",
		LabID: "lab-123",
	}
	err := service.Delete(ctx, *link)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "Link not found")
}

func TestLinkServerError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - can't force errors")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-123/links/link-123",
		httpmock.NewStringResponder(500, `{"message": "Internal server error"}`))

	service := NewLinkService(client)
	ctx := context.Background()

	_, err := service.GetByID(ctx, "lab-123", "link-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "Internal server error")
}

func TestLinkAuthError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - authentication succeeds")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-123/links/link-123",
		httpmock.NewStringResponder(401, `{"message": "Unauthorized"}`))

	service := NewLinkService(client)
	ctx := context.Background()

	_, err := service.GetByID(ctx, "lab-123", "link-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
	assert.Contains(t, err.Error(), "Unauthorized")
}

func TestLinkPermissionError(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - permissions are sufficient")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-123/links/link-123",
		httpmock.NewStringResponder(403, `{"message": "Forbidden"}`))

	service := NewLinkService(client)
	ctx := context.Background()

	_, err := service.GetByID(ctx, "lab-123", "link-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "403")
	assert.Contains(t, err.Error(), "Forbidden")
}

func TestLinkMalformedJSON(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - returns valid JSON")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab-123/links/link-123",
		httpmock.NewStringResponder(200, `{"id": "link-123"`)) // Missing closing brace

	service := NewLinkService(client)
	ctx := context.Background()

	_, err := service.GetByID(ctx, "lab-123", "link-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode response")
}

func TestLinkListMarshalJSON(t *testing.T) {
	links := linkList{
		{
			ID:    "b-uuid",
			LabID: "lab-uuid",
		},
		{
			ID:    "a-uuid",
			LabID: "lab-uuid",
		},
	}

	data, err := links.MarshalJSON()
	assert.NoError(t, err)

	// Should be sorted by ID
	expected := `[
		{
	   "id":"a-uuid", "interface_a":"", "interface_b":"", "lab_id":"lab-uuid",
	   "label":"", "link_capture_key":"", "node_a":"", "node_b":"", "slot_a":0,
	   "slot_b":0, "state":""
	  },
		{
	   "id":"b-uuid", "interface_a":"", "interface_b":"", "lab_id":"lab-uuid",
	   "label":"", "link_capture_key":"", "node_a":"", "node_b":"", "slot_a":0,
	   "slot_b":0, "state":""
	  }
	]`
	assert.JSONEq(t, expected, string(data))
}
