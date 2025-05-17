package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/apiqube/cli/internal/core/manifests/hash"
	"github.com/apiqube/cli/internal/core/manifests/parsing"
	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/ui"

	"github.com/apiqube/cli/internal/core/manifests"
)

func LoadManifests(path string) (new []manifests.Manifest, cached []manifests.Manifest, err error) {
	if isYAMLFile(path) {
		return processSingleFile(path)
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read directory: %w", err)
	}

	manifestsSet := make(map[string]struct{})
	newCount, cachedCount := 0, 0

	for _, file := range files {
		if file.IsDir() || !isYAMLFile(file.Name()) {
			continue
		}

		filePath := filepath.Join(path, file.Name())

		var fileNew, fileCached []manifests.Manifest
		fileNew, fileCached, err = processFile(filePath, manifestsSet)
		if err != nil {
			return nil, nil, err
		}

		new = append(new, fileNew...)
		cached = append(cached, fileCached...)
		newCount += len(fileNew)
		cachedCount += len(fileCached)
	}

	logResults(newCount, cachedCount)
	return new, cached, nil
}

func isYAMLFile(path string) bool {
	return strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")
}

func processSingleFile(filePath string) (new []manifests.Manifest, cached []manifests.Manifest, err error) {
	return processFile(filePath, make(map[string]struct{}))
}

func processFile(filePath string, manifestsSet map[string]struct{}) (new []manifests.Manifest, cached []manifests.Manifest, err error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	parsedManifests, err := parsing.ParseManifestsAsYAML(content)
	if err != nil {
		return nil, nil, fmt.Errorf("in file %s: %w", filepath.Base(filePath), err)
	}

	now := time.Now()
	var newManifests, cachedManifests []manifests.Manifest

	for _, m := range parsedManifests {
		manifestID := m.GetID()

		var normalized []byte
		normalized, err = normalizeYAML(m)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to normalize manifest %s: %w", manifestID, err)
		}
		var manifestHash string
		manifestHash, err = hash.CalculateHashWithContent(normalized)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to calculate hash for manifest %s: %w", manifestID, err)
		}

		var existingManifest manifests.Manifest
		existingManifest, err = store.FindManifestByHash(manifestHash)
		if err != nil && !isNotFoundError(err) {
			return nil, nil, fmt.Errorf("failed to check manifest existence: %w", err)
		}

		if existingManifest != nil {
			if _, exists := manifestsSet[existingManifest.GetID()]; !exists {
				manifestsSet[existingManifest.GetID()] = struct{}{}
				ui.Infof("Manifest %s unchanged (%s) - using cached version",
					existingManifest.GetID(), shortHash(manifestHash))
				cachedManifests = append(cachedManifests, existingManifest)
			}
			continue
		}

		if _, exists := manifestsSet[manifestID]; exists {
			ui.Warningf("Duplicate manifest ID: %s (from %s)", manifestID, filepath.Base(filePath))
			continue
		}

		meta := m.GetMeta()
		meta.SetHash(manifestHash)
		meta.SetVersion(1)
		meta.SetCreatedAt(now)
		meta.SetUpdatedAt(now)

		manifestsSet[manifestID] = struct{}{}
		newManifests = append(newManifests, m)
		ui.Successf("New manifest added: %s (h: %s)", manifestID, shortHash(manifestHash))
	}

	return newManifests, cachedManifests, nil
}

func logResults(cachedCount, newCount int) {
	if cachedCount > 0 {
		ui.Infof("Loaded %d cached manifests", cachedCount)
	}
	if newCount > 0 {
		ui.Infof("Loaded %d new manifests", newCount)
	}
	if cachedCount == 0 && newCount == 0 {
		ui.Info("No manifests found")
	}
}

func shortHash(fullHash string) string {
	if len(fullHash) > 12 {
		return fullHash[:12] + "..."
	}
	return fullHash
}

func normalizeYAML(m manifests.Manifest) ([]byte, error) {
	data, err := yaml.Marshal(m)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err = yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	sorted := sortMapKeys(raw)

	return yaml.Marshal(sorted)
}

func sortMapKeys(m map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if nested, ok := m[k].(map[string]interface{}); ok {
			res[k] = sortMapKeys(nested)
		} else {
			res[k] = m[k]
		}
	}

	return res
}

func isNotFoundError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "no matching manifest found") ||
		strings.Contains(err.Error(), "not found"))
}
