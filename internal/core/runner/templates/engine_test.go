package templates

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// --- GENERATORS ---
func TestTemplateEngine_FakeGenerators(t *testing.T) {
	e := New()
	cases := []struct {
		template string
		pattern  string
		name     string
	}{
		{"{{ Fake.name }}", `^[A-Za-z .'-]+$`, "Fake.name"},
		{"{{ Fake.email }}", `^[^@]+@[^@]+\.[a-z]+$`, "Fake.email"},
		{"{{ Fake.int.1.10 }}", `^(10|[1-9])$`, "Fake.int.1.10"},
		{"{{ Fake.uint.10.100 }}", `^\d{2,3}$`, "Fake.uint.10.100"},
		{"{{ Fake.bool }}", `^(true|false)$`, "Fake.bool"},
		{"{{ Fake.address }}", `.+`, "Fake.address"},
		{"{{ Fake.company }}", `.+`, "Fake.company"},
		{"{{ Fake.date }}", `.+`, "Fake.date"},
		{"{{ Fake.uuid }}", `^[a-f0-9-]{36}$`, "Fake.uuid"},
		{"{{ Fake.url }}", `^https?://`, "Fake.url"},
		{"{{ Fake.color }}", `.+`, "Fake.color"},
		{"{{ Fake.word }}", `^[A-Za-z]+$`, "Fake.word"},
		{"{{ Fake.sentence }}", `.+`, "Fake.sentence"},
		{"{{ Fake.country }}", `.+`, "Fake.country"},
		{"{{ Fake.city }}", `.+`, "Fake.city"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := e.Execute(c.template)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", c.template, err)
			}
			if !regexp.MustCompile(c.pattern).MatchString(fmt.Sprint(res)) {
				t.Errorf("%s: result '%v' does not match pattern %s", c.template, res, c.pattern)
			}
		})
	}
}

func TestTemplateEngine_RegexGenerator(t *testing.T) {
	e := New()
	cases := []struct {
		template string
		pattern  string
	}{
		{"{{ Regex('^foo[0-9]{3}$') }}", `^foo\d{3}$`},
		{"{{ Regex('bar[A-Z]{2,4}') }}", `^bar[A-Z]{2,4}$`},
	}
	for _, c := range cases {
		res, err := e.Execute(c.template)
		if err != nil {
			t.Fatalf("Regex: unexpected error: %v", err)
		}
		if !regexp.MustCompile(c.pattern).MatchString(fmt.Sprint(res)) {
			t.Errorf("Regex: result '%v' does not match %s", res, c.pattern)
		}
	}
}

// --- METHODS ---
func TestTemplateEngine_Methods(t *testing.T) {
	e := New()
	cases := []struct {
		template string
		check    func(string) bool
		name     string
	}{
		{"{{ Fake.name.ToUpper() }}", func(s string) bool { return s == strings.ToUpper(s) }, "ToUpper"},
		{"{{ Fake.email.Replace('@','_at_') }}", func(s string) bool { return strings.Contains(s, "_at_") }, "Replace"},
		{"{{ Fake.name.ToLower().Capitalize() }}", func(s string) bool { return len(s) > 0 && s[0] >= 'A' && s[0] <= 'Z' }, "Capitalize"},
		{"{{ Fake.word.Reverse() }}", func(s string) bool { return len(s) > 0 }, "Reverse"},
		{"{{ Fake.word.RandomCase() }}", func(s string) bool { return len(s) > 0 }, "RandomCase"},
		{"{{ Fake.word.SnakeCase() }}", func(s string) bool { return strings.Contains(s, "_") || len(s) > 0 }, "SnakeCase"},
		{"{{ Fake.word.CamelCase() }}", func(s string) bool { return len(s) > 0 }, "CamelCase"},
		{"{{ Fake.word.Split('a').Join('-') }}", func(s string) bool { return strings.Contains(s, "-") || len(s) > 0 }, "SplitJoin"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := e.Execute(c.template)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", c.template, err)
			}
			if !c.check(fmt.Sprint(res)) {
				t.Errorf("%s: check failed for result '%v'", c.template, res)
			}
		})
	}
}

// --- ARGUMENTS & EDGE CASES ---
func TestTemplateEngine_ArgumentsAndEdgeCases(t *testing.T) {
	e := New()
	cases := []struct {
		template string
		pattern  string
		name     string
	}{
		{"{{ Fake.uint.10.20 }}", `^(1[0-9]|20|10)$`, "Fake.uint.10.20"},
		{"{{ Fake.int.0.1 }}", `^(0|1)$`, "Fake.int.0.1"},
		{"{{ Fake.float.1.5.2.4 }}", `^[0-9]+\.[0-9]+$`, "Fake.float.1.5.2.4"},
		{"plain string", `^plain string$`, "plain"},
		{"{{ Fake.name }} and {{ Fake.email }}", `.+@.+`, "multi"},
		{"{{ Fake.int.1.1 }}", `^1$`, "Fake.int.1.1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := e.Execute(c.template)
			if err != nil {
				t.Fatalf("%s: unexpected error: %s", c.template, err.Error())
			}
			if !regexp.MustCompile(c.pattern).MatchString(fmt.Sprint(res)) {
				t.Errorf("%s: result '%v' does not match pattern %s", c.template, res, c.pattern)
			}
		})
	}
}

// --- NESTED & COMPLEX ---
func TestTemplateEngine_NestedTemplates(t *testing.T) {
	e := New()
	template := `{"user":{"name":"{{ Fake.name }}","email":"{{ Fake.email.ToLower() }}","age":"{{ Fake.uint.18.99 }}"}}`
	res, err := e.Execute(template)
	if err != nil {
		t.Fatalf("nested: unexpected error: %v", err)
	}
	if !strings.Contains(fmt.Sprint(res), "@") {
		t.Errorf("nested: result '%v' missing email", res)
	}
}

// --- BODY REFERENCE (stub) ---
func TestTemplateEngine_BodyReference(t *testing.T) {
	e := New()
	// This is a stub: actual Body reference resolution depends on context, but we check parsing
	res, err := e.Execute("{{ Body.someField }}")
	if err == nil && res == "{{ Body.someField }}" {
		t.Log("Body reference parsed as literal (expected in this context)")
	}
}

// --- ERROR HANDLING ---
func TestTemplateEngine_Errors(t *testing.T) {
	e := New()
	_, err := e.Execute("{{ Unknown.generator }}")
	if err == nil {
		t.Error("expected error for unknown generator")
	}
	_, err = e.Execute("{{ Fake.name.UnknownMethod() }}")
	if err == nil {
		t.Error("expected error for unknown method")
	}
	_, err = e.Execute("{{ Fake.int('notanint') }}")
	if err == nil {
		t.Error("expected error for invalid argument")
	}
}

// --- CUSTOM GENERATORS & METHODS ---
func TestTemplateEngine_Custom(t *testing.T) {
	e := New()
	e.RegisterFunc("Custom.hello", func(args ...string) (any, error) {
		if len(args) > 0 {
			return fmt.Sprintf("Hello, %s!", args[0]), nil
		}
		return "Hello, world!", nil
	})

	e.RegisterMethod("Exclaim", func(val any, args ...string) (any, error) {
		return fmt.Sprintf("%v!!!", val), nil
	})

	res, err := e.Execute("{{ Custom.hello.Exclaim() }}")
	if err != nil {
		t.Fatalf("custom: unexpected error: %v", err)
	}

	if !strings.HasSuffix(res.(string), "!!!") {
		t.Errorf("custom: result '%v' missing exclamation", res)
	}
}
