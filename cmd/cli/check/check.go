package check

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

var Cmd = &cobra.Command{
	Use:   "check",
	Short: "Check validity of manifests, plans or full configurations",
}

var cmdManifestCheck = &cobra.Command{
	Use:   "manifest",
	Short: "Validate individual manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var cmdPlanCheck = &cobra.Command{
	Use:   "plan",
	Short: "Validate a plan manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := parseCheckPlanFlags(cmd, args)
		if err != nil {
			return uiErrorf("Failed to parse provided values: %v", err)
		}

		if err := validateCheckPlanOptions(opts); err != nil {
			return uiErrorf(err.Error())
		}

		ui.Spinner(true, "Checking manifests...")
		defer ui.Spinner(false)

		loadedManifests, err := loadManifests(opts)
		if err != nil {
			return uiErrorf("Failed to load manifests: %v", err)
		}

		planManifest, err := extractPlanManifest(loadedManifests)
		if err != nil {
			return uiErrorf("Failed to check plan manifest: %v", err)
		}

		if err := validatePlan(planManifest); err != nil {
			return uiErrorf("Failed to check plan: %v", err)
		}

		ui.Successf("Successfully checked plan manifest")
		return nil
	},
}

var cmdAllCheck = &cobra.Command{
	Use:   "all",
	Short: "Validate full manifest set (plan + dependencies + tests)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	cmdManifestCheck.Flags().String("id", "", "Full manifest ID to check (namespace.kind.name)")
	cmdManifestCheck.Flags().String("kind", "", "kind of manifest (e.g., HttpTest, Server, Values)")
	cmdManifestCheck.Flags().String("name", "", "name of manifest")
	cmdManifestCheck.Flags().String("namespace", "", "namespace of manifest")
	cmdManifestCheck.Flags().String("file", "", "Path to manifest file to check")

	cmdPlanCheck.Flags().String("id", "", "Full plan ID to check (namespace.Plan.name)")
	cmdPlanCheck.Flags().StringP("name", "n", "", "name of plan")
	cmdPlanCheck.Flags().StringP("namespace", "s", "", "namespace of manifest")
	cmdPlanCheck.Flags().StringP("file", "f", "", "Path to plan.yaml")

	cmdAllCheck.Flags().String("path", ".", "Path to directory with manifests to check")

	Cmd.AddCommand(cmdManifestCheck)
	Cmd.AddCommand(cmdPlanCheck)
	Cmd.AddCommand(cmdAllCheck)
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

func uiErrorf(format string, args ...interface{}) error {
	ui.Errorf(format, args...)
	return nil
}

func validateCheckPlanOptions(opts *checkPlanOptions) error {
	if !opts.flagsSet["id"] &&
		!opts.flagsSet["name"] &&
		!opts.flagsSet["namespace"] &&
		!opts.flagsSet["file"] {
		return fmt.Errorf("at least one check plan filter must be specified")
	}
	return nil
}

func loadManifests(opts *checkPlanOptions) ([]manifests.Manifest, error) {
	switch {
	case opts.flagsSet["id"]:
		return store.Load(store.LoadOptions{
			IDs: []string{opts.id},
		})

	case opts.flagsSet["file"]:
		loadedMans, _, err := loader.LoadManifests(opts.file)
		if err == nil {
			ui.Infof("Manifests from provided path %s loaded", opts.file)
		}
		return loadedMans, err

	default:
		query := store.NewQuery()
		if opts.flagsSet["name"] {
			query.WithExactName(opts.name)
		}
		if opts.flagsSet["namespace"] {
			query.WithNamespace(opts.namespace)
		}
		return store.Search(query)
	}
}

func extractPlanManifest(mans []manifests.Manifest) (*plan.Plan, error) {
	man, err := findManifestWithKind(manifests.PlanManifestKind, mans)
	if err != nil {
		return nil, err
	}

	planManifest, ok := man.(*plan.Plan)
	if !ok {
		return nil, fmt.Errorf("manifest found but not a plan manifest")
	}
	return planManifest, nil
}

func validatePlan(planToCheck *plan.Plan) error {
	manifestIds := planToCheck.GetAllManifests()
	loadedMans, err := store.Load(store.LoadOptions{
		IDs: manifestIds,
	})
	if err != nil {
		return err
	}

	builder := runner.NewPlanManagerBuilder().WithManifests(loadedMans...)
	generator := builder.Build()
	return generator.CheckPlan(planToCheck)
}

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
