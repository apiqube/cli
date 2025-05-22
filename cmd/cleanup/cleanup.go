package cleanup

import (
	"fmt"

	"github.com/apiqube/cli/ui/cli"

	"github.com/apiqube/cli/internal/core/store"
	"github.com/spf13/cobra"
)

const keepVersionDefault = 5

var Cmd = &cobra.Command{
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

		keep := opts.keep
		if keep <= 0 {
			keep = keepVersionDefault
		}

		if err = store.CleanupOldVersions(opts.manifestID, keep); err != nil {
			cli.Errorf("Failed to cleanup old versions: %v", err)
		}

		cli.Successf("Successfully cleaned up %v to last %d versions", opts.manifestID, keep)
		return nil
	},
}

func init() {
	Cmd.Flags().IntP("keep", "k", keepVersionDefault, "Number of last versions to keep")
}

type Options struct {
	manifestID string
	keep       int
}

func parseCleanUpFlags(cmd *cobra.Command, args []string) (*Options, error) {
	opts := &Options{}

	if len(args) == 0 {
		return nil, fmt.Errorf("manifest ID is required")
	}

	opts.manifestID = args[0]
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
		opts.keep = keep
	}

	return opts, nil
}
