package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

var versionCmd = &cobra.Command{
	Use:           "version",
	Short:         "Print the version number",
	SilenceUsage:  true,
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		data := fmt.Sprintf("Qube CLI\nVersion: %s", version)

		if commit != "" {
			data += fmt.Sprintf("\nCommit: %s", commit)
		}

		if date != "" {
			data += fmt.Sprintf("\nDate: %s", date)
		}

		fmt.Println(data)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
