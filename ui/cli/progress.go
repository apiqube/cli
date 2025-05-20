package cli

import (
	"github.com/apiqube/cli/ui"
)

type progressData struct {
	total    int
	current  int
	title    string
	messages []string
}

func (u *UI) Progress() ui.ProgressReporter {
	return &progressReporter{ui: u}
}

type progressReporter struct {
	ui *UI
}

type progressMsg struct{}

func (pr *progressReporter) Start(total int, title string) {
	pr.ui.model.progressData = progressData{
		total:   total,
		current: 0,
		title:   title,
	}
	pr.ui.model.currentView = ViewProgress
	pr.ui.program.Send(progressMsg{})
}

func (pr *progressReporter) Increment(msg string) {
	pr.ui.model.progressData.current++
	pr.ui.model.progressData.messages = append(pr.ui.model.progressData.messages, msg)
	pr.ui.program.Send(progressMsg{})
}

func (pr *progressReporter) Stop() {
	pr.ui.model.currentView = ViewNone
}
