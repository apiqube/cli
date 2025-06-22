package form

import (
	"regexp"
	"strings"

	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

// DefaultReferenceResolver implements ReferenceResolver interface
type DefaultReferenceResolver struct {
	templateResolver TemplateResolver
}

func NewDefaultReferenceResolver(templateResolver TemplateResolver) *DefaultReferenceResolver {
	return &DefaultReferenceResolver{
		templateResolver: templateResolver,
	}
}

func (r *DefaultReferenceResolver) Resolve(ctx interfaces.ExecutionContext, value any, processedData map[string]any, indexStack []int) any {
	switch v := value.(type) {
	case string:
		return r.resolveStringReferences(ctx, v, processedData, indexStack)
	case map[string]any:
		return r.resolveMapReferences(ctx, v, processedData, indexStack)
	case []any:
		return r.resolveArrayReferences(ctx, v, processedData, indexStack)
	default:
		return v
	}
}

func (r *DefaultReferenceResolver) resolveStringReferences(ctx interfaces.ExecutionContext, str string, processedData map[string]any, indexStack []int) any {
	if !strings.Contains(str, "{{") {
		return str
	}

	templateRegex := regexp.MustCompile(`\{\{\s*(Body\..*?)\s*}}`)
	matches := templateRegex.FindAllStringSubmatch(str, -1)

	if len(matches) > 0 {
		// If we found Body references, resolve the entire string as template
		result, _ := r.templateResolver.Resolve(ctx, str, processedData, indexStack)
		return result
	}

	return str
}

func (r *DefaultReferenceResolver) resolveMapReferences(ctx interfaces.ExecutionContext, m map[string]any, processedData map[string]any, indexStack []int) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = r.Resolve(ctx, v, processedData, indexStack)
	}
	return result
}

func (r *DefaultReferenceResolver) resolveArrayReferences(ctx interfaces.ExecutionContext, arr []any, processedData map[string]any, indexStack []int) []any {
	result := make([]any, len(arr))
	for i, item := range arr {
		// Add current index to stack for nested array processing
		newIndexStack := append(indexStack, i)
		result[i] = r.Resolve(ctx, item, processedData, newIndexStack)
	}
	return result
}
