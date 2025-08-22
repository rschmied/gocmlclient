package cmlclient

import (
	"context"
	"fmt"
	"sort"
)

// [
//   {
//     "name": "CCNA Study Group Class of 21",
//     "description": "string",
//     "members": [
//       "90f84e38-a71c-4d57-8d90-00fa8a197385",
//       "60f84e39-ffff-4d99-8a78-00fa8aaf5666"
//     ],
//     "labs": [
//       {
//         "id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
//         "permission": "read_only"
//       }
//     ],
//     "id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
//     "created": "2021-02-28T07:33:47+00:00",
//     "modified": "2021-02-28T07:33:47+00:00"
//   }
// ]

type GroupLab struct {
	ID         string `json:"id"`
	Permission string `json:"permission"`
}
type Group struct {
	ID          string     `json:"id,omitempty"`
	Description string     `json:"description"`
	Members     []string   `json:"members"`
	Name        string     `json:"name"`
	Labs        []GroupLab `json:"labs"`
}

type GroupList []*Group

// Groups retrieves the list of all groups which exist on the controller.
func (c *Client) Groups(ctx context.Context) (GroupList, error) {
	groups := GroupList{}
	err := c.GetJSON(ctx, "groups", nil, &groups)
	if err != nil {
		return nil, err
	}
	// sort the group list by their ID
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].ID > groups[j].ID
	})
	return groups, nil
}

// GroupByName tries to get the group with the provided `name`.
func (c *Client) GroupByName(ctx context.Context, name string) (*Group, error) {
	group := Group{}
	err := c.GetJSON(ctx, fmt.Sprintf("groups/%s/id", name), nil, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GroupGet retrieves the group with the provided `id` (a UUIDv4).
func (c *Client) GroupGet(ctx context.Context, id string) (*Group, error) {
	group := Group{}
	err := c.GetJSON(ctx, fmt.Sprintf("groups/%s", id), nil, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GroupDestroy deletes the group identified by the `id` (a UUIDv4).
func (c *Client) GroupDestroy(ctx context.Context, id string) error {
	return c.DeleteJSON(ctx, fmt.Sprintf("groups/%s", id), nil)
}

// GroupCreate creates a new group on the controller based on the data provided
// in the passed group parameter.
func (c *Client) GroupCreate(ctx context.Context, group *Group) (*Group, error) {
	result := Group{}
	err := c.PostJSON(ctx, "groups", nil, group, &result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// GroupUpdate updates the given group which must exist.
func (c *Client) GroupUpdate(ctx context.Context, group *Group) (*Group, error) {
	groupID := group.ID
	group.ID = ""
	result := Group{}
	err := c.PatchJSON(ctx, fmt.Sprintf("groups/%s", groupID), group, &result)
	if err != nil {
		return nil, err
	}
	return &result, err
}
