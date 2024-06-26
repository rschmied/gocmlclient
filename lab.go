package cmlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/sync/errgroup"
)

// {
// 	"state": "STOPPED",
// 	"created": "2023-02-08T10:02:43+00:00",
// 	"modified": "2023-02-08T11:59:45+00:00",
// 	"lab_title": "sample",
// 	"lab_description": "",
// 	"lab_notes": "",
// 	"owner": "00000000-0000-4000-a000-000000000000",
// 	"owner_username": "admin",
// 	"node_count": 5,
// 	"link_count": 4,
// 	"id": "a7d20917-5e57-407f-80ea-63596c53f198",
// 	"groups": [
// 	  {
// 		"id": "bc9b796e-48bc-4369-b131-231dfa057d41",
// 		"name": "students",
// 		"permission": "read_only"
// 	  }
// 	]
// }

const (
	LabStateDefined = "DEFINED_ON_CORE"
	LabStateStopped = "STOPPED"
	LabStateStarted = "STARTED"
	LabStateBooted  = "BOOTED"
)

type LabGroup struct {
	ID         string `json:"id"`
	Name       string `json:"name,omitempty"`
	Permission string `json:"permission"`
}

type (
	IDlist        []string
	NodeMap       map[string]*Node
	InterfaceList []*Interface
	nodeList      []*Node
	linkList      []*Link
	LabGroupList  []*LabGroup
)

type labAlias struct {
	Lab
	OwnerID string `json:"owner"`
}

type labPatchPostAlias struct {
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	Notes       string       `json:"notes,omitempty"`
	Groups      LabGroupList `json:"groups,omitempty"`
}

type Lab struct {
	ID          string       `json:"id"`
	State       string       `json:"state"`
	Created     string       `json:"created"`
	Modified    string       `json:"modified"`
	Title       string       `json:"lab_title"`
	Description string       `json:"lab_description"`
	Notes       string       `json:"lab_notes"`
	Owner       *User        `json:"owner"`
	NodeCount   int          `json:"node_count"`
	LinkCount   int          `json:"link_count"`
	Nodes       NodeMap      `json:"nodes"`
	Links       linkList     `json:"links"`
	Groups      LabGroupList `json:"groups"`
}

// CanBeWiped returns `true` when all nodes in the lab are wiped.
func (l *Lab) CanBeWiped() bool {
	if len(l.Nodes) == 0 {
		return l.State != LabStateDefined
	}
	for _, node := range l.Nodes {
		if node.State != NodeStateDefined {
			return false
		}
	}
	return true
}

// Running returns `true` if at least one node is running (started or booted).
func (l *Lab) Running() bool {
	for _, node := range l.Nodes {
		if node.State != NodeStateDefined && node.State != NodeStateStopped {
			return true
		}
	}
	return false
}

// Booted returns `true` if all nodes in the lab are in booted state.
func (l *Lab) Booted() bool {
	for _, node := range l.Nodes {
		if node.State != NodeStateBooted {
			return false
		}
	}
	return true
}

// NodeByLabel returns the node of a lab identified by its `label“ or an error
// if not found.
func (l *Lab) NodeByLabel(ctx context.Context, label string) (*Node, error) {
	for _, node := range l.Nodes {
		if node.Label == label {
			return node, nil
		}
	}
	return nil, ErrElementNotFound
}

type LabImport struct {
	ID       string   `json:"id"`
	Warnings []string `json:"warnings"`
}

// LabCreate creates a new lab on the controller.
func (c *Client) LabCreate(ctx context.Context, lab Lab) (*Lab, error) {
	// TODO: inconsistent attributes lab_title vs title, ...
	postAlias := labPatchPostAlias{
		Title:       lab.Title,
		Description: lab.Description,
		Notes:       lab.Notes,
		// Groups:      lab.Groups, // can't set at create
	}

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(postAlias)
	if err != nil {
		return nil, err
	}

	la := &labAlias{}
	err = c.jsonPost(ctx, "labs", buf, &la, 0)
	if err != nil {
		return nil, err
	}

	lab.ID = la.ID
	return c.LabUpdate(ctx, lab)
}

// LabUpdate updates specific fields of a lab (title, description and notes).
func (c *Client) LabUpdate(ctx context.Context, lab Lab) (*Lab, error) {
	// TODO: inconsistent attributes lab_title vs title, ...
	patchAlias := labPatchPostAlias{
		Title:       lab.Title,
		Description: lab.Description,
		Notes:       lab.Notes,
		Groups:      lab.Groups,
	}

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(patchAlias)
	if err != nil {
		return nil, err
	}

	la := &labAlias{}
	api := fmt.Sprintf("labs/%s", lab.ID)
	err = c.jsonPatch(ctx, api, buf, &la, 0)
	if err != nil {
		return nil, err
	}

	la.Owner = &User{ID: la.OwnerID}
	return &la.Lab, nil
}

// LabImport imports a lab topology into the controller. This is expected to be
// in CML YAML topology file format.
func (c *Client) LabImport(ctx context.Context, topo string) (*Lab, error) {
	topoReader := strings.NewReader(topo)
	labImport := &LabImport{}
	err := c.jsonPost(ctx, "import", topoReader, labImport, 0)
	if err != nil {
		return nil, err
	}
	lab, err := c.LabGet(ctx, labImport.ID, true) // full/deep!
	if err != nil {
		return nil, err
	}
	return lab, nil
}

// LabStart starts all nodes of the lab identified by the `id` (a UUIDv4).
func (c *Client) LabStart(ctx context.Context, id string) error {
	return c.jsonPut(ctx, fmt.Sprintf("labs/%s/start", id), 0)
}

// HasLabConverged checks if all nodes of the lab identified by the `id`
// (a UUIDv4) have converged e.g. are in state "BOOTED".
func (c *Client) HasLabConverged(ctx context.Context, id string) (bool, error) {
	api := fmt.Sprintf("labs/%s/check_if_converged", id)
	converged := false
	err := c.jsonGet(ctx, api, &converged, 0)
	if err != nil {
		return false, err
	}
	return converged, nil
}

// LabStop stops all nodes of the lab identified by the `id` (a UUIDv4).
func (c *Client) LabStop(ctx context.Context, id string) error {
	return c.jsonPut(ctx, fmt.Sprintf("labs/%s/stop", id), 0)
}

// LabWipe wipes the lab identified by the `id` (a UUIDv4).
func (c *Client) LabWipe(ctx context.Context, id string) error {
	return c.jsonPut(ctx, fmt.Sprintf("labs/%s/wipe", id), 0)
}

// LabDestroy deletes the lab identified by the `id` (a UUIDv4).
func (c *Client) LabDestroy(ctx context.Context, id string) error {
	return c.jsonDelete(ctx, fmt.Sprintf("labs/%s", id), 0)
}

// LabGetByTitle returns the lab identified by its `title`. For the use of
// `deep` see LabGet().
func (c *Client) LabGetByTitle(ctx context.Context, title string, deep bool) (*Lab, error) {
	var data map[string]map[string]*labAlias

	err := c.jsonGet(ctx, "populate_lab_tiles", &data, 0)
	if err != nil {
		return nil, err
	}
	labs := data["lab_tiles"]
	for _, lab := range labs {
		if lab.Title == title {
			if !deep {
				lab.Owner = &User{ID: lab.OwnerID}
				return &lab.Lab, nil
			}
			return c.labFill(ctx, lab)
		}
	}

	return nil, ErrElementNotFound
}

// LabGet returns the lab identified by `id` (a UUIDv4). If `deep` is provided,
// then the nodes, their interfaces and links are also fetched from the
// controller. Also, with `deep`, the L3 IP address info is fetched for the
// given lab.
func (c *Client) LabGet(ctx context.Context, id string, deep bool) (*Lab, error) {
	api := fmt.Sprintf("labs/%s", id)
	la := &labAlias{}
	err := c.jsonGet(ctx, api, la, 0)
	if err != nil {
		return nil, err
	}
	if !deep {
		la.Owner = &User{ID: la.OwnerID}
		return &la.Lab, nil
	}
	return c.labFill(ctx, la)
}

func (c *Client) labFill(ctx context.Context, la *labAlias) (*Lab, error) {
	var err error
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		defer slog.Debug("user done")
		la.Owner, err = c.UserGet(ctx, la.OwnerID)
		if err != nil {
			return err
		}
		return nil
	})

	lab := &la.Lab

	// need to ensure that this block finishes before the others run
	ch := make(chan struct{})
	g.Go(func() error {
		defer func() {
			slog.Debug("nodes/interfaces done")
			// two sync points, we can run the API endpoints but we need to
			// wait for the node data to be read until we can add the layer3
			// info (1) and the link info (2)
			ch <- struct{}{}
			ch <- struct{}{}
		}()
		err := c.getNodesForLab(ctx, lab)
		if err != nil {
			return err
		}
		for _, node := range lab.Nodes {
			err = c.getInterfacesForNode(ctx, node)
			if err != nil {
				return err
			}
		}
		return nil
	})

	g.Go(func() error {
		defer slog.Debug("l3info done")
		l3info, err := c.getL3Info(ctx, lab.ID)
		if err != nil {
			return err
		}
		slog.Debug("l3info read")
		// wait for node data read complete
		<-ch
		// map and merge the l3 data...
		for nid, l3data := range *l3info {
			if node, found := lab.Nodes[nid]; found {
				for mac, l3i := range l3data.Interfaces {
					for _, iface := range node.Interfaces {
						if iface.MACaddress == mac {
							iface.IP4 = l3i.IP4
							iface.IP6 = l3i.IP6
							break
						}
					}
				}
			}
		}
		slog.Debug("l3info loop done")
		return nil
	})

	g.Go(func() error {
		defer slog.Debug("links done")
		idlist, err := c.getLinkIDsForLab(ctx, lab)
		if err != nil {
			return err
		}
		slog.Debug("links read")
		// wait for node data read complete
		<-ch
		return c.getLinksForLab(ctx, lab, idlist)
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}
	slog.Debug("wait done")
	return lab, nil
}
