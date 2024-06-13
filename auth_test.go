package cmlclient

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	mr "github.com/rschmied/mockresponder"
	"github.com/stretchr/testify/assert"
)

func TestClient_authenticate(t *testing.T) {
	tc := newTestAPIclient()
	tc.client.state.set(stateAuthRequired)
	tc.client.userpass = userPass{"qwe", "qwe"}

	tests := []struct {
		name      string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`{
						"description": "Not authenticated: 401 Unauthorized: No authorization token provided.",
						"code":        401
					}`),
					Code: 401,
				},
				mr.MockResp{
					Data: []byte(`{"username": "qwe", "id": "008", "token": "secret" }`),
				},
				// mr.MockResp{
				// 	Data: []byte(`"OK"`),
				// },
				mr.MockResp{
					Data: []byte(`{"username": "qwe", "id": "008", "token": "secret" }`),
				},
			},
			false,
		},
		{
			"bad",
			mr.MockRespList{
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
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc.mr.SetData(tt.responses)
			if err := tc.client.authenticate(tc.ctx, tc.client.userpass, 0); (err != nil) != tt.wantErr {
				t.Errorf("Client.authenticate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		if !tc.mr.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}

func TestClient_token_auth(t *testing.T) {
	tc := newAuthedTestAPIclient()
	tc.client.apiToken = "sometoken"
	tc.client.userpass = userPass{}

	tests := []struct {
		name      string
		responses mr.MockRespList
		wantErr   bool
		errstr    string
	}{
		{
			"goodtoken",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`{"version": "2.4.1","ready": true}`),
					Code: 200,
				},
			},
			false,
			"",
		},
		{
			"badjson",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`,,,`),
					Code: 200,
				},
			},
			true,
			"invalid character ',' looking for beginning of value",
		},
		{
			"badtoken",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`{
						"description": "No authorization token provided.",
						"code": 401
					}`),
					Code: 401,
				},
			},
			true,
			"invalid token but no credentials provided",
		},
		{
			"clienterror",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte{},
					Err:  errors.New("ka-boom"),
				},
			},
			true,
			"ka-boom",
		},
	}
	for _, tt := range tests {
		tc.mr.SetData(tt.responses)
		var err error
		t.Run(tt.name, func(t *testing.T) {
			if err = tc.client.versionCheck(tc.ctx, 0); (err != nil) != tt.wantErr {
				t.Errorf("Client.versionCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		if !tc.mr.Empty() {
			t.Error("not all data in mock client consumed")
		}
		if tt.wantErr {
			assert.ErrorContains(t, err, tt.errstr)
		}
	}
}

func TestClient_SetToken(t *testing.T) {
	c := New("https://bla.bla", true)
	c.SetToken("qwe")
	assert.Equal(t, "qwe", c.apiToken)
}

func TestClient_SetUsernamePassword(t *testing.T) {
	c := New("https://bla.bla", true)
	c.SetUsernamePassword("user", "pass")
	assert.Equal(t, "user", c.userpass.Username)
	assert.Equal(t, "pass", c.userpass.Password)
}

func TestClient_SetCACert(t *testing.T) {
	c := New("https://bla.bla", true)
	err := c.SetCACert([]byte("crapdata"))
	assert.EqualError(t, err, "failed to parse root certificate")
	testCA := "testdata/ca.pem"
	certPEMBlock, err := os.ReadFile(testCA)
	assert.NoError(t, err)
	err = c.SetCACert(certPEMBlock)
	assert.NoError(t, err)

	// can't use this with a mock responder
	mrClient, _ := mr.NewMockResponder()
	c.httpClient = mrClient
	err = c.SetCACert(certPEMBlock)
	assert.Error(t, err)
}

func TestClient_complete(t *testing.T) {
	tc := newTestAPIclient()

	tc.mr.SetData(mr.MockRespList{
		mr.MockResp{
			Data: []byte(`{"version": "2.4.1","ready": true}`),
			URL:  "system_information$",
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
			Data: []byte(`{"username": "qwe", "id": "008", "token": "secret" }`),
			URL:  "auth_extended$",
		},
		mr.MockResp{
			URL:  "labs/bla$",
			Code: 200,
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
		mr.MockResp{
			Data: []byte(`"OK"`), URL: "authok$",
		},
	})

	_, err := tc.client.LabGet(tc.ctx, "bla", false)
	assert.NoError(t, err)
	assert.True(t, tc.mr.Empty())
}

func TestClient_Race(t *testing.T) {
	// GOFLAGS="-count=1000" time go test -race -parallel 2  . -run 'Race$'
	t.Parallel()

	tc := newTestAPIclient()

	tc.mr.SetData(mr.MockRespList{
		mr.MockResp{
			Data: []byte(`{"version": "2.4.1","ready": true}`),
			URL:  "system_information$",
			Code: 200,
		},
		mr.MockResp{
			Data: []byte(`{"description": "Not authorized", "code": 401}`),
			Code: 401,
		},
		mr.MockResp{
			Data: []byte(`{"username": "qwe", "id": "008", "token": "secret" }`),
			URL:  "auth_extended$",
		},
		mr.MockResp{
			URL:  "labs/bla1$",
			Code: 200,
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
		mr.MockResp{
			URL:  "labs/bla2$",
			Code: 200,
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
		mr.MockResp{
			Data: []byte(`"OK"`), URL: "authok$",
		},
	})

	done := false
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		_, err := tc.client.LabGet(tc.ctx, "bla1", false)
		assert.NoError(t, err)
		wg.Done()
	}()

	go func() {
		_, err := tc.client.LabGet(tc.ctx, "bla2", false)
		assert.NoError(t, err)
		wg.Done()
	}()

	go func() {
		wg.Wait()
		mu.Lock()
		done = true
		mu.Unlock()
	}()

	doneCheck := func() bool {
		mu.Lock()
		defer mu.Unlock()
		return done
	}

	assert.Eventually(t, doneCheck, time.Second*2, time.Microsecond*50)
	assert.True(t, tc.mr.Empty())
}
