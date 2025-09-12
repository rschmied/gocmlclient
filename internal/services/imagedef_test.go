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

func addImageDefinitionResponders() {
	// Mock responder for listing image definitions
	httpmock.RegisterResponder("GET", "https://mock/api/v0/image_definitions",
		httpmock.NewStringResponder(200, `[
			{
				"id": "img-def-456",
				"schema_version": "1.0",
				"node_definition_id": "iosv",
				"description": "IOSv image definition",
				"label": "IOSv 15.9",
				"disk_image": "iosv-159.qcow2",
				"disk_image_2": "iosv-159-disk2.qcow2",
				"disk_image_3": null,
				"read_only": false,
				"disk_subfolder": "iosv",
				"ram": 512,
				"cpus": 1,
				"cpu_limit": 100,
				"data_volume": 0,
				"boot_disk_size": 8
			},
			{
				"id": "img-def-123",
				"schema_version": "1.0",
				"node_definition_id": "csr1000v",
				"description": "CSR1000v image definition",
				"label": "CSR1000v 17.3",
				"disk_image": "csr1000v-173.qcow2",
				"disk_image_2": null,
				"disk_image_3": null,
				"read_only": true,
				"disk_subfolder": "csr1000v",
				"ram": 4096,
				"cpus": 2,
				"cpu_limit": 50,
				"data_volume": 2,
				"boot_disk_size": 16
			},
			{
				"id": "img-def-789",
				"schema_version": "1.0",
				"node_definition_id": "nxosv",
				"description": "NX-OSv image definition",
				"label": "NX-OSv 9.3",
				"disk_image": "nxosv-93.qcow2",
				"disk_image_2": null,
				"disk_image_3": null,
				"read_only": false,
				"disk_subfolder": "nxosv",
				"ram": 2048,
				"cpus": 1,
				"cpu_limit": null,
				"data_volume": null,
				"boot_disk_size": null
			}
		]`))
}

func addImageDefinitionErrorResponders() {
	// Mock responder for 500 error
	httpmock.RegisterResponder("GET", "https://mock/api/v0/image_definitions",
		httpmock.NewStringResponder(500, `{"error": "Internal server error"}`))
}

func initImageDefinitionTest(t *testing.T, responders func()) (*api.Client, func()) {
	client, cleanup := testutil.NewAPIClient(t)
	if !testutil.IsLiveTesting() {
		responders()
	}
	return client, cleanup
}

func TestImageDefinitionService_ImageDefinitions(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initImageDefinitionTest(t, addImageDefinitionResponders)
	defer cleanup()

	service := NewImageDefinitionService(client)
	ctx := context.Background()

	imageDefs, err := service.ImageDefinitions(ctx)

	assert.NoError(t, err)
	assert.Len(t, imageDefs, 3)

	// Test that results are sorted by ID in descending order (as per the implementation)
	// The mock data has IDs: img-def-456, img-def-123, img-def-789
	// After sorting by ID descending: img-def-789, img-def-456, img-def-123
	assert.Equal(t, models.UUID("img-def-789"), imageDefs[0].ID)
	assert.Equal(t, "nxosv", imageDefs[0].NodeDefID)
	assert.Equal(t, "NX-OSv 9.3", imageDefs[0].Label)
	assert.Equal(t, "nxosv-93.qcow2", imageDefs[0].DiskImage1)
	assert.Nil(t, imageDefs[0].DiskImage2)
	assert.Nil(t, imageDefs[0].DiskImage3)
	assert.False(t, imageDefs[0].ReadOnly)
	assert.Equal(t, "nxosv", imageDefs[0].DiskSubfolder)
	assert.Equal(t, 2048, *imageDefs[0].RAM)
	assert.Equal(t, 1, *imageDefs[0].CPUs)
	assert.Nil(t, imageDefs[0].CPUlimit)
	assert.Nil(t, imageDefs[0].DataVolume)
	assert.Nil(t, imageDefs[0].BootDiskSize)

	assert.Equal(t, models.UUID("img-def-456"), imageDefs[1].ID)
	assert.Equal(t, "iosv", imageDefs[1].NodeDefID)
	assert.Equal(t, "IOSv 15.9", imageDefs[1].Label)
	assert.Equal(t, "iosv-159.qcow2", imageDefs[1].DiskImage1)
	assert.Equal(t, "iosv-159-disk2.qcow2", *imageDefs[1].DiskImage2)
	assert.Nil(t, imageDefs[1].DiskImage3)
	assert.False(t, imageDefs[1].ReadOnly)
	assert.Equal(t, "iosv", imageDefs[1].DiskSubfolder)
	assert.Equal(t, 512, *imageDefs[1].RAM)
	assert.Equal(t, 1, *imageDefs[1].CPUs)
	assert.Equal(t, 100, *imageDefs[1].CPUlimit)
	assert.Equal(t, 0, *imageDefs[1].DataVolume)
	assert.Equal(t, 8, *imageDefs[1].BootDiskSize)

	assert.Equal(t, models.UUID("img-def-123"), imageDefs[2].ID)
	assert.Equal(t, "csr1000v", imageDefs[2].NodeDefID)
	assert.Equal(t, "CSR1000v 17.3", imageDefs[2].Label)
	assert.Equal(t, "csr1000v-173.qcow2", imageDefs[2].DiskImage1)
	assert.Nil(t, imageDefs[2].DiskImage2)
	assert.Nil(t, imageDefs[2].DiskImage3)
	assert.True(t, imageDefs[2].ReadOnly)
	assert.Equal(t, "csr1000v", imageDefs[2].DiskSubfolder)
	assert.Equal(t, 4096, *imageDefs[2].RAM)
	assert.Equal(t, 2, *imageDefs[2].CPUs)
	assert.Equal(t, 50, *imageDefs[2].CPUlimit)
	assert.Equal(t, 2, *imageDefs[2].DataVolume)
	assert.Equal(t, 16, *imageDefs[2].BootDiskSize)
}

func TestImageDefinitionService_ImageDefinitions_Error(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initImageDefinitionTest(t, addImageDefinitionErrorResponders)
	defer cleanup()

	service := NewImageDefinitionService(client)
	ctx := context.Background()

	_, err := service.ImageDefinitions(ctx)
	assert.Error(t, err)
}

func TestImageDefinitionService_ImageDefinitions_EmptyList(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Register empty list responder
	httpmock.RegisterResponder("GET", "https://mock/api/v0/image_definitions",
		httpmock.NewStringResponder(200, `[]`))

	service := NewImageDefinitionService(client)
	ctx := context.Background()

	imageDefs, err := service.ImageDefinitions(ctx)

	assert.NoError(t, err)
	assert.Len(t, imageDefs, 0)
}

func TestImageDefinitionService_ImageDefinitions_SingleItem(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	// Register single item responder
	httpmock.RegisterResponder("GET", "https://mock/api/v0/image_definitions",
		httpmock.NewStringResponder(200, `[
			{
				"id": "single-img",
				"schema_version": "1.0",
				"node_definition_id": "test",
				"description": "Single image definition",
				"label": "Test Image",
				"disk_image": "test.qcow2",
				"read_only": false,
				"disk_subfolder": "test"
			}
		]`))

	service := NewImageDefinitionService(client)
	ctx := context.Background()

	imageDefs, err := service.ImageDefinitions(ctx)

	assert.NoError(t, err)
	assert.Len(t, imageDefs, 1)
	assert.Equal(t, models.UUID("single-img"), imageDefs[0].ID)
	assert.Equal(t, "test", imageDefs[0].NodeDefID)
	assert.Equal(t, "Test Image", imageDefs[0].Label)
	assert.Equal(t, "test.qcow2", imageDefs[0].DiskImage1)
	assert.False(t, imageDefs[0].ReadOnly)
	assert.Equal(t, "test", imageDefs[0].DiskSubfolder)
}

func TestImageDefinitionService_NewImageDefinitionService(t *testing.T) {
	client, cleanup := testutil.NewAPIClient(t)
	defer cleanup()

	service := NewImageDefinitionService(client)
	assert.NotNil(t, service)
	assert.Equal(t, client, service.apiClient)
}
