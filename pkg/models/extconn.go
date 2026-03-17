// Package models provides the models for Cisco Modeling Labs
// here: external connector related types
package models

// ExtConn defines the data structure for a CML external connector
type ExtConn struct {
	ID          UUID     `json:"id"`
	DeviceName  string   `json:"device_name"`
	Label       string   `json:"label"`
	Protected   bool     `json:"protected"`
	Snooped     bool     `json:"snooped"`
	Tags        []string `json:"tags"`
	Operational opdata   `json:"operational"`
}

type opdata struct {
	Forwarding string   `json:"forwarding"`
	Label      string   `json:"label"`
	MTU        int      `json:"mtu"`
	Exists     bool     `json:"exists"`
	Enabled    bool     `json:"enabled"`
	Protected  bool     `json:"protected"`
	Snooped    bool     `json:"snooped"`
	STP        bool     `json:"stp"`
	IPNetworks []string `json:"ip_networks"`
}
