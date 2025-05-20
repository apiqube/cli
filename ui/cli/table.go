package cli

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func (u *UI) Table(headers []string, rows [][]string) {
	columns := make([]table.Column, len(headers))
	for i, h := range headers {
		columns[i] = table.Column{Title: h, Width: 20}
	}

	tblRows := make([]table.Row, len(rows))
	for i, r := range rows {
		tblRows[i] = r
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(tblRows),
		table.WithFocused(true),
	)

	u.model.tableComp = t
	u.model.currentView = ViewTable
	u.program.Send(tea.ClearScreen())
}
