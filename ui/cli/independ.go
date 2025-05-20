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

func inEnabled() bool {
	return instance != nil && enabled
}

func Table(headers []string, rows [][]string) {
	if inEnabled() {
		instance.Table(headers, rows)
	}
}

func Progress() ui.ProgressReporter {
	if inEnabled() {
		return instance.Progress()
	}

	return &progressReporter{}
}

func Spinner() ui.SpinnerReporter {
	if inEnabled() {
		return instance.Spinner()
	}

	return &spinnerReporter{}
}

func Snippet() ui.SnippetReporter {
	if inEnabled() {
		return instance.Snippet()
	}
	return &snippetReporter{}
}
