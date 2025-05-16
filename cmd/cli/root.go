package cli

import (
	"fmt"
	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/internal/ui"
	"github.com/spf13/cobra"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "qube",
	Short: "ApiQube is a powerful test manager for apps and APIs",
	PreRun: func(cmd *cobra.Command, args []string) {
		fmt.Println("START !!!")
		ui.Init()
		store.Init()
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		store.Stop()
		ui.StopWithTimeout(time.Millisecond * 250)
		fmt.Println("FINISH !!!")
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
