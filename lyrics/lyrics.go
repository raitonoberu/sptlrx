package lyrics

import (
	"strconv"
	"strings"
)

const NoLyricsPlaceholder = "Hmm. We don't know the lyrics for this one."

type Provider interface {
	Lyrics(artist, track string) ([]Line, error)
}

type Line struct {
	Time  int    `json:"time"`
	Words string `json:"words"`
}

func Timesynced(lines []Line) bool {
	return len(lines) > 1 && lines[1].Time != 0
}

func IsTimestampLine(line string) bool {
	if len(line) < 10 {
		return false
	}
	return line[0] == '[' &&
		line[1] >= '0' && line[1] <= '9' &&
		line[2] >= '0' && line[2] <= '9' &&
		line[3] == ':' &&
		line[6] == '.' &&
		strings.IndexByte(line, ']') > 6
}

func ParseLrcLine(line string) Line {
	m, _ := strconv.Atoi(line[1:3])
	s, _ := strconv.Atoi(line[4:6])

	closeBracket := strings.IndexByte(line, ']')

	msStr := line[7:closeBracket]
	ms, _ := strconv.Atoi(msStr)
	if len(msStr) == 2 {
		ms *= 10
	} else if len(msStr) == 1 {
		ms *= 100
	}

	words := strings.TrimSpace(line[closeBracket+1:])

	return Line{
		Time:  m*60*1000 + s*1000 + ms,
		Words: words,
	}
}
