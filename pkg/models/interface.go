// Package models provides the models for Cisco Modeling Labs
// here: interface related types
package models

type (
	IfaceState string
	IfaceType  string
)

const (
	IfaceStateDefined IfaceState = "DEFINED_ON_CORE"
	IfaceStateStopped IfaceState = "STOPPED"
	IfaceStateStarted IfaceState = "STARTED"

	IfaceTypePhysical IfaceType = "physical"
	IfaceTypeLoopback IfaceType = "loopback"
)

type Operational struct {
	DeviceName *string `json:"device_name,omitempty"`
	MACaddress *string `json:"mac_address,omitempty"`
	SrcUDPport *int    `json:"src_udp_port,omitempty"`
	DstUDPport *int    `json:"dst_udp_port,omitempty"`
}

/*
[
  {
    "id": "6251c7bc-a273-4634-8ca9-b8d9c994ef8c",
    "is_connected": true,
    "lab_id": "20c0efde-cdaf-4dad-b6df-dd568ddf6e8d",
    "label": "eth0",
    "mac_address": null,
    "node": "c63009d9-fbb7-4dcf-a979-bcf11b2377ef",
    "slot": 0,
    "type": "physical",
    "state": "STARTED",
    "operational": {
      "device_name": "nf525400b30bed",
      "mac_address": "52:54:00:b3:0b:ed",
      "dst_udp_port": null,
      "src_udp_port": null
    }
  }
]
*/

type (
	Interface struct {
		ID          UUID         `json:"id"`
		LabID       UUID         `json:"lab_id,omitempty"`
		Node        UUID         `json:"node"`
		Label       string       `json:"label"`
		Slot        *int         `json:"slot,omitempty"`
		Type        IfaceType    `json:"type,omitempty"`
		State       IfaceState   `json:"state,omitempty"`
		IsConnected bool         `json:"is_connected"`
		Operational *Operational `json:"operational,omitempty"`
		// MACAddress  *string      `json:"mac_address,omitempty"`
		// SrcUDPPort  int          `json:"src_udp_port"`
		// DstUDPPort  int          `json:"dst_udp_port"`
		// DeviceName  *string      `json:"device_name,omitempty"`

		// extra fields not in OpenAPI spec
		IP4 []string `json:"ip4,omitempty"`
		IP6 []string `json:"ip6,omitempty"`
	}
	InterfaceList []*Interface
)

func (i Interface) Exists() bool {
	return i.State != IfaceStateDefined
}

func (i Interface) Runs() bool {
	return i.State == IfaceStateStarted
}

func (i Interface) IsPhysical() bool {
	return i.Type == IfaceTypePhysical
}
