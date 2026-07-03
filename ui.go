package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const glowTickInterval = 120 * time.Millisecond

// glowColors pulses low-to-high-to-low brightness to read as a "glowing" border.
var glowColors = []string{"24", "31", "39", "45", "51", "45", "39", "31"}

var (
	modeStyle   = lipgloss.NewStyle().Padding(0, 1)
	activeMode  = lipgloss.NewStyle().Padding(0, 1).Bold(true).Reverse(true)
	titleStyle  = lipgloss.NewStyle().Bold(true).Underline(true)
	helpStyle   = lipgloss.NewStyle().Faint(true)
	addedStyle  = lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("42"))
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	borderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
)

const helpText = "ctrl+n/ctrl+b switch mode  ·  ctrl+p proofread  ·  ctrl+k clear  ·  " +
	"any key resumes editing  ·  esc quit"

const keyVisiblePrefix = 3

type glowTickMsg struct{}

func glowTick() tea.Cmd {
	return tea.Tick(glowTickInterval, func(time.Time) tea.Msg { return glowTickMsg{} })
}

// obfuscateKey shows only the first few characters of a secret, e.g. "sk-***".
func obfuscateKey(key string) string {
	if key == "" {
		return "(none)"
	}

	if len(key) <= keyVisiblePrefix {
		return strings.Repeat("*", len(key))
	}

	return key[:keyVisiblePrefix] + "***"
}

func (m model) View() string {
	var out strings.Builder

	out.WriteString(titleStyle.Render("Mode:") + "  ")

	for i, md := range modes {
		if i == m.modeIdx {
			out.WriteString(activeMode.Render(md))
		} else {
			out.WriteString(modeStyle.Render(md))
		}
	}

	status := m.llmHost + "  ·  " + m.llmModel + "  ·  key: " + obfuscateKey(m.llmKey)

	out.WriteString("\n")
	out.WriteString(helpStyle.Render(status) + "\n\n")

	box := borderStyle
	if m.busy {
		box = box.BorderForeground(lipgloss.Color(glowColors[m.glow]))
	}

	if m.rendered != "" {
		out.WriteString(box.Render(lipgloss.NewStyle().Width(inputWidth).Render(m.rendered)))
	} else {
		out.WriteString(box.Render(m.input.View()))
	}

	out.WriteString("\n\n")

	if m.err != nil {
		out.WriteString(errStyle.Render(m.err.Error()) + "\n")
	}

	out.WriteString("\n" + helpStyle.Render(helpText))

	return out.String()
}
