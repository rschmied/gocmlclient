package cmlclient

import (
	"testing"

	mr "github.com/rschmied/mockresponder"
	"github.com/stretchr/testify/assert"
)

func TestClient_VersionCheck(t *testing.T) {
	c := New("https://bla.bla", true, useCache)
	mrClient, ctx := mr.NewMockResponder()
	c.httpClient = mrClient
	c.state.set(stateAuthenticated)

	tests := []struct {
		name     string
		wantJSON string
		wantErr  bool
	}{
		{"too old", `{"version": "2.1.0","ready": true}`, true},
		{"garbage", `{"version": "garbage","ready": true}`, true},
		{"too new", `{"version": "2.35.0","ready": true}`, true},
		{"perfect", `{"version": "2.4.0","ready": true}`, false},
		{"actual", `{"version": "2.4.0+build.1","ready": true}`, false},
		{"newer", `{"version": "2.4.1","ready": true}`, false},
		{"dev", `{"version": "2.5.0-dev0+build.3.2f7875762","ready": true}`, false},
		{"v2.5.0", `{"version": "2.5.0+build.5","ready": true}`, false},
	}
	for _, tt := range tests {
		mrClient.SetData(mr.MockRespList{{Data: []byte(tt.wantJSON)}})
		t.Run(tt.name, func(t *testing.T) {
			if err := c.versionCheck(ctx, 0); (err != nil) != tt.wantErr {
				t.Errorf("Client.VersionCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		if !mrClient.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}

func TestClient_NotReady(t *testing.T) {
	c := New("https://bla.bla", true, useCache)
	mrClient, ctx := mr.NewMockResponder()
	c.httpClient = mrClient
	c.state.set(stateAuthenticated)

	mrClient.SetData(mr.MockRespList{
		{Data: []byte(`{"version": "2.5.0","ready": false}`)},
	})

	err := c.Ready(ctx)
	assert.Error(t, err)

	if !mrClient.Empty() {
		t.Error("not all data in mock client consumed")
	}
}

func TestClient_Version(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{"empty", "", ""},
		{"some", "some", "some"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				version: tt.version,
			}
			if got := c.Version(); got != tt.want {
				t.Errorf("Client.Version() = %v, want %v", got, tt.want)
			}
		})
	}
}
