package lyrics

import "sptlrx/player"

type Provider interface {
	Lyrics(state player.State) ([]Line, error)
	Name() string
}

type Line struct {
	Time  int    `json:"time"`
	Words string `json:"words"`
}

func Timesynced(lines []Line) bool {
	return len(lines) > 1 && lines[1].Time != 0
}
