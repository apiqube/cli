package executors

import (
	"fmt"
	"reflect"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/values"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

const valuesExecutorOutputPrefix = "Executor Values:"

var _ interfaces.Executor = (*ValuesExecutor)(nil)

type ValuesExecutor struct{}

func NewValuesExecutor() *ValuesExecutor {
	return &ValuesExecutor{}
}

func (v *ValuesExecutor) Run(ctx interfaces.ExecutionContext) error {
	output := ctx.GetOutput()

	select {
	case <-ctx.Done():
		output.Logf(interfaces.DebugLevel, "%s Run cancelled\nReason: run context was canceled", valuesExecutorOutputPrefix)
		return nil
	default:
	}

	mans, err := ctx.GetManifestsByKind(manifests.ValuesManifestKind)
	if err != nil {
		output.Logf(interfaces.DebugLevel, "%s manifests with kind %s not found in run context", valuesExecutorOutputPrefix, manifests.ValuesManifestKind)
		return nil
	}

	valuesMans := make([]values.Values, 0, len(mans))
	for _, man := range mans {
		if valuesMan, ok := man.(*values.Values); ok {
			valuesMans = append(valuesMans, *valuesMan)
		}
	}

	if len(valuesMans) == 0 {
		output.Logf(interfaces.DebugLevel, "%s didn't find manually any Values manifests in run context", valuesExecutorOutputPrefix)
		return nil
	}

	for _, valMan := range valuesMans {
		for key, data := range valMan.Spec.Data {
			ctx.SetTyped(fmt.Sprintf("%s.%s", valMan.GetID(), key), data, reflect.TypeOf(data).Kind())
		}

		output.Logf(interfaces.InfoLevel, "%s data from %s Values manifests successfully loaded to run context", valuesExecutorOutputPrefix, valMan.GetName())
	}

	output.Logf(interfaces.InfoLevel, "%s run successfully finished", valuesExecutorOutputPrefix)

	return nil
}
