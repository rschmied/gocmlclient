package models

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
	NodeMap      map[string]*Node
	LinkList     []*Link
	LabGroupList []*LabGroup
)

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
	Links       LinkList     `json:"links"`
	Groups      LabGroupList `json:"groups"`
}
