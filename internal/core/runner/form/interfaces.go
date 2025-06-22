package form

import (
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

// Processor defines the interface for processing different types of values
type Processor interface {
	Process(ctx interfaces.ExecutionContext, value any, processedData map[string]any, indexStack []int) any
}

// TemplateResolver defines the interface for resolving templates
type TemplateResolver interface {
	Resolve(ctx interfaces.ExecutionContext, template string, processedData map[string]any, indexStack []int) (any, error)
}

// DirectiveExecutor defines the interface for executing directives
type DirectiveExecutor interface {
	Execute(ctx interfaces.ExecutionContext, value any, processedData map[string]any, indexStack []int) (any, error)
	CanHandle(value any) bool
}

// ReferenceResolver defines the interface for resolving references
type ReferenceResolver interface {
	Resolve(ctx interfaces.ExecutionContext, value any, processedData map[string]any, indexStack []int) any
}

// ValueExtractor defines the interface for extracting values from nested structures
type ValueExtractor interface {
	Extract(path []string, data any, indexStack []int) (any, bool)
}
