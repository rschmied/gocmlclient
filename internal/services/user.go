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
	GetByID(ctx context.Context, id string) (*models.User, error)
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

type userPatchPostAlias struct {
	Username     string   `json:"username"`
	Password     string   `json:"password,omitempty"`
	Fullname     string   `json:"fullname"`
	Email        string   `json:"email"`
	Description  string   `json:"description"`
	IsAdmin      bool     `json:"admin"`
	Groups       []string `json:"groups"`
	Labs         []string `json:"labs,omitempty"`          // can't be set
	OptIn        bool     `json:"opt_in"`                  // with 2.5.0
	ResourcePool *string  `json:"resource_pool,omitempty"` // with 2.5.0
}

func newUserAlias(user *models.User) userPatchPostAlias {
	upp := userPatchPostAlias{}

	upp.Username = user.Username
	upp.Password = user.Password
	upp.Fullname = user.Fullname
	upp.Email = user.Email
	upp.Description = user.Description
	upp.IsAdmin = user.IsAdmin
	upp.OptIn = user.OptIn
	upp.Groups = user.Groups
	upp.Labs = user.Labs
	upp.ResourcePool = user.ResourcePool

	return upp
}

// GetByID returns the user with the given `id`.
func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error) {
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
	var userID string
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
func (s *UserService) Update(ctx context.Context, user *models.User) (*models.User, error) {
	patchAlias := newUserAlias(user)
	result := models.User{}
	err := s.apiClient.PatchJSON(ctx, fmt.Sprintf("%s/%s", userAPI, user.ID), patchAlias, &result)
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, result.ID)
}

// Groups retrieves the list of all groups the user belongs to.
// THIS ENDPOINT IS DEPRECATED
// func (s *UserService) Groups(ctx context.Context, id string) (models.GroupList, error) {
// 	api := fmt.Sprintf("%s/%s/groups", userAPI, id)
// 	idList := []string{}
// 	err := s.apiClient.GetJSON(ctx, api, nil, &idList)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	groups := models.GroupList{}
// 	for _, id := range idList {
// 		group, err := s.Group.GetByID(ctx, id)
// 		if err != nil {
// 			return nil, err
// 		}
// 		groups = append(groups, group)
// 	}
// 	// sort the user list by their ID
// 	// (groups are a set so sorting is only done for test stability)
// 	sort.Slice(groups, func(i, j int) bool {
// 		return groups[i].ID > groups[j].ID
// 	})
// 	return groups, nil
// }
