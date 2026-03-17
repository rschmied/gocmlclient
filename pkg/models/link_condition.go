// Package models provides the models for Cisco Modeling Labs
// here: link conditioning related types
package models

// LinkConditionConfiguration defines the configurable parameters for link conditioning
type LinkConditionConfiguration struct {
	Bandwidth     int     `json:"bandwidth,omitempty"`      // Bandwidth in kbps (0-10000000)
	Latency       int     `json:"latency,omitempty"`        // Delay in ms (0-10000)
	DelayCorr     float64 `json:"delay_corr,omitempty"`     // Delay correlation in percent (0-100)
	Limit         int     `json:"limit,omitempty"`          // Limit in ms (0-10000)
	Loss          float64 `json:"loss,omitempty"`           // Loss in percent (0-100)
	LossCorr      float64 `json:"loss_corr,omitempty"`      // Loss correlation in percent (0-100)
	Gap           int     `json:"gap,omitempty"`            // Gap between packets in ms (0-10000)
	Duplicate     float64 `json:"duplicate,omitempty"`      // Duplicate probability in percent (0-100)
	DuplicateCorr float64 `json:"duplicate_corr,omitempty"` // Duplicate correlation in percent (0-100)
	Jitter        int     `json:"jitter,omitempty"`         // Jitter in ms (0-10000)
	ReorderProb   float64 `json:"reorder_prob,omitempty"`   // Reorder probability in percent (0-100)
	ReorderCorr   float64 `json:"reorder_corr,omitempty"`   // Reorder correlation in percent (0-100)
	CorruptProb   float64 `json:"corrupt_prob,omitempty"`   // Corruption probability in percent (0-100)
	CorruptCorr   float64 `json:"corrupt_corr,omitempty"`   // Corruption correlation in percent (0-100)
	Enabled       bool    `json:"enabled,omitempty"`        // Whether conditioning is enabled
}

// LinkConditionStricted defines operational link conditioning data (read-only, no Enabled field)
type LinkConditionStricted struct {
	Bandwidth     int     `json:"bandwidth,omitempty"`      // Bandwidth in kbps
	Latency       int     `json:"latency,omitempty"`        // Delay in ms
	DelayCorr     float64 `json:"delay_corr,omitempty"`     // Delay correlation in percent
	Limit         int     `json:"limit,omitempty"`          // Limit in ms
	Loss          float64 `json:"loss,omitempty"`           // Loss in percent
	LossCorr      float64 `json:"loss_corr,omitempty"`      // Loss correlation in percent
	Gap           int     `json:"gap,omitempty"`            // Gap between packets in ms
	Duplicate     float64 `json:"duplicate,omitempty"`      // Duplicate probability in percent
	DuplicateCorr float64 `json:"duplicate_corr,omitempty"` // Duplicate correlation in percent
	Jitter        int     `json:"jitter,omitempty"`         // Jitter in ms
	ReorderProb   float64 `json:"reorder_prob,omitempty"`   // Reorder probability in percent
	ReorderCorr   float64 `json:"reorder_corr,omitempty"`   // Reorder correlation in percent
	CorruptProb   float64 `json:"corrupt_prob,omitempty"`   // Corruption probability in percent
	CorruptCorr   float64 `json:"corrupt_corr,omitempty"`   // Corruption correlation in percent
}

// ConditionResponse represents the response from link conditioning operations
type ConditionResponse struct {
	LinkConditionConfiguration                        // Configuration fields including Enabled
	Operational                *LinkConditionStricted `json:"operational,omitempty"` // Operational data (read-only)
}
