package cmlclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestClient_InterfaceMapMarschalJSON(t *testing.T) {
// 	nm := InterfaceMap{
// 		"zzz": &Interface{ID: "zzz"},
// 		"aaa": &Interface{ID: "aaa"},
// 	}
// 	b, err := nm.MarshalJSON()
// 	assert.NoError(t, err)
// 	t.Logf("%+v", string(b))

// 	nl := []Node{}
// 	err = json.Unmarshal(b, &nl)
// 	assert.NoError(t, err)
// 	assert.Equal(t, nl[0].ID, "aaa")
// 	assert.Equal(t, nl[1].ID, "zzz")
// }

func TestClient_IfaceExists(t *testing.T) {
	iface := Interface{
		State: IfaceStateDefined,
	}
	assert.Equal(t, false, iface.Exists())
}

func TestClient_IfaceRuns(t *testing.T) {
	iface := Interface{
		State: IfaceStateStarted,
	}
	assert.Equal(t, true, iface.Runs())
}
