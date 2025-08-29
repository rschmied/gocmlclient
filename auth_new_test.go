package cmlclient_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	cmlclient "github.com/rschmied/gocmlclient"
	"github.com/rschmied/gocmlclient/internal/testutil"
	"github.com/stretchr/testify/assert"
)

var authCalls int

func authResponder(req *http.Request) (*http.Response, error) {
	_ = req
	authCalls++
	if authCalls == 1 {
		return httpmock.NewJsonResponseOrPanic(401, nil), nil
	}
	return httpmock.NewStringResponse(200, "OK"), nil
}

func setupTest(testcase int, client *cmlclient.Client) {
	switch testcase {
	case 0:
		if !testutil.IsLiveTesting() {
			httpmock.RegisterResponder("GET", "https://mock/api/v0/authok",
				httpmock.NewJsonResponderOrPanic(401, nil))
			httpmock.RegisterResponder("GET", "https://mock/api/v0/labs/lab_uuid",
				httpmock.NewStringResponder(200, `{"id":"lab_uuid","state":"DEFINED_ON_CORE","created":"2025-08-26T09:41:36+00:00","modified":"2025-08-26T09:41:36+00:00","lab_title":"this","owner":"00000000-0000-4000-a000-000000000000","owner_username":"admin","effective_permissions":["lab_admin","lab_exec","lab_edit","lab_view"]}`))
			httpmock.RegisterResponder("GET", "https://mock/api/v0/users/admin/id",
				httpmock.NewStringResponder(200, `"00000000-0000-4000-a000-000000000000"`))
			httpmock.RegisterResponder("GET", "https://mock/api/v0/users/00000000-0000-4000-a000-000000000000",
				httpmock.NewStringResponder(200, `{"id":"00000000-0000-4000-a000-000000000000","created":"2025-02-18T14:40:06+00:00","modified":"2025-08-28T12:34:09+00:00","username":"admin","password":"","fullname":"Default Admin","email":"","description":"","admin":true,"labs":[],"opt_in":true}`))
		}
	case 1:
		client.SetToken("badtoken")
		if !testutil.IsLiveTesting() {
			httpmock.RegisterResponder("GET", "https://mock/api/v0/users/admin/id",
				httpmock.NewJsonResponderOrPanic(401, nil))
		}
	case 2:
		client.SetToken("badtoken")
		client.SetUsernamePassword("admin", "badpassword")
		if !testutil.IsLiveTesting() {
			httpmock.RegisterResponder("GET", "https://mock/api/v0/users/admin/id",
				httpmock.NewJsonResponderOrPanic(401, nil))
			httpmock.RegisterResponder("POST", "https://mock/api/v0/auth_extended",
				httpmock.NewStringResponder(403, `{ "description": "Authentication failed!", "code": 403 }`))
		}
	case 3:
		client.SetUsernamePassword("admin", "badpassword")
		if !testutil.IsLiveTesting() {
			httpmock.RegisterResponder("GET", "https://mock/api/v0/authok", authResponder)
			httpmock.RegisterResponder("POST", "https://mock/api/v0/auth_extended",
				httpmock.NewStringResponder(403, `{ "description": "Authentication failed!", "code": 403 }`))
		}
	case 4:
		config := testutil.DefaultConfig()
		client.SetUsernamePassword(config.Username, config.Password)
		if !testutil.IsLiveTesting() {
			authCalls = 0
			httpmock.RegisterResponder("GET", "https://mock/api/v0/authok", authResponder)
			httpmock.RegisterResponder("POST", "https://mock/api/v0/auth_extended",
				httpmock.NewStringResponder(200, `{
					"username": "admin",
					"id": "00000000-0000-4000-a000-000000000000",
					"token": "sometoken",
					"admin": true,
					"error": null
				}`))
			httpmock.RegisterResponder("GET", "https://mock/api/v0/users/admin/id",
				httpmock.NewStringResponder(200, `"00000000-0000-4000-a000-000000000000"`))
			httpmock.RegisterResponder("GET", "https://mock/api/v0/users/00000000-0000-4000-a000-000000000000",
				httpmock.NewStringResponder(200, `{"id":"00000000-0000-4000-a000-000000000000","created":"2025-02-18T14:40:06+00:00","modified":"2025-08-28T12:34:09+00:00","username":"admin","password":"","fullname":"Default Admin","email":"","description":"","admin":true,"labs":[],"opt_in":true}`))
		}
	case 5:
		config := testutil.DefaultConfig()
		if len(config.Token) == 0 {
			config.Token = "somegoodtoken"
		}
		client.SetToken(config.Token)
		if !testutil.IsLiveTesting() {
			authCalls = 0
			httpmock.RegisterResponder("POST", "https://mock/api/v0/auth_extended",
				httpmock.NewStringResponder(200, `{
					"username": "admin",
					"id": "00000000-0000-4000-a000-000000000000",
					"token": "sometoken",
					"admin": true,
					"error": null
				}`))
			httpmock.RegisterResponder("GET", "https://mock/api/v0/users/admin/id",
				httpmock.NewStringResponder(200, `"00000000-0000-4000-a000-000000000000"`))
			httpmock.RegisterResponder("GET", "https://mock/api/v0/users/00000000-0000-4000-a000-000000000000",
				httpmock.NewStringResponder(200, `{"id":"00000000-0000-4000-a000-000000000000","created":"2025-02-18T14:40:06+00:00","modified":"2025-08-28T12:34:09+00:00","username":"admin","password":"","fullname":"Default Admin","email":"","description":"","admin":true,"labs":[],"opt_in":true}`))
		}

	}
}

func TestAuthNew(t *testing.T) {
	tests := []struct {
		name   string
		errMsg string
	}{
		{"noauth", "no credentials provided"},
		{"bad_token", "invalid token but no credentials provided"},
		{"bad_credentials1", "Authentication failed"},
		{"bad_credentials2", "Authentication failed"},
		{"good_credentials1", ""},
		{"good_token", ""},
	}
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := testutil.NewAPIClient(t)
			defer cleanup()
			setupTest(idx, client)
			user, err := client.UserByName(context.Background(), "admin")
			if len(tt.errMsg) > 0 {
				assert.ErrorContains(t, err, tt.errMsg)
				return
			}
			assert.NoError(t, err)
			if user != nil {
				assert.Equal(t, "admin", user.Username)
			}
		})
	}
}
