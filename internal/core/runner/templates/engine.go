package templates

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type TemplateEngine struct {
	funcs   map[string]TemplateFunc
	methods map[string]MethodFunc
	mu      sync.RWMutex
}

type TemplateFunc func(args ...string) (any, error)

type MethodFunc func(value any, args ...string) (any, error)

func New() *TemplateEngine {
	engine := &TemplateEngine{
		funcs:   make(map[string]TemplateFunc),
		methods: make(map[string]MethodFunc),
	}

	engine.RegisterFunc("Fake.name", fakeName)
	engine.RegisterFunc("Fake.email", fakeEmail)
	engine.RegisterFunc("Fake.password", fakePassword)
	engine.RegisterFunc("Fake.int", fakeInt)
	engine.RegisterFunc("Fake.uint", fakeUint)
	engine.RegisterFunc("Fake.float", fakeFloat)
	engine.RegisterFunc("Fake.bool", fakeBool)

	engine.RegisterMethod("ToString", methodToString)
	engine.RegisterMethod("ToUpper", methodToUpper)
	engine.RegisterMethod("ToLower", methodToLower)
	engine.RegisterMethod("Trim", methodTrim)
	engine.RegisterMethod("Clone", methodClone)
	engine.RegisterMethod("Replace", methodReplace)

	return engine
}

func (e *TemplateEngine) RegisterFunc(name string, fn TemplateFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.funcs[name] = fn
}

func (e *TemplateEngine) RegisterMethod(name string, fn MethodFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.methods[name] = fn
}

func (e *TemplateEngine) Execute(template string) (any, error) {
	if isPureDirective(template) {
		return e.processDirective(extractDirective(template))
	}

	re := regexp.MustCompile(`\{\{\s*(.*?)\s*}}`)
	var result strings.Builder
	lastIndex := 0

	for _, match := range re.FindAllStringSubmatchIndex(template, -1) {
		result.WriteString(template[lastIndex:match[0]])

		inner := strings.Trim(template[match[2]:match[3]], " \t")
		val, err := e.processDirective(inner)
		if err != nil {
			return nil, fmt.Errorf("template error: %v", err)
		}

		result.WriteString(fmt.Sprintf("%v", val))
		lastIndex = match[1]
	}

	result.WriteString(template[lastIndex:])

	return result.String(), nil
}

func (e *TemplateEngine) processDirective(directive string) (any, error) {
	parts := strings.Split(directive, ".")
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid directive format")
	}

	fmt.Printf("directive: %s parts: %v\n", directive, parts)

	var value any
	var err error
	processed := 0

	// Попытка найти генератор
	for i := 0; i < len(parts); i++ {
		current := strings.Join(parts[:i+1], ".")
		if generator, ok := e.getGenerator(current); ok {
			args := parts[i+1:]
			var genArgs []string
			for j := 0; j < len(args); j++ {
				if strings.Contains(args[j], "(") {
					break
				}
				genArgs = append(genArgs, args[j])
			}
			value, err = generator(genArgs...)
			if err != nil {
				return nil, err
			}
			processed = i + len(genArgs) + 1
			break
		}
	}

	// 🔧 Новый блок: если генератор не найден — считаем, что это литерал
	if value == nil {
		// строка до первого метода — это значение
		var literalParts []string
		for i := 0; i < len(parts); i++ {
			if strings.Contains(parts[i], "(") {
				break
			}
			literalParts = append(literalParts, parts[i])
			processed++
		}
		value = strings.Join(literalParts, ".")
	}

	// Обработка методов
	for i := processed; i < len(parts); i++ {
		if strings.Contains(parts[i], "(") {
			methodName := strings.Split(parts[i], "(")[0]
			method, ok := e.getMethod(methodName)
			if !ok {
				return nil, fmt.Errorf("unknown method: %s", methodName)
			}

			argsStr := strings.TrimSuffix(strings.SplitN(parts[i], "(", 2)[1], ")")
			args := strings.Split(argsStr, ",")
			for j := range args {
				args[j] = strings.TrimSpace(args[j])
			}

			value, err = method(value, args...)
			if err != nil {
				return nil, err
			}
		}
	}

	return value, nil
}

func (e *TemplateEngine) getGenerator(name string) (TemplateFunc, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	fn, ok := e.funcs[name]
	return fn, ok
}

func (e *TemplateEngine) getMethod(name string) (MethodFunc, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	fn, ok := e.methods[name]
	return fn, ok
}

func isPureDirective(template string) bool {
	trimmed := strings.TrimSpace(template)
	if !strings.HasPrefix(trimmed, "{{") || !strings.HasSuffix(trimmed, "}}") {
		return false
	}

	content := trimmed[2 : len(trimmed)-2]
	return !strings.Contains(content, "{{") && !strings.Contains(content, "}}")
}

func extractDirective(template string) string {
	return strings.TrimSpace(template[2 : len(template)-2])
}
