package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabImport_Unmarshal_WarningsNull(t *testing.T) {
	var li LabImport
	err := json.Unmarshal([]byte(`{"id":"lab1","warnings":null}`), &li)
	assert.NoError(t, err)
	assert.Equal(t, UUID("lab1"), li.ID)
	assert.Nil(t, li.Warnings)
}

func TestLabImport_Unmarshal_WarningsArray(t *testing.T) {
	var li LabImport
	err := json.Unmarshal([]byte(`{"id":"lab1","warnings":["w1","w2"]}`), &li)
	assert.NoError(t, err)
	assert.Equal(t, UUID("lab1"), li.ID)
	assert.Equal(t, []string{"w1", "w2"}, li.Warnings)
}
