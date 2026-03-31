package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSystemInformation_Unmarshal_WithTimeoutAndFeatures(t *testing.T) {
	var si SystemInformation
	err := json.Unmarshal([]byte(`{
		"version":"2.11.0",
		"ready":true,
		"allow_ssh_pubkey_auth":true,
		"oui":null,
		"timeout":30,
		"features":["foo","bar"]
	}`), &si)
	assert.NoError(t, err)
	assert.Equal(t, "2.11.0", si.Version)
	assert.Equal(t, 30, si.Timeout)
	assert.Equal(t, []string{"foo", "bar"}, si.Features)
}

func TestSystemInformation_Unmarshal_WithInvalidFeaturesShape(t *testing.T) {
	var si SystemInformation
	err := json.Unmarshal([]byte(`{
		"version":"2.11.0",
		"ready":true,
		"allow_ssh_pubkey_auth":true,
		"features":{"foo":true}
	}`), &si)
	assert.Error(t, err)
}
