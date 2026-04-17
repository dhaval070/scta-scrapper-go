package scraper

import (
	"reflect"
	"testing"
)

func TestNormalizeRows(t *testing.T) {
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
			name: "unexpected length (less than 7) - unchanged",
			input: [][]string{
				{"2024-01-01", "site1", "Home", "Guest", "Location", "Division"},
			},
			expected: [][]string{
				{"2024-01-01", "site1", "Home", "Guest", "Location", "Division"},
			},
		},
		{
			name: "unexpected length (more than 8) - unchanged",
			input: [][]string{
				{"2024-01-01 14:00", "site1", "Home", "Guest", "Location", "Division", "event123", "Address", "Extra"},
			},
			expected: [][]string{
				{"2024-01-01 14:00", "site1", "Home", "Guest", "Location", "Division", "event123", "Address", "Extra"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeRows(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("normalizeRows() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
