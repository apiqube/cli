package edit

import (
	"errors"
	"fmt"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/utils"
	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/internal/operations"
	uicli "github.com/apiqube/cli/ui/cli"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:           "edit",
	Short:         "Edit already saved manifests",
	SilenceErrors: true,
	SilenceUsage:  true,
	Run: func(cmd *cobra.Command, args []string) {
		opts, err := parseOptions(cmd)
		if err != nil {
			uicli.Error(err.Error())
			return
		}

		var mans []manifests.Manifest
		var man, result manifests.Manifest

		uicli.Info("Looking for manifest...")

		query := store.NewQuery()
		queryFlag := false

		if opts.flagsSet["id"] {
			mans, err = store.Load(store.LoadOptions{IDs: []string{opts.manifestID}})
		} else if opts.flagsSet["name"] {
			query.WithExactName(opts.name)
			queryFlag = true
		} else if opts.flagsSet["hash"] {
			query.WithHashPrefix(opts.hashPrefix)
			queryFlag = true
		}

		if queryFlag {
			mans, err = store.Search(store.NewQuery())
		}

		if err != nil {
			uicli.Errorf("Failed to load manifest: %s", err.Error())
			return
		} else if len(mans) == 0 {
			uicli.Info("No manifests found matching the criteria")
			return
		}

		man = mans[0]

		uicli.Successf("Manifest %s was founded", man.GetID())
		uicli.Infof("Loading %s manifest in editing context", man.GetID())

		if result, err = operations.Edit(man); err != nil {
			if errors.Is(err, operations.ErrFileNotEdited) {
				uicli.Infof("Manifest file %s was not edited", man.GetID())
				return
			}

			uicli.Errorf("Failed to edit manifest: %s", err.Error())
			return
		}

		uicli.Info("Preparing manifest for saving")

		if content, err := operations.NormalizeYAML(result); err != nil {
			uicli.Errorf("Failed to normalize manifest: %s", err.Error())
			return
		} else {
			if hash, err := utils.CalculateContentHash(content); err != nil {
				uicli.Errorf("Failed to calculate hash: %s", err.Error())
				return
			} else {
				result.GetMeta().SetHash(hash)
			}
		}

		uicli.Infof("Saving %s manifest in storage", man.GetID())

		if err = store.Save(man); err != nil {
			uicli.Errorf("Failed to save manifest: %s", err.Error())
			return
		}

		uicli.Successf("Manifest %s successfully saved", man.GetID())
	},
}

func init() {
	Cmd.Flags().StringP("id", "i", "", "Search and edit manifest by ID")
	Cmd.Flags().StringP("name", "n", "", "Search and edit manifest by name")
	Cmd.Flags().StringP("hash", "H", "", "Search and edit manifest by hash")
}

type options struct {
	manifestID string
	name       string
	hashPrefix string

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

	if markFlag("id") {
		opts.manifestID, _ = cmd.Flags().GetString("id")
	}
	if markFlag("name") {
		opts.name, _ = cmd.Flags().GetString("name")
	}
	if markFlag("hash") {
		opts.hashPrefix, _ = cmd.Flags().GetString("hash")
	}

	if opts.flagsSet["id"] && (opts.flagsSet["name"] || opts.flagsSet["hash"]) {
		return nil, fmt.Errorf("id/name and hash flags cannot be used together")
	}

	return opts, nil
}
