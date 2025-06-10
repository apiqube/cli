package generator

import (
	"fmt"
	"strings"

	"github.com/apiqube/cli/internal/core/io"
	"github.com/apiqube/cli/internal/core/manifests"
	runner "github.com/apiqube/cli/internal/core/runner/plan"
	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/internal/operations"
	"github.com/apiqube/cli/ui/cli"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var Cmd = &cobra.Command{
	Use:           "generate",
	Short:         "Generate execution plan from loaded manifests",
	SilenceErrors: true,
	SilenceUsage:  true,
	Run: func(cmd *cobra.Command, args []string) {
		opts, err := parseOptions(cmd)
		if err != nil {
			cli.Errorf("Failed to parse provided save: %v", err)
			return
		}

		cli.Info("Loading manifests...")
		loadedManifests, err := loadManifests(opts)
		if err != nil {
			cli.Errorf("Failed to load manifests: %v", err)
			return
		}

		cli.Infof("Loaded %d manifests", len(loadedManifests))
		cli.Info("Generating plan...")

		manager := runner.NewPlanManagerBuilder().
			WithManifests(loadedManifests...).Build()

		planManifest, err := manager.Generate()
		if err != nil {
			cli.Errorf("Failed to generate plan: %v", err)
			return
		}

		cli.Successf("Plan successfully generated")

		if err = savePlan(opts, planManifest); err != nil {
			cli.Errorf("Failed to save plan: %v", err)
			return
		}

		cli.Successf("Generated plan successfully saved")
	},
}

func init() {
	Cmd.Flags().StringArrayP("names", "n", []string{}, "Names of manifests to generate (comma separated)")
	Cmd.Flags().StringP("namespace", "s", "", "Namespace of manifests to generate")
	Cmd.Flags().StringArrayP("ids", "i", []string{}, "IDs of manifests to generate (comma separated)")
	Cmd.Flags().StringArrayP("hashes", "H", []string{}, "Hash prefixes for manifests (min 5 chars each)")
	Cmd.Flags().StringP("file", "f", ".", "Path to manifest directory (default: current)")
	Cmd.Flags().Bool("save", false, "Save generated plan to file")
	Cmd.Flags().Bool("print", false, "Print generated plan to stdout")
	Cmd.Flags().BoolP("output", "o", false, "Make output after generating")
	Cmd.Flags().String("output-path", "", "Output path to save the plan (default: current directory)")
	Cmd.Flags().String("output-format", "yaml", "Output format (yaml|json)")
}

type options struct {
	names     []string
	namespace string
	ids       []string
	hashes    []string

	file string

	save         bool
	print        bool
	output       bool
	outputPath   string
	outputFormat string

	flagsSet map[string]bool
}

func parseOptions(cmd *cobra.Command) (*options, error) {
	opts := &options{
		flagsSet: make(map[string]bool),
	}

	markFlag := func(name string) bool {
		if cmd.Flags().Changed(name) {
			opts.flagsSet[name] = true
			return true
		}
		return false
	}

	if markFlag("names") {
		opts.names, _ = cmd.Flags().GetStringArray("names")
	}
	if markFlag("namespace") {
		opts.namespace, _ = cmd.Flags().GetString("namespace")
	}
	if markFlag("ids") {
		opts.ids, _ = cmd.Flags().GetStringArray("ids")
	}
	if markFlag("hashes") {
		opts.hashes, _ = cmd.Flags().GetStringArray("hashes")
	}

	if markFlag("file") {
		opts.file, _ = cmd.Flags().GetString("file")
	}

	if markFlag("save") {
		opts.save, _ = cmd.Flags().GetBool("save")
	}
	if markFlag("print") {
		opts.print, _ = cmd.Flags().GetBool("print")
	}

	if markFlag("output") {
		opts.output, _ = cmd.Flags().GetBool("output")
	}
	if markFlag("output-path") {
		opts.outputPath, _ = cmd.Flags().GetString("output-path")
	}
	if markFlag("output-format") {
		opts.outputFormat, _ = cmd.Flags().GetString("output-format")
	}

	exclusiveFlags := []string{"names", "namespace", "ids", "hashes", "file"}

	var usedFlags []string
	for _, flag := range exclusiveFlags {
		if opts.flagsSet[flag] {
			usedFlags = append(usedFlags, "--"+flag)
		}
	}

	if len(usedFlags) > 1 {
		return nil, fmt.Errorf(
			"conflicting filters: %s\n"+
				"these filters cannot be used together, please use only one",
			strings.Join(usedFlags, " and "),
		)
	}

	if err := validateOptions(opts); err != nil {
		return nil, err
	}

	return opts, nil
}

func validateOptions(opts *options) error {
	if !opts.flagsSet["names"] &&
		!opts.flagsSet["namespace"] &&
		!opts.flagsSet["ids"] &&
		!opts.flagsSet["hashes"] &&
		!opts.flagsSet["file"] {
		return fmt.Errorf("at least one generate filter must be specified")
	}
	return nil
}

func loadManifests(opts *options) ([]manifests.Manifest, error) {
	switch {
	case opts.flagsSet["ids"]:
		return store.Load(store.LoadOptions{
			IDs: opts.ids,
		})

	case opts.flagsSet["file"]:
		loadedMans, cachedMans, err := io.LoadManifests(opts.file)
		if err == nil {
			cli.Infof("Manifests from provided path %s loaded", opts.file)
		}

		loadedMans = append(loadedMans, cachedMans...)
		return loadedMans, err

	default:
		query := store.NewQuery()
		if opts.flagsSet["names"] {
			for _, name := range opts.names {
				query.WithExactName(name)
			}
		}

		if opts.flagsSet["hashes"] {
			for _, hash := range opts.hashes {
				query.WithHashPrefix(hash)
			}
		}

		if opts.flagsSet["namespace"] {
			query.WithNamespace(opts.namespace)
		}

		return store.Search(query)
	}
}

func savePlan(opts *options, plan manifests.Manifest) error {
	if opts.save {
		if err := store.Save(plan); err != nil {
			return err
		}
	} else if opts.output {
		var parseFormat operations.ParseFormat
		switch opts.outputFormat {
		case "json":
			parseFormat = operations.JSONFormat
		default:
			parseFormat = operations.YAMLFormat
		}

		return io.WriteSeparate(opts.outputPath, parseFormat, plan)
	} else {
		if data, err := yaml.Marshal(plan); err == nil {
			cli.Snippet().View(plan.GetID(), data)
		} else {
			return fmt.Errorf("failed to marshal plan to yaml: %v", err)
		}
	}
	return nil
}
