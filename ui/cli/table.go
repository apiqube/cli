package cli

import (
	"github.com/pterm/pterm"
)

func (u *UI) Table(headers []string, rows [][]string) {
	tableData := pterm.TableData{headers}
	for _, row := range rows {
		tableData = append(tableData, row)
	}

	pterm.Println()

	table := pterm.DefaultTable.
		WithHasHeader(true).
		WithBoxed(false).
		WithHeaderStyle(pterm.NewStyle(pterm.FgCyan, pterm.Bold, pterm.BgDefault)).
		WithSeparator(" | ").
		WithSeparatorStyle(pterm.NewStyle(pterm.FgDarkGray)).
		WithData(tableData).
		WithLeftAlignment(true).
		WithRowSeparator("").
		WithLeftAlignment(true)

	if err := table.Render(); err != nil {
		pterm.Error.WithShowLineNumber(false).Printfln("Rendering of the table failed: %v", err)
	}
}
