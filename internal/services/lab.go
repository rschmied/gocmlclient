// Package services, lab specific
package services

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/httputil"
	"github.com/rschmied/gocmlclient/internal/logging"
	"github.com/rschmied/gocmlclient/pkg/errors"
	"github.com/rschmied/gocmlclient/pkg/models"
)

const (
	labsAPI      = "labs"
	importAPI    = "import"
	convergedAPI = "check_if_converged"
	populateAPI  = "populate_lab_tiles"
	wipeAction   = "wipe"
	startAction  = "start"
	stopAction   = "stop"
)

// Ensure LabService implements interface
var _ LabServiceInterface = (*LabService)(nil)

// LabServiceInterface defines methods needed by other services/clients.
type LabServiceInterface interface {
	Labs(ctx context.Context, showAll bool) (models.LabList, error)
	LabsWithData(ctx context.Context) ([]models.LabResponse, error)
	Create(ctx context.Context, lab models.LabCreateRequest) (models.Lab, error)
	GetByID(ctx context.Context, id models.UUID, deep bool) (models.Lab, error)
	GetByTitle(ctx context.Context, title string, deep bool) (models.Lab, error)
	Update(ctx context.Context, labID models.UUID, lab models.LabUpdateRequest) (models.Lab, error)
	Delete(ctx context.Context, id models.UUID) error
	Import(ctx context.Context, topologyYAML string) (models.Lab, error)
	Start(ctx context.Context, labID models.UUID) error
	Stop(ctx context.Context, labID models.UUID) error
	Wipe(ctx context.Context, labID models.UUID) error
	HasConverged(ctx context.Context, id models.UUID) (bool, error)
}

// LabService provides lab-related operations
type LabService struct {
	apiClient *api.Client
	User      UserServiceInterface
	Link      LinkServiceInterface
	Interface InterfaceServiceInterface
	Node      NodeServiceInterface
}

// NewLabService creates a new lab service
func NewLabService(apiClient *api.Client, iface InterfaceServiceInterface, link LinkServiceInterface, user UserServiceInterface, node NodeServiceInterface) *LabService {
	return &LabService{
		apiClient: apiClient,
		User:      user,
		Link:      link,
		Interface: iface,
		Node:      node,
	}
}

// labURL builds URL for a specific lab
func labURL(id models.UUID) string {
	return fmt.Sprintf("%s/%s", labsAPI, id)
}

// labActionURL builds URL for lab state operations
func labActionLegacyURL(labID models.UUID, action string) string {
	return fmt.Sprintf("%s/state/%s", labURL(labID), action)
}

func labActionURL(labID models.UUID, action string) string {
	return fmt.Sprintf("%s/%s", labURL(labID), action)
}

// Labs returns a list of labs.
func (s *LabService) Labs(ctx context.Context, showAll bool) (labs models.LabList, err error) {
	labs = models.LabList{}
	qb := httputil.NewQueryBuilder()
	if showAll {
		qb.Set("show_all", "true")
	}
	queryParams := qb.Build()
	err = s.apiClient.GetJSON(ctx, labsAPI, queryParams, &labs)
	return labs, err
}

// LabsWithData retrieves labs with data using the /populate_lab_tiles endpoint
func (s *LabService) LabsWithData(ctx context.Context) ([]models.LabResponse, error) {
	var labTilesResponse models.LabTilesResponse
	err := s.apiClient.GetJSON(ctx, populateAPI, nil, &labTilesResponse)
	if err != nil {
		return nil, errors.Wrap(err, "get lab tiles")
	}

	// Convert LabTiles to LabResponse format
	labs := make([]models.LabResponse, 0, len(labTilesResponse.LabTiles))
	for _, tile := range labTilesResponse.LabTiles {
		lab := models.LabResponse{
			ID:                   tile.ID,
			State:                tile.State,
			Created:              tile.Created,
			Modified:             tile.Modified,
			Title:                tile.Title,
			Description:          tile.Description,
			Notes:                tile.Notes,
			OwnerID:              tile.OwnerID,
			OwnerUsername:        tile.OwnerUsername,
			OwnerFullname:        tile.OwnerFullname,
			NodeCount:            tile.NodeCount,
			LinkCount:            tile.LinkCount,
			Groups:               tile.Groups,
			EffectivePermissions: tile.EffectivePermissions,
		}
		labs = append(labs, lab)
	}

	return labs, nil
}

// Create creates a new lab on the controller. Only certain fields from the
// full Lab model are accepted during creation. Use GetByID() to retrieve the
// complete lab object after successful creation.
func (s *LabService) Create(ctx context.Context, lab models.LabCreateRequest) (models.Lab, error) {
	var result models.Lab
	err := s.apiClient.PostJSON(ctx, labsAPI, nil, lab, &result)
	if err != nil {
		return models.Lab{}, errors.Wrap(err, "create lab")
	}
	// Update with full data (handles groups, owner, etc.)
	return s.Update(ctx, result.ID, models.LabUpdateRequest{})
}

// GetByID retrieves a lab by ID
func (s *LabService) GetByID(ctx context.Context, id models.UUID, deep bool) (models.Lab, error) {
	var result models.Lab
	err := s.apiClient.GetJSON(ctx, labURL(id), nil, &result)
	if err != nil {
		return models.Lab{}, errors.Wrapf(err, "get lab by ID %s", id)
	}

	// Set OwnerID from the API response (the "owner" field contains the UUID)
	// Note: The JSON unmarshaling will put the owner UUID into OwnerID field

	if deep {
		if err := s.fillLabData(ctx, &result); err != nil {
			return models.Lab{}, errors.Wrapf(err, "fill lab data for %s", id)
		}
	} else {
		// For shallow fetch, set basic owner info
		result.Owner = &models.User{
			UserBase: models.UserBase{
				Username: result.OwnerUsername,
				Fullname: result.OwnerFullname,
			},
			ID: result.OwnerID,
		}
	}

	return result, nil
}

// Update updates a lab's metadata
func (s *LabService) Update(ctx context.Context, id models.UUID, data models.LabUpdateRequest) (lab models.Lab, err error) {
	err = s.apiClient.PatchJSON(ctx, labURL(id), nil, data, &lab)
	return lab, err
}

// Start starts all nodes in a lab
func (s *LabService) Start(ctx context.Context, id models.UUID) error {
	if err := s.apiClient.PutJSON(ctx, labActionURL(id, startAction), nil); err != nil {
		if errors.IsNotFound(err) {
			return s.apiClient.PutJSON(ctx, labActionLegacyURL(id, startAction), nil)
		}
		return err
	}
	return nil
}

// Stop stops all nodes in a lab
func (s *LabService) Stop(ctx context.Context, id models.UUID) error {
	if err := s.apiClient.PutJSON(ctx, labActionURL(id, stopAction), nil); err != nil {
		if errors.IsNotFound(err) {
			return s.apiClient.PutJSON(ctx, labActionLegacyURL(id, stopAction), nil)
		}
		return err
	}
	return nil
}

// Delete deletes the lab identified by the `id` (a UUIDv4).
func (s *LabService) Delete(ctx context.Context, id models.UUID) error {
	return s.apiClient.DeleteJSON(ctx, labURL(id), nil)
}

// Wipe wipes the lab identified by the `id` (a UUIDv4).
func (s *LabService) Wipe(ctx context.Context, id models.UUID) error {
	if err := s.apiClient.PutJSON(ctx, labActionURL(id, wipeAction), nil); err != nil {
		if errors.IsNotFound(err) {
			return s.apiClient.PutJSON(ctx, labActionLegacyURL(id, wipeAction), nil)
		}
		return err
	}
	return nil
}

// Import imports a lab from YAML topology
func (s *LabService) Import(ctx context.Context, topology string) (models.Lab, error) {
	topoReader := strings.NewReader(topology)

	var importResponse struct {
		ID       models.UUID `json:"id"`
		Warnings []string    `json:"warnings"`
	}

	err := s.apiClient.PostJSON(ctx, importAPI, nil, topoReader, &importResponse)
	if err != nil {
		return models.Lab{}, errors.Wrap(err, "import lab")
	}

	if len(importResponse.Warnings) > 0 {
		logging.Warn("Lab import completed with warnings", "warnings", importResponse.Warnings)
	}

	// Fetch the imported lab with full data
	return s.GetByID(ctx, importResponse.ID, true)
}

// HasConverged checks if all nodes in the lab have converged (are in BOOTED state)
func (s *LabService) HasConverged(ctx context.Context, id models.UUID) (converged bool, err error) {
	err = s.apiClient.GetJSON(ctx, fmt.Sprintf("%s/%s", labURL(id), convergedAPI), nil, &converged)
	return converged, err
}

// fillLabData fetches additional lab data for deep queries
func (s *LabService) fillLabData(ctx context.Context, lab *models.Lab) error {
	g, gctx := errgroup.WithContext(ctx)

	// Fetch user concurrently (only if OwnerID is set)
	g.Go(func() error {
		if lab.OwnerID == "" {
			return nil // Skip if no owner ID
		}
		user, err := s.User.GetByID(gctx, lab.OwnerID)
		if err != nil {
			return errors.Wrapf(err, "get user %s for lab %s", lab.OwnerID, lab.ID)
		}
		lab.Owner = &user
		return nil
	})

	// Channel for synchronization
	nodeDataReady := make(chan struct{})

	// Fetch nodes concurrently
	g.Go(func() error {
		defer close(nodeDataReady) // Signal when node data is ready
		nodes, err := s.Node.GetNodesForLab(gctx, lab.ID)
		if err != nil {
			return errors.Wrapf(err, "get nodes for lab %s", lab.ID)
		}
		lab.Nodes = nodes
		return nil
	})

	// Fetch links concurrently (waits for node data)
	g.Go(func() error {
		<-nodeDataReady // Wait for node data to be ready
		links, err := s.Link.GetLinksForLab(gctx, lab.ID)
		if err != nil {
			return errors.Wrapf(err, "get links for lab %s", lab.ID)
		}
		lab.Links = links
		return nil
	})

	// Wait for concurrent operations
	if err := g.Wait(); err != nil {
		return err
	}

	// Fetch interfaces for each node (sequentially)
	for nodeID := range lab.Nodes {
		interfaces, err := s.Interface.GetInterfacesForNode(ctx, lab.ID, nodeID)
		if err != nil {
			return errors.Wrapf(err, "get interfaces for node %s", nodeID)
		}
		lab.Nodes[nodeID].Interfaces = interfaces
	}

	// Fetch and merge L3 information
	l3info, err := s.getL3Info(ctx, lab.ID)
	if err != nil {
		return errors.Wrapf(err, "get L3 info for lab %s", lab.ID)
	}

	// Merge L3 data into interfaces
	for nodeID, l3data := range *l3info {
		if node, found := lab.Nodes[models.UUID(nodeID)]; found {
			for mac, l3i := range l3data.Interfaces {
				for i := range node.Interfaces {
					if node.Interfaces[i].Operational != nil &&
						node.Interfaces[i].Operational.MACaddress != nil &&
						*node.Interfaces[i].Operational.MACaddress == mac {
						node.Interfaces[i].IP4 = l3i.IP4
						node.Interfaces[i].IP6 = l3i.IP6
						break
					}
				}
			}
			lab.Nodes[models.UUID(nodeID)] = node
		}
	}

	return nil
}

// GetByTitle returns the lab identified by its `title`.
func (s *LabService) GetByTitle(ctx context.Context, title string, deep bool) (models.Lab, error) {
	// Get all labs with data using the fast endpoint
	labs, err := s.LabsWithData(ctx)
	if err != nil {
		return models.Lab{}, errors.Wrap(err, "get labs with data")
	}

	// Find the lab with matching title
	for _, lab := range labs {
		if lab.Title == title {
			// Found it, now get full data
			return s.GetByID(ctx, lab.ID, deep)
		}
	}
	return models.Lab{}, errors.Wrapf(errors.ErrElementNotFound, "lab with title %q not found", title)
}
