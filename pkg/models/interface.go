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
	DeviceName string `json:"device_name"`
	MACaddress string `json:"mac_address"`
	SrcUDPport int    `json:"src_udp_port"`
	DstUDPport int    `json:"dst_udp_port"`
}

type (
	Interface struct {
		ID          UUID        `json:"id"`
		LabID       UUID        `json:"lab_id"`
		Node        UUID        `json:"node"`
		Label       string      `json:"label"`
		Slot        int         `json:"slot"`
		Type        IfaceType   `json:"type"`
		IsConnected bool        `json:"is_connected"`
		State       IfaceState  `json:"state"`
		Operational Operational `json:"operational"`

		// extra
		IP4 []string `json:"ip4"`
		IP6 []string `json:"ip6"`
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
