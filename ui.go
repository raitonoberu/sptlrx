package main

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

var client *spotify.SpotifyClient

const (
	// TimerIterval sets the interval for the internal timer (ms)
	TimerInterval = 200
	// StatusUpdateInterval sets the interval for updating Spotify status (ms)
	StatusUpdateInterval = 3000
)

var (
	faintStyle = gloss.NewStyle().Faint(true)
	boldStyle  = gloss.NewStyle().Bold(true)
)

type CurrentUpdateMsg *spotify.CurrentlyPlaying
type PositionUpdateMsg bool
type TimeUpdateMsg bool
type LyricsUpdateMsg []*spotify.LyricsLine

type model struct {
	w, h int

	ID       string
	Playing  bool
	Position int

	lastUpdate time.Time
	audioDelay int

	lines []*spotify.LyricsLine
	index int

	hAlignment gloss.Position
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(tickPosition(), updateCurrent(), tickTime())
}

func (m *model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		// does not work on Windows!
		m.w, m.h = msg.Width, msg.Height

	case CurrentUpdateMsg:
		if msg.ID != m.ID {
			m.index = 0
			m.lines = nil
			cmd = updateLyrics(msg.ID)
		}

		m.ID = msg.ID
		m.Playing = msg.Playing
		m.Position = msg.Position
		m.lastUpdate = time.Now()
		m.updateIndex()

	case PositionUpdateMsg:
		cmd = tickPosition()
		if m.Playing {
			now := time.Now()
			m.Position += int(now.Sub(m.lastUpdate).Milliseconds())
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

	case TimeUpdateMsg:
		cmd = tea.Batch(updateCurrent(), tickTime())

	case LyricsUpdateMsg:
		m.lines = msg
		m.updateIndex()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		// fixing delay
		case "+":
			m.audioDelay += 100
		case "-":
			m.audioDelay -= 100
		// align
		case "left":
			m.hAlignment -= 0.5
			if m.hAlignment < 0 {
				m.hAlignment = 0
			}
		case "right":
			m.hAlignment += 0.5
			if m.hAlignment > 1 {
				m.hAlignment = 1
			}
		// move
		case "up":
			if !m.Playing || (len(m.lines) > 1 && m.lines[1].Time == 0) {
				m.index -= 1
				if m.index < 0 {
					m.index = 0
				}
			}
		case "down":
			if !m.Playing || (len(m.lines) > 1 && m.lines[1].Time == 0) {
				m.index += 1
				if m.index >= len(m.lines) {
					m.index = len(m.lines) - 1
				}
			}
		}
	}

	return m, cmd
}

func (m *model) View() string {
	if len(m.lines) == 0 || m.w < 1 || m.h < 1 {
		// nothing to show
		return ""
	}

	cur := boldStyle.Width(m.w).Align(m.hAlignment).Render(m.lines[m.index].Words)
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
			line := boldStyle.Width(m.w).Align(m.hAlignment).Render(m.lines[beforeIndex].Words)
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
			line := faintStyle.Width(m.w).Align(m.hAlignment).Render(m.lines[afterIndex].Words)
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

	return gloss.JoinVertical(m.hAlignment, lines...)
}

func (m *model) updateIndex() {
	if len(m.lines) <= 1 {
		m.index = 0
		return
	}

	if !m.Playing || m.lines[1].Time == 0 {
		return
	}

	position := m.Position + m.audioDelay

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

func updateCurrent() tea.Cmd {
	return func() tea.Msg {
		current, err := client.Current()
		if err != nil {
			panic(err)
		}
		if current == nil {
			return nil
		}
		return CurrentUpdateMsg(current)
	}
}

func updateLyrics(id string) tea.Cmd {
	return func() tea.Msg {
		l, err := client.Lyrics(id)
		if err != nil {
			panic(err)
		}
		if l == nil {
			return nil
		}
		return LyricsUpdateMsg(l)
	}
}

func tickPosition() tea.Cmd {
	return tea.Tick(TimerInterval*time.Millisecond, func(t time.Time) tea.Msg {
		return PositionUpdateMsg(true)
	})
}

func tickTime() tea.Cmd {
	return tea.Tick(StatusUpdateInterval*time.Millisecond, func(t time.Time) tea.Msg {
		return TimeUpdateMsg(true)
	})
}
