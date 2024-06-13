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
	for _, tt := range tests {
		tc.mr.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			link, err := tc.client.LinkGet(tc.ctx, "qweaa", "link1", tt.deep)
			assert.NoError(t, err)
			assert.NotNil(t, link)
			if !tc.mr.Empty() {
				t.Error("not all data in mock client consumed")
			}
		})
	}
}

func TestClient_CreateLink(t *testing.T) {
	tc := newAuthedTestAPIclient()

	// same as in ifacelist1
	iface1 := []byte(`{
		"id": "iface1",
		"lab_id": "lab1",
		"node": "node1",
		"label": "eth1",
		"slot": 1,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:71",
		"is_connected": false,
		"state": "STOPPED"
	}`)

	ifaceList1 := []byte(`[{
		"id": "iface0",
		"lab_id": "lab1",
		"node": "node1",
		"label": "lo0",
		"slot": 0,
		"type": "loopback",
		"mac_address": "52:54:00:0c:e0:70",
		"is_connected": false,
		"state": "STOPPED"
	},{
		"id": "iface1",
		"lab_id": "lab1",
		"node": "node1",
		"label": "eth1",
		"slot": 1,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:71",
		"is_connected": false,
		"state": "STOPPED"
	},{
		"id": "iface2",
		"lab_id": "lab1",
		"node": "node1",
		"label": "eth2",
		"slot": 2,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:72",
		"is_connected": true,
		"state": "STOPPED"
	}]`)

	// same as in ifacelist2
	iface6 := []byte(`{
		"id": "iface2",
		"lab_id": "lab1",
		"node": "node2",
		"label": "eth2",
		"slot": 2,
		"type": "physical",
		"mac_address": "52:54:00:0c:e1:72",
		"is_connected": false,
		"state": "STOPPED"
	}`)

	ifaceList2 := []byte(`[{
		"id": "iface4",
		"lab_id": "lab1",
		"node": "node2",
		"label": "eth0",
		"slot": 0,
		"type": "physical",
		"mac_address": "52:54:00:0c:e1:70",
		"is_connected": true,
		"state": "STOPPED"
	},{
		"id": "iface5",
		"lab_id": "lab1",
		"node": "node2",
		"label": "eth1",
		"slot": 1,
		"type": "physical",
		"mac_address": "52:54:00:0c:e1:71",
		"is_connected": true,
		"state": "STOPPED"
	},{
		"id": "iface6",
		"lab_id": "lab1",
		"node": "node2",
		"label": "eth2",
		"slot": 2,
		"type": "physical",
		"mac_address": "52:54:00:0c:e1:72",
		"is_connected": false,
		"state": "STOPPED"
	}]`)

	linkn1n2 = []byte(`{
		"id": "link1",
		"interface_a": "iface1",
		"interface_b": "iface6",
		"lab_id": "lab1",
		"label": "alpine-0-eth0<->alpine-1-eth0",
		"link_capture_key": "",
		"node_a": "node1",
		"node_b": "node2",
		"state": "DEFINED_ON_CORE"
	}`)

	newiface0 := []byte(`{
		"id": "iface0",
		"lab_id": "lab1",
		"node": "node1",
		"label": "eth0",
		"slot": 0,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:70",
		"is_connected": false,
		"state": "STOPPED"
	}`)

	newiface1 := []byte(`{
		"id": "iface0",
		"lab_id": "lab1",
		"node": "node1",
		"label": "eth0",
		"slot": 0,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:70",
		"is_connected": false,
		"state": "STOPPED"
	}`)

	mockdata := mr.MockRespList{
		mr.MockResp{Data: ifaceList1, URL: `node1/interfaces\?data=true$`},
		mr.MockResp{Data: ifaceList2, URL: `node2/interfaces\?data=true$`},
		mr.MockResp{Data: linkn1n2, URL: `/links$`},
		mr.MockResp{Data: linkn1n2, URL: `/links/link1$`},
		mr.MockResp{Data: iface1, URL: `/interfaces/iface1$`},
		mr.MockResp{Data: iface6, URL: `/interfaces/iface6$`},
		mr.MockResp{Data: node1, URL: `/labs/lab1/nodes/node1$`},
		mr.MockResp{Data: node1, URL: `/labs/lab1/nodes/node2$`},
	}

	mockdata2 := mr.MockRespList{
		mr.MockResp{Data: []byte(`[]`), URL: `node1/interfaces\?data=true$`},
		mr.MockResp{Data: []byte(`[]`), URL: `node2/interfaces\?data=true$`},
		mr.MockResp{Data: newiface0, URL: `/labs/lab1/interfaces$`},
		mr.MockResp{Data: newiface1, URL: `/labs/lab1/interfaces$`},
		mr.MockResp{Data: linkn1n2, URL: `/links$`},
		mr.MockResp{Data: linkn1n2, URL: `/links/link1$`},
		mr.MockResp{Data: iface1, URL: `/interfaces/iface1$`},
		mr.MockResp{Data: iface6, URL: `/interfaces/iface6$`},
		mr.MockResp{Data: node1, URL: `/labs/lab1/nodes/node1$`},
		mr.MockResp{Data: node1, URL: `/labs/lab1/nodes/node2$`},
	}

	tests := []struct {
		name      string
		link      *Link
		responses mr.MockRespList
	}{
		{
			"link_with_slots",
			&Link{
				LabID:   "lab1",
				SrcNode: "node1",
				DstNode: "node2",
				SrcSlot: 1,
				DstSlot: 2,
			},
			mockdata,
		},
		{
			"link_no_slots_1",
			&Link{
				LabID:   "lab1",
				SrcNode: "node1",
				DstNode: "node2",
				SrcSlot: -1,
				DstSlot: -1,
			},
			mockdata,
		},
		{
			"link_no_slots_2",
			&Link{
				LabID:   "lab1",
				SrcNode: "node1",
				DstNode: "node2",
				SrcSlot: -1,
				DstSlot: -1,
			},
			mockdata2,
		},
	}
	for _, tt := range tests {
		tc.mr.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			lab, err := tc.client.LinkCreate(tc.ctx, tt.link)
			assert.NoError(t, err)
			assert.NotNil(t, lab)
			if !tc.mr.Empty() {
				t.Error("not all data in mock client consumed")
			}
		})
	}
}

func TestClient_DestroyLink(t *testing.T) {
	tc := newAuthedTestAPIclient()

	tc.mr.SetData(mr.MockRespList{
		mr.MockResp{Code: 200},
	})

	err := tc.client.LinkDestroy(tc.ctx, &Link{LabID: "lab1", ID: "link1"})
	assert.NoError(t, err)
}
