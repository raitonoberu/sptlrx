package lyrics

type Provider interface {
	Lyrics(id, query string) ([]Line, error)
}

type Line struct {
	Time  int    `json:"time"`
	Words string `json:"words"`
}

func Timesynced(lines []Line) bool {
	return len(lines) > 1 && lines[1].Time != 0
}
