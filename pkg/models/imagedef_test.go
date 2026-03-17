package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageDefinition_Unmarshal_Schema210Fields(t *testing.T) {
	var id ImageDefinition
	err := json.Unmarshal([]byte(`{
		"id":"img1",
		"node_definition_id":"iosv",
		"label":"IOSv",
		"schema_version":"1.0",
		"description":"desc",
		"disk_image":"d1.qcow2",
		"disk_image_2":null,
		"disk_image_3":null,
		"disk_image_4":null,
		"read_only":false,
		"configuration":null,
		"docker_tag":null,
		"efi_boot":false,
		"sha256":null,
		"ram":512,
		"cpus":1,
		"cpu_limit":100,
		"data_volume":0,
		"boot_disk_size":8,
		"pyats":null
	}`), &id)
	assert.NoError(t, err)
	assert.Equal(t, UUID("img1"), id.ID)
	assert.Equal(t, "iosv", id.NodeDefID)
	assert.Equal(t, "IOSv", id.Label)
	assert.Equal(t, "d1.qcow2", id.DiskImage1())
}
