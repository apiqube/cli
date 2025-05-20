package cli

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type ViewType uint8

const (
	ViewNone ViewType = iota
	ViewTable
	ViewProgress
)

const (
	logCuntsAmount = 250
)

var _ tea.Model = (*uiModel)(nil)

type uiModel struct {
	tableComp    table.Model
	progressComp progress.Model

	currentView ViewType

	progressData progressData

	logs []string
}

func (m *uiModel) Init() tea.Cmd {
	return nil
}

func (m *uiModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.QuitMsg:
		return m, tea.Quit
	case progressMsg:
		if m.currentView == ViewProgress {
			return m, nil
		}
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
	switch m.currentView {
	case ViewTable:
		return m.tableComp.View()
	case ViewProgress:
		progressPct := float64(m.progressData.current) / float64(m.progressData.total)
		view := "\n" + m.progressData.title + "\n" +
			m.progressComp.ViewAs(progressPct) + "\n\n"
		return view
	default:
		return ""
	}
}

func (m *uiModel) AddLog(log string) {
	m.logs = append(m.logs, log)

	if len(m.logs) > logCuntsAmount {
		if len(m.logs) > 100 {
			m.logs = m.logs[1:]
		}
	}
}
