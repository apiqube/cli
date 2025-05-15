package yaml

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/apiqube/cli/internal/manifest"
	"gopkg.in/yaml.v3"
)

func SaveManifestsAsCombined(mans ...manifest.Manifest) error {
	fileName := fmt.Sprintf("/combined-%s.yaml", mans[0].GetNamespace())

	filePath, err := xdg.DataFile(manifest.CombinedManifestsDirPath + fileName)
	if err != nil {
		panic(err)
	}

	if err = os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}

	var buf bytes.Buffer
	var data []byte

	for i, manifest := range mans {
		data, err = yaml.Marshal(manifest)
		if err != nil {
			return fmt.Errorf("failed to marshal manifest %d: %w", i, err)
		}

		if i > 0 {
			buf.WriteString("---\n")
		}

		buf.Write(data)
	}

	if err = os.WriteFile(filePath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
