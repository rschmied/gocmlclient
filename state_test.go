package cmlclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_clientState_String(t *testing.T) {
	tests := []struct {
		name string
		cs   clientState
		want string
	}{
		{"1", stateInitial, "INITIAL"},
		{"2", stateCheckVersion, "CHECKVERSION"},
		{"3", stateAuthRequired, "AUTHREQUIRED"},
		{"4", stateAuthenticating, "AUTHENTICATING"},
		{"5", stateAuthenticated, "AUTHENTICATED"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.String(); got != tt.want {
				t.Errorf("clientState.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_clientState_String_unknown(t *testing.T) {
	assert.Panics(t, func() {
		cs := clientState(99)
		_ = cs.String()
	})
}
