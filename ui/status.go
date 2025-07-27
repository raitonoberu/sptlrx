package ui

import (
	"strings"

	"github.com/raitonoberu/sptlrx/config"
	"github.com/raitonoberu/sptlrx/pool"

	gloss "github.com/charmbracelet/lipgloss"
)

// StatusType represents different UI states
type StatusType int

const (
	StatusLoading StatusType = iota
	StatusNoLyrics
	StatusError
	StatusReady
)

// StatusManager handles UI state messages in a modular way (V2-ready)
type StatusManager struct {
	config   *config.Config
	messages map[StatusType]string
}

// NewStatusManager creates a new status manager
func NewStatusManager(config *config.Config) *StatusManager {
	return &StatusManager{
		config: config,
		messages: map[StatusType]string{
			StatusLoading:  "Loading lyrics...",
			StatusNoLyrics: "No lyrics found for this song",
			StatusError:    "Error occurred",
			StatusReady:    "",
		},
	}
}

// SetMessage customizes a status message
func (sm *StatusManager) SetMessage(status StatusType, message string) {
	sm.messages[status] = message
}

// GetMessage retrieves a status message
func (sm *StatusManager) GetMessage(status StatusType) string {
	if msg, exists := sm.messages[status]; exists {
		return msg
	}
	return sm.messages[StatusError] // fallback
}

// RenderStatus renders a status message with proper styling
func (sm *StatusManager) RenderStatus(status StatusType, style gloss.Style, w, h int, customMsg ...string) string {
	message := sm.GetMessage(status)

	// Allow custom message override
	if len(customMsg) > 0 && customMsg[0] != "" {
		message = customMsg[0]
	}

	// Don't render empty messages
	if message == "" {
		return ""
	}

	return gloss.PlaceVertical(
		h, gloss.Center,
		style.
			Align(gloss.Center).
			Width(w).
			Render(message),
	)
}

// DetermineStatus analyzes the current state and determines UI status
func (sm *StatusManager) DetermineStatus(state pool.Update) StatusType {
	// Priority order for status determination

	// 1. Error state (highest priority)
	if state.Err != nil {
		return StatusError
	}

	// 2. No lyrics available
	if len(state.Lines) == 0 {
		return StatusNoLyrics
	}

	// 3. Ready state (has lyrics)
	return StatusReady
}

// Enhanced View method that uses StatusManager
func (m *Model) ViewWithStatusManager() string {
	if m.w < 1 || m.h < 1 {
		return ""
	}

	// Create status manager (in real implementation, this would be cached)
	statusManager := NewStatusManager(m.Config)

	// Determine current status
	status := statusManager.DetermineStatus(m.state)

	// Handle different statuses
	switch status {
	case StatusError:
		// Only show errors if not ignored
		if !m.Config.IgnoreErrors {
			return statusManager.RenderStatus(StatusError, m.styleCurrent, m.w, m.h, m.state.Err.Error())
		}
		// If errors ignored, treat as no lyrics
		return statusManager.RenderStatus(StatusNoLyrics, m.styleCurrent, m.w, m.h)

	case StatusNoLyrics:
		return statusManager.RenderStatus(StatusNoLyrics, m.styleCurrent, m.w, m.h)

	case StatusReady:
		// Render lyrics normally (existing logic)
		return m.renderLyrics()

	default:
		return statusManager.RenderStatus(StatusError, m.styleCurrent, m.w, m.h, "Unknown status")
	}
}

// renderLyrics handles the normal lyrics rendering (extracted from original View)
func (m *Model) renderLyrics() string {
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
