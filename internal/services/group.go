package services

import (
	"context"
	"fmt"
	"sort"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/pkg/models"
)

// GroupServiceInterface defines methods needed by other services
type GroupServiceInterface interface {
	GetByID(ctx context.Context, id string) (*models.Group, error)
}

const (
	groupAPI = "groups"
)

// GroupService provides group-related operations
type GroupService struct {
	apiClient *api.Client
	User      UserServiceInterface
}

// NewGroupService creates a new group service
func NewGroupService(apiClient *api.Client) *GroupService {
	return &GroupService{
		apiClient: apiClient,
	}
}

// Groups retrieves the list of all groups which exist on the controller.
func (s *GroupService) Groups(ctx context.Context) (models.GroupList, error) {
	groups := models.GroupList{}
	err := s.apiClient.GetJSON(ctx, groupAPI, nil, &groups)
	if err != nil {
		return nil, err
	}
	// sort the group list by their ID
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].ID > groups[j].ID
	})
	return groups, nil
}

// ByName tries to get the group with the provided `name`.
func (s *GroupService) ByName(ctx context.Context, name string) (*models.Group, error) {
	group := models.Group{}
	err := s.apiClient.GetJSON(ctx, fmt.Sprintf("%s/%s/id", groupAPI, name), nil, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetByID retrieves the group with the provided `id` (a UUIDv4).
func (s *GroupService) GetByID(ctx context.Context, id string) (*models.Group, error) {
	group := models.Group{}
	err := s.apiClient.GetJSON(ctx, fmt.Sprintf("%s/%s", groupAPI, id), nil, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// Delete removes the group identified by the `id` (a UUIDv4).
func (s *GroupService) Delete(ctx context.Context, id string) error {
	return s.apiClient.DeleteJSON(ctx, fmt.Sprintf("%s/%s", groupAPI, id), nil)
}

// Create creates a new group on the controller based on the data provided
// in the passed group parameter.
func (s *GroupService) Create(ctx context.Context, group *models.Group) (*models.Group, error) {
	result := models.Group{}
	err := s.apiClient.PostJSON(ctx, groupAPI, nil, group, &result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// Update updates the given group which must exist.
func (s *GroupService) Update(ctx context.Context, group *models.Group) (*models.Group, error) {
	groupID := group.ID
	group.ID = ""
	result := models.Group{}
	err := s.apiClient.PatchJSON(ctx, fmt.Sprintf("%s/%s", groupAPI, groupID), group, &result)
	if err != nil {
		return nil, err
	}
	return &result, err
}
