package apply

import (
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/loader"
	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/ui"
	"github.com/spf13/cobra"
)

func init() {
	Cmd.Flags().StringP("file", "f", ".", "Path to manifests file, by default is current")
}

var Cmd = &cobra.Command{
	Use:           "apply",
	Short:         "Apply resources from manifest file",
	SilenceErrors: true,
	SilenceUsage:  true,
	Run: func(cmd *cobra.Command, args []string) {
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			ui.Errorf("Failed to parse --file: %s", err.Error())
			return
		}

		ui.Printf("Loading manifests from: %s", file)
		ui.Spinner(true, "Applying manifests...")
		defer ui.Spinner(false)

		loadedMans, cachedMans, err := loader.LoadManifests(file)
		if err != nil {
			ui.Errorf("Failed to load manifests: %s", err.Error())
			return
		}

		printManifestsLoadResult(loadedMans, cachedMans)

		if err := store.Save(loadedMans...); err != nil {
			ui.Error("Failed to save manifests: " + err.Error())
			return
		}

		ui.Println("Manifests applied successfully")
	},
}

func printManifestsLoadResult(newMans, cachedMans []manifests.Manifest) {
	ui.Infof("Loaded %d new manifests", len(newMans))

	for _, m := range newMans {
		ui.Infof("New manifest added: %s (h: %s...)", m.GetID(), ui.ShortHash(m.GetMeta().GetHash()))
	}

	for _, m := range cachedMans {
		ui.Infof("Manifest %s unchanged (h: %s...) - using cached version", m.GetID(), ui.ShortHash(m.GetMeta().GetHash()))
	}
}
