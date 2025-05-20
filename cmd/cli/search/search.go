package search

import (
	"fmt"

	"github.com/apiqube/cli/ui/cli"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "search",
	Short: "Search for manifests using filters",
	Long: `Search for manifests with powerful filtering options including exact/wildcard matching,
time ranges, and output formatting`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := parseSearchOptions(cmd)
		if err != nil {
			return fmt.Errorf("failed to parse options: %w", err)
		}

		if err := validateSearchOptions(opts); err != nil {
			return err
		}

		manifests, err := executeSearch(opts)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(manifests) == 0 {
			cli.Info("No manifests found matching the criteria")
			return nil
		}

		return handleSearchResults(manifests, opts)
	},
}

func init() {
	// Basic filters
	Cmd.Flags().BoolP("all", "a", false, "Get all manifests")
	Cmd.Flags().StringP("name", "n", "", "Search manifest by name (exact match)")
	Cmd.Flags().StringP("name-wildcard", "W", "", "Search manifest by wildcard pattern (e.g. '*name*')")
	Cmd.Flags().StringP("name-regex", "R", "", "Search manifest by regex pattern")

	// Metadata filters
	Cmd.Flags().StringP("namespace", "s", "", "Search manifests by namespace")
	Cmd.Flags().StringP("kind", "k", "", "Search manifests by kind")
	Cmd.Flags().IntP("version", "v", 0, "Search manifests by version")
	Cmd.Flags().String("created-by", "", "Filter by exact creator username")
	Cmd.Flags().String("used-by", "", "Filter by exact user who applied")

	// Advanced filters
	Cmd.Flags().StringP("hash", "H", "", "Search manifests by hash prefix (min 5 chars)")
	Cmd.Flags().StringSliceP("depends", "d", []string{}, "Search manifests by dependencies (comma separated)")
	Cmd.Flags().StringSliceP("depends-all", "D", []string{}, "Search manifests by all dependencies (comma separated)")

	// Time filters
	Cmd.Flags().String("created-after", "", "Search manifests created after date (YYYY-MM-DD or duration like 1h30m)")
	Cmd.Flags().String("created-before", "", "Search manifests created before date/duration")
	Cmd.Flags().String("updated-after", "", "Search manifests updated after date/duration")
	Cmd.Flags().String("updated-before", "", "Search manifests updated before date/duration")
	Cmd.Flags().String("last-applied", "", "Search manifests by last applied date/duration")

	// output options
	Cmd.Flags().BoolP("output", "o", false, "Make output after searching")
	Cmd.Flags().String("output-path", "", "output path for results (default: current directory)")
	Cmd.Flags().String("output-mode", "separate", "output mode (combined|separate)")
	Cmd.Flags().String("output-format", "yaml", "File format for output (yaml|json)")

	// Sorting
	Cmd.Flags().StringSlice("sort", []string{}, "Sort by fields (e.g. --sort=kind,-name)")
}
