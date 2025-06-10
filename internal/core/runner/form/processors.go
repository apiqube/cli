package form

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

// StringProcessor handles string value processing
type StringProcessor struct {
	templateResolver TemplateResolver
}

func NewStringProcessor(templateResolver TemplateResolver) *StringProcessor {
	return &StringProcessor{
		templateResolver: templateResolver,
	}
}

func (p *StringProcessor) Process(ctx interfaces.ExecutionContext, value any, pass []*tests.Pass, processedData map[string]any, indexStack []int) any {
	str, ok := value.(string)
	if !ok {
		return value
	}

	if p.isCompleteTemplate(str) {
		result, _ := p.templateResolver.Resolve(ctx, str, pass, processedData, indexStack)
		return result
	}

	return p.processMixedString(ctx, str, pass, processedData, indexStack)
}

func (p *StringProcessor) isCompleteTemplate(str string) bool {
	str = strings.TrimSpace(str)
	return strings.HasPrefix(str, "{{") && strings.HasSuffix(str, "}}") && strings.Count(str, "{{") == 1
}

func (p *StringProcessor) processMixedString(ctx interfaces.ExecutionContext, str string, pass []*tests.Pass, processedData map[string]any, indexStack []int) any {
	templateRegex := regexp.MustCompile(`\{\{\s*(.*?)\s*}}`)
	var err error

	result := templateRegex.ReplaceAllStringFunc(str, func(match string) string {
		// Extract inner content without brackets
		inner := strings.TrimSpace(match[2 : len(match)-2])
		processed, e := p.templateResolver.Resolve(ctx, inner, pass, processedData, indexStack)
		if e != nil {
			err = e
			return match
		}
		return fmt.Sprint(processed)
	})

	if err != nil {
		return str
	}
	return result
}

// MapProcessor handles map value processing
type MapProcessor struct {
	valueProcessor   Processor
	directiveHandler DirectiveExecutor
}

func NewMapProcessor(valueProcessor Processor, directiveHandler DirectiveExecutor) *MapProcessor {
	return &MapProcessor{
		valueProcessor:   valueProcessor,
		directiveHandler: directiveHandler,
	}
}

func (p *MapProcessor) Process(ctx interfaces.ExecutionContext, value any, pass []*tests.Pass, processedData map[string]any, indexStack []int) any {
	m, ok := value.(map[string]any)
	if !ok {
		return value
	}

	// Check for directives first
	if p.directiveHandler.CanHandle(value) {
		if result, err := p.directiveHandler.Execute(ctx, value, pass, processedData, indexStack); err == nil {
			return result
		}
	}

	// Two-pass processing: first pass for non-Body fields, second for Body-dependent fields
	result := make(map[string]any, len(m))
	// Always pass merged processedData (global + current object) for every field
	for k, v := range m {
		result[k] = p.valueProcessor.Process(ctx, v, pass, mergeProcessedData(processedData, result), indexStack)
	}
	return result
}

// ArrayProcessor handles array value processing
type ArrayProcessor struct {
	valueProcessor Processor
}

func NewArrayProcessor(valueProcessor Processor) *ArrayProcessor {
	return &ArrayProcessor{
		valueProcessor: valueProcessor,
	}
}

func (p *ArrayProcessor) Process(ctx interfaces.ExecutionContext, value any, pass []*tests.Pass, processedData map[string]any, indexStack []int) any {
	arr, ok := value.([]any)
	if !ok {
		return value
	}

	result := make([]any, len(arr))
	for i, item := range arr {
		result[i] = p.valueProcessor.Process(ctx, item, pass, processedData, indexStack)
	}
	return result
}

// CompositeProcessor combines multiple processors
type CompositeProcessor struct {
	stringProcessor *StringProcessor
	mapProcessor    *MapProcessor
	arrayProcessor  *ArrayProcessor
}

func NewCompositeProcessor(templateResolver TemplateResolver, directiveHandler DirectiveExecutor) *CompositeProcessor {
	processor := &CompositeProcessor{}

	processor.stringProcessor = NewStringProcessor(templateResolver)
	processor.mapProcessor = NewMapProcessor(processor, directiveHandler)
	processor.arrayProcessor = NewArrayProcessor(processor)

	return processor
}

func (p *CompositeProcessor) Process(ctx interfaces.ExecutionContext, value any, pass []*tests.Pass, processedData map[string]any, indexStack []int) any {
	switch v := value.(type) {
	case string:
		return p.stringProcessor.Process(ctx, v, pass, processedData, indexStack)
	case map[string]any:
		return p.mapProcessor.Process(ctx, v, pass, processedData, indexStack)
	case []any:
		return p.arrayProcessor.Process(ctx, v, pass, processedData, indexStack)
	default:
		return v
	}
}
