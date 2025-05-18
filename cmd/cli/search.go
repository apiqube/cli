package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for manifests using filters",
	Long: fmt.Sprint("Search for manifests with powerful filtering options including exact/wildcard matching," +
		"\ntime ranges, and output formatting"),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := parseSearchFlags(cmd, args)
		if err != nil {
			ui.Errorf("Failed to parse provided values: %v", err)
			return err
		}

		var manifests []manifests.Manifest

		if !opts.All &&
			!opts.flagsSet["name"] &&
			!opts.flagsSet["name-wildcard"] &&
			!opts.flagsSet["name-regex"] &&
			!opts.flagsSet["kind"] &&
			!opts.flagsSet["hash"] &&
			!opts.flagsSet["version"] &&
			!opts.flagsSet["namespace"] &&
			!opts.flagsSet["created-by"] &&
			!opts.flagsSet["used-by"] &&
			!opts.flagsSet["depends"] &&
			!opts.flagsSet["depends-all"] &&
			!opts.flagsSet["created-after"] &&
			!opts.flagsSet["created-before"] &&
			!opts.flagsSet["updated-after"] &&
			!opts.flagsSet["updated-before"] &&
			!opts.flagsSet["last-applied"] {
			return fmt.Errorf("at least one search filter must be specified")
		}

		if opts.flagsSet["all"] {
			manifests, err = store.Load(store.LoadOptions{All: true})
			if err != nil {
				ui.Errorf("Failed to loadmanifests: %v", err)
				return nil
			}
		} else {
			query := store.NewQuery()

			if opts.flagsSet["name"] {
				query.WithExactName(opts.Name)
			} else if opts.flagsSet["name-wildcard"] {
				query.WithWildcardName(opts.NameWildcard)
			} else if opts.flagsSet["name-regex"] {
				query.WithRegexName(opts.NameRegex)
			}

			if opts.flagsSet["namespace"] {
				query.WithNamespace(opts.Namespace)
			}

			if opts.flagsSet["kind"] {
				query.WithKind(opts.Kind)
			}

			if opts.flagsSet["version"] {
				query.WithVersion(opts.Version)
			}

			if opts.flagsSet["created-by"] {
				query.WithCreatedBy(opts.CreatedBy)
			}

			if opts.flagsSet["user-by"] {
				query.WithUsedBy(opts.UsedBy)
			}

			if opts.flagsSet["hash"] {
				query.WithHashPrefix(opts.HashPrefix)
			}

			if opts.flagsSet["depends"] {
				query.WithDependencies(opts.DependsOn)
			} else if opts.flagsSet["depends-all"] {
				query.WithAllDependencies(opts.DependsOnAll)
			}

			if opts.flagsSet["created-after"] {
				query.WithCreatedAfter(opts.CreatedAfter)
			}

			if opts.flagsSet["created-before"] {
				query.WithCreatedBefore(opts.CreatedBefore)
			}

			if opts.flagsSet["updated-after"] {
				query.WithUpdatedAfter(opts.UpdatedAfter)
			}

			if opts.flagsSet["updated-before"] {
				query.WithUpdatedBefore(opts.UpdatedBefore)
			}

			if opts.flagsSet["last-applied"] {
				query.WithLastApplied(opts.LastApplied)
			}

			manifests, err = store.Search(query)
			if err != nil {
				ui.Errorf("Failed to search manifests: %v", err)
				return nil
			}
		}

		if len(manifests) == 0 {
			ui.Warning("No manifests found matching the criteria")
			return nil
		}

		ui.Infof("Found %d manifests", len(manifests))

		if len(opts.SortBy) > 0 {
			sortManifests(manifests, opts.SortBy)
		}

		ui.Spinner(true, "Prepare answer...")

		if opts.Output {
			if err := outputManifests(manifests, opts); err != nil {
				ui.Spinner(false)
				ui.Errorf("Failed to output manifests: %v", err)
				return nil
			}
		} else {
			displayResults(manifests)
		}

		ui.Spinner(false, "Complete")

		return nil
	},
}

func init() {
	searchCmd.Flags().BoolP("all", "a", false, "Get all manifests")

	searchCmd.Flags().StringP("name", "n", "", "Search manifest by name (exact match)")
	searchCmd.Flags().StringP("name-wildcard", "W", "", "Search manifest by wildcard pattern (e.g. '*name*')")
	searchCmd.Flags().StringP("name-regex", "R", "", "Search manifest by regex pattern")

	searchCmd.Flags().StringP("namespace", "s", "", "Search manifests by namespace")
	searchCmd.Flags().StringP("kind", "k", "", "Search manifests by kind")
	searchCmd.Flags().IntP("version", "v", 0, "Search manifests by version")
	searchCmd.Flags().String("created-by", "", "Filter by exact creator username")
	searchCmd.Flags().String("used-by", "", "Filter by exact user who applied")

	searchCmd.Flags().StringP("hash", "H", "", "Search manifests by hash prefix (min 5 chars)")
	searchCmd.Flags().StringSliceP("depends", "d", []string{}, "Search manifests by dependencies (comma separated)")
	searchCmd.Flags().StringSliceP("depends-all", "D", []string{}, "Search manifests by all dependencies (comma separated)")

	searchCmd.Flags().String("created-after", "", "Search manifests created after date (YYYY-MM-DD or duration like 1h30m)")
	searchCmd.Flags().String("created-before", "", "Search manifests created before date/duration")
	searchCmd.Flags().String("updated-after", "", "Search manifests updated after date/duration")
	searchCmd.Flags().String("updated-before", "", "Search manifests updated before date/duration")
	searchCmd.Flags().String("last-applied", "", "Search manifests by last applied date/duration")

	searchCmd.Flags().BoolP("output", "o", false, "Make output after searching")
	searchCmd.Flags().String("output-path", "", "Output path for results (default: current directory)")
	searchCmd.Flags().String("output-mode", "separate", "Output mode (combined|separate)")
	searchCmd.Flags().String("output-format", "yaml", "File format for output (yaml|json)")

	searchCmd.Flags().StringSlice("sort", []string{}, "Sort by fields (e.g. --sort=kind,-name)")

	rootCmd.AddCommand(searchCmd)
}

type SearchOptions struct {
	All bool

	Name         string
	NameWildcard string
	NameRegex    string

	Namespace string
	Kind      string
	Version   int
	CreatedBy string
	UsedBy    string

	HashPrefix   string
	DependsOn    []string
	DependsOnAll []string

	CreatedAfter   time.Time
	CreatedBefore  time.Time
	UpdatedAfter   time.Time
	UpdatedBefore  time.Time
	LastApplied    time.Time
	IsRelativeTime bool

	Output       bool
	OutputPath   string
	OutputMode   string // combined | separate
	OutputFormat string // yaml | json

	SortBy []string

	flagsSet map[string]bool
}

func parseSearchFlags(cmd *cobra.Command, _ []string) (*SearchOptions, error) {
	opts := &SearchOptions{
		flagsSet: make(map[string]bool),
	}

	markFlag := func(name string) bool {
		if cmd.Flags().Changed(name) {
			opts.flagsSet[name] = true
			return true
		}
		return false
	}

	if markFlag("all") {
		opts.All, _ = cmd.Flags().GetBool("all")
	}

	if markFlag("name") {
		opts.Name, _ = cmd.Flags().GetString("name")
	}
	if markFlag("name-wildcard") {
		opts.NameWildcard, _ = cmd.Flags().GetString("name-wildcard")
	}
	if markFlag("name-regex") {
		opts.NameRegex, _ = cmd.Flags().GetString("name-regex")
	}

	if opts.flagsSet["name"] && (opts.flagsSet["name-wildcard"] || opts.flagsSet["name-regex"]) {
		return nil, fmt.Errorf("cannot use exact name filter with wildcard/regex filters")
	}

	if markFlag("namespace") {
		opts.Namespace, _ = cmd.Flags().GetString("namespace")
	}
	if markFlag("kind") {
		opts.Kind, _ = cmd.Flags().GetString("kind")
	}
	if markFlag("version") {
		opts.Version, _ = cmd.Flags().GetInt("version")
	}
	if markFlag("created-by") {
		opts.CreatedBy, _ = cmd.Flags().GetString("created-by")
	}
	if markFlag("used-by") {
		opts.UsedBy, _ = cmd.Flags().GetString("used-by")
	}

	if markFlag("hash") {
		opts.HashPrefix, _ = cmd.Flags().GetString("hash")
		if len(opts.HashPrefix) < 5 {
			return nil, fmt.Errorf("hash prefix must be at least 5 characters")
		}
	}
	if markFlag("depends") {
		opts.DependsOn, _ = cmd.Flags().GetStringSlice("depends")
	} else if markFlag("depends-all") {
		opts.DependsOnAll, _ = cmd.Flags().GetStringSlice("depends-all")
	}

	timeFilters := map[string]*time.Time{
		"created-after":  &opts.CreatedAfter,
		"created-before": &opts.CreatedBefore,
		"updated-after":  &opts.UpdatedAfter,
		"updated-before": &opts.UpdatedBefore,
		"last-applied":   &opts.LastApplied,
	}

	for flag, target := range timeFilters {
		if markFlag(flag) {
			val, _ := cmd.Flags().GetString(flag)
			if t, err := parseTimeOrDuration(val); err == nil {
				*target = t
				opts.IsRelativeTime = isDuration(val)
			} else {
				return nil, fmt.Errorf("invalid %s value: %w", flag, err)
			}
		}
	}

	if markFlag("output") {
		opts.Output, _ = cmd.Flags().GetBool("output")
		if opts.Output {
			if markFlag("output-path") {
				opts.OutputPath, _ = cmd.Flags().GetString("output-path")
			}
			if opts.OutputPath == "" {
				opts.OutputPath = "."
			}
			if markFlag("output-mode") {
				opts.OutputMode, _ = cmd.Flags().GetString("output-mode")
				if opts.OutputMode != "combined" && opts.OutputMode != "separate" {
					return nil, fmt.Errorf("invalid output mode, must be 'combined' or 'separate'")
				}
			}
			if opts.OutputMode == "" {
				opts.OutputMode = "separate"
			}
			if markFlag("output-format") {
				opts.OutputFormat, _ = cmd.Flags().GetString("output-format")
				if opts.OutputFormat != "yaml" && opts.OutputFormat != "json" {
					return nil, fmt.Errorf("invalid output format, must be 'yaml' or 'json'")
				}
			}
			if opts.OutputFormat == "" {
				opts.OutputFormat = "yaml"
			}
		}
	}

	if markFlag("sort") {
		opts.SortBy, _ = cmd.Flags().GetStringSlice("sort")
	}

	return opts, nil
}

func parseTimeOrDuration(val string) (time.Time, error) {
	if duration, err := time.ParseDuration(val); err == nil {
		return time.Now().Add(-duration), nil
	}

	if t, err := time.Parse("2006-01-02", val); err == nil {
		return t, nil
	}

	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid time format")
}

func isDuration(val string) bool {
	_, err := time.ParseDuration(val)
	return err == nil
}

func outputManifests(manifests []manifests.Manifest, opts *SearchOptions) error {
	if err := os.MkdirAll(opts.OutputPath, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if opts.OutputMode == "combined" {
		filename := filepath.Join(opts.OutputPath, fmt.Sprintf("manifests.%s", opts.OutputFormat))
		return writeCombinedFile(filename, manifests, opts.OutputFormat)
	} else {
		for _, m := range manifests {
			filename := filepath.Join(opts.OutputPath, fmt.Sprintf("%s.%s", m.GetID(), opts.OutputFormat))
			if err := writeSingleFile(filename, m, opts.OutputFormat); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeCombinedFile(filename string, manifests []manifests.Manifest, format string) error {
	if len(manifests) == 0 {
		return fmt.Errorf("no manifests to write")
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer func() {
		_ = file.Close()
	}()

	switch strings.ToLower(format) {
	case "yaml":
		encoder := yaml.NewEncoder(file)
		for i, m := range manifests {
			if i > 0 {
				if _, err = file.WriteString("---\n"); err != nil {
					return fmt.Errorf("failed to write YAML manifest: %w", err)
				}
			}
			if err = encoder.Encode(m); err != nil {
				return fmt.Errorf("failed to encode manifest %d: %w", i+1, err)
			}
		}
	case "json":
		if _, err = file.WriteString("[\n"); err != nil {
			return fmt.Errorf("failed to write JSON manifest: %w", err)
		}

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")

		for i, m := range manifests {
			if i > 0 {
				if _, err = file.WriteString(",\n"); err != nil {
					return fmt.Errorf("failed to write YAML manifest: %w", err)
				}
			}
			if err = encoder.Encode(m); err != nil {
				return fmt.Errorf("failed to encode manifest %d: %w", i+1, err)
			}
		}
		if _, err = file.WriteString("\n]"); err != nil {
			return fmt.Errorf("failed to write JSON manifest: %w", err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	ui.Successf("Successfully wrote %d manifests to %s", len(manifests), filename)
	return nil
}

func writeSingleFile(filename string, manifest manifests.Manifest, format string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer func() {
		_ = file.Close()
	}()

	switch strings.ToLower(format) {
	case "yaml":
		encoder := yaml.NewEncoder(file)
		if err = encoder.Encode(manifest); err != nil {
			return fmt.Errorf("failed to encode manifest: %w", err)
		}
	case "json":
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err = encoder.Encode(manifest); err != nil {
			return fmt.Errorf("failed to encode manifest: %w", err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	ui.Successf("Successfully wrote manifest %s to %s", manifest.GetID(), filename)
	return nil
}

func sortManifests(manifests []manifests.Manifest, fields []string) {
	sort.Slice(manifests, func(i, j int) bool {
		for _, field := range fields {
			desc := false
			if strings.HasPrefix(field, "-") {
				desc = true
				field = field[1:]
			}

			switch field {
			case "id":
				if manifests[i].GetID() != manifests[j].GetID() {
					if desc {
						return manifests[i].GetID() > manifests[j].GetID()
					}
					return manifests[i].GetID() < manifests[j].GetID()
				}
			case "name":
				if manifests[i].GetName() != manifests[j].GetName() {
					if desc {
						return manifests[i].GetName() > manifests[j].GetName()
					}
					return manifests[i].GetName() < manifests[j].GetName()
				}
			case "kind":
				if manifests[i].GetKind() != manifests[j].GetKind() {
					if desc {
						return manifests[i].GetKind() > manifests[j].GetKind()
					}
					return manifests[i].GetKind() < manifests[j].GetKind()
				}
			case "namespace":
				if manifests[i].GetNamespace() != manifests[j].GetNamespace() {
					if desc {
						return manifests[i].GetNamespace() > manifests[j].GetNamespace()
					}
					return manifests[i].GetNamespace() < manifests[j].GetNamespace()
				}
			case "version":
				if desc {
					return manifests[i].GetMeta().GetVersion() > manifests[j].GetMeta().GetVersion()
				}
				return manifests[i].GetMeta().GetVersion() < manifests[j].GetMeta().GetVersion()
			case "created":
				if desc {
					return manifests[i].GetMeta().GetCreatedAt().After(manifests[j].GetMeta().GetCreatedAt())
				}
				return manifests[i].GetMeta().GetCreatedAt().Before(manifests[j].GetMeta().GetCreatedAt())
			case "updated":
				if desc {
					return manifests[i].GetMeta().GetUpdatedAt().After(manifests[j].GetMeta().GetUpdatedAt())
				}
				return manifests[i].GetMeta().GetUpdatedAt().Before(manifests[j].GetMeta().GetUpdatedAt())
			case "last":
				if desc {
					return manifests[i].GetMeta().GetLastApplied().After(manifests[j].GetMeta().GetLastApplied())
				}
				return manifests[i].GetMeta().GetLastApplied().Before(manifests[j].GetMeta().GetLastApplied())
			}
		}
		return false
	})
}

func displayResults(manifests []manifests.Manifest) {
	headers := []string{
		"#",
		"Hash",
		"Kind",
		"Name",
		"Namespace",
		"Version",
		"Created",
		"Updated",
		"Last Updated",
	}

	var rows [][]string
	for i, m := range manifests {
		meta := m.GetMeta()
		row := []string{
			fmt.Sprint(i + 1),
			shortHash(meta.GetHash()),
			m.GetKind(),
			m.GetName(),
			m.GetNamespace(),
			fmt.Sprint(meta.GetVersion()),
			meta.GetCreatedAt().Format(time.RFC3339),
			meta.GetUpdatedAt().Format(time.RFC3339),
			meta.GetLastApplied().Format(time.RFC3339),
		}
		rows = append(rows, row)
	}

	ui.Table(headers, rows)
}

func shortHash(fullHash string) string {
	if len(fullHash) > 8 {
		return fullHash[:8]
	}
	return fullHash
}
