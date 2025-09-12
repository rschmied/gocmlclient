package api

import (
	"maps"
	"sync"
	"time"
)

// Stats holds basic API call statistics
type Stats struct {
	mu              sync.RWMutex
	TotalCalls      int
	CallsByMethod   map[string]int
	CallsByEndpoint map[string]int
	StatusCounts    map[int]int
	ResponseTimes   []time.Duration // For histogram calculation
}

// NewStats creates a new stats instance
func NewStats() *Stats {
	return &Stats{
		CallsByMethod:   make(map[string]int),
		CallsByEndpoint: make(map[string]int),
		StatusCounts:    make(map[int]int),
		ResponseTimes:   make([]time.Duration, 0),
	}
}

// RecordCall records a single API call
func (s *Stats) RecordCall(method, endpoint string, status int, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalCalls++
	s.CallsByMethod[method]++
	s.CallsByEndpoint[endpoint]++
	s.StatusCounts[status]++
	s.ResponseTimes = append(s.ResponseTimes, duration)
}

// GetSnapshot returns a thread-safe copy of current stats
func (s *Stats) GetSnapshot() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Deep copy maps and slice
	callsByMethod := make(map[string]int)
	maps.Copy(callsByMethod, s.CallsByMethod)

	callsByEndpoint := make(map[string]int)
	maps.Copy(callsByEndpoint, s.CallsByEndpoint)

	statusCounts := make(map[int]int)
	maps.Copy(statusCounts, s.StatusCounts)

	responseTimes := make([]time.Duration, len(s.ResponseTimes))
	copy(responseTimes, s.ResponseTimes)

	return Stats{
		TotalCalls:      s.TotalCalls,
		CallsByMethod:   callsByMethod,
		CallsByEndpoint: callsByEndpoint,
		StatusCounts:    statusCounts,
		ResponseTimes:   responseTimes,
	}
}
