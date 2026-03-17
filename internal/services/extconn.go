// Package services, external connector specific
package services

import (
	"context"
	"fmt"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/pkg/models"
)

// ExtConnService provides external connector-related operations
type ExtConnService struct {
	apiClient *api.Client
}

// Ensure ExtConnService implements interface
var _ ExtConnServiceInterface = (*ExtConnService)(nil)

// ExtConnServiceInterface defines methods needed by other services/clients.
type ExtConnServiceInterface interface {
	Get(ctx context.Context, extConnID models.UUID) (models.ExtConn, error)
	List(ctx context.Context) ([]*models.ExtConn, error)
}

// NewExtConnService creates a new external connector service
func NewExtConnService(apiClient *api.Client) *ExtConnService {
	return &ExtConnService{
		apiClient: apiClient,
	}
}

// Get returns the external connector specified by the ID given
func (s *ExtConnService) Get(ctx context.Context, extConnID models.UUID) (models.ExtConn, error) {
	api := fmt.Sprintf("system/external_connectors/%s", extConnID)
	var extconn models.ExtConn
	err := s.apiClient.GetJSON(ctx, api, nil, &extconn)
	if err != nil {
		return models.ExtConn{}, err
	}
	return extconn, err
}

// List returns all external connectors on the system
func (s *ExtConnService) List(ctx context.Context) ([]*models.ExtConn, error) {
	extconnlist := make([]*models.ExtConn, 0)
	err := s.apiClient.GetJSON(ctx, "system/external_connectors", nil, &extconnlist)
	if err != nil {
		return nil, err
	}
	return extconnlist, err
}
