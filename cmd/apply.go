package cmd

import (
	"fmt"
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

		manifests, err := yaml.LoadManifestsFromDir(file)
		if err != nil {
			return err
		}

		for i, manifest := range manifests {
			fmt.Printf("%d\nKind: %s\nName: %s\nNamespace: %s\n",
				i+1, manifest.GetKind(), manifest.GetName(), manifest.GetNamespace(),
			)
		}

		if err := yaml.SaveManifests(file, manifests...); err != nil {
			return err
		}

		return nil
	},
}
