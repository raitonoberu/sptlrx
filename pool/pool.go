package pool

import (
	"sptlrx/config"
	"sptlrx/lyrics"
	"sptlrx/player"
	"time"
)

type Update struct {
	Lines   []lyrics.Line
	Index   int
	Playing bool

	Err error
}

type stateUpdate struct {
	state *player.State
	err   error
}

func Listen(
	player player.Player,
	provider lyrics.Provider,
	conf *config.Config,
	ch chan Update,
) {
	var id string
	var playing bool
	var position int

	var lines []lyrics.Line
	var index int

	var (
		timerCh  = make(chan int, 1)
		updateCh = make(chan stateUpdate, 1)
	)

	go listenTimer(timerCh, conf.TimerInterval)
	go listenUpdate(player, updateCh, conf.UpdateInterval)

	var lastUpdate = time.Now()

	for {
		var changed bool

		select {
		case update := <-updateCh:
			lastUpdate = time.Now()

			if update.err != nil {
				ch <- Update{
					Err: update.err,
				}
				break
			}

			if update.state == nil {
				if lines != nil {
					changed = true
					id = ""
					lines = nil
					playing = false
					index = 0
				}
				break
			}
			if update.state.ID != id {
				changed = true
				id = update.state.ID
				newLines, err := provider.Lyrics(id, update.state.Query)
				if err != nil {
					ch <- Update{
						Err: err,
					}
					break
				}
				lines = newLines
				index = 0
			}
			if update.state.Playing != playing {
				changed = true
				playing = update.state.Playing
			}
			position = update.state.Position
			newIndex := getIndex(position, index, lines)
			if newIndex != index {
				changed = true
				index = newIndex
			}

		case <-timerCh:
			if playing && lyrics.Timesynced(lines) {
				now := time.Now()
				position += int(now.Sub(lastUpdate).Milliseconds())
				lastUpdate = now

				newIndex := getIndex(position, index, lines)
				if newIndex != index {
					changed = true
					index = newIndex
				}
			}
		}

		if changed {
			ch <- Update{
				Lines:   lines,
				Index:   index,
				Playing: playing,
				Err:     nil,
			}
		}
	}
}

func listenTimer(ch chan int, interval int) {
	for {
		ch <- 0
		time.Sleep(time.Millisecond * time.Duration(interval))
	}
}

func listenUpdate(player player.Player, ch chan stateUpdate, interval int) {
	for {
		state, err := player.State()
		ch <- stateUpdate{
			state: state,
			err:   err,
		}
		time.Sleep(time.Millisecond * time.Duration(interval))
	}
}

// getIndex is an effective alghoritm to get current line's index
func getIndex(position, curIndex int, lines []lyrics.Line) int {
	if len(lines) <= 1 {
		return 0
	}

	if position >= lines[curIndex].Time {
		if curIndex == len(lines)-1 {
			return curIndex
		}
		if position < lines[curIndex+1].Time {
			return curIndex
		} else {
			// search after
			for i, line := range lines[curIndex:] {
				if position < line.Time {
					return curIndex + i - 1
				}
			}
		}
	}
	// search before
	for i, line := range lines {
		if position < line.Time {
			if i != 0 {
				return i - 1
			}
			return curIndex
		}
	}
	return len(lines) - 1
}
