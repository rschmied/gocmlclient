package services

import (
	"context"
	"fmt"
	"sort"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/pkg/models"
)

const groupAPI = "groups"

// Ensure GroupService implements interface
var _ GroupServiceInterface = (*GroupService)(nil)

// GroupServiceInterface defines methods needed by other services.
type GroupServiceInterface interface {
	GetByID(ctx context.Context, id models.UUID) (*models.Group, error)
}

// GroupService provides group-related operations.
type GroupService struct {
	apiClient *api.Client
}

// NewGroupService creates a new group service.
func NewGroupService(apiClient *api.Client) *GroupService {
	return &GroupService{
		apiClient: apiClient,
	}
}

// Groups retrieves the list of all groups which exist on the controller.
func (s *GroupService) Groups(ctx context.Context) (models.GroupList, error) {
	associations := models.GroupList{}
	err := s.apiClient.GetJSON(ctx, groupAPI, nil, &associations)
	if err != nil {
		return nil, err
	}
	// sort the group list by their ID
	sort.Slice(associations, func(i, j int) bool {
		return associations[i].ID > associations[j].ID
	})
	return associations, nil
}

// ByName tries to get the group with the provided `name`.
func (s *GroupService) ByName(ctx context.Context, name string) (group *models.Group, err error) {
	group = &models.Group{}
	err = s.apiClient.GetJSON(ctx, fmt.Sprintf("%s/%s/id", groupAPI, name), nil, group)
	return group, err
}

// GetByID retrieves the group with the provided `id` (a UUIDv4).
func (s *GroupService) GetByID(ctx context.Context, id models.UUID) (group *models.Group, err error) {
	group = &models.Group{}
	err = s.apiClient.GetJSON(ctx, fmt.Sprintf("%s/%s", groupAPI, id), nil, group)
	return group, err
}

// Delete removes the group identified by the `id` (a UUIDv4).
func (s *GroupService) Delete(ctx context.Context, id string) error {
	return s.apiClient.DeleteJSON(ctx, fmt.Sprintf("%s/%s", groupAPI, id), nil)
}

// Create creates a new group on the controller based on the data provided
// in the passed group parameter.
func (s *GroupService) Create(ctx context.Context, group *models.Group) (result *models.Group, err error) {
	group.ID = "" // ensure no ID
	result = &models.Group{}
	err = s.apiClient.PostJSON(ctx, groupAPI, nil, group, result)
	return result, err
}

// Update updates the given group which must exist.
func (s *GroupService) Update(ctx context.Context, group *models.Group) (result *models.Group, err error) {
	groupID := group.ID
	group.ID = "" // ensure no ID
	result = &models.Group{}
	err = s.apiClient.PatchJSON(ctx, fmt.Sprintf("%s/%s", groupAPI, groupID), group, result)
	return result, err
}
