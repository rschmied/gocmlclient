package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeDefinition_JSON(t *testing.T) {
	tests := []struct {
		name    string
		nodeDef NodeDefinition
		jsonStr string
	}{
		{
			name: "basic node definition",
			nodeDef: NodeDefinition{
				ID: "ubuntu",
				General: NodeDefinitionGeneral{
					Nature:      DeviceNatureServer,
					Description: "Ubuntu server",
					ReadOnly:    false,
				},
				Device: deviceData{
					Interfaces: interfaceData{
						SerialPorts:     0,
						Physical:        []string{"eth0", "eth1"},
						HasLoopbackZero: true,
						DefaultCount:    2,
					},
				},
				UI: NodeDefinitionUI{
					LabelPrefix: "ubuntu",
					Icon:        NodeDefIconsServer,
					Label:       "Ubuntu",
					Visible:     true,
					Group:       "Others",
					Description: "Ubuntu server",
				},
				Sim: simData{
					RAM:     1024,
					CPUs:    1,
					Console: true,
					VNC:     false,
				},
				ImageDefinitions: []string{"ubuntu-20.04"},
			},
			jsonStr: `{
				"id": "ubuntu",
				"general": {
					"nature": "server",
					"description": "Ubuntu server",
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
					"has_configuration": false,
					"show_ram": false,
					"show_cpus": false,
					"show_cpu_limit": false,
					"show_data_volume": false,
					"show_boot_disk_size": false,
					"has_config_extraction": false
				},
				"sim": {
					"ram": 1024,
					"cpus": 1,
					"console": true,
					"vnc": false
				},
				"image_definitions": ["ubuntu-20.04"]
			}`,
		},
		{
			name: "node definition with complex sim",
			nodeDef: NodeDefinition{
				ID: "iosv",
				General: NodeDefinitionGeneral{
					Nature:      DeviceNatureRouter,
					Description: "IOSv router",
					ReadOnly:    true,
				},
				Device: deviceData{
					Interfaces: interfaceData{
						SerialPorts:     2,
						Physical:        []string{"GigabitEthernet0/0"},
						HasLoopbackZero: true,
						DefaultCount:    1,
					},
				},
				UI: NodeDefinitionUI{
					LabelPrefix: "R",
					Icon:        NodeDefIconsRouter,
					Label:       "IOSv",
					Visible:     true,
					Group:       "Cisco",
					Description: "Cisco IOSv router",
				},
				Sim: simData{
					LinuxNative: LinuxNativeSimulation{
						Driver:    "qemu",
						RAM:       512,
						CPUs:      1,
						EnableRNG: true,
					},
					Parameters: NodeParameters{
						"key1": "value1",
					},
					UsageEstimations: UsageEstimations{
						CPUs: 1,
						RAM:  512,
						Disk: 100,
					},
					RAM:     512,
					CPUs:    1,
					Console: true,
					VNC:     false,
				},
			},
			jsonStr: `{
				"id": "iosv",
				"general": {
					"nature": "router",
					"description": "IOSv router",
					"read_only": true
				},
				"device": {
					"interfaces": {
						"serial_ports": 2,
						"physical": ["GigabitEthernet0/0"],
						"has_loopback_zero": true,
						"default_count": 1
					}
				},
				"ui": {
					"label_prefix": "R",
					"icon": "router",
					"label": "IOSv",
					"visible": true,
					"group": "Cisco",
					"description": "Cisco IOSv router"
				},
				"sim": {
					"linux_native": {
						"driver": "qemu",
						"ram": 512,
						"cpus": 1,
						"enable_rng": true
					},
					"parameters": {
						"key1": "value1"
					},
					"usage_estimations": {
						"cpus": 1,
						"ram": 512,
						"disk": 100
					},
					"ram": 512,
					"cpus": 1,
					"console": true,
					"vnc": false
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test unmarshaling
			var unmarshaled NodeDefinition
			err := json.Unmarshal([]byte(tt.jsonStr), &unmarshaled)
			assert.NoError(t, err)

			// Compare key fields
			assert.Equal(t, tt.nodeDef.ID, unmarshaled.ID)
			assert.Equal(t, tt.nodeDef.General, unmarshaled.General)
			assert.Equal(t, tt.nodeDef.Device, unmarshaled.Device)
			assert.Equal(t, tt.nodeDef.UI, unmarshaled.UI)
			assert.Equal(t, tt.nodeDef.Sim, unmarshaled.Sim)
			assert.Equal(t, tt.nodeDef.ImageDefinitions, unmarshaled.ImageDefinitions)
		})
	}
}

func TestNodeDefinition_Methods(t *testing.T) {
	tests := []struct {
		name        string
		nodeDef     NodeDefinition
		hasVNC      bool
		hasSerial   bool
		serialPorts int
	}{
		{
			name: "node with VNC and serial",
			nodeDef: NodeDefinition{
				Sim: simData{
					VNC: true,
				},
				Device: deviceData{
					Interfaces: interfaceData{
						SerialPorts: 2,
					},
				},
			},
			hasVNC:      true,
			hasSerial:   true,
			serialPorts: 2,
		},
		{
			name: "node without VNC or serial",
			nodeDef: NodeDefinition{
				Sim: simData{
					VNC: false,
				},
				Device: deviceData{
					Interfaces: interfaceData{
						SerialPorts: 0,
					},
				},
			},
			hasVNC:      false,
			hasSerial:   false,
			serialPorts: 0,
		},
		{
			name: "node with VNC but no serial",
			nodeDef: NodeDefinition{
				Sim: simData{
					VNC: true,
				},
				Device: deviceData{
					Interfaces: interfaceData{
						SerialPorts: 0,
					},
				},
			},
			hasVNC:      true,
			hasSerial:   false,
			serialPorts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.hasVNC, tt.nodeDef.HasVNC())
			assert.Equal(t, tt.hasSerial, tt.nodeDef.HasSerial())
			assert.Equal(t, tt.serialPorts, tt.nodeDef.SerialPorts())
		})
	}
}

func TestEnums(t *testing.T) {
	// Test DeviceNature
	assert.Equal(t, DeviceNature("server"), DeviceNatureServer)
	assert.Equal(t, DeviceNature("router"), DeviceNatureRouter)

	// Test NodeDefIcons
	assert.Equal(t, NodeDefIcons("server"), NodeDefIconsServer)
	assert.Equal(t, NodeDefIcons("router"), NodeDefIconsRouter)
}

func TestSupportingStructs_JSON(t *testing.T) {
	// Test VMProperties
	vmProps := VMProperties{
		RAM:          true,
		CPUs:         false,
		DataVolume:   true,
		BootDiskSize: false,
		CPULimit:     true,
	}
	data1, err := json.Marshal(vmProps)
	assert.NoError(t, err)
	assert.Contains(t, string(data1), `"ram":true`)
	assert.Contains(t, string(data1), `"cpus":false`)

	// Test UsageEstimations
	usage := UsageEstimations{
		CPUs: 2,
		RAM:  1024,
		Disk: 500,
	}
	data2, err := json.Marshal(usage)
	assert.NoError(t, err)
	assert.Contains(t, string(data2), `"cpus":2`)
	assert.Contains(t, string(data2), `"ram":1024`)
	assert.Contains(t, string(data2), `"disk":500`)

	// Test NodeParameters
	params := NodeParameters{
		"key1": "value1",
		"key2": 42,
	}
	data3, err := json.Marshal(params)
	assert.NoError(t, err)
	assert.Contains(t, string(data3), `"key1":"value1"`)
	assert.Contains(t, string(data3), `"key2":42`)
}
