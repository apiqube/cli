package form

import (
	"fmt"
	"strconv"

	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

// DirectiveHandler defines the interface for directive handlers
type DirectiveHandler interface {
	Execute(ctx interfaces.ExecutionContext, runner Processor, input any, processedData map[string]any, indexStack []int) (any, error)
	Name() string
	Dependencies() []string
}

// repeatDirective implements the __repeat directive
type repeatDirective struct {
	name         string
	dependencies []string
}

func newRepeatDirective() *repeatDirective {
	return &repeatDirective{
		name:         "repeat",
		dependencies: []string{"template"},
	}
}

func (r *repeatDirective) Name() string {
	return r.name
}

func (r *repeatDirective) Dependencies() []string {
	return r.dependencies
}

func (r *repeatDirective) Execute(ctx interfaces.ExecutionContext, processor Processor, input any, processedData map[string]any, indexStack []int) (any, error) {
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
	// Prepare or reuse processedData array for this field if possible
	var arrKey string
	for k := range inputMap {
		if k != "__repeat" && k != "__template" {
			arrKey = k
			break
		}
	}
	if arrKey == "" {
		arrKey = "users" // fallback
	}
	if processedData != nil {
		if _, ok = processedData[arrKey]; !ok {
			processedData[arrKey] = make([]any, count)
		}
	}
	for i := 0; i < count; i++ {
		newStack := append(indexStack[:len(indexStack):len(indexStack)], i)
		processed := processor.Process(ctx, tmpl, processedData, newStack)
		results = append(results, processed)
		// Если processed — map, сразу положить его в processedData[arrKey][i]
		if processedData != nil {
			var arr []any
			if arr, ok = processedData[arrKey].([]any); ok && i < len(arr) {
				arr[i] = processed
				processedData[arrKey] = arr
			}
		}
	}

	// После генерации всех элементов users, пройтись по каждому и вычислить вложенные map-поля
	if processedData != nil {
		var arr []any
		if arr, ok = processedData[arrKey].([]any); ok {
			for i, elem := range arr {
				var m map[string]any
				if m, ok = elem.(map[string]any); ok {
					for k, v := range m {
						if submap, is := v.(map[string]any); is {
							// Переобработать вложенную map с актуальным processedData
							m[k] = processor.Process(ctx, submap, mergeProcessedData(processedData, m), []int{i})
						}
					}
					arr[i] = m
				}
			}
			processedData[arrKey] = arr
		}
	}

	return results, nil
}
