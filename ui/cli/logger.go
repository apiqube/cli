package cli

import (
	"fmt"
	"github.com/apiqube/cli/ui"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"time"
)

type logMsg struct{}

func (u *UI) Log(level ui.LogLevel, msg string) {
	fmt.Print(formatLog(level, msg))
}

func (u *UI) Logf(level ui.LogLevel, format string, args ...any) {
	u.Log(level, fmt.Sprintf(format, args...))
}

func (u *UI) Error(err error) {
	u.Log(ui.TypeError, err.Error())
}

func (u *UI) Done(msg string) {
	u.Log(ui.TypeSuccess, msg)
}

func formatLog(level ui.LogLevel, msg string) string {
	var levelText string
	var style lipgloss.Style

	switch level {
	case ui.TypeDebug:
		levelText = "DEBUG"
		style = debugStyle
	case ui.TypeInfo:
		levelText = "INFO"
		style = infoStyle
	case ui.TypeWarning:
		levelText = "WARN"
		style = warningStyle
	case ui.TypeError:
		levelText = "ERROR"
		style = errorStyle
	case ui.TypeFatal:
		levelText = "FATAL"
		style = fatalStyle
	case ui.TypeSuccess:
		levelText = "SUCCESS"
		style = successStyle
	default:
		levelText = "INFO"
		style = infoStyle
	}

	paddedLevel := fmt.Sprintf("%-7s", levelText)
	levelStyled := style.Render(paddedLevel)

	lines := strings.Split(msg, "\n")
	baseIndent := len("14:30:45") + 1 + 7 + 2

	for i := 1; i < len(lines); i++ {
		lines[i] = strings.Repeat(" ", baseIndent) + lines[i]
	}
	msg = strings.Join(lines, "\n")

	timestamp := timestampStyle.Render(time.Now().Format("15:04:05"))
	message := logStyle.Render(msg)

	return fmt.Sprintf("%s %s %s\n", timestamp, levelStyled, message)
}
