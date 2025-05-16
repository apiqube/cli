package yaml

import (
	"bytes"
	"fmt"

	"github.com/apiqube/cli/internal/manifests"
	"github.com/apiqube/cli/internal/manifests/kinds/load"
	"github.com/apiqube/cli/internal/manifests/kinds/server"
	"github.com/apiqube/cli/internal/manifests/kinds/service"
	"github.com/apiqube/cli/internal/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/ui"
	"gopkg.in/yaml.v3"
)

type RawManifest struct {
	Kind string `yaml:"kind"`
}

func ParseManifests(data []byte) ([]manifests.Manifest, error) {
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
			var s server.Server
			if err := yaml.Unmarshal(doc, s.Default()); err != nil {
				return nil, err
			}
			m = &s

		case manifests.ServiceManifestKind:
			var s service.Service
			if err := yaml.Unmarshal(doc, s.Default()); err != nil {
				return nil, err
			}
			m = &s

		case manifests.HttpTestManifestKind:
			var h tests.Http
			if err := yaml.Unmarshal(doc, h.Default()); err != nil {
				return nil, err
			}
			m = &h

		case manifests.HttpLoadTestManifestKind:
			var h load.Http
			if err := yaml.Unmarshal(doc, h.Default()); err != nil {
				return nil, err
			}
			m = &h

		default:
			ui.Errorf("Unknown s kind %s", raw.Kind)
		}

		results = append(results, m)
	}

	return results, nil
}
