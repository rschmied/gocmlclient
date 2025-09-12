package client

import (
	"time"
)

// Stats holds basic API call statistics
type Stats struct {
	TotalCalls      int
	CallsByMethod   map[string]int
	CallsByEndpoint map[string]int
	StatusCounts    map[int]int
	ResponseTimes   []time.Duration
}
