package operations

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/apiqube/cli/internal/core/manifests/kinds/plan"

	"github.com/apiqube/cli/internal/core/manifests/kinds/values"

	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/api"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/servers"
	"github.com/apiqube/cli/internal/core/manifests/kinds/services"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/load"
	"gopkg.in/yaml.v3"
)

type rawManifest struct {
	Kind string `yaml:"kind" json:"kind"`
}

func ParseBatchAsYAML(data []byte) ([]manifests.Manifest, error) {
	docs := bytes.Split(data, []byte("\n---"))
	var results []manifests.Manifest
	var rErr error

	for _, doc := range docs {
		manifest, err := Parse(YAMLFormat, doc)
		if err != nil {
			rErr = errors.Join(rErr, err)
			continue
		}

		results = append(results, manifest)
	}

	return results, rErr
}

func ParseBatchAsJSON(data []byte) ([]manifests.Manifest, error) {
	docs := bytes.Split(data, []byte("\n\n"))

	if bytes.HasPrefix(bytes.TrimSpace(data), []byte("[")) {
		var rawManifests []json.RawMessage
		if err := json.Unmarshal(data, &rawManifests); err != nil {
			return nil, fmt.Errorf("failed to parse JSON array: %w", err)
		}

		var results []manifests.Manifest
		var rErr error
		for _, rawDoc := range rawManifests {
			manifest, err := Parse(JSONFormat, rawDoc)
			if err != nil {
				rErr = errors.Join(rErr, err)
				continue
			}
			results = append(results, manifest)
		}
		return results, rErr
	}

	var results []manifests.Manifest
	var rErr error

	for _, doc := range docs {
		doc = bytes.TrimSpace(doc)
		if len(doc) == 0 {
			continue
		}

		manifest, err := Parse(JSONFormat, doc)
		if err != nil {
			rErr = errors.Join(rErr, err)
			continue
		}
		results = append(results, manifest)
	}

	if len(results) == 0 && rErr != nil {
		return nil, rErr
	}

	return results, rErr
}

func Parse(format ParseFormat, data []byte) (manifests.Manifest, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("provided data not looks li	ke a valid manifest")
	}

	var raw rawManifest
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to recognize manifest kind: %w", err)
	}

	var manifest manifests.Manifest
	switch raw.Kind {
	case manifests.PlanKind:
		manifest = &plan.Plan{}
	case manifests.ValuesKind:
		manifest = &values.Values{}
	case manifests.ServerKind:
		manifest = &servers.Server{}
	case manifests.ServiceKind:
		manifest = &services.Service{}
	case manifests.HttpTestKind:
		manifest = &api.Http{}
	case manifests.HttpLoadTestKind:
		manifest = &load.Http{}
	default:
		return nil, fmt.Errorf("unknown manifest kind: %s", raw.Kind)
	}

	if def, ok := manifest.(manifests.Defaultable); ok {
		def.Default()
	}

	var err error
	switch format {
	case JSONFormat:
		err = json.Unmarshal(data, manifest)
	case YAMLFormat:
		err = yaml.Unmarshal(data, manifest)
	default:
		return nil, fmt.Errorf("unknown parse method: %d", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	if prep, ok := manifest.(manifests.Prepare); ok {
		prep.Prepare()
	}

	return manifest, nil
}
