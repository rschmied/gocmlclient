package cmlclient

import (
	"context"
	"os"
	"testing"

	mr "github.com/rschmied/mockresponder"
	"github.com/stretchr/testify/assert"
)

var (
	demoLab = []byte(`{
		"state": "STOPPED",
		"created": "2022-05-11T20:36:15+00:00",
		"modified": "2022-05-11T21:23:28+00:00",
		"lab_title": "vlandrop",
		"lab_description": "",
		"lab_notes": "",
		"owner": "00000000-0000-4000-a000-000000000000",
		"owner_username": "admin",
		"node_count": 2,
		"link_count": 1,
		"id": "lab1",
		"groups": []
	}`)
	ownerUser = []byte(`{
		"id": "00000000-0000-4000-a000-000000000000",
		"created": "2022-04-29T13:44:46+00:00",
		"modified": "2022-05-20T10:57:42+00:00",
		"username": "admin",
		"fullname": "",
		"email": "",
		"description": "",
		"admin": true,
		"directory_dn": "",
		"groups": [],
		"labs": ["lab1"]
	}`)
	links = []byte(`["link1"]`)
	nodes = []byte(`["node1","node2"]`)
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
	lab_layer3 = []byte(`{
		"node1": {
		  "name": "alpine-0",
		  "interfaces": {
			"52:54:00:0c:e0:69": {
			  "id": "n1i1",
			  "label": "eth0",
			  "ip4": [
				"192.168.122.173"
			  ],
			  "ip6": [
				"fe80::5054:ff:fe0c:be77"
			  ]
			}
		  }
		},
		"node2": {
		  "name": "alpine-1",
		  "interfaces": {
			"52:54:00:0c:e0:70": {
			  "id": "n2i1",
			  "label": "eth0",
			  "ip4": [
				"192.168.122.174"
			  ],
			  "ip6": [
				"fe80::5054:ff:fe0c:be88"
			  ]
			}
		  }
		}
	  }`)
	ifacesn1  = []byte(`["n1i1"]`)
	ifacesn2  = []byte(`["n2i1"]`)
	ifacen1i1 = []byte(`{
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
	ifacen2i1 = []byte(`{
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
	linkn1n2 = []byte(`{
		"id": "link1",
		"interface_a": "n1i1",
		"interface_b": "n2i1",
		"lab_id": "lab1",
		"label": "alpine-0-eth0<->alpine-1-eth0",
		"link_capture_key": "",
		"node_a": "node1",
		"node_b": "node2",
		"state": "DEFINED_ON_CORE"
	}`)
)

func TestClient_GetLab_BadAuth(t *testing.T) {
	tc := newTestAPIclient()
	response := mr.MockRespList{
		mr.MockResp{
			Data: []byte(`{"version": "2.4.0+build.1","ready": true}`),
			Code: 200,
		},
		mr.MockResp{
			Data: []byte(`{
						"description": "Not authenticated: 401 Unauthorized: No authorization token provided.",
						"code":        401
					}`),
			Code: 401,
		},
		mr.MockResp{
			Data: []byte(`"authentication failed"`),
			Code: 403,
		},
	}
	tc.mr.SetData(response)
	_, err := tc.client.LabGet(tc.ctx, "qweaa", false)
	assert.EqualError(t, err, `status: 403, "authentication failed"`)
	assert.True(t, tc.mr.Empty())
}

func TestClient_GetLab(t *testing.T) {
	tc := newAuthedTestAPIclient()

	tests := []struct {
		name      string
		responses mr.MockRespList
		deep      bool
	}{
		{
			"lab_deep",
			mr.MockRespList{
				mr.MockResp{Data: demoLab},
				mr.MockResp{Data: links, URL: `/links$`},
				mr.MockResp{Data: lab_layer3, URL: `layer3_addresses$`},
				mr.MockResp{Data: ownerUser, URL: `/users/.+$`},
				mr.MockResp{Data: nodes, URL: `/nodes$`},
				mr.MockResp{Data: node1, URL: `/nodes/node1$`},
				mr.MockResp{Data: node2, URL: `/nodes/node2$`},
				mr.MockResp{Data: ifacesn1, URL: `/node1/interfaces$`},
				mr.MockResp{Data: ifacesn2, URL: `/node2/interfaces$`},
				mr.MockResp{Data: ifacen1i1, URL: `/interfaces/n1i1$`},
				mr.MockResp{Data: ifacen2i1, URL: `/interfaces/n2i1$`},
				mr.MockResp{Data: linkn1n2, URL: `/links/link1$`},
			},
			true,
		},
		{
			"lab_shallow",
			mr.MockRespList{
				mr.MockResp{Data: demoLab},
			},
			false,
		},
	}
	for _, tt := range tests {
		tc.mr.SetData(tt.responses)
		// enforce version check
		// tc.client.versionChecked = false
		t.Run(tt.name, func(t *testing.T) {
			lab, err := tc.client.LabGet(tc.ctx, "qweaa", tt.deep)
			assert.NoError(t, err)
			assert.NotNil(t, lab)
			if tt.deep {
				assert.Len(t, lab.Links, 1)
				assert.Len(t, lab.Nodes, 2)
				if !tc.mr.Empty() {
					t.Error("not all data in mock client consumed")
				}
			}
		})
	}
}

func TestClient_ImportLab(t *testing.T) {
	tc := newAuthedTestAPIclient()

	testfile := "testdata/labimport/twonodes.yaml"
	labyaml, err := os.ReadFile(testfile)
	if err != nil {
		t.Errorf("Client.ImportLab() can't read testfile %s", testfile)
	}

	tests := []struct {
		name      string
		labyaml   string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good import",
			string(labyaml),
			// the import will also fetch the entire lab (not shallow!)
			mr.MockRespList{
				mr.MockResp{Data: []byte(`{"id": "lab-id-uuid", "warnings": [] }`)},
				mr.MockResp{Data: demoLab},
				// these responses are needed for not shallow...
				mr.MockResp{Data: links, URL: `/links$`},
				mr.MockResp{Data: lab_layer3, URL: `/layer3_addresses$`},
				mr.MockResp{Data: ownerUser, URL: `/users/.+$`},
				mr.MockResp{Data: nodes, URL: `/nodes$`},
				mr.MockResp{Data: node1, URL: `/nodes/node1$`},
				mr.MockResp{Data: node2, URL: `/nodes/node2$`},
				mr.MockResp{Data: ifacesn1, URL: `/node1/interfaces$`},
				mr.MockResp{Data: ifacesn2, URL: `/node2/interfaces$`},
				mr.MockResp{Data: ifacen1i1, URL: `/interfaces/n1i1$`},
				mr.MockResp{Data: ifacen2i1, URL: `/interfaces/n2i1$`},
				mr.MockResp{Data: linkn1n2, URL: `/links/link1$`},
			},
			false,
		},
		{
			"bad import",
			",,,", // invalid YAML
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`{
					"description": "Bad request: while parsing a block node\nexpected the node content, but found ','\n  in \"<unicode string>\", line 1, column 1:\n    ,,,\n    ^.",
					"code": 400}
					`),
					Code: 400,
				},
			},
			true,
		},
		{
			"good import, bad read",
			string(labyaml),
			// the import will also fetch the entire lab (not shallow!)
			mr.MockRespList{
				mr.MockResp{Data: []byte(`{"id": "lab-id-uuid", "warnings": [] }`)},
				mr.MockResp{Code: 404},
			},
			true,
		},
	}

	for _, tt := range tests {
		tc.mr.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			lab, err := tc.client.LabImport(tc.ctx, tt.labyaml)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.GetLab() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			assert.NotNil(t, lab)
			// TODO: when adding more tests, the node/link count needs to be
			// parametrized!!
			assert.Equal(t, lab.NodeCount, 2)
			assert.Equal(t, lab.LinkCount, 1)
		})
		if !tc.mr.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}

func TestClient_ImportLabBadAuth(t *testing.T) {
	tc := newAuthedTestAPIclient()
	tc.client.apiToken = "expiredbadtoken"
	tc.client.userpass = userPass{} // no password provided

	data := mr.MockRespList{
		mr.MockResp{
			Data: []byte(`{
				"description": "401: Unauthorized",
				"code":        401
			}`),
			Code: 401,
		},
	}
	tc.mr.SetData(data)
	lab, err := tc.client.LabImport(tc.ctx, `{}`)

	if !tc.mr.Empty() {
		t.Error("not all data in mock client consumed")
	}

	assert.NotNil(t, err)
	assert.Nil(t, lab)
	assert.EqualError(t, err, "invalid token but no credentials provided")
}

func TestClient_NodeByLabel(t *testing.T) {

	l := Lab{
		Nodes: NodeMap{
			"bla": &Node{
				Label: "test",
			},
		},
	}
	n, err := l.NodeByLabel(context.Background(), "test")
	assert.Nil(t, err)
	assert.Equal(t, "test", n.Label)

	n, err = l.NodeByLabel(context.Background(), "doesntexist")
	assert.ErrorIs(t, err, ErrElementNotFound)
	assert.Nil(t, n)
}

func TestLab_CanBeWiped(t *testing.T) {
	tests := []struct {
		name string
		lab  Lab
		want bool
	}{
		{"nonodes", Lab{}, true},
		{"oktowipe", Lab{
			Nodes: NodeMap{
				"bla": &Node{
					Label: "test",
					State: NodeStateDefined,
				},
			},
		}, true},
		{"notoktowipe", Lab{
			Nodes: NodeMap{
				"bla": &Node{
					Label: "test",
					State: NodeStateBooted,
				},
			},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lab.CanBeWiped(); got != tt.want {
				t.Errorf("Lab.CanBeWiped() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLab_Running(t *testing.T) {
	tests := []struct {
		name string
		lab  Lab
		want bool
	}{
		{"running", Lab{
			Nodes: NodeMap{
				"bla": &Node{
					Label: "test",
					State: NodeStateStarted,
				},
			},
		}, true},
		{"running 2", Lab{
			Nodes: NodeMap{
				"bla": &Node{
					Label: "test",
					State: NodeStateBooted,
				},
			},
		}, true},
		{"not running 1", Lab{
			Nodes: NodeMap{
				"bla": &Node{
					Label: "test",
					State: NodeStateStopped,
				},
			},
		}, false},
		{"not running 2", Lab{
			Nodes: NodeMap{
				"bla": &Node{
					Label: "test",
					State: NodeStateDefined,
				},
			},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lab.Running(); got != tt.want {
				t.Errorf("Lab.Running() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLab_Booted(t *testing.T) {
	tests := []struct {
		name string
		lab  Lab
		want bool
	}{
		{"not booted", Lab{
			Nodes: NodeMap{
				"bla": &Node{
					Label: "test",
					State: NodeStateStarted,
				},
				"bla2": &Node{
					Label: "test",
					State: NodeStateBooted,
				},
			},
		}, false},
		{"booted", Lab{
			Nodes: NodeMap{
				"bla": &Node{
					Label: "test",
					State: NodeStateBooted,
				},
				"bla2": &Node{
					Label: "test",
					State: NodeStateBooted,
				},
			},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lab.Booted(); got != tt.want {
				t.Errorf("Lab.Booted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_HasLabConverged(t *testing.T) {
	tc := newAuthedTestAPIclient()

	data := mr.MockRespList{mr.MockResp{Data: []byte(`true`)}}
	tc.mr.SetData(data)

	got, _ := tc.client.HasLabConverged(tc.ctx, "dummy")
	if got != true {
		t.Errorf("Client.HasLabConverged() = %v, want %v", got, true)
	}

	data = mr.MockRespList{mr.MockResp{Data: []byte(`invalid`)}}
	tc.mr.SetData(data)
	_, err := tc.client.HasLabConverged(tc.ctx, "dummy")
	assert.Error(t, err)
}

func TestClient_StartStopWipeDestroy(t *testing.T) {
	tc := newAuthedTestAPIclient()

	goodData := mr.MockRespList{
		mr.MockResp{Code: 200},
		mr.MockResp{Code: 200},
		mr.MockResp{Code: 200},
		mr.MockResp{Code: 204},
	}

	badData := mr.MockRespList{
		mr.MockResp{Code: 404},
		mr.MockResp{Code: 404},
		mr.MockResp{Code: 404},
		mr.MockResp{Code: 404},
	}

	tests := []struct {
		name string
		data mr.MockRespList
		want bool
	}{
		{"good", goodData, false},
		{"bad", badData, true},
	}

	funcs := []func(context.Context, string) error{
		tc.client.LabStart,
		tc.client.LabStop,
		tc.client.LabWipe,
		tc.client.LabDestroy,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc.mr.SetData(tt.data)
			for _, f := range funcs {
				err := f(tc.ctx, "bla")
				if tt.want {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestClient_LabGetByTitle(t *testing.T) {
	tc := newAuthedTestAPIclient()

	data := mr.MockRespList{
		mr.MockResp{
			Data: []byte(`{
				"lab_tiles": {
					"52d5c824-e10c-450a-b9c5-b700bd3bc17a": {
						"state": "DEFINED_ON_CORE",
						"created": "2022-04-29T14:17:24+00:00",
						"modified": "2022-05-04T16:43:48+00:00",
						"lab_title": "demobla",
						"lab_description": "",
						"lab_notes": "",
						"owner": "00000000-0000-4000-a000-000000000000",
						"owner_username": "admin",
						"node_count": 4,
						"link_count": 3,
						"id": "52d5c824-e10c-450a-b9c5-b700bd3bc17a",
						"groups": [],
						"topology": {
							"nodes": [],
							"interfaces": [],
							"links": []
						}
					}
				}
			}`),
		},
	}

	dataWithUser := append(data, mr.MockRespList{
		mr.MockResp{Data: links, URL: `/links$`},
		mr.MockResp{Data: lab_layer3, URL: `layer3_addresses$`},
		mr.MockResp{Data: ownerUser, URL: `/users/.+$`},
		mr.MockResp{Data: nodes, URL: `/nodes$`},
		mr.MockResp{Data: node1, URL: `/nodes/node1$`},
		mr.MockResp{Data: node2, URL: `/nodes/node2$`},
		mr.MockResp{Data: ifacesn1, URL: `/node1/interfaces$`},
		mr.MockResp{Data: ifacesn2, URL: `/node2/interfaces$`},
		mr.MockResp{Data: ifacen1i1, URL: `/interfaces/n1i1$`},
		mr.MockResp{Data: ifacen2i1, URL: `/interfaces/n2i1$`},
		mr.MockResp{Data: linkn1n2, URL: `/links/link1$`},
	}...)

	tests := []struct {
		name  string
		title string
		data  mr.MockRespList
		deep  bool
		want  bool
	}{
		{"good", "demobla", data, false, false},
		{"notfound", "doesntexist", data, false, true},
		{"error", "doesntexist", mr.MockRespList{mr.MockResp{Err: ErrElementNotFound}}, false, true},
		{"fullgood", "demobla", dataWithUser, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc.mr.SetData(tt.data)
			_, err := tc.client.LabGetByTitle(tc.ctx, tt.title, tt.deep)
			if tt.want {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestClient_LabCreate(t *testing.T) {
	tc := newAuthedTestAPIclient()

	data := mr.MockRespList{
		mr.MockResp{
			Data: []byte(`{
				"state": "DEFINED_ON_CORE",
				"created": "2022-10-14T10:05:07+00:00",
				"modified": "2022-10-14T10:05:07+00:00",
				"lab_title": "Lab at Mon 17:27 PM",
				"lab_description": "string",
				"lab_notes": "string",
				"owner": "00000000-0000-4000-a000-000000000000",
				"owner_username": "admin",
				"node_count": 0,
				"link_count": 0,
				"id": "375b41ae-dd90-41a2-858d-98948abbbd38",
				"groups": []
			}`),
		},
	}

	tests := []struct {
		name string
		lab  Lab
		data mr.MockRespList
		want bool
	}{
		{"good", Lab{}, data, false},
		{"bad", Lab{}, mr.MockRespList{mr.MockResp{Code: 405}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc.mr.SetData(tt.data)
			_, err := tc.client.LabCreate(tc.ctx, tt.lab)
			if tt.want {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_LabUpdate(t *testing.T) {
	tc := newAuthedTestAPIclient()

	data := mr.MockRespList{
		mr.MockResp{
			Data: []byte(`{
				"state": "DEFINED_ON_CORE",
				"created": "2022-10-14T10:05:07+00:00",
				"modified": "2022-10-14T10:05:07+00:00",
				"lab_title": "Lab at Mon 17:27 PM",
				"lab_description": "string",
				"lab_notes": "string",
				"owner": "00000000-0000-4000-a000-000000000000",
				"owner_username": "admin",
				"node_count": 0,
				"link_count": 0,
				"id": "375b41ae-dd90-41a2-858d-98948abbbd38",
				"groups": []
			}`),
		},
	}

	tests := []struct {
		name string
		lab  Lab
		data mr.MockRespList
		want bool
	}{
		{"good", Lab{}, data, false},
		{"bad", Lab{}, mr.MockRespList{mr.MockResp{Code: 405}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc.mr.SetData(tt.data)
			_, err := tc.client.LabUpdate(tc.ctx, tt.lab)
			if tt.want {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_CompleteCache(t *testing.T) {
	tc := newAuthedTestAPIclient()
	tc.client.useCache = true

	ifacen1i2 := []byte(`{
		"id": "n1i2",
		"lab_id": "lab1",
		"node": "node1",
		"label": "eth1",
		"slot": 1,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:69",
		"is_connected": true,
		"state": "STARTED"
	}`)
	ifacen2i2 := []byte(`{
		"id": "n2i2",
		"lab_id": "lab1",
		"node": "node2",
		"label": "eth1",
		"slot": 1,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:70",
		"is_connected": true,
		"state": "STOPPED"
	}`)
	link2n1n2 := []byte(`{
		"id": "link2",
		"interface_a": "n1i2",
		"interface_b": "n2i2",
		"lab_id": "lab1",
		"label": "alpine-0-eth1<->alpine-1-eth1",
		"link_capture_key": "",
		"node_a": "node1",
		"node_b": "node2",
		"state": "DEFINED_ON_CORE"
	}`)

	data := mr.MockRespList{
		mr.MockResp{Data: demoLab, URL: `/labs/lab1$`},
		mr.MockResp{Data: links, URL: `/links$`},
		mr.MockResp{Data: lab_layer3, URL: `layer3_addresses$`},
		mr.MockResp{Data: ownerUser, URL: `/users/.+$`},
		mr.MockResp{Data: nodes, URL: `/nodes$`},
		mr.MockResp{Data: node1, URL: `/nodes/node1$`},
		mr.MockResp{Data: node2, URL: `/nodes/node2$`},
		mr.MockResp{Data: ifacesn1, URL: `/node1/interfaces$`},
		mr.MockResp{Data: ifacesn2, URL: `/node2/interfaces$`},
		mr.MockResp{Data: ifacen1i1, URL: `/interfaces/n1i1$`},
		mr.MockResp{Data: ifacen2i1, URL: `/interfaces/n2i1$`},
		mr.MockResp{Data: linkn1n2, URL: `/links/link1$`},

		mr.MockResp{Data: ifacesn1, URL: `/node1/interfaces$`},
		mr.MockResp{Data: ifacesn2, URL: `/node2/interfaces$`},
		// mr.MockResp{Data: ifacen1i1, URL: `/interfaces/n1i1$`},
		// mr.MockResp{Data: ifacen2i1, URL: `/interfaces/n2i1$`},

		// 2 new interfaces, one new link, followed by a get
		mr.MockResp{Data: ifacen1i2, URL: `/interfaces$`},
		mr.MockResp{Data: ifacen2i2, URL: `/interfaces$`},
		mr.MockResp{Data: link2n1n2, URL: `/links$`},
		mr.MockResp{Data: link2n1n2, URL: `/links/link2$`},

		// mr.MockResp{Data: ifacen1i2, URL: `/interfaces/n1i2$`},
		// mr.MockResp{Data: ifacen2i2, URL: `/interfaces/n2i2$`},
		// mr.MockResp{Data: node1, URL: `/nodes/node1$`},
		// mr.MockResp{Data: node2, URL: `/nodes/node2$`},
	}

	tests := []struct {
		name string
		lab  Lab
		data mr.MockRespList
		want bool
	}{
		{"good", Lab{}, data, false},
		// {"bad", Lab{}, mr.MockRespList{mr.MockResp{Code: 405}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc.mr.SetData(tt.data)
			_, err := tc.client.LabGet(tc.ctx, "lab1", true)
			assert.NoError(t, err)
			// _, err := tc.client.LabGet(ctx, "labuuid", false)
			// _, err := tc.client.LabUpdate(tc.ctx, tt.lab)
			link := &Link{LabID: "lab1", SrcNode: "node1", DstNode: "node2"}
			link, err = tc.client.LinkCreate(tc.ctx, link)
			assert.Equal(t, "link2", link.ID)
			if tt.want {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.True(t, tc.mr.Empty())
		})
	}
}
