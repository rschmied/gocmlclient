package services

import (
	"context"
	"fmt"
	"sort"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/pkg/models"
)

/*
{
	"id": "e87c811d-5459-4390-8e92-317bb9dc23e8",
	"lab_id": "024fa9f4-5e5e-4e94-9f85-29f147e09689",
	"node": "f902d112-2a93-4c9f-98e6-adea6dc16fef",
	"label": "eth0",
	"slot": 0,
	"type": "physical",
	"device_name": null,
	"dst_udp_port": 21001,
	"src_udp_port": 21000,
	"mac_address": "52:54:00:1e:af:9b",
	"is_connected": true,
	"state": "STARTED"
}
*/

// InterfaceService provides interface-related operations
type InterfaceService struct {
	apiClient *api.Client
}

// InterfaceServiceInterface defines methods needed by other services
type InterfaceServiceInterface interface {
	GetInterfacesForNode(ctx context.Context, node *models.Node) error
}

// NewInterfaceService creates a new lab service
func NewInterfaceService(apiClient *api.Client) *InterfaceService {
	return &InterfaceService{
		apiClient: apiClient,
	}
}

func (s *InterfaceService) GetInterfacesForNode(ctx context.Context, node *models.Node) error {
	// with the data=true option, we get not only the list of IDs but the
	// interfaces themselves as well!
	api := fmt.Sprintf("labs/%s/nodes/%s/interfaces", node.LabID, node.ID)
	queryParm := map[string]string{
		"data": "true",
	}

	interfaceList := models.InterfaceList{}
	err := s.apiClient.GetJSON(ctx, api, queryParm, &interfaceList)
	if err != nil {
		return err
	}

	// sort the interface list by slot
	sort.Slice(interfaceList, func(i, j int) bool {
		return interfaceList[i].Slot < interfaceList[j].Slot
	})
	node.Interfaces = interfaceList
	return nil
}

// {
// 	"00da52b6-2683-49c0-ba3a-ace877dea4ca": {
// 	  "name": "alpine-0",
// 	  "interfaces": {
// 		"52:54:00:00:00:09": {
// 		  "id": "3b45184f-7041-4300-aef2-2b97d8e763a8",
// 		  "label": "eth0",
// 		  "ip4": [
// 			"192.168.122.35"
// 		  ],
// 		  "ip6": [
// 			"fe80::5054:ff:fe00:9"
// 		  ]
// 		}
// 	  }
// 	},
// 	"0df7a717-9826-4729-9fe1-bc4932498c83": {
// 	  "name": "alpine-1",
// 	  "interfaces": {
// 		"52:54:00:00:00:08": {
// 		  "id": "6bec8956-f812-4fb3-9551-aef4410807ec",
// 		  "label": "eth0",
// 		  "ip4": [
// 			"192.168.122.34"
// 		  ],
// 		  "ip6": [
// 			"fe80::5054:ff:fe00:8"
// 		  ]
// 		}
// 	  }
// 	}
// }

// GetByID returns the interface identified by its `ID` (iface.ID).
func (s *InterfaceService) GetByID(ctx context.Context, iface *models.Interface) (*models.Interface, error) {
	api := fmt.Sprintf("labs/%s/interfaces/%s", iface.LabID, iface.ID)
	err := s.apiClient.GetJSON(ctx, api, nil, iface)
	return iface, err
}

// Create creates an interface in the given lab and node.  If the slot is >= 0,
// the request creates all unallocated slots up to and including that slot.
// Conversely, if the slot is < 0 (e.g. -1), the next free slot is used.
func (s *InterfaceService) Create(ctx context.Context, labID, nodeID string, slot int) (*models.Interface, error) {
	var slotPtr *int

	if slot >= 0 {
		slotPtr = &slot
	}

	newIface := struct {
		Node string `json:"node"`
		Slot *int   `json:"slot,omitempty"`
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
		result := models.Interface{}
		err := s.apiClient.PostJSON(ctx, api, nil, newIface, &result)
		if err != nil {
			return nil, err
		}
		return &result, err
	}

	// this is when a slot has been provided; the API provides now a list of
	// interfaces
	result := []models.Interface{}
	err := s.apiClient.PostJSON(ctx, api, nil, newIface, &result)
	if err != nil {
		return nil, err
	}

	lastIface := &result[len(result)-1]
	return lastIface, nil
}
