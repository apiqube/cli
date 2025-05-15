package yaml

import (
	"fmt"
	"github.com/apiqube/cli/internal/manifests"
	"github.com/apiqube/cli/internal/manifests/load"
	"github.com/apiqube/cli/internal/manifests/server"
	"github.com/apiqube/cli/internal/manifests/service"
	"github.com/apiqube/cli/internal/manifests/tests"
	"gopkg.in/yaml.v3"
)

type RawManifest struct {
	Kind string `yaml:"kind"`
}

func ParseManifest(data []byte) (manifests.Manifest, error) {
	var raw RawManifest
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to read kind: %w", err)
	}

	switch raw.Kind {
	case manifests.ServerManifestKind:
		var m server.Server
		if err := yaml.Unmarshal(data, &m); err != nil {
			return nil, err
		}

		return m.Default(), nil
	case manifests.ServiceManifestKind:
		var m service.Service
		if err := yaml.Unmarshal(data, &m); err != nil {
			return nil, err
		}

		return m.Default(), nil

	case manifests.HttpTestManifestKind:
		var m tests.Http
		if err := yaml.Unmarshal(data, &m); err != nil {
			return nil, err
		}

		return m.Default(), nil
	case manifests.HttpLoadTestManifestKind:
		var m load.Http

		if err := yaml.Unmarshal(data, &m); err != nil {
			return nil, err
		}

		return m.Default(), nil

	default:
		return nil, fmt.Errorf("unsupported kind: %s", raw.Kind)
	}
}
