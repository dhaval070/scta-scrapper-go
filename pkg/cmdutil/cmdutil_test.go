package cmdutil

import (
	"reflect"
	"testing"
)

func TestNormalizeCSVRows(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]string
		expected [][]string
	}{
		{
			name: "7 columns to 8 columns",
			input: [][]string{
				{"2024-01-01 14:00", "site1", "Home", "Guest", "Location", "Division", "Address"},
			},
			expected: [][]string{
				{"2024-01-01 14:00", "site1", "Home", "Guest", "Location", "Division", "", "Address"},
			},
		},
		{
			name: "8 columns unchanged",
			input: [][]string{
				{"2024-01-01 14:00", "site1", "Home", "Guest", "Location", "Division", "event123", "Address"},
			},
			expected: [][]string{
				{"2024-01-01 14:00", "site1", "Home", "Guest", "Location", "Division", "event123", "Address"},
			},
		},
		{
			name: "mixed lengths",
			input: [][]string{
				{"2024-01-01 14:00", "site1", "Home", "Guest", "Location", "Division", "Address"},
				{"2024-01-01 15:00", "site1", "Home2", "Guest2", "Location2", "Division2", "event456", "Address2"},
			},
			expected: [][]string{
				{"2024-01-01 14:00", "site1", "Home", "Guest", "Location", "Division", "", "Address"},
				{"2024-01-01 15:00", "site1", "Home2", "Guest2", "Location2", "Division2", "event456", "Address2"},
			},
		},
		{
			name:     "empty input",
			input:    [][]string{},
			expected: [][]string{},
		},
		{
			name: "invalid length (less than 7) - skipped",
			input: [][]string{
				{"2024-01-01", "site1", "Home", "Guest", "Location", "Division"},
			},
			expected: [][]string{},
		},
		{
			name: "invalid length (more than 8) - skipped",
			input: [][]string{
				{"2024-01-01 14:00", "site1", "Home", "Guest", "Location", "Division", "event123", "Address", "Extra"},
			},
			expected: [][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// normalizeCSVRows is not exported, so we need to test through exported functions
			// For now, we'll just verify the logic indirectly
			// In practice, we'd export the function or test via ImportLocations/ImportEventsFromRows
			got := normalizeCSVRows(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("normalizeCSVRows() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
