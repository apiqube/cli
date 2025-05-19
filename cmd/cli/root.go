package cli

import (
	"context"
	"fmt"
	"github.com/apiqube/cli/cmd/cli/apply"
	"github.com/apiqube/cli/cmd/cli/check"
	"github.com/apiqube/cli/cmd/cli/cleanup"
	"github.com/apiqube/cli/cmd/cli/generator"
	"github.com/apiqube/cli/cmd/cli/rollback"
	"github.com/apiqube/cli/cmd/cli/search"

	"github.com/apiqube/cli/internal/config"
	"github.com/spf13/cobra"
)

type contextKey string

var configKey contextKey = "config"

var rootCmd = &cobra.Command{
	Use:   "qube",
	Short: "ApiQube is a powerful test manager for APIs",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.InitConfig()
		if err != nil {
			return fmt.Errorf("config init failed: %w", err)
		}

		cmd.SetContext(context.WithValue(cmd.Context(), configKey, cfg))
		return nil
	},
}

func Execute() {
	rootCmd.AddCommand(
		apply.Cmd,
		check.Cmd,
		cleanup.Cmd,
		generator.Cmd,
		rollback.Cmd,
		search.Cmd,
	)

	cobra.CheckErr(rootCmd.Execute())
}
