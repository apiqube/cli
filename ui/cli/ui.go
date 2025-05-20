package cli

import (
	"github.com/apiqube/cli/ui"
	"github.com/pterm/pterm"
)

var _ ui.UI = (*UI)(nil)

type UI struct {
	logger *pterm.Logger
}

func NewUI() *UI {
	return &UI{
		logger: pterm.DefaultLogger.
			WithLevel(pterm.LogLevelTrace).
			WithTime(true).
			WithTimeFormat("2006-01-02 15:04:05"),
	}
}
