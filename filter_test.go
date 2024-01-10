package kiwi_sdk

import "testing"

func TestFilter_Build(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		params   FilterParams
		expected string
	}{
		{
			name:     "Single Parameter",
			content:  "id = {:id}",
			params:   FilterParams{"id": 123},
			expected: "id=123",
		},
		{
			name:     "Multiple Parameters",
			content:  "user = {:user} && active = {:active}",
			params:   FilterParams{"user": "john", "active": "true"},
			expected: "user='john'&&active='true'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFilter(tt.content, tt.params)
			if got := filter.Build(); got != tt.expected {
				t.Errorf("Filter.Build() = %v, want %v", got, tt.expected)
			}
		})
	}
}
