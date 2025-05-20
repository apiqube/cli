package cli

import (
	"github.com/apiqube/cli/ui"
	"github.com/pterm/pterm"
)

func (u *UI) Snippet() ui.SnippetReporter {
	return &snippetReporter{ui: u}
}

var _ ui.SnippetReporter = (*snippetReporter)(nil)

type snippetReporter struct {
	ui *UI
}

func (s *snippetReporter) View(title string, data []byte) {
	pterm.Println()

	box := pterm.DefaultBox.
		WithTitle(pterm.Red(title)).
		WithTopPadding(1).
		WithLeftPadding(2).
		WithBoxStyle(pterm.NewStyle(pterm.FgGray)).
		WithTextStyle(pterm.NewStyle(pterm.FgWhite))

	box.Println(string(data))
	pterm.Println()
}
