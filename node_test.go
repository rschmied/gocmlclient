package cmlclient

import (
	"context"
	"encoding/json"
	"testing"

	mr "github.com/rschmied/mockresponder"
	"github.com/stretchr/testify/assert"
)

var (
	node1 = []byte(`{
		"id": "node1",
		"lab_id": "lab1",
		"label": "alpine-0",
		"node_definition": "alpine",
		"state": "STARTED",
		"tags": [ "tag1", "tag2" ]
	}`)

	node2 = []byte(`{
		"id": "node2",
		"lab_id": "lab1",
		"label": "alpine-1",
		"node_definition": "alpine",
		"state": "STOPPED"
	}`)
)

func TestClient_NodeMapMarschalJSON(t *testing.T) {
	nm := NodeMap{
		"zzz": &Node{ID: "zzz"},
		"aaa": &Node{ID: "aaa"},
	}
	b, err := nm.MarshalJSON()
	assert.NoError(t, err)
	t.Logf("%+v", string(b))

	nl := []Node{}
	err = json.Unmarshal(b, &nl)
	assert.NoError(t, err)
	assert.Equal(t, nl[0].ID, "aaa")
	assert.Equal(t, nl[1].ID, "zzz")
}

func TestClient_NodeCreate(t *testing.T) {
	tc := newAuthedTestAPIclient()

	dataWithUser := mr.MockRespList{
		// post returns a partial node object, need to update
		mr.MockResp{Data: node1},
		// patch returns just the node ID
		mr.MockResp{Data: []byte(`"node1"`)},
		// re-read returns the now patched node object
		mr.MockResp{Data: node1},
	}
	tc.mr.SetData(dataWithUser)

	tests := []struct {
		name      string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good", dataWithUser, false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := Node{LabID: "lab1", NodeDefinition: "server"}
			got, err := tc.client.NodeCreate(tc.ctx, &node)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.NodeCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Len(t, got.ID, 5)
		})
	}
}

func TestClient_NodeCreateFails(t *testing.T) {
	tc := newAuthedTestAPIclient()

	dataWithUser := mr.MockRespList{
		// post returns a partial node object, need to update
		mr.MockResp{Data: node1},
		// patch / update fails -- illegal data
		mr.MockResp{Data: []byte(`"node1"`), Code: 400},
		// results in a delete of the object
		mr.MockResp{Code: 204},
	}
	tc.mr.SetData(dataWithUser)

	node := Node{LabID: "lab1", NodeDefinition: "server"}
	_, err := tc.client.NodeCreate(tc.ctx, &node)
	assert.NotEqual(t, err, nil)
}

func TestClient_NodeSetConfig(t *testing.T) {
	tc := newAuthedTestAPIclient()

	dataWithUser := mr.MockRespList{
		mr.MockResp{Data: []byte("\"node1\""), URL: `/labs/lab1/nodes/node1$`},
		mr.MockResp{Data: node1, URL: `/labs/lab1/nodes/node1$`},
	}

	tests := []struct {
		name          string
		configuration string
		responses     mr.MockRespList
		wantErr       bool
	}{
		{
			"good", "hostname bla", dataWithUser, false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc.mr.SetData(tt.responses)
			node := Node{LabID: "lab1", ID: "node1"}
			err := tc.client.NodeSetConfig(tc.ctx, &node, tt.configuration)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.NodeSetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.True(t, tc.mr.Empty())
		})
	}
}

func TestClient_NodeUpdate(t *testing.T) {
	tc := newAuthedTestAPIclient()

	node99bytes := []byte(`{
		"id": "node99",
		"lab_id": "lab99",
		"label": "alpine-0",
		"node_definition": "alpine",
		"state": "DEFINED_ON_CORE",
		"tags": [ "tag1", "tag2" ]
	}`)

	goodNodeData := mr.MockRespList{
		mr.MockResp{Data: []byte("\"node99\""), URL: `/labs/lab99/nodes/node99$`},
		mr.MockResp{Data: node99bytes, URL: `/labs/lab99/nodes/node99$`},
	}
	badNodeData := mr.MockRespList{
		mr.MockResp{Code: 400, URL: `/labs/lab99/nodes/node99$`},
	}

	tests := []struct {
		name      string
		node      Node
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good", Node{RAM: 512}, goodNodeData, false,
		},
		{
			"bad", Node{}, badNodeData, true,
		},
	}

	// node99 := &Node{LabID: "lab99", ID: "node99"}
	lab := &Lab{ID: "lab99", Nodes: make(NodeMap)}
	// lab.Nodes["node99"] = node99
	tc.client.labCache["lab99"] = lab

	for _, useCache := range []bool{true, false} {
		tc.client.useCache = useCache
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tc.mr.SetData(tt.responses)
				node := Node{
					LabID: "lab99", ID: "node99", X: 100, Y: 100,
					Tags: []string{"newtag"},
				}
				resultNode, err := tc.client.NodeUpdate(tc.ctx, &node)
				_ = resultNode
				if (err != nil) != tt.wantErr {
					t.Errorf("Client.NodeUpdate() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				assert.True(t, tc.mr.Empty())
			})
		}
	}
}

func TestClient_NodeFuncs(t *testing.T) {
	tc := newAuthedTestAPIclient()

	funcs := map[string]func(context.Context, *Node) error{
		"stop":    tc.client.NodeStop,
		"start":   tc.client.NodeStart,
		"destroy": tc.client.NodeDestroy,
		"wipe":    tc.client.NodeWipe,
	}

	tests := []struct {
		name      string
		responses mr.MockRespList
		wantErr   bool
	}{
		{"good", mr.MockRespList{mr.MockResp{Code: 200}}, false},
		{"bad", mr.MockRespList{mr.MockResp{Code: 404}}, true},
	}

	node99 := &Node{LabID: "lab99", ID: "node99"}
	lab := &Lab{ID: "lab99", Nodes: make(NodeMap)}
	lab.Nodes["node99"] = node99
	tc.client.labCache["lab99"] = lab

	for _, useCache := range []bool{true, false} {
		tc.client.useCache = useCache
		for tfname, tf := range funcs {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					tc.mr.SetData(tt.responses)
					err := tf(tc.ctx, &Node{ID: "node99", LabID: "lab99"})
					if (err != nil) != tt.wantErr {
						t.Errorf("%s error = %v, wantErr %v", tfname, err, tt.wantErr)
						return
					}
					assert.True(t, tc.mr.Empty())
				})
			}
		}
	}
}
