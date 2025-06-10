package form

import (
	"fmt"
	"strconv"

	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

type DirectiveHandler interface {
	Execute(ctx interfaces.ExecutionContext, runner *Runner, input any, pass []*tests.Pass, processedData map[string]any, indexStack []int) (any, error)
	Name() string
	Dependencies() []string
}

var directiveRegistry = map[string]DirectiveHandler{
	"repeat": newRepeatDirective(),
}

func RegisterDirective(handler DirectiveHandler) {
	directiveRegistry[handler.Name()] = handler
}

type repeatDirective struct {
	name         string
	dependencies []string
}

func newRepeatDirective() *repeatDirective {
	return &repeatDirective{
		name: "repeat",
		dependencies: []string{
			"template",
		},
	}
}

func (r repeatDirective) Name() string {
	return r.name
}

func (r repeatDirective) Dependencies() []string {
	return r.dependencies
}

func (r repeatDirective) Execute(ctx interfaces.ExecutionContext, runner *Runner, input any, pass []*tests.Pass, processedData map[string]any, indexStack []int) (any, error) {
	inputMap, ok := input.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("__repeat: expected map input")
	}

	tmpl, ok := inputMap["__template"]
	if !ok {
		return nil, fmt.Errorf("__repeat: missing __template")
	}

	countRaw, ok := inputMap["__repeat"]
	if !ok {
		return nil, fmt.Errorf("__repeat: missing repeat count")
	}
	count, err := strconv.Atoi(fmt.Sprintf("%v", countRaw))
	if err != nil {
		return nil, fmt.Errorf("__repeat: invalid repeat count: %v", err)
	}

	results := make([]any, 0, count)
	for i := 0; i < count; i++ {
		newStack := append(indexStack[:len(indexStack):len(indexStack)], i)
		processed := runner.processValue(ctx, tmpl, pass, processedData, newStack)
		results = append(results, processed)
	}

	return results, nil
}
