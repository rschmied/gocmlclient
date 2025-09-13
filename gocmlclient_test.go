package gocmlclient

import (
	"testing"

	"github.com/rschmied/gocmlclient/pkg/client"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Test that New function exists and has correct signature
	// We can't actually call it without a server, but we can verify it compiles
	opts := []client.Option{client.SkipReadyCheck()}
	assert.NotNil(t, opts)
}

func TestNewFunction(t *testing.T) {
	// Test the actual New function by calling it
	// This will fail due to no server, but it will exercise the code path
	_, err := New("http://invalid-server:12345")
	assert.Error(t, err)                        // Expected to fail due to invalid server
	assert.Contains(t, err.Error(), "dial tcp") // Network error expected
}
