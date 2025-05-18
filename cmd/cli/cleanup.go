package cli

import (
	"fmt"

	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/ui"
	"github.com/spf13/cobra"
)

const keepVersionDefault = 5

var cleanupCmd = &cobra.Command{
	Use:   "cleanup [ID]",
	Short: "Cleanup old manifest versions by its id",
	Long: fmt.Sprintf("Delete all versions of the manifest,"+
		"\nleaving only the latest specified."+
		"\nBy default, the last keep amount is %d", keepVersionDefault),
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := parseCleanUpFlags(cmd, args)
		if err != nil {
			return err
		}

		keep := opts.Keep
		if keep <= 0 {
			keep = keepVersionDefault
		}

		ui.Spinner(true, "Cleaning up...")
		if err = store.CleanupOldVersions(opts.ManifestID, keep); err != nil {
			ui.Spinner(false)
			ui.Errorf("Failed to cleanup old versions: %v", err)
		}

		ui.Spinner(false)
		ui.Successf("Successfully cleaned up %v to last %d versions", opts.ManifestID, keep)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
	cleanupCmd.Flags().IntP("keep", "k", keepVersionDefault, "Number of last versions to keep")
}

type CleanUpOptions struct {
	ManifestID string
	Keep       int
}

func parseCleanUpFlags(cmd *cobra.Command, args []string) (*CleanUpOptions, error) {
	opts := &CleanUpOptions{}

	if len(args) == 0 {
		return nil, fmt.Errorf("manifest ID is required")
	}

	opts.ManifestID = args[0]
	var err error
	var keep int

	if cmd.Flags().Changed("keep") {
		keep, err = cmd.Flags().GetInt("keep")
		if err != nil {
			return nil, fmt.Errorf("invalid keep value: %w", err)
		}
		if keep < 1 {
			return nil, fmt.Errorf("keep value must be positive")
		}
		opts.Keep = keep
	}

	return opts, nil
}
