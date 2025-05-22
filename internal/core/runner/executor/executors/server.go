package executors

import (
	"fmt"
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/servers"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
	"net/http"
	"reflect"
)

const serverExecutorOutputPrefix = "Server Executor:"

var _ interfaces.Executor = (*ServerExecutor)(nil)

type ServerExecutor struct{}

func NewServerExecutor() *ServerExecutor {
	return &ServerExecutor{}
}

func (s *ServerExecutor) Run(ctx interfaces.ExecutionContext, manifest manifests.Manifest) error {
	output := ctx.GetOutput()
	var err error

	select {
	case <-ctx.Done():
		return fmt.Errorf("%s run cancelled, run context was canceled", serverExecutorOutputPrefix)
	default:
	}

	serverMan, ok := manifest.(*servers.Server)
	if !ok {
		return fmt.Errorf("%s manifest %s is not a %s kind", serverExecutorOutputPrefix, manifest.GetID(), manifests.ServerManifestKind)
	}

	if serverMan.Spec.Health != "" {
		var resp *http.Response

		resp, err = http.Get(serverMan.Spec.BaseURL + "/health")
		if err != nil || resp.StatusCode >= 400 {
			output.Logf(interfaces.WarnLevel, "%s server %s (%s) is not responding: %s", serverExecutorOutputPrefix, serverMan.GetName(), serverMan.Spec.BaseURL, err.Error())
		}

		output.Logf(interfaces.InfoLevel, "%s server %s (%s) is responding", serverExecutorOutputPrefix, serverMan.GetName(), serverMan.Spec.Health)
	}

	baseKey := serverMan.GetID()

	ctx.SetTyped(fmt.Sprintf("%s.baseUrl", baseKey), serverMan.Spec.BaseURL, reflect.String)

	for key, val := range serverMan.Spec.Headers {
		ctx.SetTyped(fmt.Sprintf("%s.headers.%s", baseKey, key), val, reflect.TypeOf(val).Kind())
	}

	output.Logf(interfaces.InfoLevel, "%s registered server: %s (%s)", serverExecutorOutputPrefix, serverMan.GetName(), serverMan.Spec.BaseURL)

	return nil
}
