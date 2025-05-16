package parsing

import (
	"bytes"
	"fmt"

	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/api"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/servers"
	"github.com/apiqube/cli/internal/core/manifests/kinds/services"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/load"
	"github.com/apiqube/cli/ui"
	"gopkg.in/yaml.v3"
)

type RawManifest struct {
	Kind string `yaml:"kind"`
}

func ParseYamlManifests(data []byte) ([]manifests.Manifest, error) {
	docs := bytes.Split(data, []byte("\n---"))
	var results []manifests.Manifest

	for _, doc := range docs {
		doc = bytes.TrimSpace(doc)
		if len(doc) == 0 {
			continue
		}

		var raw RawManifest
		if err := yaml.Unmarshal(doc, &raw); err != nil {
			return nil, fmt.Errorf("failed to decode raw s: %w", err)
		}

		var m manifests.Manifest

		switch raw.Kind {
		case manifests.ServerManifestKind:
			var s servers.Server
			if err := s.UnmarshalYAML(doc); err != nil {
				return nil, err
			}
			m = &s

		case manifests.ServiceManifestKind:
			var s services.Service
			if err := s.UnmarshalYAML(doc); err != nil {
				return nil, err
			}
			m = &s

		case manifests.HttpTestManifestKind:
			var h api.Http
			if err := h.UnmarshalYAML(doc); err != nil {
				return nil, err
			}
			m = &h

		case manifests.HttpLoadTestManifestKind:
			var h load.Http
			if err := h.UnmarshalYAML(doc); err != nil {
				return nil, err
			}
			m = &h

		default:
			ui.Errorf("Unknown manifest kind %s", raw.Kind)
		}

		results = append(results, m)
	}

	return results, nil
}
