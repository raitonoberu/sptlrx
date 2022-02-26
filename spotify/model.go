package spotify

type CurrentlyPlaying struct {
	ID       string
	Position int
	Playing  bool
}

type LyricsLine struct {
	Time  int    `json:"startTimeMs,string"`
	Words string `json:"words"`
}

type LyricsLines []*LyricsLine

func (l LyricsLines) Timesynced() bool {
	return len(l) > 1 && l[1].Time != 0
}
