// Package models provides the models for Cisco Modeling Labs
// here: node related types
package models

import (
	"encoding/json"
	"fmt"
	"sort"
)

// NodeState represents the operational state of a CML node.
type NodeState string

const (
	// NodeStateDefined indicates the node is defined on core.
	NodeStateDefined NodeState = "DEFINED_ON_CORE"
	// NodeStateStopped indicates the node is stopped.
	NodeStateStopped NodeState = "STOPPED"
	// NodeStateStarted indicates the node is started.
	NodeStateStarted NodeState = "STARTED"
	// NodeStateQueued indicates the node is queued for execution.
	NodeStateQueued NodeState = "QUEUED"
	// NodeStateBooted indicates the node is booted.
	NodeStateBooted NodeState = "BOOTED"
	// NodeStateDisconnected indicates the node is disconnected.
	NodeStateDisconnected NodeState = "DISCONNECTED"
)

// NodeConfig represents a named configuration for a node.
type NodeConfig struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// PyAtsCredentials represents node-level PyATS credentials.
// This matches the OpenAPI 2.10 schema `PyAtsCredentials`.
type PyAtsCredentials struct {
	Username       *string `json:"username"`
	Password       *string `json:"password"`
	EnablePassword *string `json:"enable_password"`
}

// BootProgress represents the boot progress state of a node.
type BootProgress string

const (
	// BootProgressNotRunning indicates the node is not running.
	BootProgressNotRunning BootProgress = "Not running"
	// BootProgressBooting indicates the node is booting.
	BootProgressBooting BootProgress = "Booting"
	// BootProgressBooted indicates the node is booted.
	BootProgressBooted BootProgress = "Booted"
)

// SerialDevice represents a serial device for a node.
type SerialDevice struct {
	ConsoleKey   UUID `json:"console_key"`
	DeviceNumber int  `json:"device_number"`
}

// SerialConsole represents a serial console for a node.
type SerialConsole struct {
	ConsoleKey   UUID   `json:"console_key"`
	DeviceNumber int    `json:"device_number"`
	Label        string `json:"label,omitempty"`
}

// NodeOperational contains operational data for a node.
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

// Node represents a CML node with its configuration and state.
type Node struct {
	ID             UUID              `json:"id"`
	LabID          UUID              `json:"lab_id"`
	Label          string            `json:"label"`
	X              int               `json:"x"`
	Y              int               `json:"y"`
	NodeDefinition string            `json:"node_definition"`
	CPUs           int               `json:"cpus"`
	Priority       *int              `json:"priority,omitempty"`
	PyATS          *PyAtsCredentials `json:"pyats,omitempty"`

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
	Operational     *NodeOperational `json:"operational,omitempty"`
	PinnedComputeID *UUID            `json:"pinned_compute_id,omitempty"`

	// Configuration can be string, array of NodeConfig, or single NodeConfig
	Configuration  any          `json:"configuration,omitempty"`
	Configurations []NodeConfig `json:"-"`

	// Interfaces and SerialDevices are not in the main schema but used internally
	Interfaces    InterfaceList  `json:"interfaces,omitempty"`
	SerialDevices []SerialDevice `json:"serial_devices,omitempty"`

	// Parameters field from schema
	Parameters any `json:"parameters,omitempty"`
}

// NodeMap is a map of node UUIDs to Node pointers.
type NodeMap map[UUID]*Node

// MarshalJSON implements json.Marshaler for NodeMap, sorting nodes by UUID for stable output.
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

// UnmarshalJSON implements json.Unmarshaler for Node, handling flexible configuration field types.
func (n *Node) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	type nodeAlias Node

	var tmpNode struct {
		nodeAlias
		Configs any `json:"configuration"`

		// Legacy/variant top-level operational fields seen on some backends.
		LegacyComputeID      *UUID           `json:"compute_id"`
		LegacyVNCkey         *UUID           `json:"vnc_key"`
		LegacyIOLAppID       *int            `json:"iol_app_id"`
		LegacyResourcePool   *string         `json:"resource_pool"`
		LegacySerialConsoles []SerialConsole `json:"serial_consoles"`
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
	case map[string]any:
		b, err := json.Marshal(thing)
		if err != nil {
			return err
		}
		var cfg NodeConfig
		if err := json.Unmarshal(b, &cfg); err != nil {
			return err
		}
		na.Configurations = []NodeConfig{cfg}
	default:
		return fmt.Errorf("unexpected type: %T", thing)
	}

	// If operational-only fields are present at top-level (legacy), migrate them.
	if tmpNode.LegacyComputeID != nil || tmpNode.LegacyVNCkey != nil || tmpNode.LegacyIOLAppID != nil ||
		tmpNode.LegacyResourcePool != nil || len(tmpNode.LegacySerialConsoles) > 0 {
		if na.Operational == nil {
			na.Operational = &NodeOperational{}
		}
		if tmpNode.LegacyComputeID != nil {
			na.Operational.ComputeID = tmpNode.LegacyComputeID
		}
		if tmpNode.LegacyVNCkey != nil {
			na.Operational.VNCkey = tmpNode.LegacyVNCkey
		}
		if tmpNode.LegacyIOLAppID != nil {
			na.Operational.IOLAppID = tmpNode.LegacyIOLAppID
		}
		if tmpNode.LegacyResourcePool != nil {
			na.Operational.ResourcePool = tmpNode.LegacyResourcePool
		}
		if len(tmpNode.LegacySerialConsoles) > 0 {
			na.Operational.SerialConsoles = tmpNode.LegacySerialConsoles
		}
	}
	*n = (Node)(na)

	return nil
}

// MarshalJSON implements json.Marshaler for Node, handling named configurations.
func (n *Node) MarshalJSON() ([]byte, error) {
	type alias Node
	if len(n.Configurations) > 0 {
		copyNode := *n
		copyNode.Configuration = nil
		return json.Marshal(&struct {
			*alias
			NamedConfig []NodeConfig `json:"configuration"`
		}{
			(*alias)(&copyNode),
			n.Configurations,
		})
	}
	return json.Marshal((*alias)(n))
}

// SameConfig compares the configuration of two nodes for equality.
func (n *Node) SameConfig(other *Node) bool {
	// Handle string configuration comparison
	if nStr, ok := n.Configuration.(string); ok {
		if otherStr, ok := other.Configuration.(string); ok {
			return nStr == otherStr
		}
		return false
	}

	// Handle nil configurations
	if n.Configuration == nil && other.Configuration == nil { //nolint:gocritic
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
