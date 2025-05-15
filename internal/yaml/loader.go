package yaml

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apiqube/cli/internal/manifests"
)

func LoadManifestsFromDir(dir string) ([]manifests.Manifest, error) {
	var mans []manifests.Manifest

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var manifest manifests.Manifest
	var content []byte

	for _, file := range files {
		if file.IsDir() || file.Name() == "combined.yaml" || (!strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml")) {
			continue
		}

		content, err = os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}

		manifest, err = ParseManifest(content)
		if err != nil {
			return nil, fmt.Errorf("in file %s: %w", file.Name(), err)
		}

		mans = append(mans, manifest)
	}

	return mans, nil
}
