// Package client provides compatibility functions for the CML client.
package client

import (
	"context"

	"github.com/rschmied/gocmlclient/pkg/models"
)

// LabGet retrieves a lab by ID with optional deep loading.
func (c *Client) LabGet(ctx context.Context, id string, deep bool) (*models.Lab, error) {
	lab, err := c.Lab.GetByID(ctx, models.UUID(id), deep)
	if err != nil {
		return nil, err
	}
	return &lab, nil
}

// LabStart starts a lab (exists in GitHub version)
func (c *Client) LabStart(ctx context.Context, id string) error {
	return c.Lab.Start(ctx, models.UUID(id))
}

// LabStop stops a lab (exists in GitHub version)
func (c *Client) LabStop(ctx context.Context, id string) error {
	return c.Lab.Stop(ctx, models.UUID(id))
}

// LabWipe wipes a lab (exists in GitHub version)
func (c *Client) LabWipe(ctx context.Context, id string) error {
	return c.Lab.Wipe(ctx, models.UUID(id))
}

// LabDestroy deletes a lab (exists in GitHub version)
func (c *Client) LabDestroy(ctx context.Context, id string) error {
	return c.Lab.Delete(ctx, models.UUID(id))
}

// LabImport imports a lab from YAML (exists in GitHub version)
func (c *Client) LabImport(ctx context.Context, topology string) (*models.Lab, error) {
	lab, err := c.Lab.Import(ctx, topology)
	if err != nil {
		return nil, err
	}
	return &lab, nil
}

// LabHasConverged checks if lab has converged (exists in GitHub version)
func (c *Client) LabHasConverged(ctx context.Context, id string) (bool, error) {
	return c.Lab.HasConverged(ctx, models.UUID(id))
}

// NodeGet retrieves a node (exists in GitHub version)
func (c *Client) NodeGet(ctx context.Context, labID, nodeID string) (*models.Node, error) {
	node, err := c.Node.GetByID(ctx, models.UUID(labID), models.UUID(nodeID))
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// NodeStart starts a node (exists in GitHub version)
func (c *Client) NodeStart(ctx context.Context, labID, nodeID string) error {
	return c.Node.Start(ctx, models.UUID(labID), models.UUID(nodeID))
}

// NodeStop stops a node (exists in GitHub version)
func (c *Client) NodeStop(ctx context.Context, labID, nodeID string) error {
	return c.Node.Stop(ctx, models.UUID(labID), models.UUID(nodeID))
}

// NodeWipe wipes a node (exists in GitHub version)
func (c *Client) NodeWipe(ctx context.Context, labID, nodeID string) error {
	return c.Node.Wipe(ctx, models.UUID(labID), models.UUID(nodeID))
}

// NodeDestroy deletes a node (exists in GitHub version)
func (c *Client) NodeDestroy(ctx context.Context, labID, nodeID string) error {
	return c.Node.Delete(ctx, models.UUID(labID), models.UUID(nodeID))
}

// Version returns the CML controller version (exists in GitHub version)
func (c *Client) Version() string {
	return c.System.Version()
}

// VersionCheck checks version compatibility (exists in GitHub version)
func (c *Client) VersionCheck(ctx context.Context, constraint string) (bool, error) {
	return c.System.VersionCheck(ctx, constraint)
}

// Ready checks if system is ready (exists in GitHub version)
func (c *Client) Ready(ctx context.Context) error {
	return c.System.Ready(ctx)
}

// UserGet retrieves a user by ID (exists in GitHub version)
func (c *Client) UserGet(ctx context.Context, id string) (*models.User, error) {
	user, err := c.User.GetByID(ctx, models.UUID(id))
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UserByName retrieves a user by username (exists in GitHub version)
func (c *Client) UserByName(ctx context.Context, username string) (*models.User, error) {
	user, err := c.User.GetByName(ctx, username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Users retrieves all users (exists in GitHub version)
func (c *Client) Users(ctx context.Context) ([]*models.User, error) {
	userList, err := c.User.Users(ctx)
	if err != nil {
		return nil, err
	}
	// Convert UserList to []*models.User
	users := make([]*models.User, len(userList))
	for i, user := range userList {
		users[i] = &user
	}
	return users, nil
}

// UserDestroy deletes a user (exists in GitHub version)
func (c *Client) UserDestroy(ctx context.Context, id string) error {
	return c.User.Delete(ctx, models.UUID(id))
}

// GroupGet retrieves a group by ID (exists in GitHub version)
func (c *Client) GroupGet(ctx context.Context, id string) (*models.Group, error) {
	group, err := c.Group.GetByID(ctx, models.UUID(id))
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// Groups retrieves all groups (exists in GitHub version)
func (c *Client) Groups(ctx context.Context) ([]*models.Group, error) {
	groups, err := c.Group.Groups(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*models.Group, len(groups))
	for i := range groups {
		result[i] = &groups[i]
	}
	return result, nil
}

// GroupByName retrieves a group by name (exists in GitHub version)
func (c *Client) GroupByName(ctx context.Context, name string) (*models.Group, error) {
	group, err := c.Group.ByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GroupDestroy deletes a group (exists in GitHub version)
func (c *Client) GroupDestroy(ctx context.Context, id string) error {
	return c.Group.Delete(ctx, id)
}

// LinkGet retrieves a link (exists in GitHub version)
func (c *Client) LinkGet(ctx context.Context, labID, linkID string) (*models.Link, error) {
	link, err := c.Link.GetByID(ctx, models.UUID(labID), models.UUID(linkID))
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// LinkDestroy deletes a link (exists in GitHub version)
func (c *Client) LinkDestroy(ctx context.Context, labID, linkID string) error {
	return c.Link.Delete(ctx, models.UUID(labID), models.UUID(linkID))
}

// InterfaceGet retrieves an interface (exists in GitHub version)
func (c *Client) InterfaceGet(ctx context.Context, labID, interfaceID string) (*models.Interface, error) {
	iface, err := c.Interface.GetByID(ctx, models.UUID(labID), models.UUID(interfaceID))
	if err != nil {
		return nil, err
	}
	return &iface, nil
}

// NodeUpdate updates the node specified by data in `node` (e.g. ID and LabID)
// with the other data provided. It returns the updated node.
func (c *Client) NodeUpdate(ctx context.Context, node *models.Node) (*models.Node, error) {
	updatedNode, err := c.Node.Update(ctx, *node)
	if err != nil {
		return nil, err
	}
	return &updatedNode, nil
}

// NodeCreate creates a new node on the controller based on the data provided
// in `node`. Label, node definition and image definition must be provided.
func (c *Client) NodeCreate(ctx context.Context, node *models.Node) (*models.Node, error) {
	createdNode, err := c.Node.Create(ctx, *node)
	if err != nil {
		return nil, err
	}
	return &createdNode, nil
}

// NodeSetConfig sets a configuration for the specified node. At least the `ID`
// of the node and the `labID` must be provided in `node`. The `node` instance
// will be updated with the current values for the node as provided by the
// controller.
func (c *Client) NodeSetConfig(ctx context.Context, node *models.Node, configuration string) error {
	return c.Node.SetConfig(ctx, node, configuration)
}

// NodeSetNamedConfigs sets a list of named configurations for the specified
// node. At least the `ID` of the node and the `labID` must be provided in
// `node`.
func (c *Client) NodeSetNamedConfigs(ctx context.Context, node *models.Node, configs []models.NodeConfig) error {
	return c.Node.SetNamedConfigs(ctx, node, configs)
}

// LabCreate creates a new lab on the controller.
func (c *Client) LabCreate(ctx context.Context, lab models.Lab) (*models.Lab, error) {
	createRequest := models.LabCreateRequest{
		Title:       lab.Title,
		Description: lab.Description,
		Notes:       lab.Notes,
	}
	createdLab, err := c.Lab.Create(ctx, createRequest)
	if err != nil {
		return nil, err
	}
	return &createdLab, nil
}

// LabUpdate updates specific fields of a lab (title, description and notes).
func (c *Client) LabUpdate(ctx context.Context, lab models.Lab) (*models.Lab, error) {
	updateRequest := models.LabUpdateRequest{
		Title:       lab.Title,
		Description: lab.Description,
		Notes:       lab.Notes,
	}
	updatedLab, err := c.Lab.Update(ctx, lab.ID, updateRequest)
	if err != nil {
		return nil, err
	}
	return &updatedLab, nil
}

// LabGetByTitle returns the lab identified by its `title`. For the use of
// `deep` see LabGet().
func (c *Client) LabGetByTitle(ctx context.Context, title string, deep bool) (*models.Lab, error) {
	lab, err := c.Lab.GetByTitle(ctx, title, deep)
	if err != nil {
		return nil, err
	}
	return &lab, nil
}

// UserCreate creates a new user on the controller based on the data provided
// in the passed user parameter.
func (c *Client) UserCreate(ctx context.Context, user *models.User) (*models.User, error) {
	createRequest := models.UserCreateRequest{
		UserBase: models.UserBase{
			Username:     user.Username,
			Fullname:     user.Fullname,
			Email:        user.Email,
			Description:  user.Description,
			IsAdmin:      user.IsAdmin,
			Groups:       user.Groups,
			ResourcePool: user.ResourcePool,
			OptIn:        user.OptIn,
		},
		Password: user.Password, // Use password from User struct for backward compatibility
	}
	createdUser, err := c.User.Create(ctx, createRequest)
	if err != nil {
		return nil, err
	}
	return &createdUser, nil
}

// UserUpdate updates the given user which must exist.
func (c *Client) UserUpdate(ctx context.Context, user *models.User) (*models.User, error) {
	var passwordUpdate *models.UpdatePassword
	if user.Password != "" {
		// If password is provided, create an update password request
		// Note: For security, the old password should be provided separately
		// For backward compatibility, we'll assume the provided password is the new one
		passwordUpdate = &models.UpdatePassword{
			New: user.Password,
			// Old password would need to be provided separately for security
		}
	}

	updateRequest := models.UserUpdateRequest{
		UserBase: models.UserBase{
			Username:     user.Username,
			Fullname:     user.Fullname,
			Email:        user.Email,
			Description:  user.Description,
			IsAdmin:      user.IsAdmin,
			Groups:       user.Groups,
			ResourcePool: user.ResourcePool,
			OptIn:        user.OptIn,
		},
		Password: passwordUpdate,
	}
	updatedUser, err := c.User.Update(ctx, user.ID, updateRequest)
	if err != nil {
		return nil, err
	}
	return &updatedUser, nil
}

// UserGroups retrieves the list of all groups the user belongs to.
func (c *Client) UserGroups(ctx context.Context, id string) ([]*models.Group, error) {
	groups, err := c.User.Groups(ctx, models.UUID(id))
	if err != nil {
		return nil, err
	}
	// Convert GroupList to []*models.Group
	result := make([]*models.Group, len(groups))
	for i := range groups {
		result[i] = &groups[i]
	}
	return result, nil
}

// UserUpdatePassword updates a user's password with old and new password verification.
func (c *Client) UserUpdatePassword(ctx context.Context, userID string, oldPassword, newPassword string) error {
	updateRequest := models.UserUpdateRequest{
		Password: &models.UpdatePassword{
			Old: oldPassword,
			New: newPassword,
		},
	}
	_, err := c.User.Update(ctx, models.UUID(userID), updateRequest)
	return err
}

// GroupCreate creates a new group on the controller based on the data provided
// in the passed group parameter.
func (c *Client) GroupCreate(ctx context.Context, group *models.Group) (*models.Group, error) {
	createdGroup, err := c.Group.Create(ctx, *group)
	if err != nil {
		return nil, err
	}
	return &createdGroup, nil
}

// GroupUpdate updates the given group which must exist.
func (c *Client) GroupUpdate(ctx context.Context, group *models.Group) (*models.Group, error) {
	updatedGroup, err := c.Group.Update(ctx, *group)
	if err != nil {
		return nil, err
	}
	return &updatedGroup, nil
}

// LinkCreate creates a link based on the data passed in `link`. Required
// fields are the `LabID` and either a pair of interfaces `SrcID` / `DstID` or
// a pair of nodes `SrcNode` / `DstNode`. With nodes it's also possible to
// provide specific slots in `SrcSlot` / `DstSlot` where the link should be
// created.
// If one or both of the provided slots aren't available, then new interfaces
// will be created. If interface creation fails or the provided Interface IDs
// can't be found, the API returns an error, otherwise the returned Link
// variable has the updated link data.
// Node: -1 for a slot means: use next free slot. Specific slots run from 0 to
// the maximum slot number -1 per the node definition of the node type.
func (c *Client) LinkCreate(ctx context.Context, link *models.Link) (*models.Link, error) {
	createdLink, err := c.Link.Create(ctx, *link)
	if err != nil {
		return nil, err
	}
	return &createdLink, nil
}

// InterfaceCreate creates an interface in the given lab and node.  If the slot
// is >= 0, the request creates all unallocated slots up to and including that
// slot. Conversely, if the slot is < 0 (e.g. -1), the next free slot is used.
func (c *Client) InterfaceCreate(ctx context.Context, labID, nodeID string, slot int) (*models.Interface, error) {
	createdInterface, err := c.Interface.Create(ctx, models.UUID(labID), models.UUID(nodeID), slot)
	if err != nil {
		return nil, err
	}
	return &createdInterface, nil
}

// UseNamedConfigs turns on the use of named configs (only with 2.7.0 and
// newer)
func (c *Client) UseNamedConfigs() {
	c.System.UseNamedConfigs()
}
