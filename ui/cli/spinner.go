package cli

import (
	"github.com/apiqube/cli/ui"
	"github.com/pterm/pterm"
)

var spinners = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func (u *UI) Spinner() ui.SpinnerReporter {
	return &spinnerReporter{ui: u}
}

var _ ui.SpinnerReporter = (*spinnerReporter)(nil)

type spinnerReporter struct {
	ui      *UI
	spinner *pterm.SpinnerPrinter
}

func (s *spinnerReporter) Start(text string) {
	spinner := pterm.DefaultSpinner.
		WithText(text).
		WithShowTimer(true).
		WithSequence(spinners...).
		WithRemoveWhenDone(true).
		WithStyle(pterm.NewStyle(pterm.FgMagenta, pterm.Bold))

	s.spinner, _ = spinner.Start(text)
}

func (s *spinnerReporter) Stop(_ string) {
	_ = s.spinner.Stop()
}
