package player

import "net/url"

type Player interface {
	State() (*State, error)
}

type TrackMetadata struct {
	// ID of the current track.
	ID string
	// URI to music file, if it is known. May be a (local) relative path.
	Uri *url.URL
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
