package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	timestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#949494"))

	logStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e4e4e4"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5fd700"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d70000")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff8700")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00afff"))

	snippetStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("236")).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("59"))

	progressBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("57")).
				Background(lipgloss.Color("236"))

	progressTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255"))

	loaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff0087"))
)

func getStyle(t MessageType) lipgloss.Style {
	switch t {
	case TypeSuccess:
		return successStyle
	case TypeError:
		return errorStyle
	case TypeWarning:
		return warningStyle
	case TypeInfo:
		return infoStyle
	case TypeSnippet:
		return snippetStyle
	default:
		return logStyle
	}
}

func printStyled(t MessageType, a ...interface{}) {
	if IsEnabled() {
		instance.queueUpdate(func(m *uiModel) {
			m.content = append(m.content, message{
				text:      fmt.Sprint(a...),
				style:     getStyle(t),
				timestamp: time.Now(),
			})
		})
	}
}

func printStyledf(t MessageType, format string, a ...interface{}) {
	if IsEnabled() {
		instance.queueUpdate(func(m *uiModel) {
			m.content = append(m.content, message{
				text:      fmt.Sprintf(format, a...),
				style:     getStyle(t),
				timestamp: time.Now(),
			})
		})
	}
}

func printStyledln(t MessageType, a ...interface{}) {
	if IsEnabled() {
		instance.queueUpdate(func(m *uiModel) {
			m.content = append(m.content, message{
				text:      fmt.Sprintln(a...),
				style:     getStyle(t),
				timestamp: time.Now(),
			})
		})
	}
}
