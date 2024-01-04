package pool

import (
	"sptlrx/config"
	"sptlrx/lyrics"
	"sptlrx/player"
	"time"
)

// Update represents the state of the lyrics.
type Update struct {
	Lines   []lyrics.Line
	Index   int
	Playing bool

	Err error
}

// Listen polls for lyrics updates and writes them to the channel.
func Listen(
	player player.Player,
	provider lyrics.Provider,
	conf *config.Config,
	ch chan Update,
) {
	stateCh := make(chan playerState)
	go listenPlayer(player, stateCh, conf.UpdateInterval)

	ticker := time.NewTicker(
		time.Millisecond * time.Duration(conf.TimerInterval),
	)

	var (
		state      playerState
		index      int
		lines      []lyrics.Line
		lastUpdate time.Time
	)

	for {
		changed := false

		select {
		case newState := <-stateCh:
			lastUpdate = time.Now()

			if newState.ID != state.ID {
				changed = true
				if newState.ID != "" {
					newLines, err := provider.Lyrics(newState.ID, newState.Query)
					if err != nil {
						state.Err = err
					}
					lines = newLines
				} else {
					lines = nil
				}
				index = 0
			}
			if newState.Playing != state.Playing {
				changed = true
			}
			state = newState
		case <-ticker.C:
			if !state.Playing || !lyrics.Timesynced(lines) {
				break
			}

			now := time.Now()
			state.Position += int(now.Sub(lastUpdate).Milliseconds())
			lastUpdate = now
		}

		newIndex := getIndex(state.Position, index, lines)
		if newIndex != index {
			changed = true
			index = newIndex
		}

		if changed {
			ch <- Update{
				Lines:   lines,
				Index:   index,
				Playing: state.Playing,
				Err:     state.Err,
			}
		}
	}
}

type playerState struct {
	player.State
	Err error
}

func listenPlayer(player player.Player, ch chan playerState, interval int) {
	for {
		state, err := player.State()

		st := playerState{Err: err}
		if state != nil {
			st.ID = state.ID
			st.Query = state.Query
			st.Playing = state.Playing
			st.Position = state.Position
		}
		ch <- st

		time.Sleep(time.Millisecond * time.Duration(interval))
	}
}

// getIndex is an effective alghoritm to get current line's index
func getIndex(position, curIndex int, lines []lyrics.Line) int {
	if len(lines) <= 1 {
		return 0
	}

	if position >= lines[curIndex].Time {
		// search after
		for i := curIndex + 1; i < len(lines); i++ {
			if position < lines[i].Time {
				return i - 1
			}
		}
		return len(lines) - 1
	}

	// search before
	for i := curIndex; i > 0; i-- {
		if position > lines[i].Time {
			return i
		}
	}
	return 0
}
