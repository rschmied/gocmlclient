package cmlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"golang.org/x/sync/errgroup"
)

// {
// 	"id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
// 	"state": "DEFINED_ON_CORE",
// 	"created": "2021-02-28T07:33:47+00:00",
// 	"modified": "2021-02-28T07:33:47+00:00",
// 	"lab_title": "Lab at Mon 17:27 PM",
// 	"owner": "90f84e38-a71c-4d57-8d90-00fa8a197385",
// 	"lab_description": "string",
// 	"node_count": 0,
// 	"link_count": 0,
// 	"lab_notes": "string",
// 	"groups": [
// 	  {
// 		"id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
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

type IDlist []string
type NodeMap map[string]*Node
type InterfaceList []*Interface
type nodeList []*Node
type linkList []*Link
type groupList []*Group

type labAlias struct {
	Lab
	OwnerID string `json:"owner"`
}

type labPatchPostAlias struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

type Lab struct {
	ID          string    `json:"id"`
	State       string    `json:"state"`
	Created     string    `json:"created"`
	Modified    string    `json:"modified"`
	Title       string    `json:"lab_title"`
	Description string    `json:"lab_description"`
	Notes       string    `json:"lab_notes"`
	Owner       *User     `json:"owner"`
	NodeCount   int       `json:"node_count"`
	LinkCount   int       `json:"link_count"`
	Nodes       NodeMap   `json:"nodes"`
	Links       linkList  `json:"links"`
	Groups      groupList `json:"groups"`

	// private
	// filled bool
}

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

func (l *Lab) Running() bool {
	for _, node := range l.Nodes {
		if node.State != NodeStateDefined && node.State != NodeStateStopped {
			return true
		}
	}
	return false
}

func (l *Lab) Booted() bool {
	for _, node := range l.Nodes {
		if node.State != NodeStateBooted {
			return false
		}
	}
	return true
}

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

func (c *Client) updateCachedLab(existingLab, updatedLab *Lab) *Lab {
	// only copy fields which can be updated
	c.mu.Lock()
	existingLab.Title = updatedLab.Title
	existingLab.Description = updatedLab.Description
	existingLab.Nodes = updatedLab.Nodes
	existingLab.State = updatedLab.State
	c.mu.Unlock()
	return existingLab
}

func (c *Client) cacheLab(lab *Lab, err error) (*Lab, error) {
	if !c.useCache || err != nil {
		return lab, err
	}

	c.mu.RLock()
	existingLab, ok := c.labCache[lab.ID]
	c.mu.RUnlock()
	if ok {
		return c.updateCachedLab(existingLab, lab), nil
	}

	lab.Nodes = make(NodeMap)
	c.mu.Lock()
	c.labCache[lab.ID] = lab
	c.mu.Unlock()
	return lab, nil
}

func (c *Client) getCachedLab(id string, deep bool) (*Lab, bool) {
	// no caching when reading deep
	if !c.useCache || deep {
		return nil, false
	}
	c.mu.RLock()
	lab, ok := c.labCache[id]
	c.mu.RUnlock()
	return lab, ok
}

func (c *Client) deleteCachedLab(id string, err error) error {
	if !c.useCache || err != nil {
		return err
	}
	c.mu.Lock()
	delete(c.labCache, id)
	c.mu.Unlock()
	return nil
}

func (c *Client) LabCreate(ctx context.Context, lab Lab) (*Lab, error) {

	// TODO: inconsistent attributes lab_title vs title, ...
	postAlias := labPatchPostAlias{
		Title:       lab.Title,
		Description: lab.Description,
		Notes:       lab.Notes,
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

	la.Owner = &User{ID: la.OwnerID}
	la.Nodes = make(NodeMap)
	return c.cacheLab(&la.Lab, nil)
}

func (c *Client) LabUpdate(ctx context.Context, lab Lab) (*Lab, error) {

	// TODO: inconsistent attributes lab_title vs title, ...
	patchAlias := labPatchPostAlias{
		Title:       lab.Title,
		Description: lab.Description,
		Notes:       lab.Notes,
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
	return c.cacheLab(&la.Lab, nil)
}

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

func (c *Client) LabStart(ctx context.Context, id string) error {
	return c.jsonPut(ctx, fmt.Sprintf("labs/%s/start", id), 0)
}

func (c *Client) HasLabConverged(ctx context.Context, id string) (bool, error) {
	api := fmt.Sprintf("labs/%s/check_if_converged", id)
	converged := false
	err := c.jsonGet(ctx, api, &converged, 0)
	if err != nil {
		return false, err
	}
	return converged, nil
}

func (c *Client) LabStop(ctx context.Context, id string) error {
	return c.jsonPut(ctx, fmt.Sprintf("labs/%s/stop", id), 0)
}

func (c *Client) LabWipe(ctx context.Context, id string) error {
	return c.jsonPut(ctx, fmt.Sprintf("labs/%s/wipe", id), 0)
}

func (c *Client) LabDestroy(ctx context.Context, id string) error {
	return c.deleteCachedLab(id, c.jsonDelete(ctx, fmt.Sprintf("labs/%s", id), 0))
}

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

func (c *Client) LabGet(ctx context.Context, id string, deep bool) (*Lab, error) {

	if lab, ok := c.getCachedLab(id, deep); ok {
		return lab, nil
	}
	api := fmt.Sprintf("labs/%s", id)
	la := &labAlias{}
	err := c.jsonGet(ctx, api, la, 0)
	if err != nil {
		return nil, err
	}
	if !deep {
		la.Owner = &User{ID: la.OwnerID}
		return c.cacheLab(&la.Lab, nil)
	}
	return c.labFill(ctx, la)
}

func (c *Client) labFill(ctx context.Context, la *labAlias) (*Lab, error) {

	var err error
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		defer log.Printf("user done")
		la.Owner, err = c.getUser(ctx, la.OwnerID)
		if err != nil {
			return err
		}
		return nil
	})

	lab := &la.Lab
	lab, _ = c.cacheLab(lab, nil)

	// need to ensure that this block finishes before the others run
	ch := make(chan struct{})
	g.Go(func() error {
		defer func() {
			log.Printf("nodes/interfaces done")
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
		defer log.Printf("l3info done")
		l3info, err := c.getL3Info(ctx, lab.ID)
		if err != nil {
			return err
		}
		log.Printf("l3info read")
		// wait for node data read complete
		<-ch
		// map and merge the l3 data...
		for nid, l3data := range *l3info {
			if node, found := lab.Nodes[nid]; found {
				for mac, l3i := range l3data.Interfaces {
					for _, iface := range node.Interfaces {
						// if iface, found := node.Interfaces[l3i.ID]; found {
						if iface.MACaddress == mac {
							iface.IP4 = l3i.IP4
							iface.IP6 = l3i.IP6
							break
						}
					}
				}
			}
		}
		log.Printf("loops done")
		return nil
	})

	g.Go(func() error {
		defer log.Printf("links done")
		idlist, err := c.getLinkIDsForLab(ctx, lab)
		if err != nil {
			return err
		}
		log.Printf("linkidlist read")
		// wait for node data read complete
		<-ch
		return c.getLinksForLab(ctx, lab, idlist)
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}
	log.Printf("wait done")
	// lab.filled = true
	// return c.cacheLab(lab, nil)
	return lab, nil
}
