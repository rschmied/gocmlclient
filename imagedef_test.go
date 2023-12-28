package cmlclient

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	mr "github.com/rschmied/mockresponder"
)

// [
//   {
//     "id": "alpine-3-13-2-base",
//     "node_definition_id": "alpine",
//     "description": "Alpine Linux and network tools",
//     "label": "Alpine 3.13.2",
//     "disk_image": "alpine-3-13-2-base.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "alpine-3-13-2-base",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "asav-9-15-1",
//     "node_definition_id": "asav",
//     "description": "ASAv 9.15.1",
//     "label": "ASAv 9.15.1",
//     "disk_image": "asav9-15-1.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "asav-9-15-1",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "coreos-2135-4-0",
//     "node_definition_id": "coreos",
//     "description": "CoreOS 2135.4.0",
//     "label": "CoreOS 2135.4.0",
//     "disk_image": "coreos_production_qemu_image.img",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "coreos-2135-4-0",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "csr1000v-170302",
//     "node_definition_id": "csr1000v",
//     "description": "CSR1000v 17.03.02",
//     "label": "CSR1000v 17.03.02",
//     "disk_image": "csr1000v-universalk9.17.03.02-serial.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "csr1000v-170302",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "desktop-3-13-2-xfce",
//     "node_definition_id": "desktop",
//     "description": "Alpine Desktop 3.13.2 XFCE",
//     "label": "Desktop 3.13.2",
//     "disk_image": "desktop-3-13-2-xfce.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "desktop-3-13-2-xfce",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "iosv-159-3-m3",
//     "node_definition_id": "iosv",
//     "description": "IOSv 15.9(3) M3",
//     "label": "IOSv 15.9(3) M3",
//     "disk_image": "vios-adventerprisek9-m.spa.159-3.m3.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "iosv-159-3-m3",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "iosvl2-2020",
//     "node_definition_id": "iosvl2",
//     "description": "IOSv L2 2020",
//     "label": "IOSv L2 2020",
//     "disk_image": "vios_l2-adventerprisek9-m.ssa.high_iron_20200929.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "iosvl2-2020",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "iosxrv-6-3-1",
//     "node_definition_id": "iosxrv",
//     "description": "IOS XRv 6.3.1",
//     "label": "IOS XRv 6.3.1",
//     "disk_image": "iosxrv-k9-demo-6.3.1.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "iosxrv-6-3-1",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "iosxrv9000-7-2-2",
//     "node_definition_id": "iosxrv9000",
//     "description": "IOS XR 9000v 7.2.2",
//     "label": "IOS XR 9000v 7.2.2",
//     "disk_image": "xrv9k-fullk9-x-7.2.2.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "iosxrv9000-7-2-2",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "nxosv-7-3-0",
//     "node_definition_id": "nxosv",
//     "description": "NX-OS 7.3.0",
//     "label": "NX-OS 7.3.0",
//     "disk_image": "titanium-final.7.3.0.d1.1.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "nxosv-7-3-0",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "nxosv9000-9-2-4",
//     "node_definition_id": "nxosv9000",
//     "description": "NX-OS 9000 9.2.4",
//     "label": "NX-OS 9000 9.2.4",
//     "disk_image": "nxosv.9.2.4.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "nxosv9000-9-2-4",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "nxosv9300-9-3-6",
//     "node_definition_id": "nxosv9000",
//     "description": "NX-OS 9300v 9.3.6",
//     "label": "NX-OS 9300v 9.3.6",
//     "disk_image": "nexus9300v.9.3.6.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "nxosv9300-9-3-6",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "nxosv9500-9-3-6",
//     "node_definition_id": "nxosv9000",
//     "description": "NX-OS 9500v 9.3.6",
//     "label": "NX-OS 9500v 9.3.6",
//     "disk_image": "nexus9500v.9.3.6.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": 4,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "nxosv9500-9-3-6",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "server-tcl-11-1",
//     "node_definition_id": "server",
//     "description": "Tiny Core Linux 11.1",
//     "label": "Server TCL 11.1",
//     "disk_image": "tcl-11-1.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "server-tcl-11-1",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "alpine-3-13-2-trex288",
//     "node_definition_id": "trex",
//     "description": "TRex 2.88 (based on Alpine 3.13.2)",
//     "label": "TRex 2.88",
//     "disk_image": "alpine-3-13-2-trex288.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "alpine-3-13-2-trex288",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "ubuntu-20-04-20210224",
//     "node_definition_id": "ubuntu",
//     "description": "20.04 - 24 Feb 2021",
//     "label": "Ubuntu 20.04 - 24 Feb 2021",
//     "disk_image": "focal-server-cloudimg-amd64.img",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "ubuntu-20-04-20210224",
//     "schema_version": "0.0.1"
//   },
//   {
//     "id": "alpine-3-13-2-wanem",
//     "node_definition_id": "wan_emulator",
//     "description": "WAN Emulator based on Alpine 3.13.2",
//     "label": "WAN Emulator 3.13.2",
//     "disk_image": "alpine-3-13-2-wanem.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "alpine-3-13-2-wanem",
//     "schema_version": "0.0.1"
//   }
// ]

func TestClient_GetImageDefs(t *testing.T) {
	tc := newAuthedTestAPIclient()

	aimgdef := `{
		"id": "alpine-3-10-base",
		"node_definition_id": "alpine",
		"description": "Alpine Linux and network tools",
		"label": "Alpine 3.10",
		"disk_image": "alpine-3-10-base.qcow2",
		"read_only": true,
		"ram": null,
		"cpus": null,
		"cpu_limit": null,
		"data_volume": null,
		"boot_disk_size": null,
		"disk_subfolder": "alpine-3-10-base",
		"schema_version": "0.0.1"
	}`
	zimgdef := `{
		"id": "zodiac-3-10-base",
		"node_definition_id": "alpine",
		"description": "Alpine Linux and network tools",
		"label": "Alpine 3.10",
		"disk_image": "alpine-3-10-base.qcow2",
		"read_only": true,
		"ram": null,
		"cpus": null,
		"cpu_limit": null,
		"data_volume": null,
		"boot_disk_size": null,
		"disk_subfolder": "alpine-3-10-base",
		"schema_version": "0.0.1"
	}`

	tests := []struct {
		name      string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good",
			mr.MockRespList{
				mr.MockResp{
					// need this to go through the sorting in the client
					// (highest -> lowest, e.g. 3.15  before 3.14)
					Data: []byte("[" + aimgdef + "," + zimgdef + "]"),
				},
			},
			false,
		},
		{
			"bad",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`"something failed!`),
					Code: http.StatusInternalServerError,
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		tc.mr.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			got, err := tc.client.ImageDefinitions(tc.ctx)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.GetImageDefs() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			expected := []ImageDefinition{}
			// this is properly sorted:
			b := bytes.NewReader([]byte("[" + zimgdef + "," + aimgdef + "]"))
			err = json.NewDecoder(b).Decode(&expected)
			if err != nil {
				t.Error("bad test data")
				return
			}
			if !reflect.DeepEqual(got, expected) {
				t.Errorf("Client.GetImageDefs() = %v, want %v", got, expected)
			}
		})
		if !tc.mr.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}
