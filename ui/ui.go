package ui

import (
	"os"
	"runtime"
	"sptlrx/spotify"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	gloss "github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

const (
	// TimerIterval sets the interval for the internal timer (ms)
	TimerInterval = 200
	// StatusUpdateInterval sets the interval for updating Spotify status (ms)
	StatusUpdateInterval = 3000
)

type currentUpdateMsg *spotify.CurrentlyPlaying
type positionUpdateMsg bool
type timeUpdateMsg bool
type lyricsUpdateMsg []*spotify.LyricsLine

type Model struct {
	Client *spotify.SpotifyClient

	StyleBefore  gloss.Style
	StyleCurrent gloss.Style
	StyleAfter   gloss.Style
	HAlignment   gloss.Position

	w, h int

	id       string
	playing  bool
	position int

	lastUpdate time.Time
	audioDelay int

	lines []*spotify.LyricsLine
	index int
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(tickPosition(), updateCurrent(m.Client), tickTime(), tea.HideCursor)
}

func (m *Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		// does not work on Windows!
		m.w, m.h = msg.Width, msg.Height

	case currentUpdateMsg:
		if msg.ID != m.id {
			m.index = 0
			m.lines = nil
			cmd = updateLyrics(m.Client, msg.ID)
		}

		m.id = msg.ID
		m.playing = msg.Playing
		m.position = msg.Position
		m.lastUpdate = time.Now()
		m.updateIndex()

	case positionUpdateMsg:
		cmd = tickPosition()
		if m.playing {
			now := time.Now()
			m.position += int(now.Sub(m.lastUpdate).Milliseconds())
			m.lastUpdate = now
			m.updateIndex()
		}

		// instead of WindowSizeMsg
		if runtime.GOOS == "windows" {
			w, h, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				m.w, m.h = w, h
			}
		}

	case timeUpdateMsg:
		cmd = tea.Batch(updateCurrent(m.Client), tickTime())

	case lyricsUpdateMsg:
		m.lines = msg
		m.updateIndex()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit

		case "+":
			m.audioDelay += 100
		case "-":
			m.audioDelay -= 100

		case "left":
			m.HAlignment -= 0.5
			if m.HAlignment < 0 {
				m.HAlignment = 0
			}
		case "right":
			m.HAlignment += 0.5
			if m.HAlignment > 1 {
				m.HAlignment = 1
			}

		case "up":
			if !m.playing || (len(m.lines) > 1 && m.lines[1].Time == 0) {
				m.index -= 1
				if m.index < 0 {
					m.index = 0
				}
			}
		case "down":
			if !m.playing || (len(m.lines) > 1 && m.lines[1].Time == 0) {
				m.index += 1
				if m.index >= len(m.lines) {
					m.index = len(m.lines) - 1
				}
			}
		}
	}

	return m, cmd
}

func (m *Model) View() string {
	if len(m.lines) == 0 || m.w < 1 || m.h < 1 {
		// nothing to show
		return ""
	}

	cur := m.StyleCurrent.Width(m.w).Align(m.HAlignment).Render(m.lines[m.index].Words)
	curLines := strings.Split(cur, "\n")
	curLen := len(curLines)
	beforeLen := (m.h - curLen) / 2
	afterLen := m.h - beforeLen - curLen

	lines := make([]string, beforeLen+curLen+afterLen)

	// fill lines before current
	var filledBefore int
	var beforeIndex = m.index - 1
	for filledBefore < beforeLen {
		index := beforeLen - filledBefore - 1
		if index >= 0 && beforeIndex >= 0 {
			line := m.StyleBefore.Width(m.w).Align(m.HAlignment).Render(m.lines[beforeIndex].Words)
			beforeIndex -= 1
			beforeLines := strings.Split(line, "\n")
			for i := len(beforeLines) - 1; i >= 0; i-- {
				lineIndex := index - i
				if lineIndex >= 0 {
					lines[lineIndex] = beforeLines[len(beforeLines)-1-i]
				}
				filledBefore += 1
			}
		} else {
			filledBefore += 1
		}
	}

	// fill current lines
	var curIndex = beforeLen
	for i, line := range curLines {
		index := curIndex + i
		if index >= 0 && index < len(lines) {
			lines[index] = line
		}
	}

	// fill lines after current
	var filledAfter int
	var afterIndex = m.index + 1
	for filledAfter < afterLen {
		index := beforeLen + curLen + filledAfter
		if index < len(lines) && afterIndex < len(m.lines) {
			line := m.StyleAfter.Width(m.w).Align(m.HAlignment).Render(m.lines[afterIndex].Words)
			afterIndex += 1
			afterLines := strings.Split(line, "\n")
			for i, line := range afterLines {
				lineIndex := index + i
				if lineIndex < len(lines) {
					lines[lineIndex] = line
				}
				filledAfter += 1
			}
		} else {
			filledAfter += 1
		}
	}

	return gloss.JoinVertical(m.HAlignment, lines...)
}

func (m *Model) updateIndex() {
	if len(m.lines) <= 1 {
		m.index = 0
		return
	}

	if !m.playing || m.lines[1].Time == 0 {
		return
	}

	position := m.position + m.audioDelay

	if position >= m.lines[m.index].Time {
		if m.index == len(m.lines)-1 {
			return
		}
		if position < m.lines[m.index+1].Time {
			return
		} else {
			// search after
			for i, line := range m.lines[m.index:] {
				if position < line.Time {
					m.index = m.index + i - 1
					return
				}
			}
		}
	}
	// search before
	for i, line := range m.lines {
		if position < line.Time {
			if i != 0 {
				m.index = i - 1
				return
			}
			return
		}
	}
	m.index = len(m.lines) - 1
}

func updateCurrent(client *spotify.SpotifyClient) tea.Cmd {
	return func() tea.Msg {
		current, err := client.Current()
		if err != nil {
			panic(err)
		}
		if current == nil {
			return nil
		}
		return currentUpdateMsg(current)
	}
}

func updateLyrics(client *spotify.SpotifyClient, id string) tea.Cmd {
	return func() tea.Msg {
		l, err := client.Lyrics(id)
		if err != nil {
			panic(err)
		}
		if l == nil {
			return nil
		}
		return lyricsUpdateMsg(l)
	}
}

func tickPosition() tea.Cmd {
	return tea.Tick(TimerInterval*time.Millisecond, func(t time.Time) tea.Msg {
		return positionUpdateMsg(true)
	})
}

func tickTime() tea.Cmd {
	return tea.Tick(StatusUpdateInterval*time.Millisecond, func(t time.Time) tea.Msg {
		return timeUpdateMsg(true)
	})
}
