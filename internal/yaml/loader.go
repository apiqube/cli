package yaml

import (
	"fmt"
	"github.com/apiqube/cli/internal/ui"
	"os"
	"path/filepath"
	"strings"

	"github.com/apiqube/cli/internal/manifests"
)

func LoadManifestsFromDir(dir string) ([]manifests.Manifest, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var manifestsSet = make(map[string]struct{})
	var parsedManifests []manifests.Manifest
	var result []manifests.Manifest

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
				ui.Errorf("Duplicate manifest: %s", m.GetID())
				continue
			}

			manifestsSet[m.GetID()] = struct{}{}
			result = append(result, m)
		}
	}

	return result, nil
}
