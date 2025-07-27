package lrclib

import (
	"bufio"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/raitonoberu/sptlrx/lyrics"
)

// LRC format patterns
var (
	// Time format: [mm:ss.xx] or [mm:ss] or [mm:ss.xxx]
	timeRegex = regexp.MustCompile(`^\[(\d{1,2}):(\d{2})(?:\.(\d{2,3}))?\](.*)$`)

	// Metadata tags: [ar:Artist] [ti:Title] [al:Album] [offset:1000] etc.
	metaRegex = regexp.MustCompile(`^\[([a-z]{2,10}):(.*)\]$`)
)

// LRCMetadata represents metadata from LRC file
type LRCMetadata struct {
	Artist string
	Title  string
	Album  string
	Offset int // Time offset in milliseconds
}

// ParseLRC parses LRC format lyrics into lyrics.Line slice
func ParseLRC(lrcContent string) ([]lyrics.Line, error) {
	if lrcContent == "" {
		return nil, fmt.Errorf("empty LRC content")
	}

	var lines []lyrics.Line
	metadata := &LRCMetadata{}

	// First pass: extract metadata
	scanner := bufio.NewScanner(strings.NewReader(lrcContent))
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if matches := metaRegex.FindStringSubmatch(text); len(matches) == 3 {
			parseMetadata(matches[1], matches[2], metadata)
		}
	}

	// Second pass: parse lyrics with metadata applied
	scanner = bufio.NewScanner(strings.NewReader(lrcContent))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		text := strings.TrimSpace(scanner.Text())

		if text == "" {
			continue // Skip empty lines
		}

		// Skip metadata tags (already processed)
		if metaRegex.MatchString(text) {
			continue
		}

		// Parse time-synced line
		if matches := timeRegex.FindStringSubmatch(text); len(matches) >= 4 {
			time, err := parseTime(matches[1], matches[2], matches[3])
			if err != nil {
				// Skip invalid lines instead of failing completely
				continue
			}

			// Apply offset if specified
			time += metadata.Offset

			// Extract lyrics text (everything after the timestamp)
			lyricText := strings.TrimSpace(matches[4])

			// For now, handle simple case - multiple timestamps will be addressed later
			lines = append(lines, lyrics.Line{
				Time:  time,
				Words: lyricText,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading LRC content: %w", err)
	}

	// Sort lines by time
	sort.Slice(lines, func(i, j int) bool {
		return lines[i].Time < lines[j].Time
	})

	// Remove duplicate timestamps and merge consecutive empty lines
	lines = cleanupLines(lines)

	return lines, nil
}

// parseTime converts time components to milliseconds
func parseTime(minutesStr, secondsStr, millisecondsStr string) (int, error) {
	minutes, err := strconv.Atoi(minutesStr)
	if err != nil {
		return 0, fmt.Errorf("invalid minutes: %w", err)
	}

	seconds, err := strconv.Atoi(secondsStr)
	if err != nil {
		return 0, fmt.Errorf("invalid seconds: %w", err)
	}

	var milliseconds int
	if millisecondsStr != "" {
		// Handle both .xx and .xxx formats
		if len(millisecondsStr) == 2 {
			// .xx format (centiseconds)
			cs, err := strconv.Atoi(millisecondsStr)
			if err != nil {
				return 0, fmt.Errorf("invalid centiseconds: %w", err)
			}
			milliseconds = cs * 10
		} else if len(millisecondsStr) == 3 {
			// .xxx format (milliseconds)
			ms, err := strconv.Atoi(millisecondsStr)
			if err != nil {
				return 0, fmt.Errorf("invalid milliseconds: %w", err)
			}
			milliseconds = ms
		}
	}

	totalMs := minutes*60*1000 + seconds*1000 + milliseconds

	if totalMs < 0 {
		return 0, fmt.Errorf("negative time not allowed")
	}

	return totalMs, nil
}

// parseMetadata processes LRC metadata tags
func parseMetadata(tag, value string, metadata *LRCMetadata) error {
	switch tag {
	case "ar":
		metadata.Artist = value
	case "ti":
		metadata.Title = value
	case "al":
		metadata.Album = value
	case "offset":
		offset, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid offset value: %w", err)
		}
		metadata.Offset = offset
	case "by", "re", "ve", "length":
		// Ignore these common metadata tags
	default:
		// Unknown tag, ignore silently
	}
	return nil
}

// cleanupLines removes duplicates and cleans up the lines
func cleanupLines(lines []lyrics.Line) []lyrics.Line {
	if len(lines) == 0 {
		return lines
	}

	var cleaned []lyrics.Line
	lastTime := -1

	for _, line := range lines {
		// Skip duplicate timestamps (keep the last one)
		if line.Time == lastTime {
			if len(cleaned) > 0 {
				// Replace the previous line with this one
				cleaned[len(cleaned)-1] = line
			}
			continue
		}

		cleaned = append(cleaned, line)
		lastTime = line.Time
	}

	return cleaned
}

// ParseLRCFromFile parses LRC content from a file path (utility function)
func ParseLRCFromFile(content string) ([]lyrics.Line, *LRCMetadata, error) {
	lines, err := ParseLRC(content)
	if err != nil {
		return nil, nil, err
	}

	// Extract metadata (this is a simplified version)
	metadata := &LRCMetadata{}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if matches := metaRegex.FindStringSubmatch(text); len(matches) == 3 {
			parseMetadata(matches[1], matches[2], metadata)
		}
	}

	return lines, metadata, nil
}
