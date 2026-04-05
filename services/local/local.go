package local

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/raitonoberu/sptlrx/lyrics"
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
	index, err := createIndex(folder)
	if err != nil {
		return nil, err
	}
	return &Client{index: index}, nil
}

// Client implements lyrics.Provider
type Client struct {
	index []*file
}

func (c *Client) Lyrics(artist, track string) ([]lyrics.Line, error) {
	query := artist + " " + track
	f := c.findFile(query)
	if f == nil {
		return nil, nil
	}

	reader, err := os.Open(f.Path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return parseLrcFile(reader), nil
}

func (c *Client) findFile(query string) *file {
	parts := splitString(query)

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
	return best
}

func createIndex(folder string) ([]*file, error) {
	if strings.HasPrefix(folder, "~/") {
		dirname, _ := os.UserHomeDir()
		folder = filepath.Join(dirname, folder[2:])
	}

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
		if !lyrics.IsTimestampLine(line) {
			continue
		}
		result = append(result, lyrics.ParseLrcLine(line))
	}
	return result
}
