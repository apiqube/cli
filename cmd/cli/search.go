package cli

import "github.com/spf13/cobra"

var searchCmd = &cobra.Command{
	Use:           "search",
	Short:         "Search for manifests using filters",
	SilenceUsage:  true,
	SilenceErrors: true,
	Run:           func(cmd *cobra.Command, args []string) {},
}

func init() {
	// All
	searchCmd.Flags().StringP("all", "a", "", "Get all manifests")

	// Exact match filters
	searchCmd.Flags().StringP("name", "n", "", "Search manifest by exact name match")
	searchCmd.Flags().StringP("namespace", "s", "default", "Search manifests by exact namespace)")
	searchCmd.Flags().StringP("kind", "k", "", "Search manifests by exact kind (e.g., Server, HttpTest, HttpLoadTest)")
	searchCmd.Flags().IntP("version", "v", 1, "Search manifests by exact version number")

	// Wildcard/partial match filters
	searchCmd.Flags().StringP("wildcard", "w", "", "Search manifests using name wildcard pattern")
	searchCmd.Flags().StringP("hash", "h", "", "Search manifests by hash prefix")

	// Dependency filters
	searchCmd.Flags().StringSliceP("depends", "d", []string{}, "Search manifests by dependencies (comma separated)")

	// Date filters
	searchCmd.Flags().String("created-after", "", "Search manifests created after date (YYYY-MM-DD)")
	searchCmd.Flags().String("created-before", "", "Search manifests created before date")
	searchCmd.Flags().String("updated-after", "", "Search manifests updated after date")
	searchCmd.Flags().String("last-applied", "", "Search manifests by last applied date")

	// Creator filter
	searchCmd.Flags().String("created-by", "", "Filter by creator username")

	rootCmd.AddCommand(searchCmd)
}
