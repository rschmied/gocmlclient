package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStats_RecordCall(t *testing.T) {
	stats := &Stats{}

	// Record some calls
	stats.RecordCall("GET", "/labs/123", 200, 50*time.Millisecond)
	stats.RecordCall("GET", "/labs/456", 200, 75*time.Millisecond)
	stats.RecordCall("POST", "/labs/123/nodes", 201, 100*time.Millisecond)

	// Check total calls
	assert.Equal(t, 3, stats.TotalCalls())

	// Check method grouping
	assert.Equal(t, 2, stats.CallsByMethod()["GET"])
	assert.Equal(t, 1, stats.CallsByMethod()["POST"])

	// Check endpoint grouping
	assert.Equal(t, 2, stats.CallsByEndpoint()["/labs/{id}"])
	assert.Equal(t, 1, stats.CallsByEndpoint()["/labs/{id}/nodes"])

	// Check status counts
	assert.Equal(t, 2, stats.StatusCounts()[200])
	assert.Equal(t, 1, stats.StatusCounts()[201])
}

func TestStats_EndpointGrouping(t *testing.T) {
	stats := &Stats{}

	// Test UUID replacement
	stats.RecordCall("GET", "/labs/a7d20917-5e57-407f-80ea-63596c53f198", 200, 50*time.Millisecond)
	stats.RecordCall("GET", "/labs/bc9b796e-48bc-4369-b131-231dfa057d41", 200, 75*time.Millisecond)

	// Both should be grouped under the same key
	assert.Equal(t, 2, stats.EndpointGroups["GET /labs/{id}"].CallCount)
	assert.Equal(t, 50*time.Millisecond, stats.EndpointGroups["GET /labs/{id}"].MinTime)
	assert.Equal(t, 75*time.Millisecond, stats.EndpointGroups["GET /labs/{id}"].MaxTime)
	assert.Equal(t, (50+75)*time.Millisecond/2, stats.EndpointGroups["GET /labs/{id}"].AvgTime)
}

func TestStats_StatusCountsPerGroup(t *testing.T) {
	stats := &Stats{}

	// Record calls with different status codes
	stats.RecordCall("GET", "/labs/123", 200, 50*time.Millisecond)
	stats.RecordCall("GET", "/labs/123", 404, 30*time.Millisecond)
	stats.RecordCall("GET", "/labs/123", 500, 80*time.Millisecond)

	group := stats.EndpointGroups["GET /labs/{id}"]
	assert.Equal(t, 1, group.StatusCounts[200])
	assert.Equal(t, 1, group.StatusCounts[404])
	assert.Equal(t, 1, group.StatusCounts[500])
}

func TestStats_EmptyStats(t *testing.T) {
	stats := &Stats{}

	assert.Equal(t, 0, stats.TotalCalls())
	assert.Empty(t, stats.CallsByMethod())
	assert.Empty(t, stats.CallsByEndpoint())
	assert.Empty(t, stats.StatusCounts())
	assert.Empty(t, stats.EndpointGroups)
}

func TestNormalizeEndpoint(t *testing.T) {
	tests := []struct {
		method   string
		endpoint string
		expected string
	}{
		{"GET", "/labs/123", "GET /labs/{id}"},
		{"GET", "/labs/a7d20917-5e57-407f-80ea-63596c53f198", "GET /labs/{id}"},
		{"POST", "/labs/123/nodes/456", "POST /labs/{id}/nodes/{id}"},
		{"GET", "/users", "GET /users"},
		{"DELETE", "/nodes/123/interfaces/456", "DELETE /nodes/{id}/interfaces/{id}"},
	}

	for _, tt := range tests {
		result := normalizeEndpoint(tt.method, tt.endpoint)
		assert.Equal(t, tt.expected, result, "method=%s, endpoint=%s", tt.method, tt.endpoint)
	}
}

func TestStats_String(t *testing.T) {
	t.Run("empty stats", func(t *testing.T) {
		stats := &Stats{}
		result := stats.String()
		assert.Equal(t, "No API calls recorded", result)
	})

	t.Run("stats with data", func(t *testing.T) {
		stats := &Stats{}

		// Record some calls
		stats.RecordCall("GET", "/labs/123", 200, 50*time.Millisecond)
		stats.RecordCall("GET", "/labs/456", 404, 75*time.Millisecond)
		stats.RecordCall("POST", "/labs/123/nodes", 201, 100*time.Millisecond)
		stats.RecordCall("GET", "/users", 200, 30*time.Millisecond)

		result := stats.String()

		// Check that the result contains expected elements
		assert.Contains(t, result, "API Statistics")
		assert.Contains(t, result, "Total Calls: 4")
		assert.Contains(t, result, "Calls by Method:")
		assert.Contains(t, result, "GET: 3")
		assert.Contains(t, result, "POST: 1")
		assert.Contains(t, result, "Endpoint Details:")
		assert.Contains(t, result, "GET /labs/{id}:")
		assert.Contains(t, result, "Calls: 2")
		assert.Contains(t, result, "Response Times:")
		assert.Contains(t, result, "Min=50ms")
		assert.Contains(t, result, "Max=75ms")
		assert.Contains(t, result, "Status Codes:")
		assert.Contains(t, result, "200: 1")
		assert.Contains(t, result, "404: 1")
		assert.Contains(t, result, "POST /labs/{id}/nodes:")
		assert.Contains(t, result, "GET /users:")
	})
}
