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
	user1 = `{
		"id": "cc42bd56-1dc6-445c-b7e7-569b0a8b0c94",
		"created": "2023-02-15T10:23:22+00:00",
		"modified": "2023-02-15T10:33:36+00:00",
		"username": "student_1",
		"fullname": "",
		"email": "",
		"lab_description": "",
		"admin": false,
		"directory_dn": "",
		"groups": [
			"bc9b796e-48bc-4369-b131-231dfa057d41"
		],
		"labs": [],
		"opt_in": false,
		"resource_pool": "79cd2ede-abf8-4041-a498-ed40ea75f6a1"
	}`
	user2 = `{
		"id": "d7dd70df-59db-4417-8362-04917b8c5d2f",
		"created": "2023-02-15T10:23:35+00:00",
		"modified": "2023-02-15T10:33:01+00:00",
		"username": "student_2",
		"fullname": "",
		"email": "",
		"lab_description": "",
		"admin": false,
		"directory_dn": "",
		"groups": [
			"bc9b796e-48bc-4369-b131-231dfa057d41"
		],
		"labs": [],
		"opt_in": false,
		"resource_pool": "f36f22d4-be5b-4589-a5d8-8a9bfa51fc49"	}`
)

func TestClient_GetUsers(t *testing.T) {

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
					Data: []byte("[" + user2 + "," + user1 + "]"),
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
			got, err := tc.client.Users(tc.ctx)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.GetUsers() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			expected := UserList{}
			// this is properly sorted:
			b := bytes.NewReader([]byte("[" + user2 + "," + user1 + "]"))
			err = json.NewDecoder(b).Decode(&expected)
			if err != nil {
				t.Error("bad test data")
				return
			}
			if !reflect.DeepEqual(got, expected) {
				t.Errorf("Client.GetUsers() = %v, want %v", got, expected)
			}
		})
		if !tc.mr.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}

func TestClient_GetUserByName(t *testing.T) {

	tc := newAuthedTestAPIclient()

	tests := []struct {
		name      string
		username  string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good",
			"student_1",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`"cc42bd56-1dc6-445c-b7e7-569b0a8b0c94"`),
				},
				mr.MockResp{
					Data: []byte(user1),
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
			got, err := tc.client.UserByName(tc.ctx, tt.username)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.GetUserByName() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			expected := User{}
			b := bytes.NewReader([]byte(user1))
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

func TestClient_GetUser(t *testing.T) {

	tc := newAuthedTestAPIclient()

	// ID of user1 from above
	uuid := "85401911-851f-4e6a-b5c3-4aa1d91fa21d"

	tests := []struct {
		name      string
		userid    string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good",
			uuid,
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(user1),
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
			got, err := tc.client.UserGet(tc.ctx, tt.userid)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.GetUser() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			expected := User{}
			b := bytes.NewReader([]byte(user1))
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

func TestClient_UserDestroy(t *testing.T) {
	tc := newAuthedTestAPIclient()
	uuid := "85401911-851f-4e6a-b5c3-4aa1d91fa21d"
	tc.mr.SetData(mr.MockRespList{
		mr.MockResp{
			Code: http.StatusNoContent,
		},
	})
	err := tc.client.UserDestroy(tc.ctx, uuid)
	assert.NoError(t, err)
	if !tc.mr.Empty() {
		t.Error("not all data in mock client consumed")
	}
}

func TestClient_UserCreate(t *testing.T) {

	tc := newAuthedTestAPIclient()
	newUser := User{Username: "doesntmatter"}

	tests := []struct {
		name      string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(user1),
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
			got, err := tc.client.UserCreate(tc.ctx, &newUser)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.UserCreate() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			expected := User{}
			b := bytes.NewReader([]byte(user1))
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

func TestClient_UserUpdate(t *testing.T) {

	tc := newAuthedTestAPIclient()
	newUser := User{Username: "doesntmatter"}

	tests := []struct {
		name      string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(user1),
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
			got, err := tc.client.UserUpdate(tc.ctx, &newUser)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.UserUpdate() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			expected := User{}
			b := bytes.NewReader([]byte(user1))
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

func TestClient_UserGroups(t *testing.T) {

	tc := newAuthedTestAPIclient()

	tests := []struct {
		name      string
		userID    string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good",
			"cc42bd56-1dc6-445c-b7e7-569b0a8b0c94",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`[
						"85401911-851f-4e6a-b5c3-4aa1d91fa21d",					
						"90f84e38-a71c-4d57-8d90-00fa8a197385"
					]`),
				},
				mr.MockResp{
					Data: []byte(group1),
				},
				mr.MockResp{
					Data: []byte(group2),
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
			got, err := tc.client.UserGroups(tc.ctx, tt.userID)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.GetUserGroups() error = %v, wantErr %v", err, tt.wantErr)
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

	// "id": "85401911-851f-4e6a-b5c3-4aa1d91fa21d",
	// "id": "90f84e38-a71c-4d57-8d90-00fa8a197385",

	// "id": "cc42bd56-1dc6-445c-b7e7-569b0a8b0c94",
}
