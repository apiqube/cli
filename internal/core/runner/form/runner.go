package form

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/goccy/go-json"

	"github.com/apiqube/cli/internal/core/runner/interfaces"
	"github.com/apiqube/cli/internal/core/runner/templates"
)

// Runner is the main form processing engine
type Runner struct {
	processor         Processor
	templateResolver  TemplateResolver
	referenceResolver ReferenceResolver
	templateEngine    *templates.TemplateEngine
}

// NewRunner creates a new form runner with all dependencies properly wired
func NewRunner() *Runner {
	templateEngine := templates.New()
	valueExtractor := NewDefaultValueExtractor()
	templateResolver := NewDefaultTemplateResolver(templateEngine, valueExtractor)
	referenceResolver := NewDefaultReferenceResolver(templateResolver)

	// Create processor with circular dependency resolution
	var processor Processor
	directiveExecutor := newDefaultDirectiveExecutor(nil) // Will be set later
	processor = NewCompositeProcessor(templateResolver, directiveExecutor)

	// Now set the processor in directive executor
	directiveExecutor.processor = processor

	return &Runner{
		processor:         processor,
		templateResolver:  templateResolver,
		referenceResolver: referenceResolver,
		templateEngine:    templateEngine,
	}
}

// RegisterDirective allows registering custom directives
func (r *Runner) RegisterDirective(handler DirectiveHandler) {
	if executor, ok := r.processor.(*CompositeProcessor).mapProcessor.directiveHandler.(*defaultDirectiveExecutor); ok {
		executor.RegisterDirective(handler)
	}
}

// Apply processes a string input with pass mappings and template resolution
func (r *Runner) Apply(ctx interfaces.ExecutionContext, input string) string {
	result := input
	result = r.applyTemplateResolution(ctx, result)
	return result
}

// ApplyBody processes a map body with full form processing capabilities
func (r *Runner) ApplyBody(ctx interfaces.ExecutionContext, body map[string]any) map[string]any {
	if body == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		return nil
	default:
	}

	processed := r.processor.Process(ctx, body, nil, []int{})
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	if processedMap, ok := processed.(map[string]any); ok {
		resolved := r.referenceResolver.Resolve(ctx, processedMap, processedMap, []int{})
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if resolvedMap, is := resolved.(map[string]any); is {
			if data, err := json.MarshalIndent(resolvedMap, "", "  "); err == nil {
				fmt.Println(string(data))
			}
			return resolvedMap
		}
	}
	return body
}

// MapHeaders processes header mappings
func (r *Runner) MapHeaders(ctx interfaces.ExecutionContext, headers map[string]string) map[string]string {
	if headers == nil {
		return nil
	}
	result := make(map[string]string, len(headers))
	for key, value := range headers {
		select {
		case <-ctx.Done():
			return result
		default:
		}

		processedKey := r.processHeaderValue(ctx, key)
		processedValue := r.processHeaderValue(ctx, value)
		result[processedKey] = processedValue
	}
	return result
}

// GetTemplateEngine returns the underlying template engine for advanced usage
func (r *Runner) GetTemplateEngine() *templates.TemplateEngine {
	return r.templateEngine
}

func (r *Runner) applyTemplateResolution(ctx interfaces.ExecutionContext, input string) string {
	reg := regexp.MustCompile(`\{\{\s*([^}\s]+)\s*}}`)
	return reg.ReplaceAllStringFunc(input, func(match string) string {
		select {
		case <-ctx.Done():
			return match
		default:
		}
		key := strings.Trim(match, "{} \t")
		if val, ok := ctx.Get(key); ok {
			return fmt.Sprintf("%v", val)
		}
		if strings.HasPrefix(key, "Fake.") {
			if val, err := r.templateEngine.Execute(match); err == nil {
				return fmt.Sprintf("%v", val)
			}
		}
		return match
	})
}

func (r *Runner) processHeaderValue(ctx interfaces.ExecutionContext, value string) string {
	select {
	case <-ctx.Done():
		return value
	default:
	}
	processed := r.processor.Process(ctx, value, nil, []int{})
	select {
	case <-ctx.Done():
		return value
	default:
	}
	if str, ok := processed.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", processed)
}
