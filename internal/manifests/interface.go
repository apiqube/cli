package manifests

const (
	DefaultNamespace = "default"
)

type Manifest interface {
	GetKind() string
	GetName() string
	GetNamespace() string
	GetDependsOn() []string
}

type Defaultable[T Manifest] interface {
	Default() T
}

type RawManifest struct {
	Kind string `yaml:"kind"`
}

type Metadata struct {
	Name      string `yaml:"name" valid:"required,alpha"`
	Namespace string `yaml:"namespace" valid:"required,alpha"`
}
