package executors

import (
	"fmt"
	"reflect"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/values"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

const valuesExecutorOutputPrefix = "Values Executor:"

var _ interfaces.Executor = (*ValuesExecutor)(nil)

type ValuesExecutor struct{}

func NewValuesExecutor() *ValuesExecutor {
	return &ValuesExecutor{}
}

func (e *ValuesExecutor) Run(ctx interfaces.ExecutionContext, manifest manifests.Manifest) error {
	output := ctx.GetOutput()

	select {
	case <-ctx.Done():
		return fmt.Errorf("%s run cancelled, run context was canceled", valuesExecutorOutputPrefix)
	default:
	}

	valueMan, ok := manifest.(*values.Values)
	if !ok {
		return fmt.Errorf("%s manifest %s is not a %s kind", valuesExecutorOutputPrefix, manifest.GetID(), manifests.ValuesKind)
	}

	for key, data := range valueMan.Spec.Data {
		ctx.SetTyped(fmt.Sprintf("%s.%s", valueMan.GetID(), key), data, reflect.TypeOf(data).Kind())
	}

	output.Logf(interfaces.InfoLevel, "%s data from %s Values manifests successfully loaded to run context", valuesExecutorOutputPrefix, valueMan.GetName())

	return nil
}
