package cmd

import (
	"slices"
	"time"

	"github.com/apiqube/cli/internal/manifests/depends"
	"github.com/apiqube/cli/internal/ui"
	"github.com/apiqube/cli/internal/yaml"
	"github.com/spf13/cobra"
)

func init() {
	applyCmd.Flags().StringP("file", "f", ".", "Path to manifest file")
	rootCmd.AddCommand(applyCmd)
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply resources from manifest file",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Init()
		defer func() {
			time.Sleep(time.Millisecond * 100)
			ui.Stop()
		}()

		file, err := cmd.Flags().GetString("file")
		if err != nil {
			ui.Errorf("Failed to parse --file: %s", err.Error())
			return err
		}

		ui.Printf("Applying manifests from: %s", file)
		ui.Spinner(true, "Loading manifests")

		mans, err := yaml.LoadManifestsFromDir(file)
		if err != nil {
			ui.Errorf("Failed to load manifests: %s", err.Error())
			return err
		}

		ui.Spinner(false)
		ui.Printf("Loaded %d manifests", len(mans))

		slices.Reverse(mans)
		for i, man := range mans {
			ui.Printf("#%d ID: %s", i+1, man.GetID())
		}

		ui.Spinner(true, "Generating execution plan")

		if err = depends.GeneratePlan("./examples/simple", mans); err != nil {
			ui.Errorf("Failed to generate plan: %s", err.Error())
			return err
		}

		ui.Spinner(false)
		ui.Print("Execution plan generated successfully")
		ui.Spinner(true, "Saving manifests...")

		if err := yaml.SaveManifests(file, mans...); err != nil {
			ui.Error("Failed to save manifests: " + err.Error())
			return err
		}

		ui.Spinner(false)
		ui.Println("Manifests applied successfully")

		return nil
	},
}
