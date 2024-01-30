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
	",", " ", ".", " ",
	"!", " ", "?", " ",
	"(", " ", ")", " ",
	"[", " ", "]", " ",
	"/", " ",
)

type file struct {
	Path      string
	NameParts []string
}

func New(folder string) (lyrics.Provider, error) {
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

func (c *Client) Lyrics(state player.State) ([]lyrics.Line, error) {
	f := c.findFile(state)

	if f == nil {
		return nil, nil
	}

	reader, err := os.Open(f.Path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	lys := parseLrcFile(reader)
	// var header []lyrics.Line
	// header = append(header, lyrics.Line{
	// 	Time:  0,
	// 	Words: "Loading from Local...",
	// })
	// if lys[0].Time < 10 {
	// 	lys[0].Time = 10
	// }
	return lys, nil
}

func (c *Client) findFile(state player.State) *file {
	possiblePath := strings.Replace(strings.Replace(state.SongPath, ".mp3", ".lrc", 1), "file://", "", 1)
	var best *file
	parts := splitString(state.Artist + " " + state.Album + " " + strconv.Itoa(state.TrackNumber) + " " + state.Title)

	existsFile, existsErr := os.Stat(possiblePath)

	if existsErr == nil && existsFile != nil {
		best = &file{
			Path:      possiblePath,
			NameParts: parts,
		}
		return best
	}

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
	if strings.Contains(best.Path, state.Artist) && strings.Contains(best.Path, state.Album) && strings.Contains(best.Path, state.Title) && strings.Contains(best.Path, strconv.Itoa(state.TrackNumber)) {
		return best
	}
	return nil
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

		parts := splitString(path)

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

func (c *Client) Name() string {
	return "LOCAL"
}
