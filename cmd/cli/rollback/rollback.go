package rollback

import (
	"fmt"

	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/ui"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
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

		targetVersion := opts.version
		if targetVersion <= 0 {
			targetVersion = 1
		}

		ui.Spinner(true, "Rolling back")
		defer ui.Spinner(false)

		if err = store.Rollback(opts.manifestID, targetVersion); err != nil {
			ui.Errorf("Error rolling back to previous version: %s", err)
			return
		}

		ui.Successf("Successfully rolled back %s to version %d\n", opts.manifestID, targetVersion)
	},
}

func init() {
	Cmd.Flags().IntP("version", "v", 0, "Target version number (defaults to previous version)")
}

type Options struct {
	manifestID string
	version    int
}

func parseRollbackFlags(cmd *cobra.Command, args []string) (*Options, error) {
	opts := &Options{}

	if len(args) == 0 {
		return nil, fmt.Errorf("manifest ID is required")
	}

	opts.manifestID = args[0]
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
		opts.version = ver
	}

	return opts, nil
}
