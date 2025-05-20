package cli

import (
	"github.com/apiqube/cli/ui"
	tea "github.com/charmbracelet/bubbletea"
	"sync"
)

var (
	instance *UI
	once     sync.Once
	enabled  bool
	//spinners = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
)

func Init() {
	once.Do(func() {
		instance = NewUI()
		enabled = true
		go func() {
			if _, err := instance.program.Run(); err != nil {
				panic(err)
			}
		}()
	})
}

func Stop() {
	if instance != nil && enabled {
		enabled = false
		instance.program.Send(tea.Quit())
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
