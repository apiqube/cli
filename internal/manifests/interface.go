package manifests

const (
	DefaultNamespace = "default"

	ServerManifestKind       = "Server"
	ServiceManifestKind      = "Service"
	HttpTestManifestKind     = "HttpTest"
	HttpLoadTestManifestKind = "HttpLoadTest"
)

type Manifest interface {
	GetID() string
	GetKind() string
	GetName() string
	GetNamespace() string
	GetDependsOn() []string
}

type Defaultable[T Manifest] interface {
	Default() T
}
