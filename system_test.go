package cmlclient

import (
	"testing"

	mr "github.com/rschmied/mockresponder"
	"github.com/stretchr/testify/assert"
)

// Common test cases for version checking
var versionTestCases = []struct {
	name       string
	version    string
	constraint string
	want       bool
	wantErr    bool
}{
	// Error cases
	{"no version", "", ">=2.4.0", false, true},
	{"garbage version", "garbage", ">=2.4.0", false, true},
	{"incomplete version", "2.4", ">=2.4.0", false, true},

	// Valid cases
	{"exact match", "2.4.0", ">=2.4.0,<3.0.0", true, false},
	{"newer version", "2.5.0", ">=2.4.0,<3.0.0", true, false},
	{"build version", "2.4.0+build.1", ">=2.4.0,<3.0.0", true, false},
	{"dev version", "2.5.0-dev0+build.3.2f7875762", ">=2.4.0,<3.0.0", true, false},
	{"too old", "2.3.9", ">=2.4.0,<3.0.0", false, false},
	{"too new", "3.0.0", ">=2.4.0,<3.0.0", false, false},

	// Named configs cases
	{"named configs supported", "2.7.0", ">=2.7.0", true, false},
	{"named configs not supported", "2.6.9", ">=2.7.0", false, false},
}

func TestClient_VersionCheck(t *testing.T) {
	c := New("https://bla.bla", true)
	mrClient, ctx := mr.NewMockResponder()
	c.SetHTTPClient(mrClient, true)
	c.useNamedConfigs = true
	c.state.set(stateAuthenticated)

	tests := []struct {
		name        string
		wantJSON    string
		wantErr     bool
		canNamedCfg bool
	}{
		// these three yield an error, useNamedConfigs is untouched
		{"too old", `{"version": "2.1.0","ready": true}`, true, true},
		{"garbage", `{"version": "garbage","ready": true}`, true, true},
		{"too new", `{"version": "3.0.0","ready": true}`, true, true},
		// the rest will reset useNamedConfigs, if needed
		{"perfect", `{"version": "2.4.0","ready": true}`, false, false},
		{"actual", `{"version": "2.4.0+build.1","ready": true}`, false, false},
		{"newer", `{"version": "2.4.1","ready": true}`, false, false},
		{"dev", `{"version": "2.5.0-dev0+build.3.2f7875762","ready": true}`, false, false},
		{"v2.5.0", `{"version": "2.5.0+build.5","ready": true}`, false, false},
		{"v2.7.0", `{"version": "2.7.0+build.8","ready": true}`, false, true},
	}
	for _, tt := range tests {
		mrClient.SetData(mr.MockRespList{{Data: []byte(tt.wantJSON)}})
		t.Run(tt.name, func(t *testing.T) {
			c.UseNamedConfigs()
			if err := c.versionCheck(ctx, 0); (err != nil) != tt.wantErr {
				t.Errorf("Client.VersionCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.canNamedCfg != c.useNamedConfigs {
				t.Errorf("Client.VersionCheck() useNamedConfigs is = %t, want %t", c.useNamedConfigs, tt.canNamedCfg)
			}
		})
		if !mrClient.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}

// Test the specific error handling in checkVersionConstraint
func TestClient_CheckVersionConstraintError(t *testing.T) {
	c := &Client{}
	c.useNamedConfigs = true

	// Test the scenario where checkVersionConstraint fails due to invalid constraint
	// This simulates what happens in versionCheck when the named configs constraint check fails
	got, err := c.checkVersionConstraint("2.4.0", "invalid constraint syntax")

	if err == nil {
		t.Error("Expected error for invalid constraint, got nil")
	}

	if got != false {
		t.Errorf("Expected false for invalid constraint, got %v", got)
	}
}

func TestClient_ExportedVersionCheck(t *testing.T) {
	_, ctx := mr.NewMockResponder()

	for _, tt := range versionTestCases {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{version: tt.version}

			got, err := c.VersionCheck(ctx, tt.constraint)

			if (err != nil) != tt.wantErr {
				t.Errorf("VersionCheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("VersionCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_CheckVersionConstraint(t *testing.T) {
	c := &Client{}

	// Test a subset of cases plus constraint-specific ones
	constraintTests := append(versionTestCases[:8], // First 8 cases
		struct {
			name       string
			version    string
			constraint string
			want       bool
			wantErr    bool
		}{"invalid constraint", "2.4.0", "invalid constraint", false, true},
	)

	for _, tt := range constraintTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.checkVersionConstraint(tt.version, tt.constraint)

			if (err != nil) != tt.wantErr {
				t.Errorf("checkVersionConstraint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("checkVersionConstraint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_NotReady(t *testing.T) {
	c := New("https://bla.bla", true)
	mrClient, ctx := mr.NewMockResponder()
	c.httpClient = mrClient
	c.do = mrClient.Do
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
