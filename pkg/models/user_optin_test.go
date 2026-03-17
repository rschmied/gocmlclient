package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_UnmarshalJSON_OptIn_StringEnum(t *testing.T) {
	var u User
	err := json.Unmarshal([]byte(`{"id":"u1","username":"x","fullname":"x","description":"","email":"","admin":false,"groups":[],"associations":[],"opt_in":"accepted","resource_pool":null,"tour_version":"","pubkey_info":""}`), &u)
	assert.NoError(t, err)
	if assert.NotNil(t, u.OptIn) {
		assert.Equal(t, OptInAccepted, *u.OptIn)
	}
}

func TestUser_UnmarshalJSON_OptIn_BoolLegacy(t *testing.T) {
	var u User
	err := json.Unmarshal([]byte(`{"id":"u1","username":"x","fullname":"x","description":"","email":"","admin":false,"groups":[],"associations":[],"opt_in":true,"resource_pool":null,"tour_version":"","pubkey_info":""}`), &u)
	assert.NoError(t, err)
	if assert.NotNil(t, u.OptIn) {
		assert.Equal(t, OptInAccepted, *u.OptIn)
	}

	var u2 User
	err = json.Unmarshal([]byte(`{"id":"u2","username":"x","fullname":"x","description":"","email":"","admin":false,"groups":[],"associations":[],"opt_in":false,"resource_pool":null,"tour_version":"","pubkey_info":""}`), &u2)
	assert.NoError(t, err)
	if assert.NotNil(t, u2.OptIn) {
		assert.Equal(t, OptInDeclined, *u2.OptIn)
	}
}

func TestUser_UnmarshalJSON_OptIn_Unset(t *testing.T) {
	var u User
	err := json.Unmarshal([]byte(`{"id":"u1","username":"x","fullname":"x","description":"","email":"","admin":false,"groups":[],"associations":[],"opt_in":"unset","resource_pool":null,"tour_version":"","pubkey_info":""}`), &u)
	assert.NoError(t, err)
	if assert.NotNil(t, u.OptIn) {
		assert.Equal(t, OptInUnset, *u.OptIn)
	}
}
