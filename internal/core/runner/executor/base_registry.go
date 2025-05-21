package executor

import (
	"github.com/apiqube/cli/internal/core/runner/interfaces"
	"sync"
)

var _ interfaces.ExecutorRegistry = (*defaultExecutorRegistry)(nil)

type defaultExecutorRegistry struct {
	sync.RWMutex
	executors map[string]interfaces.Executor
}

func (r *defaultExecutorRegistry) Register(kind string, exec interfaces.Executor) {
	r.Lock()
	defer r.Unlock()
	r.executors[kind] = exec
}

func (r *defaultExecutorRegistry) Find(kind string) (interfaces.Executor, bool) {
	r.RLock()
	defer r.RUnlock()
	exec, ok := r.executors[kind]
	return exec, ok
}
