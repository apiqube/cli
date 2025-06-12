package io

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/operations"
	"github.com/goccy/go-yaml"
)

func WriteCombined(path string, format operations.ParseFormat, mans ...manifests.Manifest) error {
	if err := ensureOutputDirectory(path); err != nil {
		return err
	}

	filename := filepath.Join(path, fmt.Sprintf("manifests.%s", format.String()))
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	switch format {
	case operations.JSONFormat:
		return writeCombinedJSON(file, mans)
	default:
		return writeCombinedYAML(file, mans)
	}
}

func WriteSeparate(path string, format operations.ParseFormat, mans ...manifests.Manifest) error {
	if err := ensureOutputDirectory(path); err != nil {
		return err
	}

	for _, m := range mans {
		filename := filepath.Join(path, fmt.Sprintf("%s.%s", m.GetID(), format.String()))
		if err := writeSingleManifest(filename, format, m); err != nil {
			return fmt.Errorf("failed to write manifest %s: %w", m.GetID(), err)
		}
	}

	return nil
}

func writeCombinedYAML(file *os.File, manifests []manifests.Manifest) error {
	encoder := yaml.NewEncoder(file)

	for _, m := range manifests {
		if err := encoder.Encode(m); err != nil {
			return fmt.Errorf("YAML encoding failed: %w", err)
		}
		if _, err := file.WriteString("---\n"); err != nil {
			return fmt.Errorf("failed to write YAML separator: %w", err)
		}
	}

	return nil
}

func writeCombinedJSON(file *os.File, manifests []manifests.Manifest) error {
	if _, err := file.WriteString("[\n"); err != nil {
		return fmt.Errorf("failed to write manifests in file: %w", err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	for i, m := range manifests {
		if i > 0 {
			if _, err := file.WriteString(",\n"); err != nil {
				return fmt.Errorf("failed to write manifests in file: %w", err)
			}
		}
		if err := encoder.Encode(m); err != nil {
			return fmt.Errorf("JSON encoding failed: %w", err)
		}
	}

	if _, err := file.WriteString("\n]"); err != nil {
		return fmt.Errorf("failed to write manifests in file: %w", err)
	}

	return nil
}

func writeSingleManifest(filename string, format operations.ParseFormat, manifest manifests.Manifest) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	switch format {
	case operations.JSONFormat:
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err = encoder.Encode(manifest); err != nil {
			return fmt.Errorf("json encoding failed: %w", err)
		}
	default:
		encoder := yaml.NewEncoder(file)
		if err = encoder.Encode(manifest); err != nil {
			return fmt.Errorf("yaml encoding failed: %w", err)
		}
	}

	return nil
}

func ensureOutputDirectory(path string) error {
	if path == "" {
		path = "."
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	return nil
}
