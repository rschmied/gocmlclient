package services

import (
	"context"
	"fmt"
	"sort"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/httputil"
	"github.com/rschmied/gocmlclient/pkg/models"
)

// Ensure InterfaceService implements interface
var _ InterfaceServiceInterface = (*InterfaceService)(nil)

// InterfaceServiceInterface defines methods needed by other services
type InterfaceServiceInterface interface {
	Create(ctx context.Context, labID, nodeID models.UUID, slot int) (models.Interface, error)
	GetByID(ctx context.Context, labID, id models.UUID) (models.Interface, error)
	GetInterfacesForNode(ctx context.Context, labID, id models.UUID) (models.InterfaceList, error)
}

// InterfaceService provides interface-related operations
type InterfaceService struct {
	apiClient *api.Client
}

// NewInterfaceService creates a new lab service
func NewInterfaceService(apiClient *api.Client) *InterfaceService {
	return &InterfaceService{
		apiClient: apiClient,
	}
}

// GetInterfacesForNode returns all interfaces for a specific node.
func (s *InterfaceService) GetInterfacesForNode(ctx context.Context, labID, id models.UUID) (models.InterfaceList, error) {
	// with the data=true option, we get not only the list of IDs but the
	// interfaces themselves as well!
	api := fmt.Sprintf("labs/%s/nodes/%s/interfaces", labID, id)
	queryParams := httputil.NewQueryBuilder().WithOperational().WithData(true).Build()

	interfaceList := models.InterfaceList{}
	err := s.apiClient.GetJSON(ctx, api, queryParams, &interfaceList)
	if err != nil {
		return nil, err
	}

	// sort the interface list by slot
	sort.Slice(interfaceList, func(i, j int) bool {
		return sortInterfacesBySlot(i, j, interfaceList)
	})
	return interfaceList, nil
}

// sortInterfacesBySlot sorts interfaces by their slot number, with nil slots coming first
func sortInterfacesBySlot(i, j int, interfaceList models.InterfaceList) bool {
	if interfaceList[i].Slot == nil && interfaceList[j].Slot == nil {
		return false
	}
	if interfaceList[i].Slot == nil {
		return true
	}
	if interfaceList[j].Slot == nil {
		return false
	}
	return *interfaceList[i].Slot < *interfaceList[j].Slot
}

// GetByID returns the interface identified by its `ID` (iface.ID).
func (s *InterfaceService) GetByID(ctx context.Context, labID, id models.UUID) (models.Interface, error) {
	api := fmt.Sprintf("labs/%s/interfaces/%s", labID, id)
	var iface models.Interface
	queryParams := httputil.NewQueryBuilder().
		WithOperational().
		Build()
	err := s.apiClient.GetJSON(ctx, api, queryParams, &iface)
	return iface, err
}

// Create creates an interface in the given lab and node.  If the slot is >= 0,
// the request creates all unallocated slots up to and including that slot.
// Conversely, if the slot is < 0 (e.g. -1), the next free slot is used.
func (s *InterfaceService) Create(ctx context.Context, labID, nodeID models.UUID, slot int) (models.Interface, error) {
	var slotPtr *int

	if slot >= 0 {
		slotPtr = &slot
	}

	newIface := struct {
		Node models.UUID `json:"node"`
		Slot *int        `json:"slot,omitempty"`
	}{
		Node: nodeID,
		Slot: slotPtr,
	}

	// This is quite awkward, not even sure if it's a good REST design practice:
	// "Returns a JSON object that identifies the interface that was created. In
	// the case of bulk interface creation, returns a JSON array of such
	// objects." <-- from the API documentation
	//
	// A list is returned when slot is defined, even if it's just creating one
	// interface

	api := fmt.Sprintf("labs/%s/interfaces", labID)
	if slotPtr == nil {
		var result models.Interface
		err := s.apiClient.PostJSON(ctx, api, nil, newIface, &result)
		if err != nil {
			return models.Interface{}, err
		}
		return result, err
	}

	// this is when a slot has been provided; the API provides now a list of
	// interfaces
	result := []models.Interface{}
	err := s.apiClient.PostJSON(ctx, api, nil, newIface, &result)
	if err != nil {
		return models.Interface{}, err
	}

	lastIface := result[len(result)-1]
	return lastIface, nil
}
