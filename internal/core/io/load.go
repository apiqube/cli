package io

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/apiqube/cli/internal/core/manifests/utils"
	"github.com/apiqube/cli/internal/operations"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/store"
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

	parsedManifests, err := operations.ParseBatchAsYAML(content)
	if err != nil {
		return nil, nil, fmt.Errorf("in file %s: %w", filepath.Base(filePath), err)
	}

	now := time.Now()
	var newManifests, cachedManifests []manifests.Manifest

	for _, m := range parsedManifests {
		manifestID := m.GetID()

		var normalized []byte
		normalized, err = operations.NormalizeYAML(m)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to normalize manifest %s: %w", manifestID, err)
		}
		var manifestHash string
		manifestHash, err = utils.CalculateContentHash(normalized)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to calculate hash for manifest %s: %w", manifestID, err)
		}

		var loadedManifests []manifests.Manifest
		var existingManifest manifests.Manifest

		loadedManifests, err = store.Load(store.LoadOptions{Hash: manifestHash})
		if err != nil && !isNotFoundError(err) {
			return nil, nil, fmt.Errorf("failed to check manifest existence: %w", err)
		}

		if len(loadedManifests) > 0 {
			existingManifest = loadedManifests[0]
		}

		if existingManifest != nil {
			if _, exists := manifestsSet[existingManifest.GetID()]; !exists {
				manifestsSet[existingManifest.GetID()] = struct{}{}
			}
			continue
		}

		if _, exists := manifestsSet[manifestID]; exists {
			continue
		}

		meta := m.GetMeta()
		meta.SetHash(manifestHash)
		meta.SetVersion(1)
		meta.SetCreatedAt(now)
		meta.SetUpdatedAt(now)

		manifestsSet[manifestID] = struct{}{}
		newManifests = append(newManifests, m)
	}

	return newManifests, cachedManifests, nil
}

func isNotFoundError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "no matching manifest found") ||
		strings.Contains(err.Error(), "not found"))
}
