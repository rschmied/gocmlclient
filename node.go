package cmlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

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

type nodePatchPostAlias struct {
	Label           string       `json:"label,omitempty"`
	X               int          `json:"x"`
	Y               int          `json:"y"`
	HideLinks       bool         `json:"hide_links"`
	NodeDefinition  string       `json:"node_definition,omitempty"`
	ImageDefinition string       `json:"image_definition,omitempty"`
	Configuration   *string      `json:"configuration,omitempty"`
	Configurations  []NodeConfig `json:"-"`
	CPUs            int          `json:"cpus,omitempty"`
	CPUlimit        int          `json:"cpu_limit,omitempty"`
	RAM             int          `json:"ram,omitempty"`
	DataVolume      int          `json:"data_volume,omitempty"`
	BootDiskSize    int          `json:"boot_disk_size,omitempty"`
	Tags            []string     `json:"tags"`
}

func newNodeAlias(node *Node, update bool) nodePatchPostAlias {
	npp := nodePatchPostAlias{}

	npp.Label = node.Label
	npp.X = node.X
	npp.Y = node.Y
	npp.HideLinks = node.HideLinks
	npp.Tags = node.Tags

	// node tags can't be null, either the tag has to be omitted or it has to
	// be an empty list. But since we can't use "omitempty" we need to ensure
	// it's an empty list if no tags are provided.
	if node.Tags == nil {
		npp.Tags = []string{}
	}

	// these can be changed but only when the node VM doesn't exist
	if node.State == NodeStateDefined {
		npp.Configuration = node.Configuration
		npp.Configurations = make([]NodeConfig, len(node.Configurations))
		copy(npp.Configurations, node.Configurations)
		npp.CPUs = node.CPUs
		npp.CPUlimit = node.CPUlimit
		npp.RAM = node.RAM
		npp.DataVolume = node.DataVolume
		npp.BootDiskSize = node.BootDiskSize
		npp.ImageDefinition = node.ImageDefinition
	}

	// node definition can only be changed at create time (eg. POST)
	if !update {
		npp.NodeDefinition = node.NodeDefinition
	}

	// slog.Warn("NODE", slog.Any("node", node), slog.Any("npp", npp))

	return npp
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

func (node *Node) MarshalJSON() ([]byte, error) {
	type alias Node
	if len(node.Configurations) > 0 {
		node.Configuration = nil
		return json.Marshal(&struct {
			*alias
			NamedConfig []NodeConfig `json:"configuration"`
		}{
			(*alias)(node),
			node.Configurations,
		})
	}
	return json.Marshal((*alias)(node))
}

func (node Node) SameConfig(other Node) bool {
	if node.Configuration != nil && other.Configuration != nil && *other.Configuration != *node.Configuration {
		return false
	}

	if len(node.Configurations) != len(other.Configurations) {
		return false
	}

	for idx, cfg := range node.Configurations {
		if cfg.Name != other.Configurations[idx].Name {
			return false
		}
		if cfg.Content != other.Configurations[idx].Content {
			return false
		}
	}
	return true
}

func (node nodePatchPostAlias) MarshalJSON() ([]byte, error) {
	type alias nodePatchPostAlias
	if len(node.Configurations) > 0 {
		node.Configuration = nil
		return json.Marshal(&struct {
			alias
			NamedConfig []NodeConfig `json:"configuration"`
		}{
			(alias)(node),
			node.Configurations,
		})
	}
	return json.Marshal((alias)(node))
}

func (c *Client) getNodesForLab(ctx context.Context, lab *Lab) error {
	api := fmt.Sprintf("labs/%s/nodes?data=true", lab.ID)

	if c.useNamedConfigs {
		api += "&operational=true&exclude_configurations=false"
	}

	nodes := &nodeList{}
	err := c.jsonGet(ctx, api, nodes, 0)
	if err != nil {
		return err
	}

	nodeMap := make(NodeMap)
	for _, node := range *nodes {
		nodeMap[node.ID] = node
	}
	c.mu.Lock()
	lab.Nodes = nodeMap
	c.mu.Unlock()

	return nil
}

func (c *Client) nodeSetConfigData(ctx context.Context, node *Node, data any) error {
	api := fmt.Sprintf("labs/%s/nodes/%s", node.LabID, node.ID)

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(data)
	if err != nil {
		return err
	}

	// API returns the node ID of the updated node
	nodeID := ""
	err = c.jsonPatch(ctx, api, buf, &nodeID, 0)
	if err != nil {
		return err
	}
	_, err = c.NodeGet(ctx, node)
	return err
}

// NodeSetConfig sets a configuration for the specified node. At least the `ID`
// of the node and the `labID` must be provided in `node`. The `node` instance
// will be updated with the current values for the node as provided by the
// controller.
func (c *Client) NodeSetConfig(ctx context.Context, node *Node, configuration string) error {
	nodeCfg := struct {
		Configuration string `json:"configuration"`
	}{configuration}
	return c.nodeSetConfigData(ctx, node, nodeCfg)
}

// NodeSetNamedConfigs sets a list of named configurations for the specified
// node. At least the `ID` of the node and the `labID` must be provided in
// `node`.
func (c *Client) NodeSetNamedConfigs(ctx context.Context, node *Node, configs []NodeConfig) error {
	nodeCfg := struct {
		NamedConfigs []NodeConfig `json:"configuration"`
	}{configs}
	return c.nodeSetConfigData(ctx, node, nodeCfg)
}

// NodeUpdate updates the node specified by data in `node` (e.g. ID and LabID)
// with the other data provided. It returns the updated node.
func (c *Client) NodeUpdate(ctx context.Context, node *Node) (*Node, error) {
	api := fmt.Sprintf("labs/%s/nodes/%s", node.LabID, node.ID)

	postAlias := newNodeAlias(node, true)

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(postAlias)
	if err != nil {
		return nil, err
	}

	// API returns "just" the node ID of the updated node
	nodeID := ""
	err = c.jsonPatch(ctx, api, buf, &nodeID, 0)
	if err != nil {
		return nil, err
	}
	return c.NodeGet(ctx, node)
}

// NodeStart starts the given node.
func (c *Client) NodeStart(ctx context.Context, node *Node) error {
	api := fmt.Sprintf("labs/%s/nodes/%s/state/start", node.LabID, node.ID)
	err := c.jsonPut(ctx, api, 0)
	if err != nil {
		return err
	}
	return nil
}

// NodeStop stops the given node.
func (c *Client) NodeStop(ctx context.Context, node *Node) error {
	api := fmt.Sprintf("labs/%s/nodes/%s/state/stop", node.LabID, node.ID)
	err := c.jsonPut(ctx, api, 0)
	if err != nil {
		return err
	}
	return nil
}

// NodeCreate creates a new node on the controller based on the data provided
// in `node`. Label, node definition and image definition must be provided.
func (c *Client) NodeCreate(ctx context.Context, node *Node) (*Node, error) {
	// TODO: inconsistent attributes lab_title vs title, ..
	node.State = NodeStateDefined
	postAlias := newNodeAlias(node, false)
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(postAlias)
	if err != nil {
		return nil, err
	}

	newNode := Node{}

	// return value of create is just
	// {
	// 	"id": "fe106ef1-cddc-49f7-9983-7ac461e96f47"
	// }

	// we want those "default" interfaces in the node
	api := fmt.Sprintf("labs/%s/nodes?populate_interfaces=true", node.LabID)
	err = c.jsonPost(ctx, api, buf, &newNode, 0)
	if err != nil {
		return nil, err
	}

	// FIX: Since the create does not use all possible values, we need to follow
	// up with a PATCH (this is an API bug, imo)
	// ram, cpu, ...

	// NodeDefinition can't be set even when the node is DEFINED_ON_CORE (since
	// the struct has them as "omitempty", this is OK)... So for the patch here,
	// it's required to be set to empty from the struct
	postAlias.NodeDefinition = ""

	buf.Reset()
	err = json.NewEncoder(buf).Encode(postAlias)
	if err != nil {
		return nil, err
	}
	api = fmt.Sprintf("labs/%s/nodes/%s", node.LabID, newNode.ID)
	// the return of the patch API is simply the node ID as a string!
	// FIX: inconsistency of patch API
	err = c.jsonPatch(ctx, api, buf, nil, 0)
	if err != nil {
		// for consistency, remove the created node that can't be updated
		// this assumes that the error was because of the provided data and
		// not because of e.g. a connectivity issue between the initial create
		// and the attempted removal.
		node.ID = newNode.ID
		c.NodeDestroy(ctx, node)
		return nil, err
	}

	node.ID = newNode.ID
	node.Interfaces = InterfaceList{}

	// fetch the node again, with all data
	return c.NodeGet(ctx, node)
}

// NodeGet returns the node identified by its `ID` and `LabID` in the provided node.
func (c *Client) NodeGet(ctx context.Context, node *Node) (*Node, error) {
	// SIMPLE-5052 -- results are different for simplified=true vs false for
	// the inherited values. In the simplified case, all values are always
	// null.

	var err error
	newNode := Node{}
	api := fmt.Sprintf("labs/%s/nodes/%s", node.LabID, node.ID)
	if c.useNamedConfigs {
		api += "?operational=true&exclude_configurations=false"
	}
	err = c.jsonGet(ctx, api, &newNode, 0)
	return &newNode, err
}

// NodeDestroy deletes the node from the controller.
func (c *Client) NodeDestroy(ctx context.Context, node *Node) error {
	api := fmt.Sprintf("labs/%s/nodes/%s", node.LabID, node.ID)
	return c.jsonDelete(ctx, api, 0)
}

// NodeWipe removes all runtime data from a node on the controller/compute.
// E.g. it will remove the actual VM and its associated disks.
func (c *Client) NodeWipe(ctx context.Context, node *Node) error {
	api := fmt.Sprintf("labs/%s/nodes/%s/wipe_disks", node.LabID, node.ID)
	return c.jsonPut(ctx, api, 0)
}
