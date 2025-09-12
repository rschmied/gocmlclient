// Package client provides compatibility functions for the CML client.
package client

import (
	"context"

	"github.com/rschmied/gocmlclient/pkg/models"
)

// LabGet retrieves a lab by ID with optional deep loading.
func (c *Client) LabGet(ctx context.Context, id string, deep bool) (*models.Lab, error) {
	return c.Lab.GetByID(ctx, models.UUID(id), deep)
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
	return c.Lab.Import(ctx, topology)
}

// LabHasConverged checks if lab has converged (exists in GitHub version)
func (c *Client) LabHasConverged(ctx context.Context, id string) (bool, error) {
	return c.Lab.HasConverged(ctx, models.UUID(id))
}

// NodeGet retrieves a node (exists in GitHub version)
func (c *Client) NodeGet(ctx context.Context, labID, nodeID string) (*models.Node, error) {
	return c.Node.GetByID(ctx, models.UUID(labID), models.UUID(nodeID))
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
	return c.User.GetByID(ctx, models.UUID(id))
}

// UserByName retrieves a user by username (exists in GitHub version)
func (c *Client) UserByName(ctx context.Context, username string) (*models.User, error) {
	return c.User.GetByName(ctx, username)
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
	return c.Group.GetByID(ctx, models.UUID(id))
}

// Groups retrieves all groups (exists in GitHub version)
func (c *Client) Groups(ctx context.Context) ([]*models.Group, error) {
	return c.Group.Groups(ctx)
}

// GroupByName retrieves a group by name (exists in GitHub version)
func (c *Client) GroupByName(ctx context.Context, name string) (*models.Group, error) {
	return c.Group.ByName(ctx, name)
}

// GroupDestroy deletes a group (exists in GitHub version)
func (c *Client) GroupDestroy(ctx context.Context, id string) error {
	return c.Group.Delete(ctx, id)
}

// LinkGet retrieves a link (exists in GitHub version)
func (c *Client) LinkGet(ctx context.Context, labID, linkID string) (*models.Link, error) {
	return c.Link.GetByID(ctx, models.UUID(labID), models.UUID(linkID))
}

// LinkDestroy deletes a link (exists in GitHub version)
func (c *Client) LinkDestroy(ctx context.Context, labID, linkID string) error {
	return c.Link.Delete(ctx, models.UUID(labID), models.UUID(linkID))
}

// InterfaceGet retrieves an interface (exists in GitHub version)
func (c *Client) InterfaceGet(ctx context.Context, labID, interfaceID string) (*models.Interface, error) {
	return c.Interface.GetByID(ctx, models.UUID(labID), models.UUID(interfaceID))
}
