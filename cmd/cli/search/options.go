package search

import (
	"fmt"
	"github.com/spf13/cobra"
	"time"
)

type Options struct {
	all            bool
	name           string
	nameWildcard   string
	nameRegex      string
	namespace      string
	kind           string
	version        int
	createdBy      string
	usedBy         string
	hashPrefix     string
	dependsOn      []string
	dependsOnAll   []string
	createdAfter   time.Time
	createdBefore  time.Time
	updatedAfter   time.Time
	updatedBefore  time.Time
	lastApplied    time.Time
	isRelativeTime bool

	// output options
	output       bool
	outputPath   string
	outputMode   string
	outputFormat string

	// Sorting
	sortBy []string

	// Internal state
	flagsSet map[string]bool
}

func parseSearchOptions(cmd *cobra.Command) (*Options, error) {
	opts := &Options{
		flagsSet: make(map[string]bool),
	}

	// Parse all flags
	if err := parseBasicFilters(cmd, opts); err != nil {
		return nil, err
	}

	if err := parseMetadataFilters(cmd, opts); err != nil {
		return nil, err
	}

	if err := parseAdvancedFilters(cmd, opts); err != nil {
		return nil, err
	}

	if err := parseTimeFilters(cmd, opts); err != nil {
		return nil, err
	}

	if err := parseOutputOptions(cmd, opts); err != nil {
		return nil, err
	}

	if err := parseSortOptions(cmd, opts); err != nil {
		return nil, err
	}

	return opts, nil
}

func validateSearchOptions(opts *Options) error {
	if !opts.all && len(opts.flagsSet) == 0 {
		return fmt.Errorf("at least one search filter must be specified")
	}

	if opts.flagsSet["name"] && (opts.flagsSet["name-wildcard"] || opts.flagsSet["name-regex"]) {
		return fmt.Errorf("cannot use exact name filter with wildcard/regex filters")
	}

	if opts.flagsSet["hash"] && len(opts.hashPrefix) < 5 {
		return fmt.Errorf("hash prefix must be at least 5 characters")
	}

	return nil
}

func parseBasicFilters(cmd *cobra.Command, opts *Options) error {
	if cmd.Flags().Changed("all") {
		opts.all, _ = cmd.Flags().GetBool("all")
		opts.flagsSet["all"] = true
	}

	if cmd.Flags().Changed("name") {
		opts.name, _ = cmd.Flags().GetString("name")
		opts.flagsSet["name"] = true
	}
	if cmd.Flags().Changed("name-wildcard") {
		opts.nameWildcard, _ = cmd.Flags().GetString("name-wildcard")
		opts.flagsSet["name-wildcard"] = true
	}
	if cmd.Flags().Changed("name-regex") {
		opts.nameRegex, _ = cmd.Flags().GetString("name-regex")
		opts.flagsSet["name-regex"] = true
	}

	return nil
}

func parseMetadataFilters(cmd *cobra.Command, opts *Options) error {
	if cmd.Flags().Changed("namespace") {
		opts.namespace, _ = cmd.Flags().GetString("namespace")
		opts.flagsSet["namespace"] = true
	}
	if cmd.Flags().Changed("kind") {
		opts.kind, _ = cmd.Flags().GetString("kind")
		opts.flagsSet["kind"] = true
	}
	if cmd.Flags().Changed("version") {
		opts.version, _ = cmd.Flags().GetInt("version")
		opts.flagsSet["version"] = true
	}
	if cmd.Flags().Changed("created-by") {
		opts.createdBy, _ = cmd.Flags().GetString("created-by")
		opts.flagsSet["created-by"] = true
	}
	if cmd.Flags().Changed("used-by") {
		opts.usedBy, _ = cmd.Flags().GetString("used-by")
		opts.flagsSet["used-by"] = true
	}
	return nil
}

func parseAdvancedFilters(cmd *cobra.Command, opts *Options) error {
	if cmd.Flags().Changed("hash") {
		opts.hashPrefix, _ = cmd.Flags().GetString("hash")
		opts.flagsSet["hash"] = true
	}
	if cmd.Flags().Changed("depends") {
		opts.dependsOn, _ = cmd.Flags().GetStringSlice("depends")
		opts.flagsSet["depends"] = true
	}
	if cmd.Flags().Changed("depends-all") {
		opts.dependsOnAll, _ = cmd.Flags().GetStringSlice("depends-all")
		opts.flagsSet["depends-all"] = true
	}
	return nil
}

func parseTimeFilters(cmd *cobra.Command, opts *Options) error {
	timeFields := map[string]*time.Time{
		"created-after":  &opts.createdAfter,
		"created-before": &opts.createdBefore,
		"updated-after":  &opts.updatedAfter,
		"updated-before": &opts.updatedBefore,
		"last-applied":   &opts.lastApplied,
	}

	for flag, target := range timeFields {
		if cmd.Flags().Changed(flag) {
			val, _ := cmd.Flags().GetString(flag)
			if t, err := parseTimeOrDuration(val); err == nil {
				*target = t
				opts.isRelativeTime = isDuration(val)
				opts.flagsSet[flag] = true
			} else {
				return fmt.Errorf("invalid %s value: %w", flag, err)
			}
		}
	}
	return nil
}

func parseOutputOptions(cmd *cobra.Command, opts *Options) error {
	if cmd.Flags().Changed("output") {
		opts.output, _ = cmd.Flags().GetBool("output")
		opts.flagsSet["output"] = true

		if opts.output {
			if cmd.Flags().Changed("output-path") {
				opts.outputPath, _ = cmd.Flags().GetString("output-path")
			}
			if opts.outputPath == "" {
				opts.outputPath = "."
			}

			if cmd.Flags().Changed("output-mode") {
				opts.outputMode, _ = cmd.Flags().GetString("output-mode")
				if opts.outputMode != "combined" && opts.outputMode != "separate" {
					return fmt.Errorf("invalid output mode: %s", opts.outputMode)
				}
			} else {
				opts.outputMode = "separate"
			}

			if cmd.Flags().Changed("output-format") {
				opts.outputFormat, _ = cmd.Flags().GetString("output-format")
				if opts.outputFormat != "yaml" && opts.outputFormat != "json" {
					return fmt.Errorf("invalid output format: %s", opts.outputFormat)
				}
			} else {
				opts.outputFormat = "yaml"
			}
		}
	}
	return nil
}

func parseSortOptions(cmd *cobra.Command, opts *Options) error {
	if cmd.Flags().Changed("sort") {
		opts.sortBy, _ = cmd.Flags().GetStringSlice("sort")
		opts.flagsSet["sort"] = true
	}
	return nil
}

func isDuration(val string) bool {
	_, err := time.ParseDuration(val)
	return err == nil
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
