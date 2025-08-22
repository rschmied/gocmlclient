// Package models provides the models for Cisco Modeling Labs
// here: user related types
package models

// {
//   "id": "00000000-0000-4000-a000-000000000000",
//   "created": "2023-02-06T09:57:56+00:00",
//   "modified": "2023-03-26T11:18:59+00:00",
//   "username": "admin",
//   "fullname": "",
//   "email": "",
//   "description": "",
//   "admin": true,
//   "directory_dn": "",
//   "groups": [],
//   "labs": [
//     "fa3269a7-1357-472a-96ed-1c000883530d",
//     "3d5251b0-0455-408b-9c41-99fb29cc3bf1",
//   ],
//   "opt_in": true,
//   "resource_pool": null
// }

type User struct {
	ID           string   `json:"id,omitempty"`
	Created      string   `json:"created,omitempty"`
	Modified     string   `json:"modified,omitempty"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	Fullname     string   `json:"fullname"`
	Email        string   `json:"email"`
	Description  string   `json:"description"`
	IsAdmin      bool     `json:"admin"`
	DirectoryDN  string   `json:"directory_dn,omitempty"`
	Groups       []string `json:"groups,omitempty"`
	Labs         []string `json:"labs,omitempty"`
	OptIn        bool     `json:"opt_in"`                  // with 2.5.0
	ResourcePool *string  `json:"resource_pool,omitempty"` // with 2.5.0
}

type UserList []*User
