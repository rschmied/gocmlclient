package cmlclient

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	mr "github.com/rschmied/mockresponder"
	"github.com/stretchr/testify/assert"
)

var (
	group1 = `{
		"name": "CCNA Study Group Class of 21",
		"description": "string",
		"members": [
			"90f84e38-a71c-4d57-8d90-00fa8a197385",
			"60f84e39-ffff-4d99-8a78-00fa8aaf5666"
		],
		"labs": [
			{
			"id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
			"permission": "read_only"
			}
		],
		"id": "85401911-851f-4e6a-b5c3-4aa1d91fa21d",
		"created": "2021-02-28T07:33:47+00:00",
		"modified": "2021-02-28T07:33:47+00:00"
	}`
	group2 = `{
		"name": "CCNA Study Group Class of 21",
		"description": "string",
		"members": [
			"90f84e38-a71c-4d57-8d90-00fa8a197385",
			"ba2ef97f-a42f-42fb-8c64-17f66fccff0a"
		],
		"labs": [
			{
			"id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
			"permission": "read_only"
			}
		],
		"id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
		"created": "2021-02-28T07:33:47+00:00",
		"modified": "2021-02-28T07:33:47+00:00"
	}`
)

func TestClient_GetGroups(t *testing.T) {

	tc := newAuthedTestAPIclient()

	tests := []struct {
		name      string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good",
			mr.MockRespList{
				mr.MockResp{
					// need this to go through the sorting in the client
					// (highest -> lowest, e.g. 3.15  before 3.14)
					Data: []byte("[" + group2 + "," + group1 + "]"),
				},
			},
			false,
		},
		{
			"bad",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`"something failed!`),
					Code: http.StatusInternalServerError,
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		tc.mr.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			got, err := tc.client.Groups(tc.ctx)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.GetImageDefs() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			expected := GroupList{}
			// this is properly sorted:
			b := bytes.NewReader([]byte("[" + group2 + "," + group1 + "]"))
			err = json.NewDecoder(b).Decode(&expected)
			if err != nil {
				t.Error("bad test data")
				return
			}
			if !reflect.DeepEqual(got, expected) {
				t.Errorf("Client.GetGroups() = %v, want %v", got, expected)
			}
		})
		if !tc.mr.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}

func TestClient_GetGroupByName(t *testing.T) {

	tc := newAuthedTestAPIclient()

	tests := []struct {
		name      string
		groupname string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good",
			"students",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(group1),
				},
			},
			false,
		},
		{
			"bad",
			"noproblemo",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`"something failed!`),
					Code: http.StatusInternalServerError,
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		tc.mr.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			got, err := tc.client.GroupByName(tc.ctx, tt.groupname)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.GetGroupByName() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			expected := Group{}
			b := bytes.NewReader([]byte(group1))
			err = json.NewDecoder(b).Decode(&expected)
			if err != nil {
				t.Error("bad test data")
				return
			}
			assert.Equal(t, got, &expected)
		})
		if !tc.mr.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}

func TestClient_GetGroup(t *testing.T) {

	tc := newAuthedTestAPIclient()

	// ID of group1 from above
	uuid := "85401911-851f-4e6a-b5c3-4aa1d91fa21d"

	tests := []struct {
		name      string
		groupid   string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good",
			uuid,
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(group1),
				},
			},
			false,
		},
		{
			"bad",
			uuid,
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`"something failed!`),
					Code: http.StatusInternalServerError,
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		tc.mr.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			got, err := tc.client.GroupGet(tc.ctx, tt.groupid)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.GetGroup() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			expected := Group{}
			b := bytes.NewReader([]byte(group1))
			err = json.NewDecoder(b).Decode(&expected)
			if err != nil {
				t.Error("bad test data")
				return
			}
			assert.Equal(t, got, &expected)
		})
		if !tc.mr.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}

func TestClient_GroupDestroy(t *testing.T) {
	tc := newAuthedTestAPIclient()
	uuid := "85401911-851f-4e6a-b5c3-4aa1d91fa21d"
	tc.mr.SetData(mr.MockRespList{
		mr.MockResp{
			Code: http.StatusNoContent,
		},
	})
	err := tc.client.GroupDestroy(tc.ctx, uuid)
	assert.NoError(t, err)
	if !tc.mr.Empty() {
		t.Error("not all data in mock client consumed")
	}
}

func TestClient_GroupCreate(t *testing.T) {

	tc := newAuthedTestAPIclient()
	newGroup := Group{Name: "doesntmatter"}

	tests := []struct {
		name      string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(group1),
				},
			},
			false,
		},
		{
			"bad",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`"something failed!`),
					Code: http.StatusInternalServerError,
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		tc.mr.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			got, err := tc.client.GroupCreate(tc.ctx, &newGroup)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.GroupCreate() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			expected := Group{}
			b := bytes.NewReader([]byte(group1))
			err = json.NewDecoder(b).Decode(&expected)
			if err != nil {
				t.Error("bad test data")
				return
			}
			assert.Equal(t, got, &expected)
		})
		if !tc.mr.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}

func TestClient_GroupUpdate(t *testing.T) {

	tc := newAuthedTestAPIclient()
	newGroup := Group{Name: "doesntmatter"}

	tests := []struct {
		name      string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(group1),
				},
			},
			false,
		},
		{
			"bad",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`"something failed!`),
					Code: http.StatusInternalServerError,
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		tc.mr.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			got, err := tc.client.GroupUpdate(tc.ctx, &newGroup)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.GroupUpdate() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			expected := Group{}
			b := bytes.NewReader([]byte(group1))
			err = json.NewDecoder(b).Decode(&expected)
			if err != nil {
				t.Error("bad test data")
				return
			}
			assert.Equal(t, got, &expected)
		})
		if !tc.mr.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}
