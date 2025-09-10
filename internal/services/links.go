package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/pkg/models"
)

// Ensure LinkService implements interface
var _ LinkServiceInterface = (*LinkService)(nil)

// LinkServiceInterface defines methods needed by other services
type LinkServiceInterface interface {
	// GetLinksForLab(ctx context.Context, lab *models.Lab) error
	GetLinksForLab(ctx context.Context, lab *models.Lab) ([]*models.Link, error)
	GetCondition(ctx context.Context, labID, linkID models.UUID) (*models.ConditionResponse, error)
	SetCondition(ctx context.Context, labID, linkID models.UUID, config *models.LinkConditionConfiguration) (*models.ConditionResponse, error)
	DeleteCondition(ctx context.Context, labID, linkID models.UUID) error
}

// LinkService provides link-related operations
type LinkService struct {
	apiClient *api.Client
	Interface InterfaceServiceInterface
	Node      NodeServiceInterface
}

// NewLinkService creates a new node service
func NewLinkService(apiClient *api.Client) *LinkService {
	return &LinkService{
		apiClient: apiClient,
	}
}

type linkList []*models.Link

func (llist linkList) MarshalJSON() ([]byte, error) {
	type slist linkList
	newlist := slist(llist)
	// we want this as a stable sort by link UUID
	sort.Slice(newlist, func(i, j int) bool {
		return newlist[i].ID < newlist[j].ID
	})
	return json.Marshal(newlist)
}

func (s *LinkService) GetLinksForLab(ctx context.Context, lab *models.Lab) ([]*models.Link, error) {
	api := fmt.Sprintf("labs/%s/links", lab.ID)

	queryParm := map[string]string{
		"data": "true",
	}

	linkList := []*models.Link{}
	err := s.apiClient.GetJSON(ctx, api, queryParm, &linkList)
	if err != nil {
		return nil, err
	}
	return linkList, nil
}

// GetByID returns the link data for the given `labID` and `linkID`. If `deep`
// is set to `true` then bot interface and node data for the given link are
// also fetched from the controller.
func (s *LinkService) GetByID(ctx context.Context, labID, linkID models.UUID) (*models.Link, error) {
	api := fmt.Sprintf("labs/%s/links/%s", labID, linkID)
	link := &models.Link{}
	err := s.apiClient.GetJSON(ctx, api, nil, link)
	if err != nil {
		return nil, err
	}
	if link.LabID != labID {
		panic("no lab ID")
	}
	return link, nil

	// link.LabID = labID

	// if deep {
	// 	var err error
	//
	// 	ifaceA := &internalInterface{}
	// 	ifaceB := &internalInterface{}
	//
	// 	// ifaceA := &internalInterface{
	// 	// 	ID:        link.SrcID,
	// 	// 	LabID:     labID,
	// 	// 	Node:      link.SrcNode,
	// 	// 	node:      &models.Node{ID: link.SrcNode, LabID: labID},
	// 	// 	Interface: models.Interface{},
	// 	// }
	// 	ifaceA.Interface, err = s.Interface.GetByID(ctx, labID, link.SrcID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	ifaceA.node, err = s.Node.GetByID(ctx, labID, link.SrcID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	ifaceB.Interface, err = s.Interface.GetByID(ctx, labID, link.DstID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	ifaceB.node, err = s.Node.GetByID(ctx, labID, link.DstID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	//
	// 	link.ifaceA = ifaceA
	// 	link.ifaceB = ifaceB
	// 	link.SrcSlot = ifaceA.Slot
	// 	link.DstSlot = ifaceB.Slot
	// }
	// return link, err
}

// Create creates a link based on the data passed in `link`. Required
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
func (s *LinkService) Create(ctx context.Context, link *models.Link) (*models.Link, error) {
	api := fmt.Sprintf("labs/%s/links", link.LabID)

	var err error

	if len(link.SrcNode) > 0 && len(link.DstNode) > 0 {

		ifaceListA, err := s.Interface.GetInterfacesForNode(ctx, link.LabID, link.SrcNode)
		if err != nil {
			return nil, err
		}

		ifaceListB, err := s.Interface.GetInterfacesForNode(ctx, link.LabID, link.DstNode)
		if err != nil {
			return nil, err
		}

		matches := func(slot int, iface *models.Interface) bool {
			return iface.IsPhysical() && !iface.IsConnected && (slot < 0 || (iface.Slot != nil && *iface.Slot == slot))
		}

		for _, iface := range ifaceListA {
			if matches(link.SrcSlot, iface) {
				iface.IsConnected = true
				link.SrcID = iface.ID
				break
			}
		}

		for _, iface := range ifaceListB {
			if matches(link.DstSlot, iface) {
				iface.IsConnected = true
				link.DstID = iface.ID
				break
			}
		}

		if len(link.SrcID) == 0 {
			iface, err := s.Interface.Create(ctx, link.LabID, link.SrcNode, link.SrcSlot)
			if err != nil {
				return nil, err
			}
			link.SrcID = iface.ID
		}

		if len(link.DstID) == 0 {
			iface, err := s.Interface.Create(ctx, link.LabID, link.DstNode, link.DstSlot)
			if err != nil {
				return nil, err
			}
			iface.IsConnected = true
			link.DstID = iface.ID
		}
	}

	newLink := struct {
		SrcInt models.UUID `json:"src_int"`
		DstInt models.UUID `json:"dst_int"`
	}{
		SrcInt: link.SrcID,
		DstInt: link.DstID,
	}

	newLinkResult := struct {
		ID models.UUID `json:"id"`
	}{}
	err = s.apiClient.PostJSON(ctx, api, nil, newLink, &newLinkResult)
	if err != nil {
		return nil, err
	}

	// FIXME: should be DEEP
	return s.GetByID(ctx, link.LabID, newLinkResult.ID)
}

// Delete removes a link from a lab identified by the Lab ID and Link ID
// provided in the link arg.
func (s *LinkService) Delete(ctx context.Context, link models.Link) error {
	api := fmt.Sprintf("labs/%s/links/%s", link.LabID, link.ID)
	return s.apiClient.DeleteJSON(ctx, api, nil)
}

// GetCondition retrieves the current link conditioning configuration
func (s *LinkService) GetCondition(ctx context.Context, labID, linkID models.UUID) (*models.ConditionResponse, error) {
	api := fmt.Sprintf("labs/%s/links/%s/condition", labID, linkID)

	queryParams := map[string]string{
		"operational": "true",
	}

	condition := &models.ConditionResponse{}
	err := s.apiClient.GetJSON(ctx, api, queryParams, condition)
	if err != nil {
		return nil, err
	}

	return condition, nil
}

// SetCondition applies link conditioning configuration
func (s *LinkService) SetCondition(ctx context.Context, labID, linkID models.UUID, config *models.LinkConditionConfiguration) (*models.ConditionResponse, error) {
	api := fmt.Sprintf("labs/%s/links/%s/condition", labID, linkID)

	queryParams := map[string]string{
		"operational": "true",
	}

	condition := &models.ConditionResponse{}
	err := s.apiClient.PatchJSON(ctx, api, queryParams, config, condition)
	if err != nil {
		return nil, err
	}

	return condition, nil
}

// DeleteCondition removes link conditioning configuration
func (s *LinkService) DeleteCondition(ctx context.Context, labID, linkID models.UUID) error {
	api := fmt.Sprintf("labs/%s/links/%s/condition", labID, linkID)
	return s.apiClient.DeleteJSON(ctx, api, nil)
}
