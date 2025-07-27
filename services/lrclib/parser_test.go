package lrclib

import (
	"reflect"
	"testing"

	"github.com/raitonoberu/sptlrx/lyrics"
)

func TestParseLRC(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []lyrics.Line
		wantErr  bool
	}{
		{
			name: "Simple LRC format",
			input: `[00:12.34]First line
[00:23.45]Second line
[00:34.56]Third line`,
			expected: []lyrics.Line{
				{Time: 12340, Words: "First line"},
				{Time: 23450, Words: "Second line"},
				{Time: 34560, Words: "Third line"},
			},
			wantErr: false,
		},
		{
			name: "LRC with metadata",
			input: `[ar:Test Artist]
[ti:Test Title]
[al:Test Album]
[00:10.00]Hello world
[00:20.50]Goodbye world`,
			expected: []lyrics.Line{
				{Time: 10000, Words: "Hello world"},
				{Time: 20500, Words: "Goodbye world"},
			},
			wantErr: false,
		},
		{
			name: "LRC without milliseconds",
			input: `[00:12]First line
[00:23]Second line`,
			expected: []lyrics.Line{
				{Time: 12000, Words: "First line"},
				{Time: 23000, Words: "Second line"},
			},
			wantErr: false,
		},
		{
			name: "LRC with three-digit milliseconds",
			input: `[00:12.340]First line
[00:23.450]Second line`,
			expected: []lyrics.Line{
				{Time: 12340, Words: "First line"},
				{Time: 23450, Words: "Second line"},
			},
			wantErr: false,
		},
		{
			name: "LRC with multiple timestamps - simplified",
			input: `[00:12.34]Line with timestamp
[01:23.45]Normal line`,
			expected: []lyrics.Line{
				{Time: 12340, Words: "Line with timestamp"},
				{Time: 83450, Words: "Normal line"},
			},
			wantErr: false,
		},
		{
			name: "LRC with offset",
			input: `[offset:1000]
[00:10.00]Line with offset`,
			expected: []lyrics.Line{
				{Time: 11000, Words: "Line with offset"},
			},
			wantErr: false,
		},
		{
			name: "Empty lines and whitespace",
			input: `
[00:12.34]First line

[00:23.45]   Second line   

[00:34.56]Third line
`,
			expected: []lyrics.Line{
				{Time: 12340, Words: "First line"},
				{Time: 23450, Words: "Second line"},
				{Time: 34560, Words: "Third line"},
			},
			wantErr: false,
		},
		{
			name:     "Empty input",
			input:    "",
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Invalid time format - should skip invalid lines",
			input: `[xx:12.34]Invalid time
[00:23.45]Valid line`,
			expected: []lyrics.Line{
				{Time: 23450, Words: "Valid line"},
			},
			wantErr: false,
		},
		{
			name: "Out of order timestamps (should be sorted)",
			input: `[00:23.45]Second line
[00:12.34]First line
[00:34.56]Third line`,
			expected: []lyrics.Line{
				{Time: 12340, Words: "First line"},
				{Time: 23450, Words: "Second line"},
				{Time: 34560, Words: "Third line"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseLRC(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLRC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseLRC() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name         string
		minutes      string
		seconds      string
		milliseconds string
		expected     int
		wantErr      bool
	}{
		{"Basic time", "1", "23", "45", 83450, false},
		{"Zero minutes", "0", "12", "34", 12340, false},
		{"No milliseconds", "2", "30", "", 150000, false},
		{"Three digit ms", "1", "23", "456", 83456, false},
		{"Invalid minutes", "xx", "12", "34", 0, true},
		{"Invalid seconds", "1", "xx", "34", 0, true},
		{"Invalid milliseconds", "1", "23", "xx", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTime(tt.minutes, tt.seconds, tt.milliseconds)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != tt.expected {
				t.Errorf("parseTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedArtist string
		expectedTrack  string
	}{
		{"Standard format", "Artist Name - Track Title", "Artist Name", "Track Title"},
		{"No separator", "Artist Track", "Artist", "Track"},
		{"Multi-word artist", "The Beatles - Hey Jude", "The Beatles", "Hey Jude"},
		{"Single word", "Track", "", "Track"},
		{"Empty query", "", "", ""},
		{"Only separator", " - ", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			artist, track := parseQuery(tt.query)

			if artist != tt.expectedArtist {
				t.Errorf("parseQuery() artist = %v, want %v", artist, tt.expectedArtist)
			}

			if track != tt.expectedTrack {
				t.Errorf("parseQuery() track = %v, want %v", track, tt.expectedTrack)
			}
		})
	}
}

func TestCleanupLines(t *testing.T) {
	tests := []struct {
		name     string
		input    []lyrics.Line
		expected []lyrics.Line
	}{
		{
			name: "Remove duplicates",
			input: []lyrics.Line{
				{Time: 1000, Words: "First"},
				{Time: 1000, Words: "First updated"},
				{Time: 2000, Words: "Second"},
			},
			expected: []lyrics.Line{
				{Time: 1000, Words: "First updated"},
				{Time: 2000, Words: "Second"},
			},
		},
		{
			name: "No duplicates",
			input: []lyrics.Line{
				{Time: 1000, Words: "First"},
				{Time: 2000, Words: "Second"},
				{Time: 3000, Words: "Third"},
			},
			expected: []lyrics.Line{
				{Time: 1000, Words: "First"},
				{Time: 2000, Words: "Second"},
				{Time: 3000, Words: "Third"},
			},
		},
		{
			name:     "Empty input",
			input:    []lyrics.Line{},
			expected: []lyrics.Line{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanupLines(tt.input)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("cleanupLines() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Benchmark for performance testing
func BenchmarkParseLRC(b *testing.B) {
	sampleLRC := `[ar:Test Artist]
[ti:Test Title]
[00:12.34]First line of lyrics
[00:23.45]Second line of lyrics
[00:34.56]Third line of lyrics
[00:45.67]Fourth line of lyrics
[00:56.78]Fifth line of lyrics
[01:07.89]Sixth line of lyrics`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseLRC(sampleLRC)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseTime(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parseTime("1", "23", "45")
		if err != nil {
			b.Fatal(err)
		}
	}
}
