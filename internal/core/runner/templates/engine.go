package templates

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// TemplateEngine is a pluggable engine for evaluating template expressions with generators and methods.
type TemplateEngine struct {
	funcs   map[string]TemplateFunc // Registered generators (e.g. Fake.name, Fake.email)
	methods map[string]MethodFunc   // Registered methods (e.g. ToUpper, Replace)
	mu      sync.RWMutex
}

// TemplateFunc is a generator function (e.g. Fake.name, Fake.int)
type TemplateFunc func(args ...string) (any, error)

// MethodFunc is a method that operates on a value (e.g. ToUpper, Replace)
type MethodFunc func(value any, args ...string) (any, error)

// New creates a new TemplateEngine with built-in Fake generators and methods.
func New() *TemplateEngine {
	engine := &TemplateEngine{
		funcs:   make(map[string]TemplateFunc),
		methods: make(map[string]MethodFunc),
	}

	// Register built-in Fake generators
	engine.RegisterFunc("Fake.name", fakeName)
	engine.RegisterFunc("Fake.email", fakeEmail)
	engine.RegisterFunc("Fake.password", fakePassword)
	engine.RegisterFunc("Fake.int", fakeInt)
	engine.RegisterFunc("Fake.uint", fakeUint)
	engine.RegisterFunc("Fake.float", fakeFloat)
	engine.RegisterFunc("Fake.bool", fakeBool)
	engine.RegisterFunc("Fake.phone", fakePhone)
	engine.RegisterFunc("Fake.address", fakeAddress)
	engine.RegisterFunc("Fake.company", fakeCompany)
	engine.RegisterFunc("Fake.date", fakeDate)
	engine.RegisterFunc("Fake.uuid", fakeUUID)
	engine.RegisterFunc("Fake.url", fakeURL)
	engine.RegisterFunc("Fake.color", fakeColor)
	engine.RegisterFunc("Fake.word", fakeWord)
	engine.RegisterFunc("Fake.sentence", fakeSentence)
	engine.RegisterFunc("Fake.country", fakeCountry)
	engine.RegisterFunc("Fake.city", fakeCity)

	// Register built-in Regex generator
	engine.RegisterFunc("Regex", regex)

	// Register built-in methods
	engine.RegisterMethod("ToString", methodToString)
	engine.RegisterMethod("ToUpper", methodToUpper)
	engine.RegisterMethod("ToLower", methodToLower)
	engine.RegisterMethod("Trim", methodTrim)
	engine.RegisterMethod("Clone", methodClone)
	engine.RegisterMethod("Replace", methodReplace)
	engine.RegisterMethod("PadLeft", methodPadLeft)
	engine.RegisterMethod("PadRight", methodPadRight)
	engine.RegisterMethod("Substring", methodSubstring)
	engine.RegisterMethod("Capitalize", methodCapitalize)
	engine.RegisterMethod("Reverse", methodReverse)
	engine.RegisterMethod("RandomCase", methodRandomCase)
	engine.RegisterMethod("SnakeCase", methodSnakeCase)
	engine.RegisterMethod("CamelCase", methodCamelCase)
	engine.RegisterMethod("Split", methodSplit)
	engine.RegisterMethod("Join", methodJoin)
	engine.RegisterMethod("Index", methodIndex)
	engine.RegisterMethod("Cut", methodCut)
	engine.RegisterMethod("ToInt", methodToInt)
	engine.RegisterMethod("ToUint", methodToUint)
	engine.RegisterMethod("ToFloat", methodToFloat)
	engine.RegisterMethod("ToBool", methodToBool)
	engine.RegisterMethod("ToArray", methodToArray)

	return engine
}

// RegisterFunc registers a new generator function.
func (e *TemplateEngine) RegisterFunc(name string, fn TemplateFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.funcs[name] = fn
}

// RegisterMethod registers a new method.
func (e *TemplateEngine) RegisterMethod(name string, fn MethodFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.methods[name] = fn
}

// Execute evaluates a template string, replacing all {{ ... }} expressions with generated values.
// Supports both pure directives (e.g. {{ Fake.name }}) and mixed strings (e.g. "User: {{ Fake.name }} <{{ Fake.email }}>").
func (e *TemplateEngine) Execute(template string) (any, error) {
	if isPureDirective(template) {
		return e.processDirective(extractDirective(template))
	}

	re := regexp.MustCompile(`\{\{\s*(.*?)\s*}}`)
	var result strings.Builder
	lastIndex := 0

	for _, match := range re.FindAllStringSubmatchIndex(template, -1) {
		result.WriteString(template[lastIndex:match[0]])
		inner := strings.TrimSpace(template[match[2]:match[3]])
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

// processDirective parses and evaluates a directive with optional methods.
// Example: "Fake.uint.10.100.ToString()" or "Regex('pattern').ToUpper()"
func (e *TemplateEngine) processDirective(directive string) (any, error) {
	// Fast path: no methods
	if !strings.Contains(directive, ".") && !strings.Contains(directive, "(") {
		generator, ok := e.getGenerator(directive)
		if !ok {
			return nil, fmt.Errorf("unknown generator: %s", directive)
		}
		return generator()
	}

	// Parse generator and method chain
	genPart, methodsPart := splitGeneratorAndMethods(directive)
	genName, genArgs := parseGeneratorNameAndArgs(genPart)

	generator, ok := e.getGenerator(genName)
	if !ok {
		return nil, fmt.Errorf("unknown generator: %s", genName)
	}
	value, err := generator(genArgs...)
	if err != nil {
		return nil, err
	}

	// Apply method chain if present
	methods := parseMethods(methodsPart)
	for _, m := range methods {
		var method MethodFunc
		method, ok = e.getMethod(m.name)
		if !ok {
			return nil, fmt.Errorf("unknown method: %s", m.name)
		}
		value, err = method(value, m.args...)
		if err != nil {
			return nil, err
		}
	}
	return value, nil
}

// splitGeneratorAndMethods splits directive into generator part and method chain part.
// Handles Fake.uint.10.100, Regex(...), and method chains like .ToUpper().Replace(...)
func splitGeneratorAndMethods(directive string) (string, string) {
	// Regex(...) special case: find closing parenthesis
	if strings.HasPrefix(directive, "Regex(") {
		paren := 0
		for i, r := range directive {
			if r == '(' {
				paren++
			} else if r == ')' {
				paren--
				if paren == 0 {
					return directive[:i+1], directive[i+1:]
				}
			}
		}
		return directive, ""
	}
	// For Fake.*, Body.*, and other dot generators
	paren := 0
	for i := 0; i < len(directive); i++ {
		r := directive[i]
		switch r {
		case '(':
			paren++
		case ')':
			if paren > 0 {
				paren--
			}
		case '.':
			if paren == 0 && i+1 < len(directive) {
				next := directive[i+1]
				if next >= 'A' && next <= 'Z' {
					return directive[:i], directive[i:]
				}
			}
		}
	}
	return directive, ""
}

// parseGeneratorNameAndArgs parses generator name and arguments from the generator part.
// For Fake.uint.10.100 returns (Fake.uint, [10, 100])
func parseGeneratorNameAndArgs(genPart string) (string, []string) {
	if strings.HasPrefix(genPart, "Fake.") || strings.HasPrefix(genPart, "Body.") {
		parts := strings.Split(genPart, ".")
		if len(parts) > 2 {
			return strings.Join(parts[:2], "."), parts[2:]
		}
		return genPart, nil
	}
	if idx := strings.Index(genPart, "("); idx != -1 && strings.HasSuffix(genPart, ")") {
		name := genPart[:idx]
		argsStr := genPart[idx+1 : len(genPart)-1]
		if len(argsStr) > 0 {
			return name, splitArgs(argsStr)
		}
		return name, nil
	}
	return genPart, nil
}

// splitArgs splits arguments by comma, respecting quotes.
func splitArgs(s string) []string {
	var args []string
	var cur strings.Builder
	inQuotes := false
	quoteChar := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c == '\'' || c == '"') && (i == 0 || s[i-1] != '\\') {
			if inQuotes && c == quoteChar {
				inQuotes = false
			} else if !inQuotes {
				inQuotes = true
				quoteChar = c
			}
		}
		if c == ',' && !inQuotes {
			args = append(args, cur.String())
			cur.Reset()
		} else {
			cur.WriteByte(c)
		}
	}
	if cur.Len() > 0 {
		args = append(args, cur.String())
	}
	return args
}

type methodCall struct {
	name string
	args []string
}

// parseMethods parses method chain from a string (e.g. .ToUpper().Replace(' ','_'))
func parseMethods(s string) []methodCall {
	var methods []methodCall
	i := 0
	for i < len(s) {
		if s[i] != '.' {
			i++
			continue
		}
		j := i + 1
		for j < len(s) && ((s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z')) {
			j++
		}
		name := s[i+1 : j]
		var args []string
		if j < len(s) && s[j] == '(' {
			k := j + 1
			paren := 1
			for k < len(s) && paren > 0 {
				if s[k] == '(' {
					paren++
				} else if s[k] == ')' {
					paren--
				}
				k++
			}
			if paren == 0 {
				argStr := s[j+1 : k-1]
				args = splitArgs(argStr)
				j = k
			}
		}
		methods = append(methods, methodCall{name, args})
		i = j
	}
	return methods
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

// isPureDirective checks if the template is a single directive (e.g. {{ Fake.name }})
func isPureDirective(template string) bool {
	trimmed := strings.TrimSpace(template)
	return strings.HasPrefix(trimmed, "{{") && strings.HasSuffix(trimmed, "}}") &&
		strings.Count(trimmed, "{{") == 1 && strings.Count(trimmed, "}}") == 1
}

// extractDirective extracts the inner directive from a template string.
func extractDirective(template string) string {
	return strings.TrimSpace(template[2 : len(template)-2])
}
