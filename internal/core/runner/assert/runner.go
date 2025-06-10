package assert

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/goccy/go-json"

	"github.com/apiqube/cli/internal/core/runner/templates"

	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
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

func (r *Runner) Assert(ctx interfaces.ExecutionContext, asserts []*tests.Assert, resp *http.Response, body []byte) error {
	var err error

	for _, assert := range asserts {
		switch assert.Target {
		case Status.String():
			err = errors.Join(err, r.assertStatus(ctx, assert, resp))
		case Body.String():
			err = errors.Join(err, r.assertBody(ctx, assert, resp, body))
		case Headers.String():
			err = errors.Join(err, r.assertHeaders(ctx, assert, resp))
		default:
			return fmt.Errorf("assert failed: unknown assert target %s", assert.Target)
		}
	}

	return err
}

func (r *Runner) assertStatus(_ interfaces.ExecutionContext, assert *tests.Assert, resp *http.Response) error {
	if assert.Equals != nil {
		var expectedCode int
		switch val := assert.Equals.(type) {
		case uint:
			expectedCode = int(val)
		case uint64:
			expectedCode = int(val)
		case int64:
			expectedCode = int(val)
		case int:
			expectedCode = val
		default:
			return fmt.Errorf("expected status type [int] got %T", assert.Equals)
		}

		if resp.StatusCode != expectedCode {
			return fmt.Errorf("expected status code %v, got %v", expectedCode, resp.StatusCode)
		}
	}

	if assert.Contains != "" {
		if !strings.Contains(resp.Status, assert.Contains) {
			return fmt.Errorf("expected %v to contain %q", resp.Status, assert.Contains)
		}
	}

	return nil
}

func (r *Runner) assertBody(_ interfaces.ExecutionContext, assert *tests.Assert, _ *http.Response, body []byte) error {
	if assert.Exists {
		if len(body) == 0 {
			return fmt.Errorf("expected not null body")
		}

		return nil
	}

	if assert.Template != "" {
		tplResult, err := r.templateEngine.Execute(assert.Template)
		if err != nil {
			return fmt.Errorf("template execution error: %v", err)
		}

		expected, err := json.Marshal(tplResult)
		if err != nil {
			return fmt.Errorf("template marshal error: %v", err)
		}

		if !reflect.DeepEqual(body, expected) {
			return fmt.Errorf("body doesn't match template\nexpected: %s\nactual: %s", expected, body)
		}

		return nil
	}

	if assert.Equals != nil {
		var expected any
		if err := json.Unmarshal([]byte(fmt.Sprintf("%v", assert.Equals)), &expected); err != nil {
			return fmt.Errorf("invalid Equals in body target value: %v", err)
		}

		if !reflect.DeepEqual(body, expected) {
			return fmt.Errorf("expected body %v to equal %v", string(body), expected)
		}
		return nil
	}

	if assert.Contains != "" {
		if !bytes.Contains(body, []byte(assert.Contains)) {
			return fmt.Errorf("expected '%v' in body", assert.Contains)
		}

		return nil
	}

	return nil
}

func (r *Runner) assertHeaders(_ interfaces.ExecutionContext, assert *tests.Assert, resp *http.Response) error {
	if assert.Equals != nil {
		if equals, ok := assert.Equals.(map[string]any); ok {
			return fmt.Errorf("expected map type assertion got %T", assert.Equals)
		} else {
			for key, expectedVal := range equals {
				actualVal := resp.Header.Get(key)
				if fmt.Sprintf("%v", expectedVal) != actualVal {
					return fmt.Errorf("expected header value %v, got %v", expectedVal, actualVal)
				}
			}
		}
	}

	if assert.Contains != "" {
		found := false
		for _, values := range resp.Header {
			for _, val := range values {
				if strings.Contains(val, assert.Contains) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			return fmt.Errorf("expected header contains %v but not found", assert.Contains)
		}
	}

	if assert.Exists {
		if len(resp.Header) == 0 {
			return fmt.Errorf("expected some headers in response")
		}
	}

	return nil
}
