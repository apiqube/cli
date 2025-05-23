package cli

import "github.com/charmbracelet/lipgloss"

var (
	TimestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#949494"))

	LogStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e4e4e4"))

	DebugStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#005fff"))

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00afff"))

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff8700"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d70000")).
			Bold(true)

	FatalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5fd700"))
)
