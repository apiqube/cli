package yaml

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/apiqube/cli/internal/manifests"
	"gopkg.in/yaml.v3"
)

func SaveManifests(dir string, manifests ...manifests.Manifest) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create dir: %w", err)
	}

	var buf bytes.Buffer

	for i, manifest := range manifests {
		data, err := yaml.Marshal(manifest)
		if err != nil {
			return fmt.Errorf("failed to marshal manifest %d: %w", i, err)
		}

		if i > 0 {
			buf.WriteString("---\n")
		}

		buf.Write(data)
	}

	outputPath := filepath.Join(dir, "combined.yaml")
	if err := os.WriteFile(outputPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
