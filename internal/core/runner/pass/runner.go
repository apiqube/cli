package pass

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
	"github.com/apiqube/cli/internal/core/runner/templates"
)

type Runner struct {
	fakeEngine *templates.TemplateEngine
}

func NewRunner() *Runner {
	return &Runner{
		fakeEngine: templates.New(),
	}
}

func (r *Runner) Apply(ctx interfaces.ExecutionContext, input string, pass []tests.Pass) string {
	result := input
	for _, p := range pass {
		if p.Map != nil {
			for placeholder, mapKey := range p.Map {
				if strings.Contains(result, placeholder) {
					if val, ok := ctx.Get(mapKey); ok {
						result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", val))
					}
				}
			}
		}
	}

	re := regexp.MustCompile(`\{\{\s*([^}\s]+)\s*}}`)
	result = re.ReplaceAllStringFunc(result, func(match string) string {
		key := strings.Trim(match, "{} \t")

		if val, ok := ctx.Get(key); ok {
			return fmt.Sprintf("%v", val)
		}

		if strings.HasPrefix(key, "Fake.") {
			if val, err := r.fakeEngine.Execute(match); err == nil {
				return fmt.Sprintf("%v", val)
			}
		}

		return match
	})

	return result
}

func (r *Runner) ApplyBody(ctx interfaces.ExecutionContext, body map[string]any, pass []tests.Pass) map[string]any {
	if body == nil {
		return nil
	}

	result := make(map[string]any, len(body))

	for key, value := range body {
		switch v := value.(type) {
		case string:
			result[key] = r.renderTemplate(ctx, v)
		case map[string]any:
			result[key] = r.ApplyBody(ctx, v, pass)
		case []any:
			result[key] = r.applyArray(ctx, v, pass)
		default:
			result[key] = v
		}
	}

	return result
}

func (r *Runner) MapHeaders(ctx interfaces.ExecutionContext, headers map[string]string, pass []tests.Pass) map[string]string {
	if headers == nil {
		return nil
	}

	result := make(map[string]string, len(headers))

	for key, value := range headers {
		processedKey := r.Apply(ctx, key, pass)
		processedValue := r.Apply(ctx, value, pass)
		result[processedKey] = processedValue
	}

	return result
}

func (r *Runner) renderTemplate(_ interfaces.ExecutionContext, raw string) any {
	if !strings.Contains(raw, "{{") {
		return raw
	}

	if strings.Contains(raw, "Fake") {
		result, err := r.fakeEngine.Execute(raw)
		if err != nil {
			return raw
		}

		return result
	}

	return raw
}

func (r *Runner) applyArray(ctx interfaces.ExecutionContext, arr []any, pass []tests.Pass) []any {
	if arr == nil {
		return nil
	}

	result := make([]any, len(arr))
	for i, item := range arr {
		switch v := item.(type) {
		case string:
			result[i] = r.renderTemplate(ctx, v)
		case map[string]any:
			result[i] = r.ApplyBody(ctx, v, pass)
		case []any:
			result[i] = r.applyArray(ctx, v, pass)
		default:
			result[i] = v
		}
	}
	return result
}
