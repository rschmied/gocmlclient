package cmlclient

import (
	"testing"

	mr "github.com/rschmied/mockresponder"
	"github.com/stretchr/testify/assert"
)

func TestClient_GetNodeDefs(t *testing.T) {

	tc := newAuthedTestAPIclient()

	nodeDefs := []byte(`[
		{
			"id": "alpine",
			"configuration": {
				"generator": {
					"driver": "alpine"
				},
				"provisioning": {
					"volume_name": "cfg",
					"media_type": "iso",
					"files": [
						{
							"name": "node.cfg",
							"content": "bla",
							"editable": true
						}
					]
				}
			},
			"device": {
				"interfaces": {
					"has_loopback_zero": false,
					"default_count": 1,
					"physical": [
						"eth0",
						"eth1",
						"eth2",
						"eth3"
					],
					"serial_ports": 1
				}
			},
			"inherited": {
				"image": {
					"ram": true,
					"cpus": true,
					"cpu_limit": true,
					"data_volume": true,
					"boot_disk_size": true
				},
				"node": {
					"ram": true,
					"cpus": true,
					"cpu_limit": true,
					"data_volume": true,
					"boot_disk_size": true
				}
			},
			"general": {
				"description": "Alpine Linux",
				"nature": "server",
				"read_only": true
			},
			"schema_version": "0.0.1",
			"sim": {
				"linux_native": {
					"cpus": 1,
					"disk_driver": "virtio",
					"driver": "server",
					"libvirt_domain_driver": "kvm",
					"nic_driver": "virtio",
					"ram": 512,
					"boot_disk_size": 16,
					"video": {
						"memory": 16
					},
					"cpu_limit": 100
				}
			},
			"boot": {
				"completed": [
					"Welcome to Alpine Linux"
				],
				"timeout": 60
			},
			"pyats": {
				"os": "linux"
			},
			"ui": {
				"description": "Alpine Linux\n\n512 MB DRAM, 1 vCPU, 16GB base",
				"group": "Others",
				"icon": "server",
				"label": "Alpine",
				"label_prefix": "alpine-",
				"visible": true
			}
		}
	]`)

	tests := []struct {
		name    string
		data    mr.MockRespList
		wantErr bool
	}{
		{"good", mr.MockRespList{mr.MockResp{Data: nodeDefs}}, false},
		{"badjson", mr.MockRespList{mr.MockResp{Data: []byte(`///`)}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc.mr.SetData(tt.data)
			got, err := tc.client.GetNodeDefs(tc.ctx)
			if err != nil {
				if !tt.wantErr {
					t.Error("unexpected error!")
				}
				return
			}
			assert.NoError(t, err)
			assert.Len(t, got, 1)
			got["alpine"].hasSerial()
			got["alpine"].hasVNC()
			got["alpine"].serialPorts()
		})
	}
}
