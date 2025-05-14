package menifests

type Manifest interface {
	GetVersion() string
	GetKind() string
	GetName() string
	Hash() string
}

type RawManifest struct {
	Kind string `yaml:"kind"`
}

type Metadata struct {
	Name  string `yaml:"name"`
	Group string `yaml:"group,omitempty"`
}
