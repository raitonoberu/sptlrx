package pool

import (
	"sptlrx/spotify"
	"time"
)

const (
	// TimerIterval sets the interval for the internal timer (ms)
	TimerInterval = 200
	// StatusUpdateInterval sets the interval for updating Spotify status (ms)
	StatusUpdateInterval = 3000
)

type Update struct {
	Lines   spotify.LyricsLines
	Index   int
	Playing bool

	Err error
}

type statusUpdate struct {
	status *spotify.CurrentlyPlaying
	err    error
}

func Listen(client *spotify.SpotifyClient, ch chan Update) {
	var id string
	var playing bool
	var position int

	var lines spotify.LyricsLines
	var index int

	var (
		timerCh  = make(chan int, 1)
		updateCh = make(chan statusUpdate, 1)
	)

	go listenTimer(timerCh)
	go listenUpdate(client, updateCh)

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

			if update.status == nil {
				if lines != nil {
					changed = true
					id = ""
					lines = nil
					playing = false
					index = 0
				}
				break
			}
			if update.status.ID != id {
				changed = true
				id = update.status.ID
				newLines, err := client.Lyrics(id)
				if err != nil {
					ch <- Update{
						Err: err,
					}
					break
				}
				lines = newLines
				index = 0
			}
			if update.status.Playing != playing {
				changed = true
				playing = update.status.Playing
			}
			position = update.status.Position
			newIndex := getIndex(position, index, lines)
			if newIndex != index {
				changed = true
				index = newIndex
			}

		case <-timerCh:
			if playing && lines.Timesynced() {
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

func listenTimer(ch chan int) {
	for {
		ch <- 0
		time.Sleep(time.Millisecond * TimerInterval)
	}
}

func listenUpdate(client *spotify.SpotifyClient, ch chan statusUpdate) {
	for {
		status, err := client.Current()
		ch <- statusUpdate{
			status: status,
			err:    err,
		}
		time.Sleep(time.Millisecond * StatusUpdateInterval)
	}
}

func getIndex(position, curIndex int, lines []*spotify.LyricsLine) int {
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
