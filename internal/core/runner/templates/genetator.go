package templates

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type TemplateEngine struct {
	generators map[string]GeneratorFunc
	methods    map[string]MethodFunc
	mu         sync.RWMutex
	cache      *sync.Map
}

type GeneratorFunc func(args ...string) (any, error)

type MethodFunc func(value any, args ...string) (any, error)

func New() *TemplateEngine {
	engine := &TemplateEngine{
		generators: make(map[string]GeneratorFunc),
		methods:    make(map[string]MethodFunc),
		cache:      &sync.Map{},
	}

	engine.RegisterGenerator("Fake.name", fakeName)
	engine.RegisterGenerator("Fake.email", fakeEmail)
	engine.RegisterGenerator("Fake.password", fakePassword)
	engine.RegisterGenerator("Fake.int", fakeInt)
	engine.RegisterGenerator("Fake.uint", fakeUint)
	engine.RegisterGenerator("Fake.float", fakeFloat)
	engine.RegisterGenerator("Fake.bool", fakeBool)

	engine.RegisterMethod("ToUpper", methodToUpper)
	engine.RegisterMethod("ToLower", methodToLower)
	engine.RegisterMethod("Trim", methodTrim)

	return engine
}

func (e *TemplateEngine) RegisterGenerator(name string, fn GeneratorFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.generators[name] = fn
}

func (e *TemplateEngine) RegisterMethod(name string, fn MethodFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.methods[name] = fn
}

func (e *TemplateEngine) Execute(template string) (any, error) {
	if isPureDirective(template) {
		directive := extractDirective(template)
		return e.processDirective(directive)
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

func (e *TemplateEngine) processDirective(directive string) (any, error) {
	if val, ok := e.cache.Load(directive); ok {
		return val, nil
	}

	parts := strings.Split(directive, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid directive format")
	}

	var value any
	var err error
	processed := 0

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

	if value == nil {
		return nil, fmt.Errorf("no generator found for directive")
	}

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

	e.cache.Store(directive, value)
	return value, nil
}

func (e *TemplateEngine) getGenerator(name string) (GeneratorFunc, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	fn, ok := e.generators[name]
	return fn, ok
}

func (e *TemplateEngine) getMethod(name string) (MethodFunc, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	fn, ok := e.methods[name]
	return fn, ok
}
