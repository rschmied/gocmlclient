// Package services, smart annotation specific
package services

import (
	"context"
	"fmt"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/pkg/models"
)

const smartAnnotationsAPI = "smart_annotations"

// Ensure SmartAnnotationService implements interface
var _ SmartAnnotationServiceInterface = (*SmartAnnotationService)(nil)

// SmartAnnotationServiceInterface defines methods needed by other services/clients.
type SmartAnnotationServiceInterface interface {
	List(ctx context.Context, labID models.UUID) ([]models.SmartAnnotation, error)
	Get(ctx context.Context, labID, id models.UUID) (models.SmartAnnotation, error)
	Update(ctx context.Context, labID, id models.UUID, in models.SmartAnnotationUpdate) (models.SmartAnnotation, error)
}

// SmartAnnotationService provides smart annotation operations.
type SmartAnnotationService struct {
	apiClient *api.Client
}

func NewSmartAnnotationService(apiClient *api.Client) *SmartAnnotationService {
	return &SmartAnnotationService{apiClient: apiClient}
}

func smartAnnotationsURL(labID models.UUID) string {
	return fmt.Sprintf("%s/%s/%s", labsAPI, labID, smartAnnotationsAPI)
}

func smartAnnotationURL(labID, id models.UUID) string {
	return fmt.Sprintf("%s/%s", smartAnnotationsURL(labID), id)
}

func (s *SmartAnnotationService) List(ctx context.Context, labID models.UUID) ([]models.SmartAnnotation, error) {
	apiPath := smartAnnotationsURL(labID)
	var out []models.SmartAnnotation
	if err := s.apiClient.GetJSON(ctx, apiPath, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *SmartAnnotationService) Get(ctx context.Context, labID, id models.UUID) (models.SmartAnnotation, error) {
	apiPath := smartAnnotationURL(labID, id)
	var out models.SmartAnnotation
	if err := s.apiClient.GetJSON(ctx, apiPath, nil, &out); err != nil {
		return models.SmartAnnotation{}, err
	}
	return out, nil
}

func (s *SmartAnnotationService) Update(ctx context.Context, labID, id models.UUID, in models.SmartAnnotationUpdate) (models.SmartAnnotation, error) {
	apiPath := smartAnnotationURL(labID, id)
	var out models.SmartAnnotation
	if err := s.apiClient.PatchJSON(ctx, apiPath, nil, in, &out); err != nil {
		return models.SmartAnnotation{}, err
	}
	return out, nil
}
