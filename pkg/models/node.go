// Package models provides the models for Cisco Modeling Labs
// here: node related types
package models

import (
	"encoding/json"
	"fmt"
	"sort"
)

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
	ConsoleKey   UUID `json:"console_key"`
	DeviceNumber int  `json:"device_number"`
}

type Node struct {
	ID              UUID           `json:"id"`
	LabID           UUID           `json:"lab_id"`
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
	VNCkey          UUID           `json:"vnc_key"`
	SerialDevices   []SerialDevice `json:"serial_devices"`
	ComputeID       UUID           `json:"compute_id"`

	// Configurations is not exported, it's overloaded within the API
}

// {
// 	"boot_disk_size": 0,
// 	"compute_id": "9c2519bf-dda6-4d31-942e-8068a6349b5e",
// 	"configuration": "bridge0",
// 	"cpu_limit": 100,
// 	"cpus": 0,
// 	"data_volume": 0,
// 	"hide_links": false,
// 	"id": "9efb1503-7e2a-4d2a-959e-865209f1acc0",
// 	"image_definition": null,
// 	"lab_id": "52d5c824-e10c-450a-b9c5-b700bd3bc17a",
// 	"label": "ext-conn-0",
// 	"node_definition": "external_connector",
// 	"ram": 0,
// 	"tags": [],
// 	"vnc_key": "",
// 	"x": 317,
// 	"y": 341,
// 	"config_filename": "noname",
// 	"config_mediatype": "ISO",
// 	"config_image_path": "/var/local/virl2/images/52d5c824-e10c-450a-b9c5-b700bd3bc17a/9efb1503-7e2a-4d2a-959e-865209f1acc0/config.img",
// 	"cpu_model": null,
// 	"data_image_path": "/var/local/virl2/images/52d5c824-e10c-450a-b9c5-b700bd3bc17a/9efb1503-7e2a-4d2a-959e-865209f1acc0/data.img",
// 	"disk_image": null,
// 	"disk_image_2": null,
// 	"disk_image_3": null,
// 	"disk_image_path": null,
// 	"disk_image_path_2": null,
// 	"disk_image_path_3": null,
// 	"disk_driver": null,
// 	"driver_id": "external_connector",
// 	"efi_boot": false,
// 	"image_dir": "/var/local/virl2/images/52d5c824-e10c-450a-b9c5-b700bd3bc17a/9efb1503-7e2a-4d2a-959e-865209f1acc0",
// 	"libvirt_image_dir": "/var/lib/libvirt/images/virl-base-images",
// 	"nic_driver": null,
// 	"number_of_serial_devices": 0,
// 	"serial_devices": [],
// 	"video_memory": 0,
// 	"video_model": null,
// 	"state": "BOOTED",
// 	"boot_progress": "Booted"
//   }

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
	if n.Configuration != nil && other.Configuration != nil && *other.Configuration != *n.Configuration {
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
