package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/httputil"
	"github.com/rschmied/gocmlclient/pkg/models"
)

const (
	linksAPI     = "links"
	conditionAPI = "condition"
)

// Ensure LinkService implements interface
var _ LinkServiceInterface = (*LinkService)(nil)

// LinkServiceInterface defines methods needed by other services
type LinkServiceInterface interface {
	GetLinksForLab(ctx context.Context, labID models.UUID) ([]models.Link, error)
	Delete(ctx context.Context, labID, linkID models.UUID) error
	GetCondition(ctx context.Context, labID, linkID models.UUID) (models.ConditionResponse, error)
	SetCondition(ctx context.Context, labID, linkID models.UUID, config *models.LinkConditionConfiguration) (models.ConditionResponse, error)
	DeleteCondition(ctx context.Context, labID, linkID models.UUID) error
}

// LinkService provides link-related operations
type LinkService struct {
	apiClient *api.Client
	Interface InterfaceServiceInterface
	Node      NodeServiceInterface
}

// NewLinkService creates a new link service
func NewLinkService(apiClient *api.Client) *LinkService {
	return &LinkService{
		apiClient: apiClient,
	}
}

// linksURL builds base URL for links in a lab
func linksURL(labID models.UUID) string {
	return fmt.Sprintf("%s/%s/%s", labsAPI, labID, linksAPI)
}

// linkURL builds URL for a specific link
func linkURL(labID, linkID models.UUID) string {
	return fmt.Sprintf("%s/%s", linksURL(labID), linkID)
}

// linkConditionURL builds URL for link condition operations
func linkConditionURL(labID, linkID models.UUID) string {
	return fmt.Sprintf("%s/%s", linkURL(labID, linkID), conditionAPI)
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

func (s *LinkService) GetLinksForLab(ctx context.Context, labID models.UUID) ([]models.Link, error) {
	api := linksURL(labID)

	queryParams := httputil.NewQueryBuilder().WithData(true).Build()

	var linkList []models.Link
	err := s.apiClient.GetJSON(ctx, api, queryParams, &linkList)
	if err != nil {
		return nil, err
	}
	return linkList, nil
}

// GetByID returns the link data for the given `labID` and `linkID`.
func (s *LinkService) GetByID(ctx context.Context, labID, linkID models.UUID) (models.Link, error) {
	api := linkURL(labID, linkID)
	var link models.Link
	err := s.apiClient.GetJSON(ctx, api, nil, &link)
	if err != nil {
		return models.Link{}, err
	}
	if link.LabID != labID {
		return models.Link{}, fmt.Errorf("link lab ID mismatch: expected %s, got %s", labID, link.LabID)
	}
	return link, nil
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
func (s *LinkService) Create(ctx context.Context, link models.Link) (models.Link, error) {
	api := linksURL(link.LabID)

	if len(link.SrcNode) > 0 && len(link.DstNode) > 0 {
		ifaceListA, err := s.Interface.GetInterfacesForNode(ctx, link.LabID, link.SrcNode)
		if err != nil {
			return models.Link{}, err
		}

		ifaceListB, err := s.Interface.GetInterfacesForNode(ctx, link.LabID, link.DstNode)
		if err != nil {
			return models.Link{}, err
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
				return models.Link{}, err
			}
			link.SrcID = iface.ID
		}

		if len(link.DstID) == 0 {
			iface, err := s.Interface.Create(ctx, link.LabID, link.DstNode, link.DstSlot)
			if err != nil {
				return models.Link{}, err
			}
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
	err := s.apiClient.PostJSON(ctx, api, nil, newLink, &newLinkResult)
	if err != nil {
		return models.Link{}, err
	}

	resultLink, err := s.GetByID(ctx, link.LabID, newLinkResult.ID)
	if err != nil {
		return models.Link{}, err
	}
	return resultLink, nil
}

// Delete removes a link from a lab identified by the Lab ID and Link ID.
func (s *LinkService) Delete(ctx context.Context, labID, linkID models.UUID) error {
	api := linkURL(labID, linkID)
	return s.apiClient.DeleteJSON(ctx, api, nil)
}

// GetCondition retrieves the current link conditioning configuration
func (s *LinkService) GetCondition(ctx context.Context, labID, linkID models.UUID) (models.ConditionResponse, error) {
	api := linkConditionURL(labID, linkID)

	queryParams := httputil.NewQueryBuilder().
		WithOperational().
		Build()

	var condition models.ConditionResponse
	err := s.apiClient.GetJSON(ctx, api, queryParams, &condition)
	if err != nil {
		return models.ConditionResponse{}, err
	}

	return condition, nil
}

// SetCondition applies link conditioning configuration
func (s *LinkService) SetCondition(ctx context.Context, labID, linkID models.UUID, config *models.LinkConditionConfiguration) (models.ConditionResponse, error) {
	api := linkConditionURL(labID, linkID)

	queryParams := httputil.NewQueryBuilder().
		WithOperational().
		Build()

	var condition models.ConditionResponse
	err := s.apiClient.PatchJSON(ctx, api, queryParams, config, &condition)
	if err != nil {
		return models.ConditionResponse{}, err
	}

	return condition, nil
}

// DeleteCondition removes link conditioning configuration
func (s *LinkService) DeleteCondition(ctx context.Context, labID, linkID models.UUID) error {
	api := linkConditionURL(labID, linkID)
	return s.apiClient.DeleteJSON(ctx, api, nil)
}
