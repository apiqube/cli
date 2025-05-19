package cli

import "github.com/spf13/cobra"

var generateCmd = &cobra.Command{
	Use:           "generate",
	Short:         "Generate manifests with provided flags",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
