package cli

import (
	"github.com/apiqube/cli/ui"
	"github.com/pterm/pterm"
)

func (u *UI) Progress() ui.ProgressReporter {
	return &progressReporter{ui: u}
}

type progressReporter struct {
	ui  *UI
	bar *pterm.ProgressbarPrinter
}

func (pr *progressReporter) Start(total int, title string) {
	p := pterm.DefaultProgressbar.
		WithTotal(total).
		WithTitle(title).
		WithMaxWidth(75).
		WithRemoveWhenDone(true).
		WithBarCharacter("█").
		WithBarFiller("░").
		WithBarStyle(pterm.NewStyle(pterm.FgLightMagenta)).
		WithTitleStyle(pterm.NewStyle(pterm.FgWhite, pterm.Bold))

	pr.bar, _ = p.Start()
}

func (pr *progressReporter) Increment(msg string) {
	if msg != "" {
		pr.bar.UpdateTitle(msg)
	}
	pr.bar.Increment()
}

func (pr *progressReporter) Stop() {
	_, _ = pr.bar.Stop()
}
