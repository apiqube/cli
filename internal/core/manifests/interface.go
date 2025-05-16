package manifests

import (
	"time"
)

const (
	DefaultNamespace = "default"

	PlanManifestKind            = "Plan"
	ValuesManifestLind          = "Values"
	ServerManifestKind          = "Server"
	ServiceManifestKind         = "Service"
	HttpTestManifestKind        = "HttpTest"
	HttpLoadTestManifestKind    = "HttpLoadTest"
	GRPCTestManifestKind        = "GRPCTest"
	GRPCLoadTestManifestKind    = "GRPCLoadTest"
	WSTestManifestKind          = "WSTest"
	WSLoadTestManifestKind      = "WSLoadTest"
	GRAPHQLTestManifestKind     = "GraphQLTest"
	GRAPHQLLoadTestManifestKind = "GraphQLLoadTest"
)

type Manifest interface {
	GetID() string
	GetKind() string
	GetName() string
	GetNamespace() string
}

type Dependencies interface {
	GetDependsOn() []string
}

type MetaTable interface {
	GetMeta() Meta
}

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

type Defaultable interface {
	Default()
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
