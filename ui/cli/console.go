package cli

import (
	"fmt"

	"github.com/apiqube/cli/ui"
)

func Print(msg string) {
	if isEnabled() {
		instance.Log(ui.TypeLog, msg)
	}
}

func Printf(format string, a ...any) {
	if isEnabled() {
		instance.Logf(ui.TypeLog, format, a...)
	}
}

func Println(a ...any) {
	if isEnabled() {
		instance.Log(ui.TypeLog, fmt.Sprintln(a...))
	}
}

func Debug(msg string) {
	if isEnabled() {
		instance.Log(ui.TypeDebug, msg)
	}
}

func Debugf(format string, a ...any) {
	if isEnabled() {
		instance.Logf(ui.TypeDebug, format, a...)
	}
}

func Info(msg string) {
	if isEnabled() {
		instance.Log(ui.TypeInfo, msg)
	}
}

func Infof(format string, a ...any) {
	if isEnabled() {
		instance.Logf(ui.TypeInfo, format, a...)
	}
}

func Warning(msg string) {
	if isEnabled() {
		instance.Log(ui.TypeWarning, msg)
	}
}

func Warningf(format string, a ...any) {
	if isEnabled() {
		instance.Logf(ui.TypeWarning, format, a...)
	}
}

func Error(msg string) {
	if isEnabled() {
		instance.Log(ui.TypeError, msg)
	}
}

func Errorf(format string, a ...any) {
	if isEnabled() {
		instance.Logf(ui.TypeError, format, a...)
	}
}

func Fatal(msg string) {
	if isEnabled() {
		instance.Log(ui.TypeFatal, msg)
	}
}

func Fatalf(format string, a ...any) {
	if isEnabled() {
		instance.Logf(ui.TypeFatal, format, a...)
	}
}

func Success(msg string) {
	if isEnabled() {
		instance.Log(ui.TypeSuccess, msg)
	}
}

func Successf(format string, a ...any) {
	if isEnabled() {
		instance.Logf(ui.TypeSuccess, format, a...)
	}
}

func Done(msg string) {
	if isEnabled() {
		instance.Done(msg)
	}
}
