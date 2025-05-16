package manifests

import (
	"github.com/apiqube/cli/internal/manifests/kinds"
	"time"
)

const (
	CombinedManifestsDirPath = "qube/manifests/combined"
	ExecutionPlansDirPath    = "qube/plans/"
)

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

var _ kinds.Meta

type Meta interface {
	GetHash() string
	SetHash(hash string)

	GetVersion() uint8
	SetVersion(version uint8)
	IncVersion()

	GetCreatedAt() time.Time
	SetCreatedAt(createdAt time.Time)

	GetCreatedBy() string
	SetCreatedBy(createdBy string)

	GetUpdatedAt() time.Time
	SetUpdatedAt(updatedAt time.Time)

	GetUpdatedBy() string
	SetUpdatedBy(updatedBy string)

	GetUsedBy() string
	SetUsedBy(usedBy string)

	GetLastApplied() time.Time
	SetLastApplied(lastApplied time.Time)
}

type Defaultable[T Manifest] interface {
	Default() T
}

type Prepare interface {
	Prepare()
}

type Marshaler interface {
	MarshalYAML() ([]byte, error)
	MarshalJSON() ([]byte, error)
}

type Unmarshaler interface {
	UnmarshalYAML([]byte) error
	UnmarshalJSON([]byte) error
}
