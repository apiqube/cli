package assert

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
	"github.com/tidwall/gjson"
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (a *Runner) Assert(_ interfaces.ExecutionContext, assert *tests.Assert, _ *http.Response, raw []byte, _ any) error {
	for _, el := range assert.Assertions {
		val := gjson.GetBytes(raw, el.Target).Value()

		if el.Equals != nil && !reflect.DeepEqual(val, el.Equals) {
			return fmt.Errorf("expected %v to equal %v", val, el.Equals)
		}
		if el.Contains != "" {
			if s, ok := val.(string); !ok || !strings.Contains(s, el.Contains) {
				return fmt.Errorf("expected %v to contain %q", val, el.Contains)
			}
		}
		if el.Exists && val == nil {
			return fmt.Errorf("expected %v to exist", el.Target)
		}
	}
	return nil
}
