package assert

import (
	"fmt"
	"net/http"
	"strings"

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

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Assert(ctx interfaces.ExecutionContext, asserts []*tests.Assert, resp *http.Response, body []byte) error {
	for _, assert := range asserts {
		switch assert.Target {
		case Status.String():
			return r.assertStatus(ctx, assert, resp)
		case Body.String():
			return r.assertBody(ctx, assert, resp, body)
		case Headers.String():
			return r.assertHeaders(ctx, assert, resp)
		default:
			return fmt.Errorf("assert failed: unknown assert target %s", assert.Target)
		}
	}
	return nil
}

func (r *Runner) assertStatus(_ interfaces.ExecutionContext, assert *tests.Assert, resp *http.Response) error {
	if assert.Equals != nil {
		expectedCode, ok := assert.Equals.(int)
		if !ok {
			return fmt.Errorf("assertion failed, expected status correct value got %v", assert.Equals)
		}

		if resp.StatusCode != expectedCode {
			return fmt.Errorf("assertion failed, expected status code %v, got %v", expectedCode, resp.StatusCode)
		}
	}

	if assert.Contains != "" {
		if !strings.Contains(resp.Status, assert.Contains) {
			return fmt.Errorf("assertion failed, expected %v to contain %q", resp.Status, assert.Contains)
		}
	}

	return nil
}

func (r *Runner) assertBody(_ interfaces.ExecutionContext, _ *tests.Assert, _ *http.Response, _ []byte) error {
	return nil
}

func (r *Runner) assertHeaders(_ interfaces.ExecutionContext, _ *tests.Assert, _ *http.Response) error {
	return nil
}
