package assert

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/goccy/go-json"

	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
	"github.com/apiqube/cli/internal/core/runner/templates"
)

type Type string

const (
	Status  Type = "status"
	Body    Type = "body"
	Headers Type = "headers"
)

func (t Type) String() string {
	return string(t)
}

type Runner struct {
	templateEngine *templates.TemplateEngine
}

func NewRunner() *Runner {
	return &Runner{
		templateEngine: templates.New(),
	}
}

// Assert runs all assertions and aggregates errors.
func (r *Runner) Assert(ctx interfaces.ExecutionContext, asserts []*tests.Assert, resp *http.Response, body []byte) error {
	var err error
	for _, a := range asserts {
		switch a.Target {
		case Status.String():
			err = errors.Join(err, r.assertStatus(ctx, a, resp))
		case Body.String():
			err = errors.Join(err, r.assertBody(ctx, a, resp, body))
		case Headers.String():
			err = errors.Join(err, r.assertHeaders(ctx, a, resp))
		default:
			return fmt.Errorf("assert failed: unknown assert target %s", a.Target)
		}
	}
	return err
}

func (r *Runner) assertStatus(_ interfaces.ExecutionContext, a *tests.Assert, resp *http.Response) error {
	if a.Equals != nil {
		expectedCode, err := toInt(a.Equals)
		if err != nil {
			return fmt.Errorf("expected status type [int] got %T", a.Equals)
		}
		if resp.StatusCode != expectedCode {
			return fmt.Errorf("expected status code %v, got %v", expectedCode, resp.StatusCode)
		}
	}
	if a.Contains != "" {
		if !strings.Contains(resp.Status, a.Contains) {
			return fmt.Errorf("expected %v to contain %q", resp.Status, a.Contains)
		}
	}
	return nil
}

func (r *Runner) assertBody(_ interfaces.ExecutionContext, a *tests.Assert, _ *http.Response, body []byte) error {
	if a.Exists {
		if len(body) == 0 {
			return fmt.Errorf("expected not null body")
		}
		return nil
	}
	if a.Template != "" {
		tplResult, err := r.templateEngine.Execute(a.Template)
		if err != nil {
			return fmt.Errorf("template execution error: %v", err)
		}
		expected, err := json.Marshal(tplResult)
		if err != nil {
			return fmt.Errorf("template marshal error: %v", err)
		}
		if !bytes.Equal(body, expected) {
			return fmt.Errorf("body doesn't match template\nexpected: %s\nactual: %s", expected, body)
		}
		return nil
	}
	if a.Equals != nil {
		var expected any
		// Try to unmarshal as JSON, fallback to string compare
		if err := json.Unmarshal([]byte(fmt.Sprintf("%v", a.Equals)), &expected); err == nil {
			var actual any
			if marshalErr := json.Unmarshal(body, &actual); marshalErr == nil {
				if !reflect.DeepEqual(actual, expected) {
					return fmt.Errorf("expected body %v to equal %v", string(body), expected)
				}
				return nil
			}
		}
		// Fallback: compare as string
		if string(body) != fmt.Sprint(a.Equals) {
			return fmt.Errorf("expected body %v to equal %v", string(body), a.Equals)
		}
		return nil
	}
	if a.Contains != "" {
		if !bytes.Contains(body, []byte(a.Contains)) {
			return fmt.Errorf("expected '%v' in body", a.Contains)
		}
		return nil
	}
	return nil
}

func (r *Runner) assertHeaders(_ interfaces.ExecutionContext, a *tests.Assert, resp *http.Response) error {
	if a.Equals != nil {
		equals, ok := a.Equals.(map[string]any)
		if !ok {
			return fmt.Errorf("expected map[string]any for header equals, got %T", a.Equals)
		}
		for key, expectedVal := range equals {
			actualVal := resp.Header.Get(key)
			if fmt.Sprintf("%v", expectedVal) != actualVal {
				return fmt.Errorf("expected header %v value %v, got %v", key, expectedVal, actualVal)
			}
		}
	}
	if a.Contains != "" {
		found := false
		for _, values := range resp.Header {
			for _, val := range values {
				if strings.Contains(val, a.Contains) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return fmt.Errorf("expected header contains %v but not found", a.Contains)
		}
	}
	if a.Exists {
		if len(resp.Header) == 0 {
			return fmt.Errorf("expected some headers in response")
		}
	}
	return nil
}

// toInt tries to convert interface{} to int for status code assertions.
func toInt(val any) (int, error) {
	switch v := val.(type) {
	case uint:
		return int(v), nil
	case uint64:
		return int(v), nil
	case int64:
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", val)
	}
}
