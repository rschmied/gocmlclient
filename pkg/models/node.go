// Package models provides the models for Cisco Modeling Labs
// here: node related types
package models

import (
	"encoding/json"
	"fmt"
	"sort"
)

type NodeState string

const (
	NodeStateDefined      NodeState = "DEFINED_ON_CORE"
	NodeStateStopped      NodeState = "STOPPED"
	NodeStateStarted      NodeState = "STARTED"
	NodeStateQueued       NodeState = "QUEUED"
	NodeStateBooted       NodeState = "BOOTED"
	NodeStateDisconnected NodeState = "DISCONNECTED"
)

type NodeConfig struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type BootProgress string

const (
	BootProgressNotRunning BootProgress = "Not running"
	BootProgressBooting    BootProgress = "Booting"
	BootProgressBooted     BootProgress = "Booted"
)

type SerialDevice struct {
	ConsoleKey   UUID `json:"console_key"`
	DeviceNumber int  `json:"device_number"`
}

type SerialConsole struct {
	ConsoleKey   UUID   `json:"console_key"`
	DeviceNumber int    `json:"device_number"`
	Label        string `json:"label,omitempty"`
}

type NodeOperational struct {
	BootDiskSize    *int            `json:"boot_disk_size"`
	CPUlimit        *int            `json:"cpu_limit"`
	CPUs            *int            `json:"cpus"`
	DataVolume      *int            `json:"data_volume"`
	RAM             *int            `json:"ram"`
	ComputeID       *UUID           `json:"compute_id"`
	ImageDefinition *string         `json:"image_definition"`
	VNCkey          *UUID           `json:"vnc_key"`
	ResourcePool    *string         `json:"resource_pool,omitempty"`
	IOLAppID        *int            `json:"iol_app_id,omitempty"`
	SerialConsoles  []SerialConsole `json:"serial_consoles,omitempty"`
}

type Node struct {
	ID             UUID   `json:"id"`
	LabID          UUID   `json:"lab_id"`
	Label          string `json:"label"`
	X              int    `json:"x"`
	Y              int    `json:"y"`
	NodeDefinition string `json:"node_definition"`
	CPUs           int    `json:"cpus"`

	// Optional fields with proper null handling
	ImageDefinition *string          `json:"image_definition,omitempty"`
	RAM             *int             `json:"ram,omitempty"`
	CPUlimit        *int             `json:"cpu_limit,omitempty"`
	DataVolume      *int             `json:"data_volume,omitempty"`
	BootDiskSize    *int             `json:"boot_disk_size,omitempty"`
	HideLinks       *bool            `json:"hide_links,omitempty"`
	Tags            []string         `json:"tags,omitempty"`
	State           NodeState        `json:"state,omitempty"`
	BootProgress    BootProgress     `json:"boot_progress,omitempty"`
	ComputeID       *UUID            `json:"compute_id,omitempty"`
	IOLAppID        *int             `json:"iol_app_id,omitempty"`
	Operational     *NodeOperational `json:"operational,omitempty"`
	ResourcePool    *string          `json:"resource_pool,omitempty"`
	VNCkey          *UUID            `json:"vnc_key,omitempty"`
	PinnedComputeID *UUID            `json:"pinned_compute_id,omitempty"`
	SerialConsoles  []SerialConsole  `json:"serial_consoles,omitempty"`

	// Configuration can be string, array of NodeConfig, or single NodeConfig
	Configuration  any          `json:"configuration,omitempty"`
	Configurations []NodeConfig `json:"-"`

	// Interfaces and SerialDevices are not in the main schema but used internally
	Interfaces    InterfaceList  `json:"interfaces,omitempty"`
	SerialDevices []SerialDevice `json:"serial_devices,omitempty"`

	// Parameters field from schema
	Parameters any `json:"parameters,omitempty"`
}

func (nmap NodeMap) MarshalJSON() ([]byte, error) {
	nodeList := []*Node{}
	for _, node := range nmap {
		nodeList = append(nodeList, node)
	}
	// we want this as a stable sort by node UUID
	sort.Slice(nodeList, func(i, j int) bool {
		return nodeList[i].ID < nodeList[j].ID
	})

	return json.Marshal(nodeList)
}

func (n *Node) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	type nodeAlias Node

	var tmpNode struct {
		nodeAlias
		Configs any `json:"configuration"`
	}

	// Unmarshal the JSON into the tmpNode struct.
	if err := json.Unmarshal(data, &tmpNode); err != nil {
		return err
	}

	na := tmpNode.nodeAlias

	switch thing := tmpNode.Configs.(type) {
	case nil:
		na.Configuration = nil
	case string:
		na.Configuration = &thing
	case []any:
		b, err := json.Marshal(thing)
		if err != nil {
			return err
		}
		err = json.Unmarshal(b, &na.Configurations)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unexpected type: %T", thing)
	}
	*n = (Node)(na)

	return nil
}

func (n *Node) MarshalJSON() ([]byte, error) {
	type alias Node
	if len(n.Configurations) > 0 {
		n.Configuration = nil
		return json.Marshal(&struct {
			*alias
			NamedConfig []NodeConfig `json:"configuration"`
		}{
			(*alias)(n),
			n.Configurations,
		})
	}
	return json.Marshal((*alias)(n))
}

func (n *Node) SameConfig(other *Node) bool {
	// Handle string configuration comparison
	if nStr, ok := n.Configuration.(string); ok {
		if otherStr, ok := other.Configuration.(string); ok {
			return nStr == otherStr
		}
		return false
	}

	// Handle nil configurations
	if n.Configuration == nil && other.Configuration == nil {
		// Both are nil, check named configurations
	} else if n.Configuration != nil && other.Configuration != nil {
		// Both are non-nil but not strings, they should be equal
		return n.Configuration == other.Configuration
	} else {
		// One is nil, other is not
		return false
	}

	if len(n.Configurations) != len(other.Configurations) {
		return false
	}

	for idx, cfg := range n.Configurations {
		if cfg.Name != other.Configurations[idx].Name {
			return false
		}
		if cfg.Content != other.Configurations[idx].Content {
			return false
		}
	}
	return true
}
