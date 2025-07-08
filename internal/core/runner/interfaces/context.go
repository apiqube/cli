package interfaces

import (
	"context"

	"github.com/apiqube/cli/internal/core/manifests"
)

type ExecutorRegistry interface {
	Register(kind string, exec Executor)
	Find(kind string) (Executor, bool)
}

type Executor interface {
	Run(ctx ExecutionContext, manifest manifests.Manifest) error
}

type PlanRunner interface {
	Run(ctx ExecutionContext, plan manifests.Manifest) error
}

type ExecutionContext interface {
	context.Context
	ManifestStore
	DataStore
	PassStore
	OutputStore
}
