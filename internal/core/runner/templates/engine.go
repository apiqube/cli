package templates

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// Precompiled regex for template expressions
var templateExprRe = regexp.MustCompile(`\{\{\s*(.*?)\s*}}`)

// TemplateEngine evaluates template expressions with generators and methods.
type TemplateEngine struct {
	funcs   map[string]TemplateFunc
	methods map[string]MethodFunc
	mu      sync.RWMutex
}

type (
	TemplateFunc func(args ...string) (any, error)
	MethodFunc   func(value any, args ...string) (any, error)
)

func New() *TemplateEngine {
	e := &TemplateEngine{
		funcs:   make(map[string]TemplateFunc),
		methods: make(map[string]MethodFunc),
	}
	// Register built-in generators
	e.RegisterFunc("Fake.name", fakeName)
	e.RegisterFunc("Fake.email", fakeEmail)
	e.RegisterFunc("Fake.password", fakePassword)
	e.RegisterFunc("Fake.int", fakeInt)
	e.RegisterFunc("Fake.uint", fakeUint)
	e.RegisterFunc("Fake.float", fakeFloat)
	e.RegisterFunc("Fake.bool", fakeBool)
	e.RegisterFunc("Fake.phone", fakePhone)
	e.RegisterFunc("Fake.address", fakeAddress)
	e.RegisterFunc("Fake.company", fakeCompany)
	e.RegisterFunc("Fake.date", fakeDate)
	e.RegisterFunc("Fake.uuid", fakeUUID)
	e.RegisterFunc("Fake.url", fakeURL)
	e.RegisterFunc("Fake.color", fakeColor)
	e.RegisterFunc("Fake.word", fakeWord)
	e.RegisterFunc("Fake.sentence", fakeSentence)
	e.RegisterFunc("Fake.country", fakeCountry)
	e.RegisterFunc("Fake.city", fakeCity)
	e.RegisterFunc("Regex", regex)
	// Register built-in methods
	e.RegisterMethod("ToString", methodToString)
	e.RegisterMethod("ToUpper", methodToUpper)
	e.RegisterMethod("ToLower", methodToLower)
	e.RegisterMethod("Trim", methodTrimSpace)
	e.RegisterMethod("Replace", methodReplace)
	e.RegisterMethod("PadLeft", methodPadLeft)
	e.RegisterMethod("PadRight", methodPadRight)
	e.RegisterMethod("Substring", methodSubstring)
	e.RegisterMethod("Capitalize", methodCapitalize)
	e.RegisterMethod("Reverse", methodReverse)
	e.RegisterMethod("RandomCase", methodRandomCase)
	e.RegisterMethod("SnakeCase", methodSnakeCase)
	e.RegisterMethod("CamelCase", methodCamelCase)
	e.RegisterMethod("Split", methodSplit)
	e.RegisterMethod("Join", methodJoin)
	e.RegisterMethod("Index", methodIndex)
	e.RegisterMethod("Cut", methodCut)
	e.RegisterMethod("ToInt", methodToInt)
	e.RegisterMethod("ToUint", methodToUint)
	e.RegisterMethod("ToFloat", methodToFloat)
	e.RegisterMethod("ToBool", methodToBool)
	e.RegisterMethod("ToArray", methodToArray)
	return e
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

// Execute replaces all {{ ... }} expressions with generated values.
func (e *TemplateEngine) Execute(template string) (any, error) {
	if isPureDirective(template) {
		return e.processDirective(extractDirective(template))
	}
	var b strings.Builder
	last := 0
	for _, m := range templateExprRe.FindAllStringSubmatchIndex(template, -1) {
		b.WriteString(template[last:m[0]])
		inner := strings.TrimSpace(template[m[2]:m[3]])
		val, err := e.processDirective(inner)
		if err != nil {
			return nil, fmt.Errorf("template error: %v", err)
		}
		b.WriteString(fmt.Sprint(val))
		last = m[1]
	}
	b.WriteString(template[last:])
	return b.String(), nil
}

// processDirective parses and evaluates a directive with optional methods.
func (e *TemplateEngine) processDirective(directive string) (any, error) {
	genPart, methodsPart := splitGeneratorAndMethods(directive)
	genName, genArgs := parseGeneratorNameAndArgs(genPart)
	generator, ok := e.getGenerator(genName)
	if !ok {
		return nil, fmt.Errorf("unknown generator: %s", genName)
	}
	val, err := generator(genArgs...)
	if err != nil {
		return nil, err
	}
	for _, m := range parseMethods(methodsPart) {
		method, ok := e.getMethod(m.name)
		if !ok {
			return nil, fmt.Errorf("unknown method: %s", m.name)
		}
		val, err = method(val, m.args...)
		if err != nil {
			return nil, err
		}
	}
	return val, nil
}

// splitGeneratorAndMethods splits directive into generator part and method chain part.
func splitGeneratorAndMethods(directive string) (string, string) {
	if strings.HasPrefix(directive, "Regex(") {
		paren := 0
		for i, r := range directive {
			switch r {
			case '(':
				paren++
			case ')':
				paren--
			}
			if paren == 0 && r == ')' {
				return directive[:i+1], directive[i+1:]
			}
		}
		return directive, ""
	}
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
				switch s[k] {
				case '(':
					paren++
				case ')':
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

func isPureDirective(template string) bool {
	trimmed := strings.TrimSpace(template)
	return strings.HasPrefix(trimmed, "{{") && strings.HasSuffix(trimmed, "}}") &&
		strings.Count(trimmed, "{{") == 1 && strings.Count(trimmed, "}}") == 1
}

func extractDirective(template string) string {
	return strings.TrimSpace(template[2 : len(template)-2])
}
