package apply

import (
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/loader"
	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/ui/cli"
	"github.com/spf13/cobra"
	"math/rand/v2"
	"time"
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
			cli.Errorf("Failed to parse --file: %s", err.Error())
			return
		}

		cli.Infof("Loading manifests from: %s", file)

		loadedMans, cachedMans, err := loader.LoadManifests(file)
		if err != nil {
			cli.Errorf("Failed to load manifests: %s", err.Error())
			return
		}

		printManifestsLoadResult(loadedMans, cachedMans)

		if err := store.Save(loadedMans...); err != nil {
			cli.Infof("Failed to save manifests: %s", err.Error())
			return
		}

		cli.Println("Manifests applied successfully")
		cli.Debug("Manifests applied successfully")
		cli.Info("Manifests applied successfully")
		cli.Warning("Manifests applied successfully")
		cli.Error("Manifests applied successfully")
		cli.Fatal("Manifests applied successfully")
		cli.Success("Manifests applied successfully")

		p := cli.Progress()
		p.Start(100, "Some work...")
		for i := 0; i < 100; i++ {
			p.Increment("")
			time.Sleep(time.Duration(rand.IntN(100)) * time.Millisecond)
		}
		p.Stop()

		for i := 100; i > 0; i-- {
			cli.Errorf("Error test just of\nRaason: %s\nDescription: %s",
				"Just test reason",
				"Just test description",
			)
			time.Sleep(time.Duration(rand.IntN(50)) * time.Millisecond)
		}

		p = cli.Progress()
		p.Start(100, "Finishing...")
		for i := 0; i < 100; i++ {
			p.Increment("")
			time.Sleep(time.Duration(rand.IntN(100)) * time.Millisecond)
		}
		p.Stop()

		cli.Success("Manifests applied successfully")
	},
}

func printManifestsLoadResult(newMans, cachedMans []manifests.Manifest) {
	cli.Infof("Loaded %d new manifests", len(newMans))

	for _, m := range newMans {
		cli.Infof("New manifest added: %s (h: %s...)", m.GetID(), cli.ShortHash(m.GetMeta().GetHash()))
	}

	for _, m := range cachedMans {
		cli.Infof("Manifest %s unchanged (h: %s...) - using cached version", m.GetID(), cli.ShortHash(m.GetMeta().GetHash()))
	}
}
