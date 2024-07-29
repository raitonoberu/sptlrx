package local

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sptlrx/lyrics"
	"sptlrx/player"
	"strconv"
	"strings"
)

var replacer = strings.NewReplacer(
	"_", " ", "-", " ",
	",", "", ".", "",
	"!", "", "?", "",
	"(", "", ")", "",
	"[", "", "]", "",
)

type file struct {
	Path      string
	NameParts []string
}

func New(folder string) (*Client, error) {
	var expandedFolder string
	if strings.HasPrefix(folder, "~/") {
		dirname, _ := os.UserHomeDir()
		expandedFolder = filepath.Join(dirname, folder[2:])
	}

	index, err := createIndex(expandedFolder)
	if err != nil {
		return nil, err
	}
	return &Client{folder: expandedFolder, index: index}, nil
}

// Client implements lyrics.Provider
type Client struct {
	folder string
	index []*file
}

func (c *Client) Lyrics(track *player.TrackMetadata) ([]lyrics.Line, error) {
	f := c.findFile(track)
	if f == "" {
		return nil, nil
	}

	reader, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return parseLrcFile(reader), nil
}

func (c *Client) findFile(track *player.TrackMetadata) string {
	if track == nil {
		return ""
	}

	// If it is a local track, try for similarly named .lrc file first
	if track.Uri != "" {
		var absUri string
		if filepath.IsAbs(track.Uri) {
			// Uri is already absolute
			absUri = track.Uri
		} else {
			// Uri is relative to local music directory
			absUri = filepath.Join(c.folder, track.Uri)
		}
		absLyricsUri := strings.TrimSuffix(absUri, filepath.Ext(absUri)) + ".lrc"
		if _, err := os.Stat(absLyricsUri); err == nil {
			return absLyricsUri
		}
	}

	parts := splitString(track.Query)

	var best *file
	var maxScore int
	for _, f := range c.index {
		var score int
		for _, part := range parts {
			for _, namePart := range f.NameParts {
				if namePart == part {
					score++
					break
				}
			}
		}
		if score > maxScore {
			maxScore = score
			best = f
			if score == len(parts) {
				break
			}
		}
	}
	if best == nil {
		return ""
	}
	return best.Path
}

func createIndex(folder string) ([]*file, error) {
	index := []*file{}
	return index, filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
		if d == nil {
			return fmt.Errorf("invalid path: %s", path)
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".lrc") {
			return nil
		}
		name := strings.TrimSuffix(d.Name(), ".lrc")
		parts := splitString(name)

		index = append(index, &file{
			Path:      path,
			NameParts: parts,
		})
		return nil
	})
}

func splitString(s string) []string {
	s = strings.ToLower(s)
	s = replacer.Replace(s)
	return strings.Fields(s)
}

func parseLrcFile(reader io.Reader) []lyrics.Line {
	result := []lyrics.Line{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "[") || len(line) < 10 {
			continue
		}
		result = append(result, parseLrcLine(line))
	}
	return result
}

func parseLrcLine(line string) lyrics.Line {
	// [00:00.00]text -> {"time": 0, "words": "text"}
	h, _ := strconv.Atoi(line[1:3])
	m, _ := strconv.Atoi(line[4:6])
	s, _ := strconv.Atoi(line[7:9])

	return lyrics.Line{
		Time:  h*60*1000 + m*1000 + s*10,
		Words: line[10:],
	}
}
