package api

import (
	"maps"
	"sync"
	"time"

	"github.com/rschmied/gocmlclient/pkg/models"
)

// Stats holds basic API call statistics
type Stats struct {
	models.Stats
	mu sync.RWMutex
}

// NewStats creates a new stats instance
func NewStats() *Stats {
	return &Stats{
		models.Stats{
			EndpointGroups: make(map[string]*models.EndpointStats),
		},
		sync.RWMutex{},
	}
}

// RecordCall records a single API call
func (s *Stats) RecordCall(method, endpoint string, status int, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Stats.RecordCall(method, endpoint, status, duration)
}

// GetSnapshot returns a thread-safe copy of current stats
func (s *Stats) GetSnapshot() *models.Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a deep copy of the stats
	snapshot := &models.Stats{
		EndpointGroups: make(map[string]*models.EndpointStats),
	}

	for key, group := range s.EndpointGroups {
		groupCopy := &models.EndpointStats{
			CallCount:    group.CallCount,
			MinTime:      group.MinTime,
			MaxTime:      group.MaxTime,
			AvgTime:      group.AvgTime,
			TotalTime:    group.TotalTime,
			StatusCounts: make(map[int]int),
		}
		maps.Copy(groupCopy.StatusCounts, group.StatusCounts)
		snapshot.EndpointGroups[key] = groupCopy
	}

	return snapshot
}
