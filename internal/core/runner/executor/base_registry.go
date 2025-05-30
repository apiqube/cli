package executor

import (
	"sync"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/runner/executor/executors"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

var DefaultRegistry = &DefaultExecutorRegistry{
	executors: map[string]interfaces.Executor{
		manifests.ValuesKind:   executors.NewValuesExecutor(),
		manifests.ServerKind:   executors.NewServerExecutor(),
		manifests.HttpTestKind: executors.NewHTTPExecutor(),
	},
}

var _ interfaces.ExecutorRegistry = (*DefaultExecutorRegistry)(nil)

type DefaultExecutorRegistry struct {
	sync.RWMutex
	executors map[string]interfaces.Executor
}

func NewDefaultExecutorRegistry() *DefaultExecutorRegistry {
	return DefaultRegistry
}

func (r *DefaultExecutorRegistry) Register(kind string, exec interfaces.Executor) {
	r.Lock()
	defer r.Unlock()
	r.executors[kind] = exec
}

func (r *DefaultExecutorRegistry) Find(kind string) (interfaces.Executor, bool) {
	r.RLock()
	defer r.RUnlock()
	exec, ok := r.executors[kind]
	return exec, ok
}
