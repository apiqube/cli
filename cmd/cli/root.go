package cli

import (
	"context"
	"fmt"
	"github.com/apiqube/cli/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "qube",
	Short: "ApiQube is a powerful test manager for APIs",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.InitConfig()
		if err != nil {
			return fmt.Errorf("config init failed: %w", err)
		}

		cmd.SetContext(context.WithValue(cmd.Context(), "config", cfg))
		return nil
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
