package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Stats holds basic API call statistics
type Stats struct {
	// Primary data structure - everything computed from this
	EndpointGroups map[string]*EndpointStats // Key: "METHOD /path/{id}/subpath/{id}"
}

// EndpointStats contains all metrics for a grouped endpoint
type EndpointStats struct {
	CallCount    int
	MinTime      time.Duration
	MaxTime      time.Duration
	AvgTime      time.Duration
	TotalTime    time.Duration
	StatusCounts map[int]int // Status code counts for this endpoint group
}

// normalizeEndpoint replaces UUIDs and numeric IDs with placeholders
func normalizeEndpoint(method, endpoint string) string {
	// Replace UUID patterns: 8-4-4-4-12 hex digits
	uuidPattern := regexp.MustCompile(`[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}`)
	normalized := uuidPattern.ReplaceAllString(endpoint, "{id}")

	// Replace numeric IDs (assuming they're at the end of path segments)
	numericPattern := regexp.MustCompile(`/\d+`)
	normalized = numericPattern.ReplaceAllString(normalized, "/{id}")

	return fmt.Sprintf("%s %s", method, normalized)
}

// RecordCall records a single API call with endpoint grouping
func (s *Stats) RecordCall(method, endpoint string, status int, duration time.Duration) {
	groupedKey := normalizeEndpoint(method, endpoint)

	if s.EndpointGroups == nil {
		s.EndpointGroups = make(map[string]*EndpointStats)
	}

	group, exists := s.EndpointGroups[groupedKey]
	if !exists {
		group = &EndpointStats{
			StatusCounts: make(map[int]int),
		}
		s.EndpointGroups[groupedKey] = group
	}

	// Update metrics
	group.CallCount++
	group.TotalTime += duration
	group.StatusCounts[status]++

	if group.CallCount == 1 || duration < group.MinTime {
		group.MinTime = duration
	}
	if duration > group.MaxTime {
		group.MaxTime = duration
	}
	group.AvgTime = group.TotalTime / time.Duration(group.CallCount)
}

// TotalCalls returns the total number of API calls
func (s *Stats) TotalCalls() int {
	total := 0
	for _, group := range s.EndpointGroups {
		total += group.CallCount
	}
	return total
}

// CallsByMethod returns call counts grouped by HTTP method
func (s *Stats) CallsByMethod() map[string]int {
	result := make(map[string]int)
	for key, group := range s.EndpointGroups {
		method := strings.Split(key, " ")[0]
		result[method] += group.CallCount
	}
	return result
}

// CallsByEndpoint returns call counts grouped by normalized endpoint
func (s *Stats) CallsByEndpoint() map[string]int {
	result := make(map[string]int)
	for key, group := range s.EndpointGroups {
		endpoint := strings.SplitN(key, " ", 2)[1]
		result[endpoint] += group.CallCount
	}
	return result
}

// StatusCounts returns total status code counts across all endpoints
func (s *Stats) StatusCounts() map[int]int {
	result := make(map[int]int)
	for _, group := range s.EndpointGroups {
		for status, count := range group.StatusCounts {
			result[status] += count
		}
	}
	return result
}

// String returns a formatted string representation of the statistics
func (s *Stats) String() string {
	if len(s.EndpointGroups) == 0 {
		return "No API calls recorded"
	}

	var builder strings.Builder
	builder.WriteString("API Statistics\n")
	builder.WriteString("==============\n")
	builder.WriteString(fmt.Sprintf("Total Calls: %d\n\n", s.TotalCalls()))

	// Calls by method
	methods := s.CallsByMethod()
	if len(methods) > 0 {
		builder.WriteString("Calls by Method:\n")
		for method, count := range methods {
			builder.WriteString(fmt.Sprintf("  %s: %d\n", method, count))
		}
		builder.WriteString("\n")
	}

	// Endpoint details
	builder.WriteString("Endpoint Details:\n")
	for endpoint, stats := range s.EndpointGroups {
		builder.WriteString(fmt.Sprintf("  %s:\n", endpoint))
		builder.WriteString(fmt.Sprintf("    Calls: %d\n", stats.CallCount))
		builder.WriteString(fmt.Sprintf("    Response Times: Min=%v, Max=%v, Avg=%v\n",
			stats.MinTime, stats.MaxTime, stats.AvgTime))

		if len(stats.StatusCounts) > 0 {
			builder.WriteString("    Status Codes:\n")
			for status, count := range stats.StatusCounts {
				builder.WriteString(fmt.Sprintf("      %d: %d\n", status, count))
			}
		}
		builder.WriteString("\n")
	}

	return strings.TrimSpace(builder.String())
}
