package cli

import "github.com/charmbracelet/lipgloss"

var (
	timestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#949494"))

	logStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e4e4e4"))

	debugStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#005fff"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00afff"))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff8700"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d70000")).
			Bold(true)

	fatalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5fd700"))
)
