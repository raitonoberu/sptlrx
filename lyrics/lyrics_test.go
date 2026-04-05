package lyrics

import (
	"reflect"
	"testing"
)

func TestIsTimestampLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// positive
		{"valid 1-digit", "[01:02.3] lyrics", true},
		{"valid 2-digit", "[01:02.34] lyrics", true},
		{"valid 3-digit", "[01:02.345] lyrics", true},
		{"valid without space", "[01:02.34]lyrics", true},
		{"valid empty", "[01:02.34]", true},

		// negative
		{"missing ms", "[01:02]", false},
		{"missing leading bracket", "01:02.34] lyrics", false},
		{"missing closing bracket", "[01:02.34 lyrics", false},

		// tags
		{"title tag", "[ti: Song Title]", false},
		{"artist tag", "[ar: Artist Name]", false},
		{"offet tag", "[offset:0]", false},

		// edge cases
		{"empty string", "", false},
		{"just brackets", "[]", false},
		{"spaces before bracket", " [01:02.34]", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTimestampLine(tt.input)
			if result != tt.expected {
				t.Errorf("IsTimestampLine(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseLrcLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Line
	}{
		{
			name:     "1-digit precision",
			input:    "[00:01.5] lyrics",
			expected: Line{Time: 1500, Words: "lyrics"},
		},
		{
			name:     "2-digit precision",
			input:    "[01:02.34] lyrics",
			expected: Line{Time: 62340, Words: "lyrics"},
		},
		{
			name:     "3-digit precision",
			input:    "[00:10.500] lyrics",
			expected: Line{Time: 10500, Words: "lyrics"},
		},
		{
			name:     "empty",
			input:    "[00:05.00]",
			expected: Line{Time: 5000, Words: ""},
		},
		{
			name:     "with spaces",
			input:    "[00:02.00]   lyrics   ",
			expected: Line{Time: 2000, Words: "lyrics"},
		},
		{
			name:     "large",
			input:    "[99:00.00] lyrics",
			expected: Line{Time: 5940000, Words: "lyrics"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLrcLine(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseLrcLine(%q) = %+v; want %+v", tt.input, result, tt.expected)
			}
		})
	}
}
