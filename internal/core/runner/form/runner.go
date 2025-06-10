package form

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
	"github.com/apiqube/cli/internal/core/runner/templates"
)

type Runner struct {
	templateEngine *templates.TemplateEngine
}

func NewRunner() *Runner {
	return &Runner{
		templateEngine: templates.New(),
	}
}

func (r *Runner) Apply(ctx interfaces.ExecutionContext, input string, pass []*tests.Pass) string {
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

	reg := regexp.MustCompile(`\{\{\s*([^}\s]+)\s*}}`)
	result = reg.ReplaceAllStringFunc(result, func(match string) string {
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

	return result
}

func (r *Runner) ApplyBody(ctx interfaces.ExecutionContext, body map[string]any, pass []*tests.Pass) map[string]any {
	if body == nil {
		return nil
	}

	var indexStack []int

	for key := range body {
		if strings.HasPrefix(key, "__") {
			val, err := r.handleDirective(ctx, body, pass, nil, indexStack)
			if err != nil {
				fmt.Printf("[ERROR] Failed to execute directive %s: %v\n", key, err)
				return nil
			}
			if m, ok := val.(map[string]any); ok {
				return m
			}
		}
	}

	result := make(map[string]any, len(body))
	for key, value := range body {
		result[key] = r.processValue(ctx, value, pass, result, indexStack)
	}

	for key, value := range result {
		result[key] = r.resolveReferences(ctx, value, result, pass, indexStack)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))

	return result
}

func (r *Runner) MapHeaders(ctx interfaces.ExecutionContext, headers map[string]string, pass []*tests.Pass) map[string]string {
	if headers == nil {
		return nil
	}

	result := make(map[string]string, len(headers))
	for key, value := range headers {
		processedKey := r.processValue(ctx, key, pass, nil, []int{}).(string)
		processedValue := r.processValue(ctx, value, pass, nil, []int{}).(string)
		result[processedKey] = processedValue
	}
	return result
}

func (r *Runner) handleDirective(ctx interfaces.ExecutionContext, value any, pass []*tests.Pass, processedData map[string]any, indexStack []int) (any, error) {
	if valMap, is := value.(map[string]any); is {
		for k := range valMap {
			if strings.HasPrefix(k, "__") {
				dirName := strings.TrimPrefix(k, "__")
				if handler, ok := directiveRegistry[dirName]; ok {
					for _, dep := range handler.Dependencies() {
						if _, ok = valMap["__"+dep]; !ok {
							// TODO: emit warning or log missing dependency __ + dep
							fmt.Printf("[WARN] Directive __%s requires __%s\n", dirName, dep)
						}
					}

					return handler.Execute(ctx, r, valMap, pass, processedData, indexStack)
				}
			}
		}
	}
	return value, nil
}

func (r *Runner) processValue(ctx interfaces.ExecutionContext, value any, pass []*tests.Pass, processedData map[string]any, indexStack []int) any {
	if m, ok := value.(map[string]any); ok {
		for k := range m {
			if strings.HasPrefix(k, "__") {
				if directiveVal, err := r.handleDirective(ctx, value, pass, processedData, indexStack); err == nil {
					return directiveVal
				}
			}
		}
	}

	switch v := value.(type) {
	case string:
		return r.processStringValue(ctx, v, pass, processedData, indexStack)
	case map[string]any:
		return r.processMapValue(ctx, v, pass, processedData, indexStack)
	case []any:
		return r.processArrayValue(ctx, v, pass, processedData, indexStack)
	default:
		return v
	}
}

func (r *Runner) processStringValue(ctx interfaces.ExecutionContext, str string, pass []*tests.Pass, processedData map[string]any, indexStack []int) any {
	if isCompleteTemplate(str) {
		res, _ := r.processTemplate(ctx, str, pass, processedData, []int{})
		return res
	}
	return r.processMixedString(ctx, str, pass, processedData, indexStack)
}

func isCompleteTemplate(str string) bool {
	str = strings.TrimSpace(str)
	return strings.HasPrefix(str, "{{") && strings.HasSuffix(str, "}}") && strings.Count(str, "{{") == 1
}

func (r *Runner) processMixedString(ctx interfaces.ExecutionContext, str string, pass []*tests.Pass, processedData map[string]any, indexStack []int) any {
	templateRegex := regexp.MustCompile(`\{\{\s*(.*?)\s*}}`)
	var err error

	for {
		if !templateRegex.MatchString(str) {
			break
		}

		str = templateRegex.ReplaceAllStringFunc(str, func(match string) string {
			inner := strings.TrimSpace(match[2 : len(match)-2]) // убираем {{ }}
			processed, e := r.processTemplate(ctx, "{{ "+inner+" }}", pass, processedData, indexStack)
			if e != nil {
				err = e
				return match // оставляем как есть
			}
			return fmt.Sprintf("%v", processed)
		})
	}

	if err != nil {
		return str
	}
	return str
}

// Главное: indexStack — стек индексов для #, чтобы корректно получать элементы массива
func (r *Runner) processTemplate(ctx interfaces.ExecutionContext, templateStr string, _ []*tests.Pass, processedData map[string]any, indexStack []int) (any, error) {
	templateStr = strings.TrimSpace(templateStr)
	if !strings.HasPrefix(templateStr, "{{") || !strings.HasSuffix(templateStr, "}}") {
		return templateStr, nil
	}

	content := strings.Trim(templateStr, "{} \t")

	if val, ok := ctx.Get(content); ok {
		return val, nil
	}

	if strings.HasPrefix(content, "Body.") {
		expr := strings.TrimPrefix(content, "Body.")
		pathParts, funcsPart := splitPathAndFuncsString(expr)

		// Теперь получаем значение по пути (уже с индексами)
		val, exists := r.getNestedValueWithIndex(pathParts, processedData, indexStack)
		if !exists {
			return templateStr, nil
		}

		// Формируем шаблон с уже подставленным значением (чтобы применить функции, если есть)
		finalTemplate := fmt.Sprintf("{{ %s%s }}", val, funcsPart)

		result, err := r.templateEngine.Execute(finalTemplate)
		if err != nil {
			return val, err
		}

		return result, nil
	}

	result, err := r.templateEngine.Execute(templateStr)
	if err != nil {
		return templateStr, err
	}
	return result, nil
}

// Разделяем путь и цепочку функций:
// Пример: "users.#.name.ToUpper().Substring(1,3)"
// pathParts = ["users", "#", "name"]
// funcsPart = ".ToUpper().Substring(1,3)"
func splitPathAndFuncsString(expr string) (pathParts []string, funcsPart string) {
	// Будем парсить аккуратно, учитывая скобки функций
	// Для простоты: находим первый индекс, где начинается функция — это точка перед заглавной буквой или частью с "()"
	parts := strings.Split(expr, ".")

	idxFuncStart := len(parts)
	for i, part := range parts {
		if isFuncPart(part) {
			idxFuncStart = i
			break
		}
	}

	pathParts = parts[:idxFuncStart]
	if idxFuncStart < len(parts) {
		funcsPart = "." + strings.Join(parts[idxFuncStart:], ".")
	} else {
		funcsPart = ""
	}

	return pathParts, funcsPart
}

func isFuncPart(part string) bool {
	// Проверяем, содержит ли часть вызов функции
	// Пример: ToUpper(), ToString(), Substring(1,3)
	if strings.Contains(part, "(") && strings.Contains(part, ")") {
		return true
	}
	// Или если часть начинается с заглавной буквы и не ключ пути
	if len(part) > 0 && unicode.IsUpper(rune(part[0])) {
		return true
	}

	return false
}

func (r *Runner) getNestedValueWithIndex(parts []string, data any, indexStack []int) (any, bool) {
	current := data
	for _, part := range parts {
		part = strings.TrimSpace(part)
		switch val := current.(type) {
		case map[string]any:
			v, ok := val[part]
			if !ok {
				return nil, false
			}
			current = v

		case []any:
			if part == "#" {
				if len(indexStack) == 0 {
					return nil, false
				}
				idx := indexStack[0]
				indexStack = indexStack[1:]
				if idx < 0 || idx >= len(val) {
					return nil, false
				}
				current = val[idx]
			} else {
				idx, err := strconv.Atoi(part)
				if err != nil || idx < 0 || idx >= len(val) {
					return nil, false
				}
				current = val[idx]
			}

		default:
			// Если есть еще части пути, а мы в конце типа — ошибочно
			if part != parts[len(parts)-1] {
				return nil, false
			}
		}
	}
	return current, true
}

func (r *Runner) processMapValue(ctx interfaces.ExecutionContext, m map[string]any, pass []*tests.Pass, processedData map[string]any, indexStack []int) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = r.processValue(ctx, v, pass, processedData, indexStack)
	}
	return result
}

func (r *Runner) processArrayValue(ctx interfaces.ExecutionContext, arr []any, pass []*tests.Pass, processedData map[string]any, indexStack []int) []any {
	result := make([]any, len(arr))
	for i, item := range arr {
		result[i] = r.processValue(ctx, item, pass, processedData, indexStack)
	}
	return result
}

func (r *Runner) resolveReferences(ctx interfaces.ExecutionContext, value any, processedData map[string]any, pass []*tests.Pass, indexStack []int) any {
	switch v := value.(type) {
	case string:
		return r.resolveStringReferences(ctx, v, processedData, pass, indexStack)
	case map[string]any:
		return r.resolveMapReferences(ctx, v, processedData, pass, indexStack)
	case []any:
		return r.resolveArrayReferences(ctx, v, processedData, pass, indexStack)
	default:
		return v
	}
}

func (r *Runner) resolveStringReferences(ctx interfaces.ExecutionContext, str string, processedData map[string]any, _ []*tests.Pass, indexStack []int) any {
	if !strings.Contains(str, "{{") {
		return str
	}

	templateRegex := regexp.MustCompile(`\{\{\s*(Body\..*?)\s*}}`)
	result := str

	for _, match := range templateRegex.FindAllStringSubmatch(str, -1) {
		if len(match) < 2 {
			continue
		}

		res, _ := r.processTemplate(ctx, str, nil, processedData, indexStack)
		return res
	}

	return result
}

func (r *Runner) resolveMapReferences(ctx interfaces.ExecutionContext, m map[string]any, processedData map[string]any, pass []*tests.Pass, indexStack []int) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = r.resolveReferences(ctx, v, processedData, pass, indexStack)
	}
	return result
}

func (r *Runner) resolveArrayReferences(ctx interfaces.ExecutionContext, arr []any, processedData map[string]any, pass []*tests.Pass, indexStack []int) []any {
	result := make([]any, len(arr))
	for i, item := range arr {
		result[i] = r.resolveReferences(ctx, item, processedData, pass, append(indexStack, i))
	}
	return result
}
