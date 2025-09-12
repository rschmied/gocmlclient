// Package services, image definition specific
package services

import (
	"context"
	"sort"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/pkg/models"
)

// ImageDefinitionService provides image definition-related operations
type ImageDefinitionService struct {
	apiClient *api.Client
}

// NewImageDefinitionService creates a new image definition service
func NewImageDefinitionService(apiClient *api.Client) *ImageDefinitionService {
	return &ImageDefinitionService{
		apiClient: apiClient,
	}
}

// ImageDefinitions returns a list of image definitions known to the controller.
func (s *ImageDefinitionService) ImageDefinitions(ctx context.Context) ([]models.ImageDefinition, error) {
	imgDef := []models.ImageDefinition{}
	err := s.apiClient.GetJSON(ctx, "image_definitions", nil, &imgDef)
	if err != nil {
		return nil, err
	}

	// sort the image list by their ID
	sort.Slice(imgDef, func(i, j int) bool {
		return imgDef[i].ID > imgDef[j].ID
	})

	return imgDef, nil
}
