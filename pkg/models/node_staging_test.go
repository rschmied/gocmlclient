package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeStaging_JSON(t *testing.T) {
	js := `{"enabled":true,"start_remaining":true,"abort_on_failure":false}`
	var ns NodeStaging
	err := json.Unmarshal([]byte(js), &ns)
	assert.NoError(t, err)
	assert.True(t, ns.Enabled)
	assert.True(t, ns.StartRemaining)
	assert.False(t, ns.AbortOnFailure)

	b, err := json.Marshal(ns)
	assert.NoError(t, err)
	var m map[string]any
	err = json.Unmarshal(b, &m)
	assert.NoError(t, err)
	assert.Equal(t, true, m["enabled"])
	assert.Equal(t, true, m["start_remaining"])
	assert.Equal(t, false, m["abort_on_failure"])
}
