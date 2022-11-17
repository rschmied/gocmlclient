package cmlclient

import (
	"encoding/json"
	"testing"

	mr "github.com/rschmied/mockresponder"
	"github.com/stretchr/testify/assert"
)

func TestClient_LinkMapMarshalJSON(t *testing.T) {
	ll := linkList{
		&Link{
			ID: "zzz",
			ifaceA: &Interface{
				ID:   "za",
				IP4:  []string{},
				IP6:  []string{},
				node: &Node{ID: "za"},
			},
			ifaceB: &Interface{
				ID:   "zb",
				IP4:  []string{},
				IP6:  []string{},
				node: &Node{ID: "zb"},
			},
		},
		&Link{
			ID: "aaa",
			ifaceA: &Interface{
				ID:   "aa",
				IP4:  []string{},
				IP6:  []string{},
				node: &Node{ID: "aa"},
			},
			ifaceB: &Interface{
				ID:   "ab",
				IP4:  []string{},
				IP6:  []string{},
				node: &Node{ID: "bb"},
			},
		},
		// &Link{ID: "aaa", ifaceA: &Interface{ID: "aa"}, ifaceB: &Interface{ID: "ab"}},
	}
	_ = ll
	b, err := ll.MarshalJSON()
	assert.NoError(t, err)
	t.Logf("%+v", string(b))

	nl := []Link{}
	err = json.Unmarshal(b, &nl)
	assert.NoError(t, err)
	assert.Equal(t, nl[0].ID, "aaa")
	assert.Equal(t, nl[1].ID, "zzz")
}

func TestClient_GetLink(t *testing.T) {
	tc := newAuthedTestAPIclient()

	ifacen1i1 := []byte(`{
		"id": "n1i1",
		"lab_id": "lab1",
		"node": "node1",
		"label": "eth0",
		"slot": 0,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:69",
		"is_connected": true,
		"state": "STARTED"
	}`)
	ifacen2i1 := []byte(`{
		"id": "n2i1",
		"lab_id": "lab1",
		"node": "node2",
		"label": "eth0",
		"slot": 0,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:70",
		"is_connected": true,
		"state": "STOPPED"
	}`)

	tests := []struct {
		name      string
		responses mr.MockRespList
		deep      bool
	}{
		{
			"link_deep",
			mr.MockRespList{
				mr.MockResp{Data: linkn1n2, URL: `/links/link1$`},
				mr.MockResp{Data: ifacen1i1, URL: `/interfaces/n1i1$`},
				mr.MockResp{Data: ifacen2i1, URL: `/interfaces/n2i1$`},
				mr.MockResp{Data: node1, URL: `/nodes/node1$`},
				mr.MockResp{Data: node2, URL: `/nodes/node2$`},
			},
			true,
		},
		{
			"link_shallow",
			mr.MockRespList{
				mr.MockResp{Data: linkn1n2, URL: `/links/link1$`},
			},
			false,
		},
	}
	for _, useCache := range []bool{false, true} {
		for _, tt := range tests {
			tc.mr.SetData(tt.responses)
			tc.client.useCache = useCache
			t.Run(tt.name, func(t *testing.T) {
				lab, err := tc.client.LinkGet(tc.ctx, "qweaa", "link1", tt.deep)
				assert.NoError(t, err)
				assert.NotNil(t, lab)
				// if tt.deep {
				// 	assert.Len(t, lab.Links, 1)
				// 	assert.Len(t, lab.Nodes, 2)
				if !(useCache || tc.mr.Empty()) {
					t.Errorf("not all data in mock client consumed: %v", useCache)
				}
				// }
			})
		}
	}
}
