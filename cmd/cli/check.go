package cli

import (
	"fmt"
	"strings"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/plan"
	"github.com/apiqube/cli/internal/core/manifests/loader"
	runner "github.com/apiqube/cli/internal/core/runner/plan"
	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/ui"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check validity of manifests, plans or full configurations",
}

var checkManifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "Validate individual manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var checkPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Validate a plan manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := parseCheckPlanFlags(cmd, args)
		if err != nil {
			ui.Errorf("Failed to parse provided values: %v", err)
			return err
		}

		if !opts.flagsSet["id"] &&
			!opts.flagsSet["name"] &&
			!opts.flagsSet["namespace"] &&
			!opts.flagsSet["file"] {
			ui.Errorf("At least one check plan filter must be specified")
			return nil
		}

		ui.Spinner(true, "Checking manifests...")
		defer ui.Spinner(false)

		var loadedMans []manifests.Manifest
		var man manifests.Manifest
		query := store.NewQuery()
		withQuery := false

		if opts.flagsSet["id"] {
			if loadedMans, err = store.Load(store.LoadOptions{
				IDs: []string{opts.id},
			}); err != nil {
				ui.Errorf("Failed to load manifest: %v", err)
				return nil
			}
		} else if opts.flagsSet["name"] {
			query.WithExactName(opts.name)
			withQuery = true
		} else if opts.flagsSet["namespace"] {
			query.WithNamespace(opts.namespace)
			withQuery = true
		} else if opts.flagsSet["file"] {
			if loadedMans, _, err = loader.LoadManifests(opts.file); err != nil {
				ui.Errorf("Failed to load manifest: %v", err)
				return nil
			}

			ui.Infof("Manifests from provieded path %s loaded", opts.file)
		}

		if withQuery {
			loadedMans, err = store.Search(query)
			if err != nil {
				ui.Errorf("Failed to search plan manifests: %v", err)
				return nil
			}
		}

		if man, err = findManifestWithKind(manifests.PlanManifestKind, loadedMans); err != nil {
			ui.Errorf("Failed to check plan manifest: %v", err)
			return nil
		}

		if planToCheck, ok := man.(*plan.Plan); ok {
			manifestIds := planToCheck.GetAllManifests()

			if loadedMans, err = store.Load(store.LoadOptions{
				IDs: manifestIds,
			}); err != nil {
				ui.Errorf("Failed to load plan manifests: %v", err)
			}

			builder := runner.NewPlanManagerBuilder().WithManifests(loadedMans...)
			generator := builder.Build()

			if generator.CheckPlan(planToCheck) != nil {
				ui.Errorf("Failed to check plan: %s", err)
				return nil
			}
		} else {
			ui.Errorf("Failed to check plan, manifest found but not plan manifest")
			return nil
		}

		ui.Successf("Successfully checked plan manifest")
		return nil
	},
}

var checkAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Validate full manifest set (plan + dependencies + tests)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	checkManifestCmd.Flags().String("id", "", "Full manifest ID to check (namespace.kind.name)")
	checkManifestCmd.Flags().String("kind", "", "kind of manifest (e.g., HttpTest, Server, Values)")
	checkManifestCmd.Flags().String("name", "", "name of manifest")
	checkManifestCmd.Flags().String("namespace", "", "namespace of manifest")
	checkManifestCmd.Flags().String("file", "", "Path to manifest file to check")

	checkPlanCmd.Flags().String("id", "", "Full plan ID to check (namespace.Plan.name)")
	checkPlanCmd.Flags().StringP("name", "n", "", "name of plan")
	checkPlanCmd.Flags().StringP("namespace", "s", "", "namespace of manifest")
	checkPlanCmd.Flags().StringP("file", "f", "", "Path to plan.yaml")

	checkAllCmd.Flags().String("path", ".", "Path to directory with manifests to check")

	checkCmd.AddCommand(checkManifestCmd)
	checkCmd.AddCommand(checkPlanCmd)
	checkCmd.AddCommand(checkAllCmd)

	rootCmd.AddCommand(checkCmd)
}

type (
	checkPlanOptions struct {
		id        string
		name      string
		namespace string
		file      string

		flagsSet map[string]bool
	}
)

func parseCheckPlanFlags(cmd *cobra.Command, _ []string) (*checkPlanOptions, error) {
	opts := &checkPlanOptions{
		flagsSet: make(map[string]bool),
	}

	markFlag := func(name string) bool {
		if cmd.Flags().Changed(name) {
			opts.flagsSet[name] = true
			return true
		}
		return false
	}

	if markFlag("id") {
		opts.id, _ = cmd.Flags().GetString("id")
	}
	if markFlag("name") {
		opts.name, _ = cmd.Flags().GetString("name")
	}
	if markFlag("namespace") {
		opts.namespace, _ = cmd.Flags().GetString("namespace")
	}

	if markFlag("file") {
		var file string
		file, _ = cmd.Flags().GetString("file")
		if strings.HasSuffix(file, ".yml") || strings.HasSuffix(file, ".yaml") {
			opts.file = file
		} else {
			return nil, fmt.Errorf("--file flag must end with .yml or .yaml")
		}
	}

	if opts.flagsSet["id"] || (opts.flagsSet["name"] || (opts.flagsSet["namespace"] && opts.flagsSet["file"])) {
		return nil, fmt.Errorf("cannot use all filters at the same time")
	}

	return opts, nil
}

func findManifestWithKind(kind string, mans []manifests.Manifest) (manifests.Manifest, error) {
	for i, man := range mans {
		if man.GetKind() == kind {
			return mans[i], nil
		}
	}

	return nil, fmt.Errorf("expected manifest with %s kind not found", kind)
}
