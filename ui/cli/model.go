package cli

import (
	"fmt"
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

	progressData progressData

	logHistory []string

	width  int
	height int
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
	case logMsg:
		return m, nil
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
	var view strings.Builder

	view.WriteString(m.renderLogArea())
	view.WriteString(m.renderContentArea())

	return view.String()
}

func (m *uiModel) AddLog(log string) {
	m.logHistory = append(m.logHistory, log)
}

func (m *uiModel) renderLogArea() string {
	var area strings.Builder

	for _, log := range m.logHistory {
		area.WriteString(log)
	}

	return area.String()
}

func (m *uiModel) renderContentArea() string {
	switch m.currentView {
	case ViewTable:
		return m.tableComp.View()
	case ViewProgress:
		progressPct := float64(m.progressData.current) / float64(m.progressData.total)
		return fmt.Sprintf("\n%s\n%s\n",
			m.progressData.title,
			m.progressComp.ViewAs(progressPct))
	default:
		return ""
	}
}
