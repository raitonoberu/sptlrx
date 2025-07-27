package lrclib

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/raitonoberu/sptlrx/lyrics"
)

// LRC parser for converting LRC format to sptlrx lyrics.Line format
// Supports both simple and extended LRC formats

var (
	// LRC time tag pattern: [mm:ss.xx] or [mm:ss]
	lrcTimePattern = regexp.MustCompile(`^\[(\d{1,2}):(\d{2})(?:\.(\d{1,3}))?\](.*)$`)

	// Metadata patterns
	metadataPatterns = map[string]*regexp.Regexp{
		"title":  regexp.MustCompile(`^\[ti:(.+)\]$`),
		"artist": regexp.MustCompile(`^\[ar:(.+)\]$`),
		"album":  regexp.MustCompile(`^\[al:(.+)\]$`),
		"offset": regexp.MustCompile(`^\[offset:(-?\d+)\]$`),
		"length": regexp.MustCompile(`^\[length:(\d{2}:\d{2}\.\d{2})\]$`),
	}
)

// LRCMetadata holds parsed LRC file metadata
type LRCMetadata struct {
	Title  string
	Artist string
	Album  string
	Offset int // milliseconds
	Length string
	Raw    map[string]string // all metadata tags
}

// ParsedLRC represents a completely parsed LRC file
type ParsedLRC struct {
	Metadata LRCMetadata
	Lines    []lyrics.Line
}

// ParseLRC parses LRC format text into lyrics.Line slice with metadata
func ParseLRC(lrcText string) (*ParsedLRC, error) {
	if strings.TrimSpace(lrcText) == "" {
		return nil, fmt.Errorf("empty LRC content")
	}

	lines := strings.Split(lrcText, "\n")

	parsed := &ParsedLRC{
		Metadata: LRCMetadata{Raw: make(map[string]string)},
		Lines:    make([]lyrics.Line, 0),
	}

	var lyricsLines []timeLyricPair

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to parse as metadata
		if parseMetadata(line, &parsed.Metadata) {
			continue
		}

		// Try to parse as timed lyric
		if pairs := parseTimedLyric(line); len(pairs) > 0 {
			lyricsLines = append(lyricsLines, pairs...)
			continue
		}

		// Unrecognized line format
		if strings.HasPrefix(line, "[") {
			// Probably malformed tag, log but continue
			continue
		}

		// Plain text line without timestamp - add with 0 time
		lyricsLines = append(lyricsLines, timeLyricPair{
			time: 0,
			text: line,
			line: lineNum + 1,
		})
	}

	// Sort by time and convert to lyrics.Line
	sort.Slice(lyricsLines, func(i, j int) bool {
		return lyricsLines[i].time < lyricsLines[j].time
	})

	// Apply offset if specified
	offset := time.Duration(parsed.Metadata.Offset) * time.Millisecond

	for _, pair := range lyricsLines {
		adjustedTime := pair.time + offset
		if adjustedTime < 0 {
			adjustedTime = 0
		}

		parsed.Lines = append(parsed.Lines, lyrics.Line{
			Time:  int(adjustedTime.Milliseconds()),
			Words: pair.text,
		})
	}

	return parsed, nil
}

// timeLyricPair represents a lyric line with its timestamp during parsing
type timeLyricPair struct {
	time time.Duration
	text string
	line int // source line number for debugging
}

// parseMetadata extracts metadata from LRC tags
func parseMetadata(line string, metadata *LRCMetadata) bool {
	for key, pattern := range metadataPatterns {
		if matches := pattern.FindStringSubmatch(line); matches != nil {
			value := matches[1]
			metadata.Raw[key] = value

			switch key {
			case "title":
				metadata.Title = value
			case "artist":
				metadata.Artist = value
			case "album":
				metadata.Album = value
			case "offset":
				if offset, err := strconv.Atoi(value); err == nil {
					metadata.Offset = offset
				}
			case "length":
				metadata.Length = value
			}
			return true
		}
	}
	return false
}

// parseTimedLyric parses timed lyric lines, handling multiple timestamps per line
func parseTimedLyric(line string) []timeLyricPair {
	var pairs []timeLyricPair
	remaining := line

	// First, find all timestamps and extract the final text
	var timestamps []time.Duration
	textAfterTimestamps := remaining

	for {
		matches := lrcTimePattern.FindStringSubmatch(textAfterTimestamps)
		if matches == nil {
			break
		}

		// Parse time components
		minutes, _ := strconv.Atoi(matches[1])
		seconds, _ := strconv.Atoi(matches[2])

		var milliseconds int
		if matches[3] != "" {
			// Pad or truncate to 3 digits
			msStr := matches[3]
			if len(msStr) == 1 {
				msStr += "00"
			} else if len(msStr) == 2 {
				msStr += "0"
			} else if len(msStr) > 3 {
				msStr = msStr[:3]
			}
			milliseconds, _ = strconv.Atoi(msStr)
		}

		// Calculate total time
		totalTime := time.Duration(minutes)*time.Minute +
			time.Duration(seconds)*time.Second +
			time.Duration(milliseconds)*time.Millisecond

		timestamps = append(timestamps, totalTime)

		// Move to the text after this timestamp
		textAfterTimestamps = strings.TrimSpace(matches[4])

		// Check if there are more timestamps
		if !strings.HasPrefix(textAfterTimestamps, "[") {
			break
		}
	}

	// Create pairs for each timestamp with the final text
	for _, timestamp := range timestamps {
		pairs = append(pairs, timeLyricPair{
			time: timestamp,
			text: textAfterTimestamps,
		})
	}

	return pairs
}

// SimpleLRCParse provides a simple parsing function for basic use cases
func SimpleLRCParse(lrcText string) ([]lyrics.Line, error) {
	parsed, err := ParseLRC(lrcText)
	if err != nil {
		return nil, err
	}
	return parsed.Lines, nil
}

// ValidateLRC checks if the provided text is valid LRC format
func ValidateLRC(lrcText string) error {
	if strings.TrimSpace(lrcText) == "" {
		return fmt.Errorf("empty content")
	}

	lines := strings.Split(lrcText, "\n")
	hasTimedLines := false

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if it's a valid metadata tag
		isMetadata := false
		for _, pattern := range metadataPatterns {
			if pattern.MatchString(line) {
				isMetadata = true
				break
			}
		}

		if isMetadata {
			continue
		}

		// Check if it's a valid timed line
		if lrcTimePattern.MatchString(line) {
			hasTimedLines = true
			continue
		}

		// Check for unknown tags
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
			// Probably an unknown tag, which is okay
			continue
		}

		// Plain text lines are okay if we have some timed lines
		if !strings.HasPrefix(line, "[") {
			continue
		}

		return fmt.Errorf("invalid LRC format at line %d: %s", i+1, line)
	}

	if !hasTimedLines {
		return fmt.Errorf("no timed lyrics found")
	}

	return nil
}

// Example usage and test cases
func ExampleLRCFormats() {
	examples := map[string]string{
		"basic": `[ti:Example Song]
[ar:Example Artist]
[al:Example Album]

[00:12.34]First line of lyrics
[00:15.67]Second line of lyrics
[00:18.90]Third line with multiple timestamps
[00:22.11][00:25.44]Chorus line that repeats`,

		"with_offset": `[ti:Delayed Song]
[ar:Test Artist]
[offset:500]

[00:10.00]This line is delayed by 500ms
[00:15.00]So is this one`,

		"plain_text": `Just plain lyrics
Without any timestamps
Still valid for some use cases`,
	}

	for name, lrc := range examples {
		fmt.Printf("=== %s ===\n", name)
		if err := ValidateLRC(lrc); err != nil {
			fmt.Printf("Validation error: %v\n", err)
			continue
		}

		parsed, err := ParseLRC(lrc)
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			continue
		}

		fmt.Printf("Title: %s\n", parsed.Metadata.Title)
		fmt.Printf("Artist: %s\n", parsed.Metadata.Artist)
		fmt.Printf("Lines: %d\n", len(parsed.Lines))
		for _, line := range parsed.Lines {
			fmt.Printf("  [%d] %s\n", line.Time, line.Words)
		}
		fmt.Println()
	}
}
