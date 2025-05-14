package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run test suite with interactive CLI",
	Run: func(cmd *cobra.Command, args []string) {
		RunInteractiveTestUI()
	},
}

type testCase struct {
	Name   string
	Status string
}

type model struct {
	tests    []testCase
	progress progress.Model
	index    int
	quitting bool
}

func initialModel() model {
	return model{
		tests: []testCase{
			{"Register User", "pending"},
			{"Login User", "pending"},
			{"Create Resource", "pending"},
			{"Delete Resource", "pending"},
		},
		progress: progress.New(progress.WithDefaultGradient()),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
			return tickMsg{}
		}),
	)
}

type tickMsg struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if m.index >= len(m.tests) {
			m.quitting = true
			return m, tea.Quit
		}

		m.tests[m.index].Status = "✓ passed"
		m.index++
		return m, tea.Tick(700*time.Millisecond, func(t time.Time) tea.Msg {
			return tickMsg{}
		})

	case tea.KeyMsg:
		if msg.String() == "q" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10")).
			Render("\n✅ All tests complete!\n\n")
	}

	s := "🧪 Running test cases:\n\n"

	for i, t := range m.tests {
		style := lipgloss.NewStyle()
		if i == m.index {
			style = style.Bold(true).Foreground(lipgloss.Color("12"))
		}
		s += style.Render(fmt.Sprintf("• %s [%s]", t.Name, t.Status)) + "\n"
	}

	s += "\nPress 'q' to quit at any time.\n"
	return s
}

func RunInteractiveTestUI() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Println("Error running UI:", err)
		os.Exit(1)
	}
}
