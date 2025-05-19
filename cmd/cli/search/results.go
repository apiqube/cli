package search

import (
	"encoding/json"
	"fmt"
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/ui"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func sortManifests(manifests []manifests.Manifest, fields []string) {
	sort.Slice(manifests, func(i, j int) bool {
		for _, field := range fields {
			desc := false
			if strings.HasPrefix(field, "-") {
				desc = true
				field = field[1:]
			}

			switch field {
			case "id":
				if manifests[i].GetID() != manifests[j].GetID() {
					if desc {
						return manifests[i].GetID() > manifests[j].GetID()
					}
					return manifests[i].GetID() < manifests[j].GetID()
				}
			case "name":
				if manifests[i].GetName() != manifests[j].GetName() {
					if desc {
						return manifests[i].GetName() > manifests[j].GetName()
					}
					return manifests[i].GetName() < manifests[j].GetName()
				}
			case "kind":
				if manifests[i].GetKind() != manifests[j].GetKind() {
					if desc {
						return manifests[i].GetKind() > manifests[j].GetKind()
					}
					return manifests[i].GetKind() < manifests[j].GetKind()
				}
			case "namespace":
				if manifests[i].GetNamespace() != manifests[j].GetNamespace() {
					if desc {
						return manifests[i].GetNamespace() > manifests[j].GetNamespace()
					}
					return manifests[i].GetNamespace() < manifests[j].GetNamespace()
				}
			case "version":
				if desc {
					return manifests[i].GetMeta().GetVersion() > manifests[j].GetMeta().GetVersion()
				}
				return manifests[i].GetMeta().GetVersion() < manifests[j].GetMeta().GetVersion()
			case "created":
				if desc {
					return manifests[i].GetMeta().GetCreatedAt().After(manifests[j].GetMeta().GetCreatedAt())
				}
				return manifests[i].GetMeta().GetCreatedAt().Before(manifests[j].GetMeta().GetCreatedAt())
			case "updated":
				if desc {
					return manifests[i].GetMeta().GetUpdatedAt().After(manifests[j].GetMeta().GetUpdatedAt())
				}
				return manifests[i].GetMeta().GetUpdatedAt().Before(manifests[j].GetMeta().GetUpdatedAt())
			case "last":
				if desc {
					return manifests[i].GetMeta().GetLastApplied().After(manifests[j].GetMeta().GetLastApplied())
				}
				return manifests[i].GetMeta().GetLastApplied().Before(manifests[j].GetMeta().GetLastApplied())
			}
		}
		return false
	})
}

func displayResults(manifests []manifests.Manifest) {
	headers := []string{
		"#",
		"Hash",
		"kind",
		"name",
		"namespace",
		"version",
		"Created",
		"Updated",
		"Last Updated",
	}

	var rows [][]string
	for i, m := range manifests {
		meta := m.GetMeta()
		row := []string{
			fmt.Sprint(i + 1),
			ui.ShortHash(meta.GetHash()),
			m.GetKind(),
			m.GetName(),
			m.GetNamespace(),
			fmt.Sprint(meta.GetVersion()),
			meta.GetCreatedAt().Format(time.RFC3339),
			meta.GetUpdatedAt().Format(time.RFC3339),
			meta.GetLastApplied().Format(time.RFC3339),
		}
		rows = append(rows, row)
	}

	ui.Table(headers, rows)
}

func handleSearchResults(manifests []manifests.Manifest, opts *Options) error {
	ui.Infof("Found %d manifests", len(manifests))

	if len(opts.sortBy) > 0 {
		sortManifests(manifests, opts.sortBy)
	}

	ui.Spinner(true, "Preparing results...")
	defer ui.Spinner(false)

	if opts.output {
		if err := outputResults(manifests, opts); err != nil {
			return fmt.Errorf("output failed: %w", err)
		}
	} else {
		displayResults(manifests)
	}

	ui.Success("Search completed")
	return nil
}

func outputResults(manifests []manifests.Manifest, opts *Options) error {
	if err := ensureOutputDirectory(opts.outputPath); err != nil {
		return err
	}

	if opts.outputMode == "combined" {
		return writeCombinedOutput(manifests, opts)
	}
	return writeSeparateOutputs(manifests, opts)
}

func writeCombinedOutput(manifests []manifests.Manifest, opts *Options) error {
	filename := filepath.Join(opts.outputPath, fmt.Sprintf("manifests.%s", opts.outputFormat))
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	switch opts.outputFormat {
	case "yaml":
		return writeCombinedYAML(file, manifests)
	case "json":
		return writeCombinedJSON(file, manifests)
	default:
		return fmt.Errorf("unsupported format: %s", opts.outputFormat)
	}
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
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	for i, m := range manifests {
		if i > 0 {
			if _, err := file.WriteString(",\n"); err != nil {
				return err
			}
		}
		if err := encoder.Encode(m); err != nil {
			return err
		}
	}

	_, err := file.WriteString("\n]")
	return err
}

func ensureOutputDirectory(path string) error {
	if path == "" {
		path = "."
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		ui.Infof("Creating output directory: %s", path)
		if err = os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}
	return nil
}

func writeSeparateOutputs(manifests []manifests.Manifest, opts *Options) error {
	for _, m := range manifests {
		filename := filepath.Join(opts.outputPath, fmt.Sprintf("%s.%s", m.GetID(), opts.outputFormat))
		if err := writeSingleManifest(filename, m, opts.outputFormat); err != nil {
			return fmt.Errorf("failed to write manifest %s: %w", m.GetID(), err)
		}
	}
	ui.Successf("Successfully wrote %d manifests to %s", len(manifests), opts.outputPath)
	return nil
}

func writeSingleManifest(filename string, manifest manifests.Manifest, format string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	switch strings.ToLower(format) {
	case "yaml":
		encoder := yaml.NewEncoder(file)
		if err := encoder.Encode(manifest); err != nil {
			return fmt.Errorf("yaml encoding failed: %w", err)
		}
	case "json":
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(manifest); err != nil {
			return fmt.Errorf("json encoding failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
	return nil
}
