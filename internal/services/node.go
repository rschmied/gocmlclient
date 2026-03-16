package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/httputil"
	"github.com/rschmied/gocmlclient/pkg/models"
)

const (
	nodesAPI = "nodes"
)

// Ensure NodeService implements interface
var _ NodeServiceInterface = (*NodeService)(nil)

// NodeServiceInterface defines methods needed by other services
type NodeServiceInterface interface {
	GetNodesForLab(ctx context.Context, labID models.UUID) (models.NodeMap, error)
	GetByID(ctx context.Context, labID, id models.UUID) (models.Node, error)
}

// NodeService provides node-related operations
type NodeService struct {
	apiClient       *api.Client
	useNamedConfigs bool
	excludeConfigs  *bool
}

// NewNodeService creates a new node service
func NewNodeService(apiClient *api.Client, useNamedConfigs bool) *NodeService {
	return &NodeService{
		apiClient:       apiClient,
		useNamedConfigs: useNamedConfigs,
	}
}

// SetExcludeConfigurations forces exclude_configurations query behavior for node
// read operations (GET/LIST). If v is nil, the parameter is omitted.
func (s *NodeService) SetExcludeConfigurations(v *bool) {
	s.excludeConfigs = v
}

// nodesURL builds base URL for nodes in a lab
func nodesURL(labID models.UUID) string {
	return fmt.Sprintf("%s/%s/%s", labsAPI, labID, nodesAPI)
}

// nodeURL builds URL for a specific node
func nodeURL(labID, nodeID models.UUID) string {
	return fmt.Sprintf("%s/%s", nodesURL(labID), nodeID)
}

// nodeStateURL builds URL for node state operations
func nodeStateURL(labID, nodeID models.UUID, action string) string {
	return fmt.Sprintf("%s/state/%s", nodeURL(labID, nodeID), action)
}

// nodeWipeURL builds URL for disk wipe operations
func nodeWipeURL(labID, nodeID models.UUID) string {
	return fmt.Sprintf("%s/wipe_disks", nodeURL(labID, nodeID))
}

type (
	nodePatchPostAlias struct {
		Label           string              `json:"label,omitempty"`
		X               int                 `json:"x"`
		Y               int                 `json:"y"`
		HideLinks       *bool               `json:"hide_links,omitempty"`
		NodeDefinition  string              `json:"node_definition,omitempty"`
		ImageDefinition *string             `json:"image_definition,omitempty"`
		Configuration   *string             `json:"configuration,omitempty"`
		Configurations  []models.NodeConfig `json:"-"`
		CPUs            int                 `json:"cpus,omitempty"`
		CPUlimit        *int                `json:"cpu_limit,omitempty"`
		RAM             *int                `json:"ram,omitempty"`
		DataVolume      *int                `json:"data_volume,omitempty"`
		BootDiskSize    *int                `json:"boot_disk_size,omitempty"`
		Tags            []string            `json:"tags"`
	}
)

func newNodeAlias(node *models.Node, update bool) nodePatchPostAlias {
	npp := nodePatchPostAlias{}

	npp.Label = node.Label
	npp.X = node.X
	npp.Y = node.Y
	npp.Tags = node.Tags

	// Handle pointer types
	npp.HideLinks = node.HideLinks

	// node tags can't be null, either the tag has to be omitted or it has to
	// be an empty list. But since we can't use "omitempty" we need to ensure
	// it's an empty list if no tags are provided.
	if node.Tags == nil {
		npp.Tags = []string{}
	}

	// these can be changed but only when the node VM doesn't exist
	if node.State == models.NodeStateDefined {
		// Handle configuration which can be string, NodeConfig, or []NodeConfig
		if configStr, ok := node.Configuration.(string); ok {
			npp.Configuration = &configStr
		}
		npp.Configurations = make([]models.NodeConfig, len(node.Configurations))
		copy(npp.Configurations, node.Configurations)
		npp.CPUs = node.CPUs

		// Handle pointer types
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

	return npp
}

func (node nodePatchPostAlias) MarshalJSON() ([]byte, error) {
	type alias nodePatchPostAlias
	if len(node.Configurations) > 0 {
		node.Configuration = nil
		return json.Marshal(&struct {
			alias
			NamedConfig []models.NodeConfig `json:"configuration"`
		}{
			(alias)(node),
			node.Configurations,
		})
	}
	return json.Marshal((alias)(node))
}

func (s *NodeService) GetNodesForLab(ctx context.Context, labID models.UUID) (models.NodeMap, error) {
	api := nodesURL(labID)

	queryParams := httputil.NewQueryBuilder().
		WithData(true).
		WithNamedConfigs(s.useNamedConfigs).
		WithExcludeConfigurations(s.excludeConfigs).
		Build()

	// First unmarshal into a slice of nodes
	var nodes []models.Node
	err := s.apiClient.GetJSON(ctx, api, queryParams, &nodes)
	if err != nil {
		return nil, err
	}

	// Convert slice to NodeMap
	nodeMap := make(models.NodeMap)
	for i := range nodes {
		nodeMap[nodes[i].ID] = &nodes[i]
	}

	return nodeMap, nil
}

func (s *NodeService) setConfigData(ctx context.Context, node *models.Node, data any) error {
	api := nodeURL(node.LabID, node.ID)

	// API returns the node ID of the updated node
	var nodeID models.UUID = ""
	err := s.apiClient.PatchJSON(ctx, api, nil, data, &nodeID)
	if err != nil {
		return err
	}
	updatedNode, err := s.GetByID(ctx, node.LabID, nodeID)
	if err != nil {
		return err
	}
	// Update the original node with the new data
	*node = updatedNode
	return nil
}

// SetConfig sets a configuration for the specified node. At least the `ID`
// of the node and the `labID` must be provided in `node`. The `node` instance
// will be updated with the current values for the node as provided by the
// controller.
func (s *NodeService) SetConfig(ctx context.Context, node *models.Node, configuration string) error {
	nodeCfg := struct {
		Configuration string `json:"configuration"`
	}{configuration}
	return s.setConfigData(ctx, node, nodeCfg)
}

// SetNamedConfigs sets a list of named configurations for the specified
// node. At least the `ID` of the node and the `labID` must be provided in
// `node`.
func (s *NodeService) SetNamedConfigs(ctx context.Context, node *models.Node, configs []models.NodeConfig) error {
	nodeCfg := struct {
		NamedConfigs []models.NodeConfig `json:"configuration"`
	}{configs}
	return s.setConfigData(ctx, node, nodeCfg)
}

// Update updates the node specified by data in `node` (e.g. ID and LabID) with
// the other data provided. It returns the updated node.
func (s *NodeService) Update(ctx context.Context, node models.Node) (models.Node, error) {
	api := nodeURL(node.LabID, node.ID)

	postAlias := newNodeAlias(&node, true)

	// API returns "just" the node ID of the updated node
	var nodeID models.UUID = ""
	err := s.apiClient.PatchJSON(ctx, api, nil, postAlias, &nodeID)
	if err != nil {
		return models.Node{}, err
	}
	return s.GetByID(ctx, node.LabID, nodeID)
}

// Start starts the given node.
func (s *NodeService) Start(ctx context.Context, labID, nodeID models.UUID) error {
	api := nodeStateURL(labID, nodeID, "start")
	return s.apiClient.PutJSON(ctx, api, 0)
}

// Stop stops the given node.
func (s *NodeService) Stop(ctx context.Context, labID, nodeID models.UUID) error {
	api := nodeStateURL(labID, nodeID, "stop")
	return s.apiClient.PutJSON(ctx, api, 0)
}

// Create creates a new node on the controller based on the data provided
// in `node`. Label, node definition and image definition must be provided.
func (s *NodeService) Create(ctx context.Context, node models.Node) (models.Node, error) {
	// TODO: inconsistent attributes lab_title vs title, ..
	node.State = models.NodeStateDefined
	postAlias := newNodeAlias(&node, false)

	var newNode models.Node

	// we want those "default" interfaces in the node
	queryParams := httputil.NewQueryBuilder().
		WithPopulateInterfaces().
		Build()
	api := nodesURL(node.LabID)
	err := s.apiClient.PostJSON(ctx, api, queryParams, postAlias, &newNode)
	if err != nil {
		return models.Node{}, err
	}

	// FIX: Since the create does not use all possible values, we need to follow
	// up with a PATCH (this is an API bug, imo)
	// ram, cpu, ...

	// NodeDefinition can't be set even when the node is DEFINED_ON_CORE (since
	// the struct has them as "omitempty", this is OK)... So for the patch here,
	// it's required to be set to empty from the struct
	postAlias.NodeDefinition = ""

	api = nodeURL(node.LabID, newNode.ID)

	// the return of the patch API is simply the node ID as a string!
	// FIX: inconsistency of patch API
	err = s.apiClient.PatchJSON(ctx, api, nil, postAlias, nil)
	if err != nil {
		// for consistency, remove the created node that can't be updated this
		// assumes that the error was because of the provided data and not because
		// of e.g. a connectivity issue between the initial create and the
		// attempted removal.
		node.ID = newNode.ID
		s.Delete(ctx, node.LabID, node.ID)
		return models.Node{}, err
	}

	node.ID = newNode.ID
	node.Interfaces = models.InterfaceList{}

	// fetch the node again, with all data
	return s.GetByID(ctx, node.LabID, node.ID)
}

// GetByID returns the node identified by its `ID` and `LabID` in the provided node.
func (s *NodeService) GetByID(ctx context.Context, labID, id models.UUID) (models.Node, error) {
	// SIMPLE-5052 -- results are different for simplified=true vs false for
	// the inherited values. In the simplified case, all values are always
	// null.

	var err error
	var newNode models.Node
	api := nodeURL(labID, id)
	queryParams := httputil.NewQueryBuilder().
		WithNamedConfigs(s.useNamedConfigs).
		WithExcludeConfigurations(s.excludeConfigs).
		Build()
	err = s.apiClient.GetJSON(ctx, api, queryParams, &newNode)
	return newNode, err
}

// Delete deletes the node from the controller.
func (s *NodeService) Delete(ctx context.Context, labID, nodeID models.UUID) error {
	api := nodeURL(labID, nodeID)
	return s.apiClient.DeleteJSON(ctx, api, nil)
}

// Wipe removes all runtime data from a node on the controller/compute. E.g. it
// will remove the actual VM and its associated disks.
func (s *NodeService) Wipe(ctx context.Context, labID, nodeID models.UUID) error {
	api := nodeWipeURL(labID, nodeID)
	return s.apiClient.PutJSON(ctx, api, nil)
}
