package cmlclient

import (
	"testing"

	mr "github.com/rschmied/mockresponder"
	"github.com/stretchr/testify/assert"
)

func TestClient_ExtConns(t *testing.T) {
	tc := newAuthedTestAPIclient()

	extConn := []byte(`
	  {
		"id": "58568fbb-e1f8-4b83-a1f8-148c656eed39",
		"device_name": "virbr0",
		"label": "NAT",
		"protected": false,
		"snooped": true,
		"tags": [
		  "NAT"
		],
		"operational": {
		  "forwarding": "NAT",
		  "label": "NAT",
		  "mtu": 1500,
		  "exists": true,
		  "enabled": true,
		  "protected": false,
		  "snooped": true,
		  "stp": false,
		  "interface": null,
		  "ip_networks": []
		}
	  }
	`)
	extConns := append([]byte(`[`), extConn...)
	extConns = append(extConns, []byte(`]`)...)

	tests := []struct {
		name    string
		data    mr.MockRespList
		wantErr bool
	}{
		{"good", mr.MockRespList{mr.MockResp{Data: extConns}}, false},
		{"badjson", mr.MockRespList{mr.MockResp{Data: []byte(`///`)}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc.mr.SetData(tt.data)
			got, err := tc.client.ExtConnectors(tc.ctx)
			if err != nil {
				if !tt.wantErr {
					t.Error("unexpected error!", err)
				}
				return
			}
			assert.NoError(t, err)
			assert.Len(t, got, 1)
			assert.True(t, got[0].Label == "NAT")
		})
	}

	t.Run("single", func(t *testing.T) {
		tc.mr.SetData(mr.MockRespList{mr.MockResp{Data: extConn}})
		got, err := tc.client.ExtConnGet(tc.ctx, "58568fbb-e1f8-4b83-a1f8-148c656eed39")
		assert.NoError(t, err)
		assert.Equal(t, got.ID, "58568fbb-e1f8-4b83-a1f8-148c656eed39")
		assert.True(t, got.Label == "NAT")
	})
}
