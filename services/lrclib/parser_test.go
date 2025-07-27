package lrclib

import (
	"testing"
	"time"

	"github.com/raitonoberu/sptlrx/lyrics"
)

func TestSimpleLRCParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []lyrics.Line
		wantErr  bool
	}{
		{
			name: "basic_lrc",
			input: `[ti:Test Song]
[ar:Test Artist]

[00:12.34]First line
[00:15.67]Second line
[00:18.90]Third line`,
			expected: []lyrics.Line{
				{Time: 12340, Words: "First line"},
				{Time: 15670, Words: "Second line"},
				{Time: 18900, Words: "Third line"},
			},
			wantErr: false,
		},
		{
			name: "multiple_timestamps",
			input: `[00:12.34][00:45.67]Chorus line
[00:15.00]Regular line`,
			expected: []lyrics.Line{
				{Time: 12340, Words: "Chorus line"},
				{Time: 15000, Words: "Regular line"},
				{Time: 45670, Words: "Chorus line"},
			},
			wantErr: false,
		},
		{
			name: "with_offset",
			input: `[offset:500]
[00:10.00]Delayed line`,
			expected: []lyrics.Line{
				{Time: 10500, Words: "Delayed line"},
			},
			wantErr: false,
		},
		{
			name:    "empty_input",
			input:   "",
			wantErr: true,
		},
		{
			name: "no_timestamps",
			input: `Just plain text
No timestamps here`,
			expected: []lyrics.Line{
				{Time: 0, Words: "Just plain text"},
				{Time: 0, Words: "No timestamps here"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SimpleLRCParse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(result))
				return
			}

			for i, line := range result {
				if line.Time != tt.expected[i].Time || line.Words != tt.expected[i].Words {
					t.Errorf("Line %d: expected {%d, %q}, got {%d, %q}",
						i, tt.expected[i].Time, tt.expected[i].Words, line.Time, line.Words)
				}
			}
		})
	}
}

func TestValidateLRC(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid_lrc",
			input: `[ti:Song]
[00:12.34]Line 1
[00:15.67]Line 2`,
			wantErr: false,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name: "only_metadata",
			input: `[ti:Song]
[ar:Artist]`,
			wantErr: true, // no timed lines
		},
		{
			name: "mixed_valid",
			input: `[ti:Song]
[00:12.34]Timed line
Plain text line`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLRC(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLRC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseTimedLyric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []timeLyricPair
	}{
		{
			name:  "single_timestamp",
			input: "[00:12.34]Hello world",
			expected: []timeLyricPair{
				{time: 12*time.Second + 340*time.Millisecond, text: "Hello world"},
			},
		},
		{
			name:  "multiple_timestamps",
			input: "[00:12.34][00:45.67]Chorus",
			expected: []timeLyricPair{
				{time: 12*time.Second + 340*time.Millisecond, text: "Chorus"},
				{time: 45*time.Second + 670*time.Millisecond, text: "Chorus"},
			},
		},
		{
			name:     "no_timestamp",
			input:    "Just text",
			expected: []timeLyricPair{},
		},
		{
			name:  "milliseconds_variations",
			input: "[00:12.3]One digit ms",
			expected: []timeLyricPair{
				{time: 12*time.Second + 300*time.Millisecond, text: "One digit ms"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTimedLyric(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d pairs, got %d", len(tt.expected), len(result))
				return
			}

			for i, pair := range result {
				if pair.time != tt.expected[i].time || pair.text != tt.expected[i].text {
					t.Errorf("Pair %d: expected {%v, %q}, got {%v, %q}",
						i, tt.expected[i].time, tt.expected[i].text, pair.time, pair.text)
				}
			}
		})
	}
}

func TestLRCLibClient_convertToLines(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name     string
		response *LRCLibResponse
		wantErr  bool
	}{
		{
			name: "instrumental",
			response: &LRCLibResponse{
				Instrumental: true,
			},
			wantErr: false,
		},
		{
			name: "synced_lyrics",
			response: &LRCLibResponse{
				SyncedLyrics: "[00:12.34]Test line",
			},
			wantErr: false,
		},
		{
			name: "plain_lyrics",
			response: &LRCLibResponse{
				PlainLyrics: "Line 1\nLine 2",
			},
			wantErr: false,
		},
		{
			name: "no_lyrics",
			response: &LRCLibResponse{
				Instrumental: false,
				SyncedLyrics: "",
				PlainLyrics:  "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.convertToLines(tt.response)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) == 0 {
				t.Errorf("Expected some lyrics lines, got none")
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkSimpleLRCParse(b *testing.B) {
	lrcContent := `[ti:Benchmark Song]
[ar:Test Artist]
[al:Test Album]

[00:12.34]First line of lyrics
[00:15.67]Second line of lyrics
[00:18.90]Third line of lyrics
[00:22.11]Fourth line of lyrics
[00:25.44]Fifth line of lyrics`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := SimpleLRCParse(lrcContent)
		if err != nil {
			b.Fatalf("Parse error: %v", err)
		}
	}
}

func BenchmarkValidateLRC(b *testing.B) {
	lrcContent := `[ti:Benchmark Song]
[00:12.34]Test line
[00:15.67]Another line`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ValidateLRC(lrcContent)
		if err != nil {
			b.Fatalf("Validation error: %v", err)
		}
	}
}
