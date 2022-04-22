package ui

import (
	"os"
	"runtime"
	"sptlrx/config"
	"sptlrx/pool"
	"sptlrx/spotify"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	gloss "github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type updateMsg pool.Update

type Model struct {
	Client *spotify.SpotifyClient
	Config *config.Config

	styleBefore  gloss.Style
	styleCurrent gloss.Style
	styleAfter   gloss.Style
	hAlignment   gloss.Position

	w, h int

	channel chan pool.Update

	lines   spotify.LyricsLines
	index   int
	playing bool
	err     error
}

func (m *Model) Init() tea.Cmd {
	m.styleBefore = m.Config.Style.Before.Parse()
	m.styleCurrent = m.Config.Style.Current.Parse()
	m.styleAfter = m.Config.Style.After.Parse()

	m.hAlignment = 0.5
	switch m.Config.Style.HAlignment {
	case "left":
		m.hAlignment = 0
	case "right":
		m.hAlignment = 1
	}

	m.channel = make(chan pool.Update)
	go pool.Listen(m.Client, m.Config, m.channel)
	return tea.Batch(waitForUpdate(m.channel), tea.HideCursor)
}

func (m *Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		// does not work on Windows!
		m.w, m.h = msg.Width, msg.Height

	case updateMsg:
		m.lines = msg.Lines
		m.index = msg.Index
		m.playing = msg.Playing
		m.err = msg.Err

		if runtime.GOOS == "windows" {
			w, h, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				m.w, m.h = w, h
			}
		}
		cmd = waitForUpdate(m.channel)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			cmd = tea.Quit

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

		case "up":
			if !m.playing || !m.lines.Timesynced() {
				m.index -= 1
				if m.index < 0 {
					m.index = 0
				}
			}
		case "down":
			if !m.playing || !m.lines.Timesynced() {
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
	if m.w < 1 || m.h < 1 {
		return ""
	}
	if m.err != nil {
		return gloss.PlaceVertical(
			m.h, gloss.Center,
			m.styleCurrent.
				Align(gloss.Center).
				Width(m.w).
				Render(m.err.Error()),
		)
	}
	if len(m.lines) == 0 {
		return ""
	}

	cur := m.styleCurrent.
		Width(m.w).
		Align(m.hAlignment).
		Render(m.lines[m.index].Words)
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
			line := m.styleBefore.
				Width(m.w).
				Align(m.hAlignment).
				Render(m.lines[beforeIndex].Words)
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
			line := m.styleAfter.
				Width(m.w).
				Align(m.hAlignment).
				Render(m.lines[afterIndex].Words)
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

func waitForUpdate(ch chan pool.Update) tea.Cmd {
	return func() tea.Msg {
		return updateMsg(<-ch)
	}
}
