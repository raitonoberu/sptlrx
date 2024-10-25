package ui

import (
	"os"
	"runtime"
	"sptlrx/config"
	"sptlrx/lyrics"
	"sptlrx/pool"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	gloss "github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type Model struct {
	Config  *config.Config
	Channel chan pool.Update

	state pool.Update
	w, h  int

	styleBefore  gloss.Style
	styleCurrent gloss.Style
	styleAfter   gloss.Style
	hAlignment   gloss.Position
}

func (m *Model) Init() tea.Cmd {
	m.styleBefore = m.Config.Style.Before.Parse()
	m.styleCurrent = m.Config.Style.Current.Parse()
	m.styleAfter = m.Config.Style.After.Parse()

	switch m.Config.Style.HAlignment {
	case "left":
		m.hAlignment = 0
	case "right":
		m.hAlignment = 1
	default:
		m.hAlignment = 0.5
	}

	return tea.Batch(waitForUpdate(m.Channel), tea.HideCursor)
}

func (m *Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		// does not work on Windows!
		m.w, m.h = msg.Width, msg.Height

	case pool.Update:
		m.state = msg

		if runtime.GOOS == "windows" {
			w, h, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				m.w, m.h = w, h
			}
		}
		cmd = waitForUpdate(m.Channel)

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
			if m.state.Playing && lyrics.Timesynced(m.state.Lines) {
				break
			}
			m.state.Index -= 1
			if m.state.Index < 0 {
				m.state.Index = 0
			}
		case "down":
			if m.state.Playing && lyrics.Timesynced(m.state.Lines) {
				break
			}
			m.state.Index += 1
			if m.state.Index >= len(m.state.Lines) {
				m.state.Index = len(m.state.Lines) - 1
			}
		}
	}
	return m, cmd
}

func (m *Model) View() string {
	if m.w < 1 || m.h < 1 {
		return ""
	}
	if m.state.Err != nil && !m.Config.IgnoreErrors {
		return gloss.PlaceVertical(
			m.h, gloss.Center,
			m.styleCurrent.
				Align(gloss.Center).
				Width(m.w).
				Render(m.state.Err.Error()),
		)
	}
	if len(m.state.Lines) == 0 {
        placeholder := "<" + m.state.ID + ">" + "\n\nNo lyrics found"

        return gloss.PlaceVertical(
            m.h, gloss.Center,
            m.styleAfter.
                Align(gloss.Center).
                Width(m.w).
                Render( placeholder ),
        )
	}

	curLine := m.styleCurrent.
		Width(m.w).
		Align(m.hAlignment).
		Render(m.state.Lines[m.state.Index].Words)
	curLines := strings.Split(curLine, "\n")

	curLen := len(curLines)
	beforeLen := (m.h - curLen) / 2
	afterLen := m.h - beforeLen - curLen

	lines := make([]string, beforeLen+curLen+afterLen)

	// fill lines before current
	var filledBefore int
	var beforeIndex = m.state.Index - 1
	for filledBefore < beforeLen {
		index := beforeLen - filledBefore - 1
		if index < 0 || beforeIndex < 0 {
			filledBefore += 1
			continue
		}
		line := m.styleBefore.
			Width(m.w).
			Align(m.hAlignment).
			Render(m.state.Lines[beforeIndex].Words)
		beforeIndex -= 1
		beforeLines := strings.Split(line, "\n")
		for i := len(beforeLines) - 1; i >= 0; i-- {
			lineIndex := index - i
			if lineIndex >= 0 {
				lines[lineIndex] = beforeLines[len(beforeLines)-1-i]
			}
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
	var afterIndex = m.state.Index + 1
	for filledAfter < afterLen {
		index := beforeLen + curLen + filledAfter
		if index >= len(lines) || afterIndex >= len(m.state.Lines) {
			filledAfter += 1
			continue
		}
		line := m.styleAfter.
			Width(m.w).
			Align(m.hAlignment).
			Render(m.state.Lines[afterIndex].Words)
		afterIndex += 1
		afterLines := strings.Split(line, "\n")
		for i, line := range afterLines {
			lineIndex := index + i
			if lineIndex < len(lines) {
				lines[lineIndex] = line
			}
			filledAfter += 1
		}
	}

	return gloss.JoinVertical(m.hAlignment, lines...)
}

func waitForUpdate(ch chan pool.Update) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}
