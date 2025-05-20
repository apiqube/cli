package operations

import "github.com/apiqube/cli/internal/core/manifests"

type ParseFormat uint8

const (
	JSONFormat ParseFormat = iota + 1
	YAMLFormat
)

func (f ParseFormat) String() string {
	switch f {
	case JSONFormat:
		return "json"
	default:
		return "yaml"
	}
}

type Parser interface {
	Parse(format ParseFormat, data []byte) (manifests.Manifest, error)
	ParseBatch(format ParseFormat, data []byte) ([]manifests.Manifest, error)
}

type Editor interface {
	Edit(format ParseFormat, manifest manifests.Manifest) (manifests.Manifest, error)
}
