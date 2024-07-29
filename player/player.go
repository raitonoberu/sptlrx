package player

type Player interface {
	State() (*State, error)
}

type TrackMetadata struct {
	// ID of the current track.
	ID string
	// URI is the path to the local music file, if it exists.
	// May be either absolute or relative to the local music directory (configured in "local" source).
	Uri string
	// Query is a string that can be used to find lyrics.
	Query string
}

type State struct {
	Track TrackMetadata
	// Position of the current track in ms.
	Position int
	// Playing means whether the track is playing at the moment.
	Playing bool
}
