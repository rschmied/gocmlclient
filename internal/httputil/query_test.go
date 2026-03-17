package httputil

import (
	"testing"
)

func TestQueryBuilder(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *QueryBuilder
		expected map[string]string
	}{
		{
			name: "empty builder",
			builder: func() *QueryBuilder {
				return NewQueryBuilder()
			},
			expected: map[string]string{},
		},
		{
			name: "with data true",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().WithData(true)
			},
			expected: map[string]string{
				"data": "true",
			},
		},
		{
			name: "with data false",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().WithData(false)
			},
			expected: map[string]string{},
		},
		{
			name: "with operational",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().WithOperational()
			},
			expected: map[string]string{
				"operational": "true",
			},
		},
		{
			name: "with named configs true",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().WithNamedConfigs(true)
			},
			expected: map[string]string{
				"operational":            "true",
				"exclude_configurations": "false",
			},
		},
		{
			name: "with named configs false",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().WithNamedConfigs(false)
			},
			expected: map[string]string{},
		},
		{
			name: "with populate interfaces",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().WithPopulateInterfaces()
			},
			expected: map[string]string{
				"populate_interfaces": "true",
			},
		},
		{
			name: "with custom parameter",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().Set("limit", "10")
			},
			expected: map[string]string{
				"limit": "10",
			},
		},
		{
			name: "chained methods",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().
					WithData(true).
					WithOperational().
					Set("limit", "20")
			},
			expected: map[string]string{
				"data":        "true",
				"operational": "true",
				"limit":       "20",
			},
		},
		{
			name: "complex example",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().
					WithData(true).
					WithNamedConfigs(true).
					WithPopulateInterfaces().
					Set("sort", "name").
					Set("order", "asc")
			},
			expected: map[string]string{
				"data":                   "true",
				"operational":            "true",
				"exclude_configurations": "false",
				"populate_interfaces":    "true",
				"sort":                   "name",
				"order":                  "asc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d parameters, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				actualValue, exists := result[key]
				if !exists {
					t.Errorf("expected parameter %s to exist", key)
				}
				if actualValue != expectedValue {
					t.Errorf("expected parameter %s to be %s, got %s", key, expectedValue, actualValue)
				}
			}

			// Ensure no extra parameters
			for key := range result {
				if _, exists := tt.expected[key]; !exists {
					t.Errorf("unexpected parameter %s", key)
				}
			}
		})
	}
}

func TestQueryBuilderImmutability(t *testing.T) {
	builder := NewQueryBuilder().WithData(true)
	original := builder.Build()

	// Modify the original map (should not affect builder)
	original["extra"] = "value"

	// Build again - should not include the extra parameter
	result := builder.Build()

	if len(result) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(result))
	}

	if result["data"] != "true" {
		t.Errorf("expected data=true, got %s", result["data"])
	}

	if _, exists := result["extra"]; exists {
		t.Error("builder should not include extra parameter from modified map")
	}
}

func TestQueryBuilderMethodChaining(t *testing.T) {
	// Test that methods return the builder for chaining
	builder := NewQueryBuilder()

	// These should all return the same builder instance
	result1 := builder.WithData(true)
	result2 := result1.WithOperational()
	result3 := result2.Set("key", "value")

	if result1 != result2 || result2 != result3 || result3 != builder {
		t.Error("method chaining should return the same builder instance")
	}

	result := builder.Build()
	expected := map[string]string{
		"data":        "true",
		"operational": "true",
		"key":         "value",
	}

	if len(result) != len(expected) {
		t.Errorf("expected %d parameters, got %d", len(expected), len(result))
	}

	for key, expectedValue := range expected {
		if actualValue := result[key]; actualValue != expectedValue {
			t.Errorf("expected %s=%s, got %s=%s", key, expectedValue, key, actualValue)
		}
	}
}

func TestQueryBuilder_WithExcludeConfigurations(t *testing.T) {
	trueVal := true
	falseVal := false

	got := NewQueryBuilder().WithExcludeConfigurations(nil).Build()
	if len(got) != 0 {
		t.Fatalf("expected empty map, got %v", got)
	}

	got = NewQueryBuilder().WithExcludeConfigurations(&trueVal).Build()
	if got["exclude_configurations"] != "true" {
		t.Fatalf("expected exclude_configurations=true, got %v", got)
	}
	if got["operational"] != "true" {
		t.Fatalf("expected operational=true, got %v", got)
	}

	got = NewQueryBuilder().WithExcludeConfigurations(&falseVal).Build()
	if got["exclude_configurations"] != "false" {
		t.Fatalf("expected exclude_configurations=false, got %v", got)
	}
	if got["operational"] != "true" {
		t.Fatalf("expected operational=true, got %v", got)
	}
}
