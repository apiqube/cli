package cmd

import (
	"fmt"
	"github.com/apiqube/cli/internal/manifests/depends"
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
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}
		if file == "" {
			return fmt.Errorf("no manifest file provided (use -f or --file)")
		}

		fmt.Println("Applying manifest from:", file)

		mans, err := yaml.LoadManifestsFromDir(file)
		if err != nil {
			return err
		}

		if err = depends.GeneratePlan("./examples", mans); err != nil {
			return err
		}

		graph, _, err := depends.BuildDependencyGraph(mans)
		if err != nil {
			return err
		}

		order, err := depends.TopoSort(graph)
		if err != nil {
			return err
		}

		sortedMans, err := depends.SortManifestsByExecutionOrder(mans, order)
		if err != nil {
			return err
		}

		for _, m := range sortedMans {
			fmt.Printf("ManifestID: %s\n", m.GetID())
		}

		if err = depends.SaveExecutionPlan("./examples", order); err != nil {
			return err
		}

		if err := yaml.SaveManifests(file, mans...); err != nil {
			return err
		}

		return nil
	},
}
