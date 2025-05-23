package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/apiqube/cli/ui"
	"github.com/charmbracelet/lipgloss"
)

type LogPair struct {
	Message string
	Style   *lipgloss.Style
}

func (p LogPair) String() string {
	if p.Style == nil {
		return LogStyle.Render(p.Message)
	}

	return p.Style.Render(p.Message)
}

func (u *UI) Log(level ui.LogLevel, msg string) {
	fmt.Print(formatLog(level, msg))
}

func (u *UI) Logf(level ui.LogLevel, format string, args ...any) {
	u.Log(level, fmt.Sprintf(format, args...))
}

func (u *UI) LogStyled(level ui.LogLevel, pairs ...LogPair) {
	var logBuilder strings.Builder

	var levelText string
	var style lipgloss.Style

	switch level {
	case ui.TypeDebug:
		levelText = "DEBUG"
		style = DebugStyle
	case ui.TypeInfo:
		levelText = "INFO"
		style = InfoStyle
	case ui.TypeWarning:
		levelText = "WARN"
		style = WarningStyle
	case ui.TypeError:
		levelText = "ERROR"
		style = ErrorStyle
	case ui.TypeFatal:
		levelText = "FATAL"
		style = FatalStyle
	case ui.TypeSuccess:
		levelText = "SUCCESS"
		style = SuccessStyle
	default:
		levelText = "INFO"
		style = InfoStyle
	}

	paddedLevel := fmt.Sprintf("%-7s", levelText)
	levelStyled := style.Render(paddedLevel)
	timestamp := TimestampStyle.Render(time.Now().Format("15:04:05"))

	for _, pair := range pairs {
		msg := pair.Message
		if pair.Style != nil {
			msg = pair.Style.Render(msg)
		} else {
			msg = LogStyle.Render(msg)
		}
		logBuilder.WriteString(msg)
	}

	msg := logBuilder.String()
	lines := strings.Split(msg, "\n")
	baseIndent := len("14:30:45") + 1 + 7 + 2

	for i := 1; i < len(lines); i++ {
		lines[i] = strings.Repeat(" ", baseIndent) + lines[i]
	}
	msg = strings.Join(lines, "\n")

	message := LogStyle.Render(msg)
	fmt.Printf("%s %s %s\n", timestamp, levelStyled, message)
}

func (u *UI) LogStyledf(level ui.LogLevel, format string, pairs ...LogPair) {
	args := make([]any, len(pairs))
	for i, pair := range pairs {
		msg := pair.Message
		if pair.Style != nil {
			msg = pair.Style.Render(msg)
		} else {
			msg = LogStyle.Render(msg)
		}
		args[i] = msg
	}

	formattedMsg := fmt.Sprintf(format, args...)

	u.LogStyled(level, LogPair{Message: formattedMsg})
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
		style = DebugStyle
	case ui.TypeInfo:
		levelText = "INFO"
		style = InfoStyle
	case ui.TypeWarning:
		levelText = "WARN"
		style = WarningStyle
	case ui.TypeError:
		levelText = "ERROR"
		style = ErrorStyle
	case ui.TypeFatal:
		levelText = "FATAL"
		style = FatalStyle
	case ui.TypeSuccess:
		levelText = "SUCCESS"
		style = SuccessStyle
	default:
		levelText = "INFO"
		style = InfoStyle
	}

	paddedLevel := fmt.Sprintf("%-7s", levelText)
	levelStyled := style.Render(paddedLevel)

	lines := strings.Split(msg, "\n")
	baseIndent := len("14:30:45") + 1 + 7 + 2

	for i := 1; i < len(lines); i++ {
		lines[i] = strings.Repeat(" ", baseIndent) + lines[i]
	}
	msg = strings.Join(lines, "\n")

	timestamp := TimestampStyle.Render(time.Now().Format("15:04:05"))
	message := LogStyle.Render(msg)

	return fmt.Sprintf("%s %s %s\n", timestamp, levelStyled, message)
}
