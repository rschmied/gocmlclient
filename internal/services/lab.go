// Package services, lab specific
package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/pkg/models"
)

const (
	labAPI    = "labs"
	importAPI = "import"
)

// LabService provides lab-related operations
type LabService struct {
	apiClient *api.Client
	Interface InterfaceServiceInterface
	Node      NodeServiceInterface
	User      UserServiceInterface
}

// NewLabService creates a new lab service
func NewLabService(apiClient *api.Client) *LabService {
	return &LabService{
		apiClient: apiClient,
	}
}

type labAlias struct {
	models.Lab
	OwnerID string `json:"owner"`
}

// type labPatchPostAlias struct {
// 	Title       string `json:"title,omitempty"`
// 	Description string `json:"description,omitempty"`
// 	Notes       string `json:"notes,omitempty"`
// 	// Groups      LabGroupList `json:"groups,omitempty"`
// }

// API request/response types (internal to the service)
type labCreateRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

type labUpdateRequest struct {
	Title       string             `json:"title,omitempty"`
	Description string             `json:"description,omitempty"`
	Notes       string             `json:"notes,omitempty"`
	Groups      []*models.LabGroup `json:"groups,omitempty"`
}

type labResponse struct {
	ID          string             `json:"id"`
	State       string             `json:"state"`
	Created     string             `json:"created"`
	Modified    string             `json:"modified"`
	Title       string             `json:"lab_title"`
	Description string             `json:"lab_description"`
	Notes       string             `json:"lab_notes"`
	OwnerID     string             `json:"owner"`
	NodeCount   int                `json:"node_count"`
	LinkCount   int                `json:"link_count"`
	Groups      []*models.LabGroup `json:"groups"`
}

func (r *labResponse) toLab() *models.Lab {
	return &models.Lab{
		ID:          r.ID,
		State:       r.State,
		Created:     r.Created,
		Modified:    r.Modified,
		Title:       r.Title,
		Description: r.Description,
		Notes:       r.Notes,
		Owner:       &models.User{ID: r.OwnerID},
		NodeCount:   r.NodeCount,
		LinkCount:   r.LinkCount,
		Groups:      r.Groups,
		Nodes:       make(models.NodeMap),
		Links:       []*models.Link{},
	}
}

// Create creates a new lab on the controller
func (s *LabService) Create(ctx context.Context, lab *models.Lab) (*models.Lab, error) {
	// Use the alias type for API communication (handles inconsistent field names)
	postData := labCreateRequest{
		Title:       lab.Title,
		Description: lab.Description,
		Notes:       lab.Notes,
	}

	var response labResponse
	err := s.apiClient.PostJSON(ctx, "labs", nil, postData, &response)
	if err != nil {
		return nil, fmt.Errorf("create lab: %w", err)
	}

	// Update the lab with the ID from the response
	lab.ID = response.ID

	// Update with full data (handles groups, owner, etc.)
	return s.Update(ctx, lab)
}

// GetByID retrieves a lab by ID
func (s *LabService) GetByID(ctx context.Context, id string, deep bool) (*models.Lab, error) {
	api := fmt.Sprintf("%s/%s", labAPI, id)
	la := &labAlias{}
	err := s.apiClient.GetJSON(ctx, api, nil, la)
	if err != nil {
		return nil, err
	}
	if !deep {
		la.Owner = &models.User{ID: la.OwnerID}
		return &la.Lab, nil
	}
	return s.fillLabData(ctx, la)
}

// Update updates a lab's metadata
func (s *LabService) Update(ctx context.Context, lab *models.Lab) (*models.Lab, error) {
	endpoint := fmt.Sprintf("%s/%s", labAPI, lab.ID)

	updateData := labUpdateRequest{
		Title:       lab.Title,
		Description: lab.Description,
		Notes:       lab.Notes,
		Groups:      lab.Groups,
	}

	var response labResponse
	err := s.apiClient.PatchJSON(ctx, endpoint, updateData, &response)
	if err != nil {
		return nil, fmt.Errorf("update lab %s: %w", lab.ID, err)
	}

	return response.toLab(), nil
}

// Start starts all nodes in a lab
func (s *LabService) Start(ctx context.Context, id string) error {
	endpoint := fmt.Sprintf("%s/%s/start", labAPI, id)
	err := s.apiClient.PutJSON(ctx, endpoint, nil)
	if err != nil {
		return fmt.Errorf("start lab %s: %w", id, err)
	}
	return nil
}

// Stop stops all nodes in a lab
func (s *LabService) Stop(ctx context.Context, id string) error {
	endpoint := fmt.Sprintf("%s/%s/stop", labAPI, id)
	err := s.apiClient.PutJSON(ctx, endpoint, nil)
	if err != nil {
		return fmt.Errorf("stop lab %s: %w", id, err)
	}
	return nil
}

// Import imports a lab from YAML topology
func (s *LabService) Import(ctx context.Context, topology string) (*models.Lab, error) {
	topoReader := strings.NewReader(topology)

	var importResponse struct {
		ID       string   `json:"id"`
		Warnings []string `json:"warnings"`
	}

	err := s.apiClient.PostJSON(ctx, importAPI, nil, topoReader, &importResponse)
	if err != nil {
		return nil, fmt.Errorf("import lab: %w", err)
	}

	if len(importResponse.Warnings) > 0 {
		slog.Warn("Lab import completed with warnings", "warnings", importResponse.Warnings)
	}

	// Fetch the imported lab with full data
	return s.GetByID(ctx, importResponse.ID, true)
}

// HasConverged checks if all nodes in the lab have converged (are in BOOTED state)
func (s *LabService) HasConverged(ctx context.Context, id string) (bool, error) {
	endpoint := fmt.Sprintf("%s/%s/check_if_converged", labAPI, id)

	var converged bool
	err := s.apiClient.GetJSON(ctx, endpoint, nil, &converged)
	if err != nil {
		return false, fmt.Errorf("check lab convergence %s: %w", id, err)
	}

	return converged, nil
}

// fillLabData fetches additional lab data for deep queries
func (s *LabService) fillLabData(ctx context.Context, la *labAlias) (*models.Lab, error) {
	var err error
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		defer slog.Debug("user done")
		// retrieve the user by ID
		la.Owner, err = s.User.GetByID(ctx, la.OwnerID)
		if err != nil {
			return err
		}
		// FIXME: endpoint is deprecated...!?
		// // fill the groups the user is member of
		// groups, err := s.User.Groups(ctx, la.OwnerID)
		// if err != nil {
		// 	return err
		// }
		// for _, group := range groups {
		// 	la.Owner.Groups = append(la.Owner.Groups, group.ID)
		// }
		return nil
	})

	lab := &la.Lab

	// need to ensure that this block finishes before the others run
	// ch := make(chan struct{})
	g.Go(func() error {
		defer func() {
			slog.Debug("nodes/interfaces done")
			// two sync points, we can run the API endpoints but we need to
			// wait for the node data to be read until we can add the layer3
			// info (1) and the link info (2)
			// ch <- struct{}{}
			// ch <- struct{}{}
		}()
		slog.Warn("get nodes")
		err := s.Node.GetNodesForLab(ctx, lab)
		if err != nil {
			slog.Error("get nodes", "err", err)
			return err
		}
		slog.Warn("get interfaces")
		for _, node := range lab.Nodes {
			err = s.Interface.GetInterfacesForNode(ctx, node)
			if err != nil {
				slog.Error("get interfaces", "err", err)
				return err
			}
		}
		return nil
	})
	//
	// g.Go(func() error {
	// 	defer slog.Debug("l3info done")
	// 	l3info, err := c.getL3Info(ctx, lab.ID)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	slog.Debug("l3info read")
	// 	// wait for node data read complete
	// 	<-ch
	// 	// map and merge the l3 data...
	// 	for nid, l3data := range *l3info {
	// 		if node, found := lab.Nodes[nid]; found {
	// 			for mac, l3i := range l3data.Interfaces {
	// 				for _, iface := range node.Interfaces {
	// 					if iface.MACaddress == mac {
	// 						iface.IP4 = l3i.IP4
	// 						iface.IP6 = l3i.IP6
	// 						break
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// 	slog.Debug("l3info loop done")
	// 	return nil
	// })
	//
	// g.Go(func() error {
	// 	defer slog.Debug("links done")
	// 	idlist, err := c.getLinkIDsForLab(ctx, lab)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	slog.Debug("links read")
	// 	// wait for node data read complete
	// 	<-ch
	// 	return c.getLinksForLab(ctx, lab, idlist)
	// })

	if err := g.Wait(); err != nil {
		slog.Error("error", "err", err)
		return nil, err
	}
	slog.Debug("wait done")
	return lab, nil
}
