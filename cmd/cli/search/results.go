package search

import (
	"fmt"
	"sort"
	"strings"

	"github.com/apiqube/cli/internal/core/io"
	"github.com/apiqube/cli/internal/operations"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/ui/cli"
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
		"Kind",
		"Name",
		"Namespace",
		"Version",
		"Created",
		"Updated",
		"Last Applied",
	}

	var rows [][]string
	for i, m := range manifests {
		meta := m.GetMeta()
		row := []string{
			fmt.Sprint(i + 1),
			cli.ShortHash(meta.GetHash()),
			m.GetKind(),
			m.GetName(),
			m.GetNamespace(),
			fmt.Sprint(meta.GetVersion()),
			meta.GetCreatedAt().Format("2006-01-02 15:04:05"),
			meta.GetUpdatedAt().Format("2006-01-02 15:04:05"),
			meta.GetLastApplied().Format("2006-01-02 15:04:05"),
		}
		rows = append(rows, row)
	}

	cli.Table(headers, rows)
}

func handleSearchResults(manifests []manifests.Manifest, opts *Options) error {
	cli.Success("Search completed")
	cli.Infof("Found %d manifests", len(manifests))

	if len(opts.sortBy) > 0 {
		sortManifests(manifests, opts.sortBy)
	}

	if opts.output {
		var parseFormat operations.ParseFormat
		switch opts.outputFormat {
		case "json":
			parseFormat = operations.JSONFormat
		default:
			parseFormat = operations.YAMLFormat
		}

		if opts.outputMode == "combined" {
			return io.WriteCombined(opts.outputPath, parseFormat, manifests...)
		} else {
			return io.WriteSeparate(opts.outputPath, parseFormat, manifests...)
		}
	} else {
		displayResults(manifests)
	}

	return nil
}
