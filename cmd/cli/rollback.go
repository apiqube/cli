package cli

import (
	"fmt"

	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/ui"
	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback [ID]",
	Short: "Rollback to previous manifest version",
	Long: fmt.Sprint("Rollback to specific version of manifest." +
		"\nIf version is not specified, rolls back to previous one."),
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		opts, err := parseRollbackFlags(cmd, args)
		if err != nil {
			return
		}

		targetVersion := opts.Version
		if targetVersion <= 0 {
			targetVersion = 1
		}

		ui.Spinner(true, "Rolling back")
		if err = store.Rollback(opts.ManifestID, targetVersion); err != nil {
			ui.Spinner(false, "Failed to rollback")
			ui.Errorf("Error rolling back to previous version: %s", err)
			return
		}

		ui.Spinner(false)
		ui.Successf("Successfully rolled back %s to version %d\n", opts.ManifestID, targetVersion)
	},
}

func init() {
	rollbackCmd.Flags().IntP("version", "v", 0, "Target version number (defaults to previous version)")

	rootCmd.AddCommand(rollbackCmd)
}

type RollbackOptions struct {
	ManifestID string
	Version    int
}

func parseRollbackFlags(cmd *cobra.Command, args []string) (*RollbackOptions, error) {
	opts := &RollbackOptions{}

	if len(args) == 0 {
		return nil, fmt.Errorf("manifest ID is required")
	}

	opts.ManifestID = args[0]
	var err error
	var ver int

	if cmd.Flags().Changed("version") {
		ver, err = cmd.Flags().GetInt("version")
		if err != nil {
			return nil, fmt.Errorf("invalid version: %w", err)
		}
		if ver < 1 {
			return nil, fmt.Errorf("version must be positive")
		}
		opts.Version = ver
	}

	return opts, nil
}
