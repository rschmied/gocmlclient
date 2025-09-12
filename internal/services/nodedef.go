// Package services, node definition specific
package services

import (
	"context"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/pkg/models"
)

// NodeDefinitionService provides node definition-related operations
type NodeDefinitionService struct {
	apiClient *api.Client
}

// NewNodeDefinitionService creates a new node definition service
func NewNodeDefinitionService(apiClient *api.Client) *NodeDefinitionService {
	return &NodeDefinitionService{
		apiClient: apiClient,
	}
}

// NodeDefinitions returns the list of node definitions available on the CML
// controller. The key of the map is the definition type name (e.g. "alpine" or
// "ios"). The node def data structure matches the OpenAPI schema.
func (s *NodeDefinitionService) NodeDefinitions(ctx context.Context) (models.NodeDefinitionMap, error) {
	nd := []models.NodeDefinition{}
	err := s.apiClient.GetJSON(ctx, "simplified_node_definitions", nil, &nd)
	if err != nil {
		return nil, err
	}

	nodeDefMap := make(models.NodeDefinitionMap)
	for _, nodeDef := range nd {
		nodeDefMap[nodeDef.ID] = nodeDef
	}
	return nodeDefMap, nil
}
