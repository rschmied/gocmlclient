package services

import (
	"context"
	"fmt"

	"github.com/rschmied/gocmlclient/pkg/models"
)

const layer3Action = "layer3_addresses"

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

func (s *LabService) getL3Info(ctx context.Context, id models.UUID) (nodes *l3nodes, err error) {
	nodes = &l3nodes{}
	err = s.apiClient.GetJSON(ctx, fmt.Sprintf("%s/%s", labURL(id), layer3Action), nil, nodes)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}
