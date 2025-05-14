package plugins

import (
	"fmt"

	"github.com/apiqube/cli/core/plan"
)

type Plugin interface {
	Name() string
	Execute(step plan.StepConfig, ctx interface{}) (PluginResult, error)
}

type PluginResult struct {
	Success bool
	Code    int
	Message string
}

var registry = map[string]Plugin{}

func Register(p Plugin) {
	registry[p.Name()] = p
}

func GetPlugin(name string) (Plugin, error) {
	p, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}
	return p, nil
}
