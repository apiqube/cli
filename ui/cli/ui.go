package cli

import (
	"github.com/apiqube/cli/ui"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pterm/pterm"
)

var _ ui.UI = (*UI)(nil)

type UI struct {
	program *tea.Program
	model   *uiModel
	logger  *pterm.Logger
}

func NewUI() *UI {
	model := &uiModel{
		progressComp: progress.New(
			progress.WithWidth(50),
			progress.WithDefaultScaledGradient(),
		),
		currentView: ViewNone,
	}

	return &UI{
		program: tea.NewProgram(model),
		model:   model,
		logger: pterm.DefaultLogger.
			WithLevel(pterm.LogLevelTrace).
			WithTime(true).
			WithTimeFormat("2006-01-02 15:04:05"),
	}
}
