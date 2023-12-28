package cmlclient

import (
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"

	mr "github.com/rschmied/mockresponder"
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

func TestClient_IfaceWithSlots(t *testing.T) {
	tc := newAuthedTestAPIclient()
	tc.client.useCache = true

	ifaceList := []byte(`[{
			"id": "n2i0",
			"lab_id": "lab1",
			"node": "node1",
			"label": "eth0",
			"slot": 0,
			"type": "physical",
			"mac_address": "52:54:00:0c:e0:70",
			"is_connected": true,
			"state": "STOPPED"
		},
		{
			"id": "n2i1",
			"lab_id": "lab1",
			"node": "node1",
			"label": "eth1",
			"slot": 1,
			"type": "physical",
			"mac_address": "52:54:00:0c:e0:71",
			"is_connected": true,
			"state": "STOPPED"
		}]
	`)

	data := mr.MockRespList{
		mr.MockResp{Data: ifaceList, URL: `/interfaces$`},
	}
	tc.mr.SetData(data)
	iface, err := tc.client.InterfaceCreate(tc.ctx, "lab1", "node1", 1)
	if assert.NoError(t, err) {
		assert.Equal(t, "eth1", iface.Label)
	}
}

func TestClient_IfaceDelete(t *testing.T) {
	tc := newAuthedTestAPIclient()
	tc.client.useCache = true

	tests := []struct {
		name      string
		lab       Lab
		node      Node
		ifaceList InterfaceList
		preErr    error
		want      bool
	}{
		{
			"error before",
			Lab{
				ID:    "different",
				Nodes: make(NodeMap),
			},
			Node{},
			InterfaceList{},
			errors.New("some error"),
			false,
		},
		{
			"nolab",
			Lab{
				ID:    "different",
				Nodes: make(NodeMap),
			},
			Node{},
			InterfaceList{},
			nil,
			false,
		},
		{
			"nolab",
			Lab{
				ID:    "lab1",
				Nodes: make(NodeMap),
			},
			Node{ID: "node2"},
			InterfaceList{},
			nil,
			false,
		},
		{
			"good",
			Lab{
				ID:    "lab1",
				Nodes: make(NodeMap),
			},
			Node{ID: "node1"},
			InterfaceList{
				&Interface{ID: "iface0"},
				&Interface{ID: "iface1"},
				&Interface{ID: "iface2"},
				&Interface{ID: "iface3"},
			},
			nil,
			false,
		},
	}

	iface := &Interface{ID: "iface2", LabID: "lab1", Node: "node1"}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.node.Interfaces = tt.ifaceList
			tt.lab.Nodes[tt.node.ID] = &tt.node
			tc.client.labCache[tt.lab.ID] = &tt.lab
			err := tc.client.deleteCachedIface(iface, tt.preErr)
			assert.Equal(t, tt.preErr, err)
			if err == nil && tt.name == "good" {
				assert.Len(t, tt.node.Interfaces, 3)
			}
		})
	}
}

func Test_Race(t *testing.T) {
	tc := newAuthedTestAPIclient()
	tc.client.useCache = true

	iface0 := []byte(`{
		"id": "iface0",
		"lab_id": "lab1",
		"node": "node1",
		"label": "eth0",
		"slot": 0,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:70",
		"is_connected": true,
		"state": "STOPPED"
	}`)
	ifaceList := []byte(`[{
		"id": "iface0",
		"lab_id": "lab1",
		"node": "node1",
		"label": "eth0",
		"slot": 0,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:70",
		"is_connected": true,
		"state": "STOPPED"
	},{
		"id": "iface1",
		"lab_id": "lab1",
		"node": "node1",
		"label": "eth1",
		"slot": 1,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:70",
		"is_connected": true,
		"state": "STOPPED"
	},{
		"id": "iface2",
		"lab_id": "lab1",
		"node": "node1",
		"label": "eth2",
		"slot": 2,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:70",
		"is_connected": true,
		"state": "STOPPED"
	},{
		"id": "iface3",
		"lab_id": "lab1",
		"node": "node1",
		"label": "eth3",
		"slot": 3,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:70",
		"is_connected": true,
		"state": "STOPPED"
	},{
		"id": "iface4",
		"lab_id": "lab1",
		"node": "node1",
		"label": "eth4",
		"slot": 4,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:71",
		"is_connected": true,
		"state": "STOPPED"
	}]`)

	data := mr.MockRespList{
		mr.MockResp{Data: iface0, URL: `/interfaces/iface0$`},
		mr.MockResp{Data: ifaceList, URL: `/interfaces\?data=true$`},
	}
	tc.mr.SetData(data)
	wg := sync.WaitGroup{}
	lab := Lab{
		ID:    "lab1",
		Nodes: make(NodeMap),
	}
	node := Node{ID: "node1", LabID: lab.ID}
	lab.Nodes[node.ID] = &node
	tc.client.labCache[lab.ID] = &lab
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 50; i++ {
		tc.mr.Reset()
		wg.Add(2)
		go func() {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(20)))
			_ = tc.client.getInterfacesForNode(tc.ctx, &node)
			wg.Done()
		}()
		go func() {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(20)))
			iface := Interface{LabID: lab.ID, Node: node.ID, ID: "iface0"}
			tc.client.InterfaceGet(tc.ctx, &iface)
			wg.Done()
		}()
		wg.Wait()
	}
}
