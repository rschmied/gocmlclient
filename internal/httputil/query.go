// Package httputil provides shared HTTP request building utilities
package httputil

import "maps"

// QueryBuilder helps build query parameter maps with common patterns
type QueryBuilder struct {
	params map[string]string
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		params: make(map[string]string),
	}
}

// WithData adds data=true when showAll is true
func (qb *QueryBuilder) WithData(showAll bool) *QueryBuilder {
	if showAll {
		qb.params["data"] = "true"
	}
	return qb
}

// WithOperational adds operational=true
func (qb *QueryBuilder) WithOperational() *QueryBuilder {
	qb.params["operational"] = "true"
	return qb
}

// WithNamedConfigs adds operational=true and exclude_configurations=false when useNamedConfigs is true
func (qb *QueryBuilder) WithNamedConfigs(useNamedConfigs bool) *QueryBuilder {
	if useNamedConfigs {
		qb.params["operational"] = "true"
		qb.params["exclude_configurations"] = "false"
	}
	return qb
}

// WithPopulateInterfaces adds populate_interfaces=true
func (qb *QueryBuilder) WithPopulateInterfaces() *QueryBuilder {
	qb.params["populate_interfaces"] = "true"
	return qb
}

// Set adds a custom parameter
func (qb *QueryBuilder) Set(key, value string) *QueryBuilder {
	qb.params[key] = value
	return qb
}

// Build returns the built query parameter map
func (qb *QueryBuilder) Build() map[string]string {
	// Return a copy to prevent external modification
	result := make(map[string]string, len(qb.params))
	maps.Copy(result, qb.params)
	return result
}
