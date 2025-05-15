package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"sync"
	"time"
)

type uiModel struct {
	elements     []Element
	content      []message
	mu           sync.Mutex
	spinnerIndex int
}

type message struct {
	text      string
	style     lipgloss.Style
	timestamp time.Time
}

type updateFunc func(*uiModel)

type forceRefresh struct{}

func (m *uiModel) Init() tea.Cmd {
	return nil
}

func (m *uiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case forceRefresh:
		return m, nil
	}
	return m, nil
}

func (m *uiModel) View() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var sb strings.Builder

	m.spinnerIndex = (m.spinnerIndex + 1) % len(spinners)

	for _, msg := range m.content {
		sb.WriteString(timestampStyle.Render(msg.timestamp.Format("15:04:05")) + " " + msg.style.Render(msg.text))
		sb.WriteString("\n")
	}

	for _, elem := range m.elements {
		switch elem.elementType {
		case TypeProgress:
			sb.WriteString(renderProgressBar(elem.progress, elem.progressText))
			sb.WriteString("\n\n")

		case TypeLoader:
			if elem.showLoader {
				sb.WriteString(renderLoader(elem.loaderText))
				sb.WriteString("\n\n")
			}

		case TypePackage:
			sb.WriteString(renderPackage(elem.action, elem.packageName, elem.status))
			sb.WriteString("\n")

		case TypeRealtime:
			sb.WriteString(renderRealtime(elem.content))
			sb.WriteString("\n")

		case TypeSpinner:
			if elem.showSpinner {
				sb.WriteString(renderSpinner(m.spinnerIndex, elem.spinnerText))
				sb.WriteString("\n")
			}

		case TypeStopwatch:
			sb.WriteString(renderStopwatch(elem.startTime, elem.content))
			sb.WriteString("\n")

		case TypeTable:
			sb.WriteString(renderTable(elem.tableHeaders, elem.tableData))
			sb.WriteString("\n")

		case TypeSnippet:
			sb.WriteString(renderSnippet(elem.content))
			sb.WriteString("\n\n")
		default:
		}
	}

	return sb.String()
}
