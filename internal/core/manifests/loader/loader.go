package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/apiqube/cli/internal/core/manifests/hash"
	"github.com/apiqube/cli/internal/core/manifests/parsing"
	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/ui"

	"github.com/apiqube/cli/internal/core/manifests"
)

func LoadManifestsFromDir(dir string) ([]manifests.Manifest, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var (
		manifestsList []manifests.Manifest
		manifestsSet  = make(map[string]struct{})
		newCounter    int
	)

	for _, file := range files {
		if file.IsDir() || (!strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml")) {
			continue
		}

		var content []byte
		filePath := filepath.Join(dir, file.Name())
		content, err = os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
		}

		var fileHash string
		fileHash, err = hash.CalculateHashWithPath(filePath, content)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate hash for %s: %w", filePath, err)
		}

		var existingManifest manifests.Manifest
		existingManifest, err = store.FindManifestByHash(fileHash)
		if err != nil && !strings.Contains(err.Error(), "no matching manifest found") {
			return nil, fmt.Errorf("failed to check manifest existence: %w", err)
		}

		if existingManifest != nil {
			if _, exists := manifestsSet[existingManifest.GetID()]; !exists {
				manifestsSet[existingManifest.GetID()] = struct{}{}
				manifestsList = append(manifestsList, existingManifest)
			}

			ui.Infof("Manifest file %s unchanged (%s) - using cached version", file.Name(), shortHash(fileHash))
			continue
		}

		var parsedManifests []manifests.Manifest
		parsedManifests, err = parsing.ParseManifestsAsYAML(content)
		if err != nil {
			return nil, fmt.Errorf("in file %s: %w", file.Name(), err)
		}

		for _, m := range parsedManifests {
			manifestID := m.GetID()

			if _, exists := manifestsSet[manifestID]; exists {
				ui.Warningf("Duplicate manifest ID: %s (from %s)", manifestID, file.Name())
				continue
			}

			if metaTable, ok := m.(manifests.MetaTable); ok {
				meta := metaTable.GetMeta()
				meta.SetHash(fileHash)
				meta.SetVersion(1)
				now := time.Now()
				meta.SetCreatedAt(now)
				meta.SetUpdatedAt(now)
			}

			manifestsSet[manifestID] = struct{}{}
			manifestsList = append(manifestsList, m)
			newCounter++

			ui.Successf("New manifest added: %s (h: %s)", manifestID, shortHash(fileHash))
		}
	}

	if newCounter == 0 {
		ui.Info("No new manifests found in directory")
	} else {
		ui.Infof("Loaded %d new manifests", newCounter)
	}

	return manifestsList, nil
}

func shortHash(fullHash string) string {
	if len(fullHash) > 12 {
		return fullHash[:12] + "..."
	}
	return fullHash
}
