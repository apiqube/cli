package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "qube",
	Short: "ApiQube is a powerful test manager for apps and APIs",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
