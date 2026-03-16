// Package services, annotation specific
package services

import (
	"context"
	"fmt"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/pkg/models"
)

const annotationsAPI = "annotations"

// Ensure AnnotationService implements interface
var _ AnnotationServiceInterface = (*AnnotationService)(nil)

// AnnotationServiceInterface defines methods needed by other services/clients.
type AnnotationServiceInterface interface {
	List(ctx context.Context, labID models.UUID) ([]models.Annotation, error)
	Create(ctx context.Context, labID models.UUID, in models.AnnotationCreate) (models.Annotation, error)
	Get(ctx context.Context, labID, annotationID models.UUID) (models.Annotation, error)
	Update(ctx context.Context, labID, annotationID models.UUID, in models.AnnotationUpdate) (models.Annotation, error)
	Delete(ctx context.Context, labID, annotationID models.UUID) error
}

// AnnotationService provides classic annotation operations.
type AnnotationService struct {
	apiClient *api.Client
}

func NewAnnotationService(apiClient *api.Client) *AnnotationService {
	return &AnnotationService{apiClient: apiClient}
}

func annotationsURL(labID models.UUID) string {
	return fmt.Sprintf("%s/%s/%s", labsAPI, labID, annotationsAPI)
}

func annotationURL(labID, annotationID models.UUID) string {
	return fmt.Sprintf("%s/%s", annotationsURL(labID), annotationID)
}

func (s *AnnotationService) List(ctx context.Context, labID models.UUID) ([]models.Annotation, error) {
	apiPath := annotationsURL(labID)
	var out []models.Annotation
	if err := s.apiClient.GetJSON(ctx, apiPath, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AnnotationService) Create(ctx context.Context, labID models.UUID, in models.AnnotationCreate) (models.Annotation, error) {
	apiPath := annotationsURL(labID)
	var out models.Annotation
	if err := s.apiClient.PostJSON(ctx, apiPath, nil, in, &out); err != nil {
		return models.Annotation{}, err
	}
	return out, nil
}

func (s *AnnotationService) Get(ctx context.Context, labID, annotationID models.UUID) (models.Annotation, error) {
	apiPath := annotationURL(labID, annotationID)
	var out models.Annotation
	if err := s.apiClient.GetJSON(ctx, apiPath, nil, &out); err != nil {
		return models.Annotation{}, err
	}
	return out, nil
}

func (s *AnnotationService) Update(ctx context.Context, labID, annotationID models.UUID, in models.AnnotationUpdate) (models.Annotation, error) {
	apiPath := annotationURL(labID, annotationID)
	var out models.Annotation
	if err := s.apiClient.PatchJSON(ctx, apiPath, nil, in, &out); err != nil {
		return models.Annotation{}, err
	}
	return out, nil
}

func (s *AnnotationService) Delete(ctx context.Context, labID, annotationID models.UUID) error {
	apiPath := annotationURL(labID, annotationID)
	return s.apiClient.DeleteJSON(ctx, apiPath, nil)
}
