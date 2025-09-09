// Package services, lab specific
package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

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
	Link      LinkServiceInterface
	Node      NodeServiceInterface
	User      UserServiceInterface
}

// NewLabService creates a new lab service
func NewLabService(apiClient *api.Client, iface InterfaceServiceInterface, link LinkServiceInterface, user UserServiceInterface, node NodeServiceInterface) *LabService {
	return &LabService{
		apiClient: apiClient,
	}
}

// Create creates a new lab on the controller. Only certain fields from the
// full Lab model are accepted during creation. Use GetByID() to retrieve the
// complete lab object after successful creation.
func (s *LabService) Create(ctx context.Context, lab models.LabCreateRequest) (*models.Lab, error) {
	result := &models.Lab{}
	err := s.apiClient.PostJSON(ctx, "labs", nil, lab, result)
	if err != nil {
		return nil, fmt.Errorf("create lab: %w", err)
	}
	// Update with full data (handles groups, owner, etc.)
	return s.Update(ctx, *result)
}

// GetByID retrieves a lab by ID
func (s *LabService) GetByID(ctx context.Context, id models.UUID, deep bool) (*models.Lab, error) {
	api := fmt.Sprintf("%s/%s", labAPI, id)
	result := &models.Lab{}
	err := s.apiClient.GetJSON(ctx, api, nil, result)
	if err != nil {
		return nil, err
	}
	_ = deep
	// if !deep {
	// 	// la.Owner = &models.User{ID: la.OwnerID}
	// 	return &result, nil
	// }
	// return s.fillLabData(ctx, la)
	return result, nil
}

// Update updates a lab's metadata
func (s *LabService) Update(ctx context.Context, lab models.Lab) (*models.Lab, error) {
	updateData := models.LabCreateRequest{
		Title:       lab.Title,
		Description: lab.Description,
		Notes:       lab.Notes,
		Owner:       lab.Owner,
		// Associations: lab.EffectivePermissions,
	}

	// var result labResponse
	result := &models.Lab{}
	labID := lab.ID
	lab.ID = "" // ensure no ID
	endpoint := fmt.Sprintf("%s/%s", labAPI, labID)
	err := s.apiClient.PatchJSON(ctx, endpoint, updateData, result)
	if err != nil {
		// return nil, fmt.Errorf("update lab %s: %w", labID, err)
		return nil, err
	}

	return result, nil
	// return response.toLab(), nil
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
		ID       models.UUID `json:"id"`
		Warnings []string    `json:"warnings"`
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

// Wipe wipes the lab identified by the `id` (a UUIDv4).
func (s *LabService) Wipe(ctx context.Context, id models.UUID) error {
	return s.apiClient.PutJSON(ctx, fmt.Sprintf("labs/%s/wipe", id), nil)
}

// Delete deletes the lab identified by the `id` (a UUIDv4).
func (s *LabService) Delete(ctx context.Context, id models.UUID) error {
	return s.apiClient.DeleteJSON(ctx, fmt.Sprintf("labs/%s", id), nil)
}

// HasConverged checks if all nodes in the lab have converged (are in BOOTED state)
func (s *LabService) HasConverged(ctx context.Context, id models.UUID) (bool, error) {
	endpoint := fmt.Sprintf("%s/%s/check_if_converged", labAPI, id)

	var converged bool
	err := s.apiClient.GetJSON(ctx, endpoint, nil, &converged)
	if err != nil {
		return false, fmt.Errorf("check lab convergence %s: %w", id, err)
	}

	return converged, nil
}

// fillLabData fetches additional lab data for deep queries
// func (s *LabService) fillLabData(ctx context.Context, la *labAlias) (*models.Lab, error) {
// 	var err error
// 	g, ctx := errgroup.WithContext(ctx)
//
// 	g.Go(func() error {
// 		defer slog.Debug("user done")
// 		// retrieve the user by ID
// 		la.Owner, err = s.User.GetByID(ctx, la.OwnerID)
// 		if err != nil {
// 			return err
// 		}
// 		// FIXME: endpoint is deprecated...!?
// 		// // fill the groups the user is member of
// 		// groups, err := s.User.Groups(ctx, la.OwnerID)
// 		// if err != nil {
// 		// 	return err
// 		// }
// 		// for _, group := range groups {
// 		// 	la.Owner.Groups = append(la.Owner.Groups, group.ID)
// 		// }
// 		return nil
// 	})
//
// 	lab := &la.Lab
//
// 	// need to ensure that this block finishes before the others run
// 	ch := make(chan struct{})
// 	g.Go(func() error {
// 		defer func() {
// 			slog.Debug("nodes/interfaces done")
// 			// two sync points, we can run the API endpoints but we need to
// 			// wait for the node data to be read until we can add the layer3
// 			// info (1) and the link info (2)
// 			ch <- struct{}{}
// 			ch <- struct{}{}
// 		}()
// 		slog.Warn("get nodes")
// 		err := s.Node.GetNodesForLab(ctx, lab)
// 		if err != nil {
// 			slog.Error("get nodes", "err", err)
// 			return err
// 		}
// 		slog.Warn("get interfaces")
// 		for _, node := range lab.Nodes {
// 			ifaceList, err := s.Interface.GetInterfacesForNode(ctx, lab.ID, node.ID)
// 			if err != nil {
// 				slog.Error("get interfaces", "err", err)
// 				return err
// 			}
// 			node.Interfaces = ifaceList
// 		}
// 		return nil
// 	})
//
// 	g.Go(func() error {
// 		defer slog.Debug("l3info done")
// 		l3info, err := s.getL3Info(ctx, lab.ID)
// 		if err != nil {
// 			return err
// 		}
// 		slog.Debug("l3info read")
// 		// wait for node data read complete
// 		<-ch
// 		// map and merge the l3 data...
// 		for nid, l3data := range *l3info {
// 			if node, found := lab.Nodes[nid]; found {
// 				for mac, l3i := range l3data.Interfaces {
// 					for _, iface := range node.Interfaces {
// 						if iface.MACaddress == mac {
// 							iface.IP4 = l3i.IP4
// 							iface.IP6 = l3i.IP6
// 							break
// 						}
// 					}
// 				}
// 			}
// 		}
// 		slog.Debug("l3info loop done")
// 		return nil
// 	})
//
// 	g.Go(func() error {
// 		defer slog.Debug("links done")
// 		// wait for node data read complete
// 		<-ch
// 		linkList, err := s.Link.GetLinksForLab(ctx, lab)
// 		if err != nil {
// 			return err
// 		}
// 		lab.Links = linkList
// 		return nil
// 	})
//
// 	if err := g.Wait(); err != nil {
// 		slog.Error("error", "err", err)
// 		return nil, err
// 	}
// 	slog.Debug("wait done")
// 	return lab, nil
// }

type l3nodes map[string]*l3node

type l3node struct {
	Name       string                 `json:"name"`
	Interfaces map[string]l3interface `json:"interfaces"`
}

type l3interface struct {
	ID    string   `json:"id"`
	Label string   `json:"label"`
	IP4   []string `json:"ip4"`
	IP6   []string `json:"ip6"`
}

func (s *LabService) getL3Info(ctx context.Context, id string) (*l3nodes, error) {
	api := fmt.Sprintf("labs/%s/layer3_addresses", id)
	l3n := &l3nodes{}
	err := s.apiClient.GetJSON(ctx, api, nil, l3n)
	if err != nil {
		return nil, err
	}
	return l3n, nil
}
