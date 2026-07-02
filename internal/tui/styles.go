package tui

import "github.com/charmbracelet/lipgloss"

var (
	AppStyle = lipgloss.NewStyle().Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFF")).
			Background(lipgloss.Color("#1a73e8")).
			Padding(0, 2)

	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFF")).
			Background(lipgloss.Color("#1a73e8")).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Foreground(lipgloss.AdaptiveColor{Light: "#555", Dark: "#999"}).
				Background(lipgloss.AdaptiveColor{Light: "#DDD", Dark: "#333"})

	ConnectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#2e7d32", Dark: "#66bb6a"}).
			Bold(true)

	DisconnectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#c62828", Dark: "#ef5350"})

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4444")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1a73e8")).
			Bold(true)

	SubtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#999", Dark: "#666"})

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777"})

	CodeStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.AdaptiveColor{Light: "#222", Dark: "#ddd"}).
			Background(lipgloss.AdaptiveColor{Light: "#F5F5F5", Dark: "#1a1a2e"})

	TreeStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFF")).
				Background(lipgloss.Color("#333")).
				Padding(0, 1)

	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	TableAltCellStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(lipgloss.AdaptiveColor{Light: "#EEE", Dark: "#252525"})
)
