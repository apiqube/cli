package cli

import (
	"sync"

	"github.com/apiqube/cli/ui"
)

var (
	instance *UI
	once     sync.Once
	enabled  bool
)

func Init() {
	once.Do(func() {
		instance = NewUI()
		enabled = true
	})
}

func Stop() {
	if instance != nil && enabled {
		enabled = false
		instance = nil
	}
}

func Instance() *UI {
	if isEnabled() {
		return instance
	}

	Init()

	if isEnabled() {
		return instance
	}

	return NewUI()
}

func Table(headers []string, rows [][]string) {
	if isEnabled() {
		instance.Table(headers, rows)
	}
}

func Progress() ui.ProgressReporter {
	if isEnabled() {
		return instance.Progress()
	}

	return &progressReporter{}
}

func Spinner() ui.SpinnerReporter {
	if isEnabled() {
		return instance.Spinner()
	}

	return &spinnerReporter{}
}

func Snippet() ui.SnippetReporter {
	if isEnabled() {
		return instance.Snippet()
	}
	return &snippetReporter{}
}

func isEnabled() bool {
	return instance != nil && enabled
}
