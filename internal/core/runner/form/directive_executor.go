package form

import (
	"fmt"
	"strings"

	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

// defaultDirectiveExecutor implements DirectiveExecutor interface
type defaultDirectiveExecutor struct {
	registry  map[string]DirectiveHandler
	processor Processor
}

func newDefaultDirectiveExecutor(processor Processor) *defaultDirectiveExecutor {
	executor := &defaultDirectiveExecutor{
		registry:  make(map[string]DirectiveHandler),
		processor: processor,
	}

	// Register built-in directives
	executor.RegisterDirective(newRepeatDirective())

	return executor
}

func (e *defaultDirectiveExecutor) RegisterDirective(handler DirectiveHandler) {
	e.registry[handler.Name()] = handler
}

func (e *defaultDirectiveExecutor) CanHandle(value any) bool {
	valMap, ok := value.(map[string]any)
	if !ok {
		return false
	}

	for k := range valMap {
		if strings.HasPrefix(k, "__") {
			return true
		}
	}
	return false
}

func (e *defaultDirectiveExecutor) Execute(ctx interfaces.ExecutionContext, value any, pass []*tests.Pass, processedData map[string]any, indexStack []int) (any, error) {
	valMap, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("directive executor: expected map input")
	}

	for k := range valMap {
		if strings.HasPrefix(k, "__") {
			dirName := strings.TrimPrefix(k, "__")
			if handler, exists := e.registry[dirName]; exists {
				// Check dependencies
				e.checkDependencies(handler, valMap)

				return handler.Execute(ctx, e.processor, valMap, pass, processedData, indexStack)
			}
		}
	}

	return value, nil
}

func (e *defaultDirectiveExecutor) checkDependencies(handler DirectiveHandler, valMap map[string]any) {
	for _, dep := range handler.Dependencies() {
		if _, ok := valMap["__"+dep]; !ok {
			fmt.Printf("[WARN] Directive __%s requires __%s\n", handler.Name(), dep)
		}
	}
}
