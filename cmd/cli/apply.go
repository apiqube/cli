package cli

import (
	"github.com/apiqube/cli/internal/core/manifests/depends"
	"github.com/apiqube/cli/internal/core/manifests/loader"
	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/ui"
	"github.com/spf13/cobra"
)

func init() {
	applyCmd.Flags().StringP("file", "f", ".", "Path to manifest file")
	rootCmd.AddCommand(applyCmd)
}

var applyCmd = &cobra.Command{
	Use:           "apply",
	Short:         "Apply resources from manifest file",
	SilenceErrors: true,
	SilenceUsage:  true,
	Run: func(cmd *cobra.Command, args []string) {
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			ui.Errorf("Failed to parse --file: %s", err.Error())
			return
		}

		ui.Printf("Applying manifests from: %s", file)
		ui.Spinner(true, "Loading manifests")

		mans, err := loader.LoadManifestsFromDir(file)
		if err != nil {
			ui.Spinner(false)
			ui.Errorf("Failed to load manifests: %s", err.Error())
			return
		}

		ui.Spinner(false)

		var result *depends.GraphResult
		if result, err = depends.BuildGraphWithPriority(mans); err != nil {
			ui.Errorf("Failed to generate plan: %s", err.Error())
			return
		}

		for i, order := range result.ExecutionOrder {
			ui.Printf("#Order %d %s", i+1, order)
		}

		ui.Spinner(false)
		ui.Print("Execution plan generated successfully")
		ui.Spinner(true, "Saving manifests...")

		if err := store.SaveManifests(mans...); err != nil {
			ui.Error("Failed to save manifests: " + err.Error())
			return
		}

		ui.Spinner(false)
		ui.Println("Manifests applied successfully")
	},
}
