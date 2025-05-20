package cli

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type ViewType uint8

const (
	ViewNone ViewType = iota
	ViewTable
	ViewProgress
)

var _ tea.Model = (*uiModel)(nil)

type uiModel struct {
	tableComp    table.Model
	progressComp progress.Model

	currentView ViewType
}

func (m *uiModel) Init() tea.Cmd {
	return nil
}

func (m *uiModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.QuitMsg:
		return m, tea.Quit
	}

	switch m.currentView {
	case ViewTable:
		var cmd tea.Cmd
		m.tableComp, cmd = m.tableComp.Update(message)
		return m, cmd
	default:
	}

	return m, nil
}

func (m *uiModel) View() string {
	var view strings.Builder

	switch m.currentView {
	case ViewTable:
		view.WriteString(m.tableComp.View())
	default:
		return ""
	}

	return view.String()
}
