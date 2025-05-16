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
		return nil, err
	}

	existingHashes, err := store.LoadManifestHashes()
	if err != nil {
		return nil, fmt.Errorf("failed to load hashes: %w", err)
	}

	hashCache := make(map[string]bool)
	for _, h := range existingHashes {
		hashCache[h] = true
	}

	var (
		mans                      []manifests.Manifest
		existingIDs               []string
		manifestsSet              = make(map[string]struct{})
		processedHashes           = make(map[string]bool)
		exists                    bool
		newCounter, existsCounter int
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
			return nil, fmt.Errorf("failed to calculate h for %s: %w", filePath, err)
		}

		if hashCache[fileHash] {
			ui.Infof("Manifest file %s unchanged (%s) - using cache", file.Name(), shortHash(fileHash))
			exists = true
			existsCounter++
		} else {
			newCounter++
		}

		var parsedManifests []manifests.Manifest

		parsedManifests, err = parsing.ParseYamlManifests(content)
		if err != nil {
			return nil, fmt.Errorf("in file %s: %w", file.Name(), err)
		}

		for _, m := range parsedManifests {
			if !exists {
				manifestID := m.GetID()

				if _, ok := manifestsSet[manifestID]; ok {
					ui.Warningf("Manifest: %s (from %s) already processed", manifestID, file.Name())
					continue
				}

				if meta, ok := m.(manifests.MetaTable); ok {
					meta.GetMeta().SetHash(fileHash)
					now := time.Now()
					meta.GetMeta().SetCreatedAt(now)
					meta.GetMeta().SetUpdatedAt(now)
				}

				manifestsSet[manifestID] = struct{}{}
				mans = append(mans, m)
				processedHashes[fileHash] = true

				ui.Successf("New manifest added: %s (h: %s)", manifestID, shortHash(fileHash))
			} else {
				existingIDs = append(existingIDs, m.GetID())
			}
		}
	}

	for h := range processedHashes {
		if err = store.SaveManifestHash(h); err != nil {
			ui.Errorf("Failed to save manifest h: %s", err.Error())
		}
	}

	var existingManifests []manifests.Manifest
	if len(existingHashes) > 0 {
		existingManifests, err = store.LoadManifests(existingIDs...)
		if err != nil {
			ui.Warningf("Failed to load existing manifests: %v", err)
		} else {
			mans = append(existingManifests, mans...)
		}
	}

	if newCounter == 0 {
		ui.Infof("Loaded %d manifests, new manifests not found", len(mans))
	} else {
		ui.Infof("Loaded %d manifests (%d new, %d from cache)",
			len(mans),
			newCounter,
			existsCounter,
		)
	}

	return mans, nil
}

func shortHash(fullHash string) string {
	if len(fullHash) > 12 {
		return fullHash[:12] + "..."
	}
	return fullHash
}
