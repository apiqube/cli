package cli

import (
	"context"
	"github.com/apiqube/cli/ui/cli"
	"os"
	"os/signal"
	"syscall"

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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg, err := config.InitConfig()
		if err != nil {
			cli.Errorf("Error initializing config: %s", err.Error())
		}

		cmd.SetContext(configureContext(context.WithValue(cmd.Context(), configKey, cfg)))
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

func configureContext(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()
	return ctx
}
