package form

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
	"github.com/apiqube/cli/internal/core/runner/templates"
)

// DefaultTemplateResolver implements TemplateResolver interface
type DefaultTemplateResolver struct {
	templateEngine *templates.TemplateEngine
	valueExtractor ValueExtractor
}

func NewDefaultTemplateResolver(templateEngine *templates.TemplateEngine, valueExtractor ValueExtractor) *DefaultTemplateResolver {
	return &DefaultTemplateResolver{
		templateEngine: templateEngine,
		valueExtractor: valueExtractor,
	}
}

func (r *DefaultTemplateResolver) Resolve(ctx interfaces.ExecutionContext, templateStr string, pass []*tests.Pass, processedData map[string]any, indexStack []int) (any, error) {
	templateStr = strings.TrimSpace(templateStr)
	content := templateStr
	// If input was wrapped in {{ }}, remove them for processing
	if strings.HasPrefix(templateStr, "{{") && strings.HasSuffix(templateStr, "}}") {
		content = strings.Trim(templateStr, "{} \t")
	}

	// Try to get value from context first
	if val, ok := ctx.Get(content); ok {
		return val, nil
	}

	// Handle Body references
	if strings.HasPrefix(content, "Body.") {
		val, err := r.resolveBodyReference(content, processedData, indexStack)
		return val, err
	}

	// Handle Fake.* templates: always wrap in {{ ... }} for template engine
	if strings.HasPrefix(content, "Fake.") {
		wrapped := "{{ " + content + " }}"
		result, err := r.templateEngine.Execute(wrapped)
		if err != nil {
			return wrapped, err
		}
		return result, nil
	}

	// Use template engine for other cases
	result, err := r.templateEngine.Execute(templateStr)
	if err != nil {
		return templateStr, err
	}
	return result, nil
}

func (r *DefaultTemplateResolver) resolveBodyReference(content string, processedData map[string]any, indexStack []int) (any, error) {
	expr := strings.TrimPrefix(content, "Body.")
	pathParts, funcsPart := r.splitPathAndFuncs(expr)

	// Get value by path with index support
	val, exists := r.valueExtractor.Extract(pathParts, processedData, indexStack)
	if !exists {
		return fmt.Sprintf("{{Body.%s}}", expr), nil
	}

	// Apply functions if present
	if funcsPart != "" {
		finalTemplate := fmt.Sprintf("{{ %s%s }}", val, funcsPart)
		result, err := r.templateEngine.Execute(finalTemplate)
		if err != nil {
			return val, err
		}
		return result, nil
	}

	return val, nil
}

// splitPathAndFuncs separates path parts from function calls
//
// Example: "users.#.name.ToUpper().Substring(1,3)"
//
// Returns: pathParts = ["users", "#", "name"], funcsPart = ".ToUpper().Substring(1,3)"
func (r *DefaultTemplateResolver) splitPathAndFuncs(expr string) (pathParts []string, funcsPart string) {
	parts := strings.Split(expr, ".")

	idxFuncStart := len(parts)
	for i, part := range parts {
		if r.isFuncPart(part) {
			idxFuncStart = i
			break
		}
	}

	pathParts = parts[:idxFuncStart]
	if idxFuncStart < len(parts) {
		funcsPart = "." + strings.Join(parts[idxFuncStart:], ".")
	}

	return pathParts, funcsPart
}

func (r *DefaultTemplateResolver) isFuncPart(part string) bool {
	// Check if part contains function call
	if strings.Contains(part, "(") && strings.Contains(part, ")") {
		return true
	}
	// Check if part starts with uppercase letter (method name)
	if len(part) > 0 && unicode.IsUpper(rune(part[0])) {
		return true
	}
	return false
}
