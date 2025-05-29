package manifests

import (
	"time"
)

const (
	DefaultNamespace = "default"
)

const (
	V1 = "v1"
)

const (
	PlanKind            = "Plan"
	ValuesKind          = "Values"
	ServerKind          = "Server"
	ServiceKind         = "Service"
	HttpTestKind        = "HttpTest"
	HttpLoadTestKind    = "HttpLoadTest"
	GRPCTestKind        = "GRPCTest"
	GRPCLoadTestKind    = "GRPCLoadTest"
	WSTestKind          = "WSTest"
	WSLoadTestKind      = "WSLoadTest"
	GRAPHQLTestKind     = "GraphQLTest"
	GRAPHQLLoadTestKind = "GraphQLLoadTest"
)

type Manifest interface {
	GetID() string
	GetKind() string
	GetName() string
	GetNamespace() string
	Index() any
	GetMeta() Meta
}

type Dependencies interface {
	GetDependsOn() []string
}

type Meta interface {
	GetHash() string
	SetHash(hash string)

	GetVersion() uint8
	SetVersion(version uint8)
	IncVersion()

	GetIsCurrent() bool
	SetIsCurrent(isCurrent bool)

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
