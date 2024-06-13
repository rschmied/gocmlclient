package cmlclient

import (
	"context"
	"testing"

	mr "github.com/rschmied/mockresponder"
	"github.com/stretchr/testify/assert"
)

type testClient struct {
	client *Client
	mr     *mr.MockResponder
	ctx    context.Context
}

func newTestAPIclient() testClient {
	c := New("https://controller", true)
	mrClient, ctx := mr.NewMockResponder()
	c.httpClient = mrClient
	c.SetUsernamePassword("user", "pass")
	return testClient{c, mrClient, ctx}
}

func newAuthedTestAPIclient() testClient {
	c := newTestAPIclient()
	c.client.state.set(stateAuthenticated)
	return c
}

func TestClient_methoderror(t *testing.T) {
	c := New("", true)
	err := c.jsonReq(context.Background(), "ü", "###", nil, nil, 0)
	assert.Error(t, err)
}

func TestClient_compatibility(t *testing.T) {
	response := mr.MockRespList{
		mr.MockResp{Data: []byte(`{"version": "2.1.0","ready": true}`)},
	}

	tc := newTestAPIclient()
	tc.mr.SetData(response)

	err := tc.client.jsonGet(tc.ctx, "/api/v0/labs", nil, 0)
	assert.Error(t, err)
}

func TestClient_putpatch(t *testing.T) {
	putResponse := mr.MockRespList{
		mr.MockResp{Code: 204},
	}

	patchResponse := mr.MockRespList{
		mr.MockResp{Data: []byte("\"OK\"")},
	}

	tc := newAuthedTestAPIclient()
	tc.mr.SetData(putResponse)

	err := tc.client.jsonPut(tc.ctx, "###", 0)
	assert.NoError(t, err)

	tc.mr.SetData(patchResponse)
	var result string
	err = tc.client.jsonPatch(tc.ctx, "###", nil, &result, 0)
	assert.NoError(t, err)
	assert.Equal(t, result, "OK")
}
