package cmd

import (
	"github.com/apiqube/cli/internal/manifests/depends"
	"github.com/apiqube/cli/internal/ui"
	"github.com/apiqube/cli/internal/yaml"
	"github.com/spf13/cobra"
	"time"
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
		ui.Init()

		file, err := cmd.Flags().GetString("file")
		if err != nil {
			ui.Errorf("Failed to parse --file: %s", err.Error())
			return
		}

		ui.Printf("Applying manifests from: %s", file)
		ui.Spinner(true, "Loading manifests")

		mans, err := yaml.LoadManifestsFromDir(file)
		if err != nil {
			ui.Spinner(false)
			ui.Errorf("Failed to load manifests: %s", err.Error())
			return
		}

		ui.Printf("Loaded %d manifests", len(mans))

		if err = depends.GeneratePlan(mans); err != nil {
			ui.Errorf("Failed to generate plan: %s", err.Error())
			return
		}

		ui.Spinner(false)
		ui.Print("Execution plan generated successfully")
		ui.Spinner(true, "Saving manifests...")

		if err := yaml.SaveManifestsAsCombined(mans...); err != nil {
			ui.Error("Failed to save manifests: " + err.Error())
			return
		}

		ui.Spinner(false)
		ui.Println("Manifests applied successfully")
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		time.Sleep(time.Millisecond * 500)
		ui.Stop()
	},
}
