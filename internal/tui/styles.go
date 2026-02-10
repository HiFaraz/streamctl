package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("62")  // Purple
	secondaryColor = lipgloss.Color("241") // Gray
	successColor   = lipgloss.Color("42")  // Green
	warningColor   = lipgloss.Color("214") // Orange
	errorColor     = lipgloss.Color("196") // Red
	infoColor      = lipgloss.Color("39")  // Cyan

	// State colors
	pendingColor    = lipgloss.Color("214") // Orange
	inProgressColor = lipgloss.Color("39")  // Cyan
	blockedColor    = lipgloss.Color("196") // Red
	doneColor       = lipgloss.Color("42")  // Green

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")). // Yellow
			Background(lipgloss.Color("57"))   // Purple

	normalStyle = lipgloss.NewStyle()

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(secondaryColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	stateStyles = map[string]lipgloss.Style{
		"pending":     lipgloss.NewStyle().Foreground(pendingColor),
		"in_progress": lipgloss.NewStyle().Foreground(inProgressColor),
		"blocked":     lipgloss.NewStyle().Foreground(blockedColor),
		"done":        lipgloss.NewStyle().Foreground(doneColor),
	}
)

// StateStyle returns the style for a given state
func StateStyle(state string) lipgloss.Style {
	if style, ok := stateStyles[state]; ok {
		return style
	}
	return normalStyle
}
