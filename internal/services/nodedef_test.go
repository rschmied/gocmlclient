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

func initNodeDefTest(t *testing.T) (*api.Client, func()) {
	client, cleanup := testutil.NewAPIClient(t)
	if !testutil.IsLiveTesting() {
		// Mock responses will be registered in individual tests
	}
	return client, cleanup
}

func TestNodeDefinitionService_NodeDefinitions(t *testing.T) {
	if testutil.IsLiveTesting() {
		t.Skip("Skipping on live server - test expects specific mock data")
	}

	client, cleanup := initNodeDefTest(t)
	defer cleanup()

	if !testutil.IsLiveTesting() {
		// Mock response for node definitions
		nodeDefsResponse := `[
			{
				"id": "ubuntu",
				"general": {
					"nature": "server",
					"description": "Ubuntu server node",
					"read_only": false
				},
				"device": {
					"interfaces": {
						"serial_ports": 0,
						"physical": ["eth0", "eth1"],
						"has_loopback_zero": true,
						"default_count": 2
					}
				},
				"ui": {
					"label_prefix": "ubuntu",
					"icon": "server",
					"label": "Ubuntu",
					"visible": true,
					"group": "Others",
					"description": "Ubuntu server",
					"has_configuration": true,
					"show_ram": true,
					"show_cpus": true,
					"show_cpu_limit": true,
					"show_data_volume": true,
					"show_boot_disk_size": true,
					"has_config_extraction": false
				},
				"sim": {
					"parameters": {},
					"ram": 1024,
					"cpus": 1,
					"cpu_limit": 100,
					"data_volume": 0,
					"boot_disk_size": 0,
					"console": true,
					"simulate": false,
					"custom_mac": false,
					"vnc": false
				},
				"image_definitions": ["ubuntu-20.04"]
			},
			{
				"id": "iosv",
				"general": {
					"nature": "router",
					"description": "IOSv router",
					"read_only": true
				},
				"device": {
					"interfaces": {
						"serial_ports": 2,
						"physical": ["GigabitEthernet0/0", "GigabitEthernet0/1"],
						"has_loopback_zero": true,
						"default_count": 2
					}
				},
				"ui": {
					"label_prefix": "R",
					"icon": "router",
					"label": "IOSv",
					"visible": true,
					"group": "Cisco",
					"description": "Cisco IOSv router",
					"has_configuration": true,
					"show_ram": true,
					"show_cpus": true,
					"show_cpu_limit": false,
					"show_data_volume": false,
					"show_boot_disk_size": false,
					"has_config_extraction": true
				},
				"sim": {
					"parameters": {},
					"ram": 512,
					"cpus": 1,
					"cpu_limit": 100,
					"data_volume": 0,
					"boot_disk_size": 0,
					"console": true,
					"simulate": false,
					"custom_mac": false,
					"vnc": false
				},
				"image_definitions": ["iosv-15.9"]
			}
		]`

		httpmock.RegisterResponder("GET", "https://mock/api/v0/simplified_node_definitions",
			httpmock.NewStringResponder(200, nodeDefsResponse))
	}

	ctx := context.Background()
	nodeDefService := NewNodeDefinitionService(client)

	nodeDefs, err := nodeDefService.NodeDefinitions(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, nodeDefs)
	assert.Len(t, nodeDefs, 2)

	// Check first node definition
	ubuntuDef, exists := nodeDefs["ubuntu"]
	assert.True(t, exists)
	assert.Equal(t, models.UUID("ubuntu"), ubuntuDef.ID)
	assert.Equal(t, models.DeviceNatureServer, ubuntuDef.General.Nature)
	assert.Equal(t, "Ubuntu server node", ubuntuDef.General.Description)
	assert.False(t, ubuntuDef.General.ReadOnly)
	assert.Equal(t, 0, ubuntuDef.Device.Interfaces.SerialPorts)
	assert.Len(t, ubuntuDef.Device.Interfaces.Physical, 2)
	assert.True(t, ubuntuDef.Device.Interfaces.HasLoopbackZero)
	assert.Equal(t, 2, ubuntuDef.Device.Interfaces.DefaultCount)
	assert.Equal(t, "ubuntu", ubuntuDef.UI.LabelPrefix)
	assert.Equal(t, models.NodeDefIconsServer, ubuntuDef.UI.Icon)
	assert.Equal(t, "Ubuntu", ubuntuDef.UI.Label)
	assert.True(t, ubuntuDef.UI.Visible)
	assert.Equal(t, "Others", ubuntuDef.UI.Group)
	assert.Equal(t, "Ubuntu server", ubuntuDef.UI.Description)
	assert.True(t, ubuntuDef.UI.HasConfiguration)
	assert.True(t, ubuntuDef.UI.ShowRAM)
	assert.True(t, ubuntuDef.UI.ShowCPUs)
	assert.True(t, ubuntuDef.UI.ShowCPULimit)
	assert.True(t, ubuntuDef.UI.ShowDataVolume)
	assert.True(t, ubuntuDef.UI.ShowBootDiskSize)
	assert.False(t, ubuntuDef.UI.HasConfigExtraction)
	assert.Equal(t, 1024, ubuntuDef.Sim.RAM)
	assert.Equal(t, 1, ubuntuDef.Sim.CPUs)
	assert.Equal(t, 100, ubuntuDef.Sim.CPULimit)
	assert.Equal(t, 0, ubuntuDef.Sim.DataVolume)
	assert.Equal(t, 0, ubuntuDef.Sim.BootDiskSize)
	assert.True(t, ubuntuDef.Sim.Console)
	assert.False(t, ubuntuDef.Sim.Simulate)
	assert.False(t, ubuntuDef.Sim.CustomMAC)
	assert.False(t, ubuntuDef.Sim.VNC)
	assert.Len(t, ubuntuDef.Sim.Parameters, 0)
	// UsageEstimations is a struct, not a slice, so check if zero
	assert.Equal(t, models.UsageEstimations{}, ubuntuDef.Sim.UsageEstimations)
	assert.Len(t, ubuntuDef.ImageDefinitions, 1)
	assert.Equal(t, "ubuntu-20.04", ubuntuDef.ImageDefinitions[0])

	// Check second node definition
	iosvDef, exists := nodeDefs["iosv"]
	assert.True(t, exists)
	assert.Equal(t, models.UUID("iosv"), iosvDef.ID)
	assert.Equal(t, models.DeviceNatureRouter, iosvDef.General.Nature)
	assert.Equal(t, "IOSv router", iosvDef.General.Description)
	assert.True(t, iosvDef.General.ReadOnly)
	assert.Equal(t, 2, iosvDef.Device.Interfaces.SerialPorts)
	assert.Len(t, iosvDef.Device.Interfaces.Physical, 2)
	assert.True(t, iosvDef.Device.Interfaces.HasLoopbackZero)
	assert.Equal(t, 2, iosvDef.Device.Interfaces.DefaultCount)
	assert.Equal(t, "R", iosvDef.UI.LabelPrefix)
	assert.Equal(t, models.NodeDefIconsRouter, iosvDef.UI.Icon)
	assert.Equal(t, "IOSv", iosvDef.UI.Label)
	assert.True(t, iosvDef.UI.Visible)
	assert.Equal(t, "Cisco", iosvDef.UI.Group)
	assert.Equal(t, "Cisco IOSv router", iosvDef.UI.Description)
	assert.True(t, iosvDef.UI.HasConfiguration)
	assert.True(t, iosvDef.UI.ShowRAM)
	assert.True(t, iosvDef.UI.ShowCPUs)
	assert.False(t, iosvDef.UI.ShowCPULimit)
	assert.False(t, iosvDef.UI.ShowDataVolume)
	assert.False(t, iosvDef.UI.ShowBootDiskSize)
	assert.True(t, iosvDef.UI.HasConfigExtraction)
	assert.Equal(t, 512, iosvDef.Sim.RAM)
	assert.Equal(t, 1, iosvDef.Sim.CPUs)
	assert.Equal(t, 100, iosvDef.Sim.CPULimit)
	assert.Equal(t, 0, iosvDef.Sim.DataVolume)
	assert.Equal(t, 0, iosvDef.Sim.BootDiskSize)
	assert.True(t, iosvDef.Sim.Console)
	assert.False(t, iosvDef.Sim.Simulate)
	assert.False(t, iosvDef.Sim.CustomMAC)
	assert.False(t, iosvDef.Sim.VNC)
	assert.Len(t, iosvDef.Sim.Parameters, 0)
	// UsageEstimations is a struct, not a slice, so check if zero
	assert.Equal(t, models.UsageEstimations{}, iosvDef.Sim.UsageEstimations)
	assert.Len(t, iosvDef.ImageDefinitions, 1)
	assert.Equal(t, "iosv-15.9", iosvDef.ImageDefinitions[0])

	// Test HasVNC and HasSerial methods
	assert.False(t, ubuntuDef.HasVNC())
	assert.True(t, iosvDef.HasSerial())
	assert.Equal(t, 2, iosvDef.SerialPorts())
}
