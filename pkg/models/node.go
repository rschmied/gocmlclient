// Package models provides the models for Cisco Modeling Labs
// here: node related types
package models

const (
	NodeStateDefined = "DEFINED_ON_CORE"
	NodeStateStopped = "STOPPED"
	NodeStateStarted = "STARTED"
	NodeStateBooted  = "BOOTED"
)

type NodeConfig struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type SerialDevice struct {
	ConsoleKey   string `json:"console_key"`
	DeviceNumber int    `json:"device_number"`
}

type Node struct {
	ID              string         `json:"id"`
	LabID           string         `json:"lab_id"`
	Label           string         `json:"label"`
	X               int            `json:"x"`
	Y               int            `json:"y"`
	HideLinks       bool           `json:"hide_links"`
	NodeDefinition  string         `json:"node_definition"`
	ImageDefinition string         `json:"image_definition"`
	Configuration   *string        `json:"configuration"`
	Configurations  []NodeConfig   `json:"-"`
	CPUs            int            `json:"cpus"`
	CPUlimit        int            `json:"cpu_limit"`
	RAM             int            `json:"ram"`
	State           string         `json:"state"`
	DataVolume      int            `json:"data_volume"`
	BootDiskSize    int            `json:"boot_disk_size"`
	Interfaces      InterfaceList  `json:"interfaces,omitempty"`
	Tags            []string       `json:"tags"`
	VNCkey          string         `json:"vnc_key"`
	SerialDevices   []SerialDevice `json:"serial_devices"`
	ComputeID       string         `json:"compute_id"`

	// Configurations is not exported, it's overloaded within the API
}
