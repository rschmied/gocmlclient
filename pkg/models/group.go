// Package models provides the models for Cisco Modeling Labs
// here: group related types
package models

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
