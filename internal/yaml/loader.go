package yaml

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apiqube/cli/internal/ui"

	"github.com/apiqube/cli/internal/manifest"
)

func LoadManifestsFromDir(dir string) ([]manifest.Manifest, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	manifestsSet := make(map[string]struct{})
	var parsedManifests []manifest.Manifest
	var result []manifest.Manifest
	var counter int

	var content []byte

	for _, file := range files {
		if file.IsDir() || (!strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml")) {
			continue
		}

		content, err = os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}

		parsedManifests, err = ParseManifests(content)
		if err != nil {
			return nil, fmt.Errorf("in file %s: %w", file.Name(), err)
		}

		for _, m := range parsedManifests {
			if _, ok := manifestsSet[m.GetID()]; ok {
				ui.Infof("Manifest: %s loaded", m.GetID())
			} else {
				ui.Infof("Manifest: %s cached", m.GetID())
				manifestsSet[m.GetID()] = struct{}{}
				result = append(result, m)
			}

			counter++
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("manifests not found in %s", dir)
	}

	return result, nil
}
