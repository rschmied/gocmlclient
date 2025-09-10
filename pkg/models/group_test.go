package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroup_JSON(t *testing.T) {
	tests := []struct {
		name    string
		group   Group
		jsonStr string
	}{
		{
			name: "complete group",
			group: Group{
				ID:          "123e4567-e89b-12d3-a456-426614174000",
				Description: "Test group",
				Name:        "test-group",
				Members:     []UUID{"user1", "user2"},
				Created:     "2023-01-01T00:00:00Z",
				Modified:    "2023-01-02T00:00:00Z",
				DirectoryDN: "cn=test-group,ou=groups,dc=example,dc=com",
			},
			jsonStr: `{
				"id": "123e4567-e89b-12d3-a456-426614174000",
				"description": "Test group",
				"name": "test-group",
				"members": ["user1", "user2"],
				"created": "2023-01-01T00:00:00Z",
				"modified": "2023-01-02T00:00:00Z",
				"directory_dn": "cn=test-group,ou=groups,dc=example,dc=com"
			}`,
		},
		{
			name: "minimal group",
			group: Group{
				Name: "minimal-group",
			},
			jsonStr: `{
				"name": "minimal-group"
			}`,
		},
		{
			name: "group with directory exists",
			group: Group{
				Name:            "dir-group",
				DirectoryExists: boolPtr(true),
			},
			jsonStr: `{
				"name": "dir-group",
				"directory_exists": true
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			_, err := json.Marshal(tt.group)
			assert.NoError(t, err)

			// Test unmarshaling
			var unmarshaled Group
			err = json.Unmarshal([]byte(tt.jsonStr), &unmarshaled)
			assert.NoError(t, err)

			// Compare key fields
			assert.Equal(t, tt.group.ID, unmarshaled.ID)
			assert.Equal(t, tt.group.Name, unmarshaled.Name)
			assert.Equal(t, tt.group.Description, unmarshaled.Description)
			assert.Equal(t, tt.group.Members, unmarshaled.Members)
			assert.Equal(t, tt.group.Created, unmarshaled.Created)
			assert.Equal(t, tt.group.Modified, unmarshaled.Modified)
			assert.Equal(t, tt.group.DirectoryDN, unmarshaled.DirectoryDN)
			assert.Equal(t, tt.group.DirectoryExists, unmarshaled.DirectoryExists)
		})
	}
}

func TestGroupList_JSON(t *testing.T) {
	groups := GroupList{
		{
			ID:   "group1",
			Name: "Group 1",
		},
		{
			ID:   "group2",
			Name: "Group 2",
		},
	}

	// Test marshaling
	data, err := json.Marshal(groups)
	assert.NoError(t, err)

	// Test unmarshaling
	var unmarshaled GroupList
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Len(t, unmarshaled, 2)
	assert.Equal(t, "group1", string(unmarshaled[0].ID))
	assert.Equal(t, "Group 1", unmarshaled[0].Name)
	assert.Equal(t, "group2", string(unmarshaled[1].ID))
	assert.Equal(t, "Group 2", unmarshaled[1].Name)
}

// Helper function for creating bool pointers
func boolPtr(b bool) *bool {
	return &b
}
