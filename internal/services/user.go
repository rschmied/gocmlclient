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
func NewUserService(apiClient *api.Client) *UserService {
	return &UserService{
		apiClient: apiClient,
	}
}

// GetByID returns the user with the given `id`.
func (s *UserService) GetByID(ctx context.Context, id models.UUID) (*models.User, error) {
	api := fmt.Sprintf("%s/%s", userAPI, id)
	user := &models.User{}
	err := s.apiClient.GetJSON(ctx, api, nil, user)
	if err != nil {
		return nil, err
	}
	return user, nil
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
func (s *UserService) Delete(ctx context.Context, id string) error {
	return s.apiClient.DeleteJSON(ctx, fmt.Sprintf("%s/%s", userAPI, id), nil)
}

// UserCreate creates a new user on the controller based on the data provided
// in the passed user parameter.
func (s *UserService) UserCreate(ctx context.Context, user *models.User) (*models.User, error) {
	result := models.User{}
	err := s.apiClient.PostJSON(ctx, userAPI, nil, user, &result)
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, result.ID)
}

// Update updates the given user which must exist.
func (s *UserService) Update(ctx context.Context, user models.User) (*models.User, error) {
	result := models.User{}
	userID := user.ID
	user.ID = "" // ensure no ID
	err := s.apiClient.PatchJSON(ctx, fmt.Sprintf("%s/%s", userAPI, userID), user, &result)
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, result.ID)
}
