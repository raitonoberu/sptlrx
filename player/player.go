package player

type Player interface {
	State() (*State, error)
}

type State struct {
	// ID of the current track.
	ID string
	// Artist is the name of the artist(s).
	Artist string
	// Track is the name of the track.
	Track string
	// Position of the current track in ms.
	Position int
	// Playing means whether the track is playing at the moment.
	Playing bool
}
