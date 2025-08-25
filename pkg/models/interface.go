// Package models provides the models for Cisco Modeling Labs
// here: interface related types
package models

/*
{
	"id": "e87c811d-5459-4390-8e92-317bb9dc23e8",
	"lab_id": "024fa9f4-5e5e-4e94-9f85-29f147e09689",
	"node": "f902d112-2a93-4c9f-98e6-adea6dc16fef",
	"label": "eth0",
	"slot": 0,
	"type": "physical",
	"device_name": null,
	"dst_udp_port": 21001,
	"src_udp_port": 21000,
	"mac_address": "52:54:00:1e:af:9b",
	"is_connected": true,
	"state": "STARTED"
}
*/

const (
	IfaceStateDefined = "DEFINED_ON_CORE"
	IfaceStateStopped = "STOPPED"
	IfaceStateStarted = "STARTED"

	IfaceTypePhysical = "physical"
	IfaceTypeLoopback = "loopback"
)

type (
	Interface struct {
		ID          string `json:"id"`
		LabID       string `json:"lab_id"`
		Node        string `json:"node"`
		Label       string `json:"label"`
		Slot        int    `json:"slot"`
		Type        string `json:"type"`
		DeviceName  string `json:"device_name"`
		SrcUDPport  int    `json:"src_udp_port"`
		DstUDPport  int    `json:"dst_udp_port"`
		MACaddress  string `json:"mac_address"`
		IsConnected bool   `json:"is_connected"`
		State       string `json:"state"`

		// extra
		IP4 []string `json:"ip4"`
		IP6 []string `json:"ip6"`

		// needed for internal linking
		node *Node
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
