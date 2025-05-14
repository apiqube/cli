package cmd

import (
	"fmt"
	"github.com/apiqube/cli/internal/config"
	"github.com/apiqube/cli/internal/yaml"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run [test file]",
	Short: "Run test scenarios with provided configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		testFile := args[0]

		tea.Printf("Configuration file path: %s\n", testFile)()

		cfg, err := yaml.LoadConfig[config.Config](testFile)
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		p := tea.NewProgram(testModel{config: cfg})

		if _, err := p.Run(); err != nil {
			fmt.Println("Error starting program:", err)
		}
	},
}

type testModel struct {
	config *config.Config
	cursor int
}

func (m testModel) Init() tea.Cmd {
	return nil
}

func (m testModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch message := msg.(type) {
	case tea.KeyMsg:
		switch message.String() {
		case "ctrl+c":
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m testModel) View() string {
	var out string
	out += fmt.Sprintf("Test Manager - Version %s\n", m.config.Version)
	out += "Press 'q' to quit.\n\n"

	for _, test := range m.config.Tests {
		out += fmt.Sprintf("Test: %s\n", test.Name)
		out += fmt.Sprintf("Description: %s\n", test.Description)
		if len(test.Flags) > 0 {
			out += "Flags: "
			for _, flag := range test.Flags {
				out += flag + " "
			}
			out += "\n"
		}
		out += "\n"
	}

	return out
}
