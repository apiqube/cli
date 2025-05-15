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
			return nil, fmt.Errorf("failed to decode raw manifest: %w", err)
		}

		var manifest manifests.Manifest

		switch raw.Kind {
		case manifests.ServerManifestKind:
			var m server.Server
			if err := yaml.Unmarshal(doc, m.Default()); err != nil {
				return nil, err
			}
			manifest = &m

		case manifests.ServiceManifestKind:
			var m service.Service
			if err := yaml.Unmarshal(doc, m.Default()); err != nil {
				return nil, err
			}
			manifest = &m

		case manifests.HttpTestManifestKind:
			var m tests.Http
			if err := yaml.Unmarshal(doc, m.Default()); err != nil {
				return nil, err
			}
			manifest = &m

		case manifests.HttpLoadTestManifestKind:
			var m load.Http
			if err := yaml.Unmarshal(doc, m.Default()); err != nil {
				return nil, err
			}
			manifest = &m

		default:
			ui.Errorf("Unknown manifest kind %s", raw.Kind)
		}

		results = append(results, manifest)
	}

	return results, nil
}
