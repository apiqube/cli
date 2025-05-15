package manifests

const (
	DefaultNamespace = "default"

	ServerManifestKind       = "Server"
	ServiceManifestKind      = "Service"
	HttpTestManifestKind     = "HttpTest"
	HttpLoadTestManifestKind = "HttpLoadTest"
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

type Metadata struct {
	Name      string `yaml:"name" valid:"required,alpha"`
	Namespace string `yaml:"namespace" valid:"required,alpha"`
}
