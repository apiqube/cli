package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/apiqube/cli/cmd/qube/apply"
	"github.com/apiqube/cli/cmd/qube/check"
	"github.com/apiqube/cli/cmd/qube/cleanup"
	"github.com/apiqube/cli/cmd/qube/edit"
	"github.com/apiqube/cli/cmd/qube/generator"
	"github.com/apiqube/cli/cmd/qube/rollback"
	"github.com/apiqube/cli/cmd/qube/search"

	"github.com/apiqube/cli/ui/cli"

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
		edit.Cmd,
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
