package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrettyPrintError(t *testing.T) {
	tests := []struct {
		name     string
		errStr   string
		expected error
	}{
		{
			name:     "JSON error response",
			errStr:   `{"code":400,"description":"Bad Request"}`,
			expected: nil,
		},
		{
			name:     "JSON error with object description",
			errStr:   `{"code":500,"description":{"message":"Internal Server Error","details":"Something went wrong"}}`,
			expected: nil,
		},
		{
			name:     "JSON error with array description",
			errStr:   `{"code":422,"description":["Field is required","Invalid format"]}`,
			expected: nil,
		},
		{
			name:     "JSON error with nested object in description",
			errStr:   `{"code":500,"description":"{\"inner\":\"value\"}"}`,
			expected: nil,
		},
		{
			name:     "JSON error with invalid inner JSON",
			errStr:   `{"code":500,"description":"{invalid json}"}`,
			expected: nil,
		},
		{
			name:     "non-JSON error",
			errStr:   "plain text error",
			expected: nil,
		},
		{
			name:     "malformed JSON",
			errStr:   `{"code":400,"description":}`,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &testError{msg: tt.errStr}
			result := PrettyPrintError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// testError implements error interface for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
