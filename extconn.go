package cmlclient

import (
	"context"
	"fmt"
)

/*
[
  {
    "id": "58568fbb-e1f8-4b83-a1f8-148c656eed39",
    "device_name": "virbr0",
    "label": "NAT",
    "protected": false,
    "snooped": true,
    "tags": [
      "NAT"
    ],
    "operational": {
      "forwarding": "NAT",
      "label": "NAT",
      "mtu": 1500,
      "exists": true,
      "enabled": true,
      "protected": false,
      "snooped": true,
      "stp": false,
      "interface": null,
      "ip_networks": []
    }
  }
]
*/

// Operational data struct
type opdata struct {
	Forwarding string `json:"forwarding"`
	Label      string `json:"label"`
	MTU        int    `json:"mtu"`
	Exists     bool   `json:"exits"`
	Enabled    bool   `json:"enabled"`
	Protected  bool   `json:"protected"`
	Snooped    bool   `json:"snooped"`
	STP        bool   `json:"stp"`
	// Interface  string   `json:"interface"`
	IPNetworks []string `json:"ip_networks"`
}

// ExtConn defines the data structure for a CML external connector
type ExtConn struct {
	ID          string   `json:"id"`
	DeviceName  string   `json:"device_name"`
	Label       string   `json:"label"`
	Protected   bool     `json:"protected"`
	Snooped     bool     `json:"snooped"`
	Tags        []string `json:"tags"`
	Operational opdata   `json:"operational"`
}

// ExtConnGet returns the external connector specified by the ID given
func (c *Client) ExtConnGet(ctx context.Context, extConnID string) (*ExtConn, error) {
	api := fmt.Sprintf("system/external_connectors/%s", extConnID)
	extconn := &ExtConn{}
	err := c.jsonGet(ctx, api, extconn, 0)
	if err != nil {
		return nil, err
	}
	return extconn, err
}

// ExtConnectors returns all external connectors on the system
func (c *Client) ExtConnectors(ctx context.Context) ([]*ExtConn, error) {
	api := fmt.Sprintf("system/external_connectors")
	extconnlist := make([]*ExtConn, 0)
	err := c.jsonGet(ctx, api, &extconnlist, 0)
	if err != nil {
		return nil, err
	}
	return extconnlist, err
}
