package interfaces

import (
	"context"
	"github.com/apiqube/cli/internal/core/manifests/kinds/plan"
)

type ExecutorRegistry interface {
	Register(kind string, exec Executor)
	Find(kind string) (Executor, bool)
}

type Executor interface {
	Run(ctx ExecutionContext) error
}

type PlanRunner interface {
	RunPlan(ctx ExecutionContext, plan *plan.Plan) error
}

type ExecutionContext interface {
	context.Context
	ManifestStore
	DataStore
	PassStore
	OutputStore
}
