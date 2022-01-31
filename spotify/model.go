package spotify

type LyricsLine struct {
	Time  int    `json:"startTimeMs,string"`
	Words string `json:"words"`
}

type CurrentlyPlaying struct {
	ID       string
	Position int
	Playing  bool
}
