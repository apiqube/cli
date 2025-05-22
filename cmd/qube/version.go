package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)

var versionCmd = &cobra.Command{
	Use:           "version",
	Short:         "Print the Version number",
	SilenceUsage:  true,
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		data := fmt.Sprintf("Qube CLI\nVersion: %s", Version)

		if Commit != "" {
			data += fmt.Sprintf("\nCommit: %s", Commit)
		}

		if Date != "" {
			data += fmt.Sprintf("\nDate: %s", Date)
		}

		fmt.Println(data)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
