// Package services, system specific
package services

import (
	"context"
	"fmt"
	"sort"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/pkg/models"
)

const userAPI string = "users"

// Ensure UserService implements interface
var _ UserServiceInterface = (*UserService)(nil)

// UserServiceInterface defines methods needed by other services
type UserServiceInterface interface {
	GetByID(ctx context.Context, id models.UUID) (*models.User, error)
}

// UserService provides user-related operations
type UserService struct {
	apiClient *api.Client
	Group     GroupServiceInterface
}

// NewUserService creates a new user service
func NewUserService(apiClient *api.Client, group GroupServiceInterface) *UserService {
	return &UserService{
		apiClient: apiClient,
		Group:     group,
	}
}

// GetByID returns the user with the given `id`.
func (s *UserService) GetByID(ctx context.Context, id models.UUID) (user *models.User, err error) {
	api := fmt.Sprintf("%s/%s", userAPI, id)
	user = &models.User{}
	err = s.apiClient.GetJSON(ctx, api, nil, user)
	return user, err
}

// GetByName returns the user with the given username `name`.
func (s *UserService) GetByName(ctx context.Context, name string) (*models.User, error) {
	api := fmt.Sprintf("%s/%s/id", userAPI, name)
	var userID models.UUID
	err := s.apiClient.GetJSON(ctx, api, nil, &userID)
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, userID)
}

// Users retrieves the list of all users which exist on the controller.
func (s *UserService) Users(ctx context.Context) (models.UserList, error) {
	users := models.UserList{}
	err := s.apiClient.GetJSON(ctx, userAPI, nil, &users)
	if err != nil {
		return nil, err
	}
	// sort the user list by their ID
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID > users[j].ID
	})
	return users, nil
}

// Delete removes the user identified by the `id` (a UUIDv4).
func (s *UserService) Delete(ctx context.Context, id models.UUID) error {
	return s.apiClient.DeleteJSON(ctx, fmt.Sprintf("%s/%s", userAPI, id), nil)
}

// Create creates a new user on the controller based on the data provided in
// the passed user parameter.
func (s *UserService) Create(ctx context.Context, user models.UserCreateRequest) (*models.User, error) {
	result := models.User{}
	err := s.apiClient.PostJSON(ctx, userAPI, nil, user, &result)
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, result.ID)
}

// Update updates the given user which must exist.
func (s *UserService) Update(ctx context.Context, id models.UUID, user models.UserUpdateRequest) (*models.User, error) {
	result := models.User{}
	err := s.apiClient.PatchJSON(ctx, fmt.Sprintf("%s/%s", userAPI, id), nil, user, &result)
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, result.ID)
}

// Groups retrieves the list of all groups the user belongs to.
func (s *UserService) Groups(ctx context.Context, id models.UUID) (models.GroupList, error) {
	api := fmt.Sprintf("users/%s/groups", id)
	idList := []models.UUID{}
	err := s.apiClient.GetJSON(ctx, api, nil, &idList)
	if err != nil {
		return nil, err
	}

	groups := models.GroupList{}
	for _, gid := range idList {
		group, err := s.Group.GetByID(ctx, gid)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}

	// sort the group list by their ID
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].ID > groups[j].ID
	})
	return groups, nil
}
