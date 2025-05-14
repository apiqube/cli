package http

import (
	"bytes"
	"fmt"
	"github.com/apiqube/cli/core/plan"
	"github.com/apiqube/cli/plugins"
	"net/http"
)

func init() {
	plugins.Register(Plugin{})
}

type Plugin struct{}

func (p Plugin) Name() string {
	return "http"
}

func (p Plugin) Execute(step plan.StepConfig, ctx interface{}) (plugins.PluginResult, error) {
	req, _ := http.NewRequest(step.Method, step.URL, bytes.NewBuffer([]byte(step.Body)))
	for k, v := range step.Headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return plugins.PluginResult{Success: false}, err
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body %v\n", err)
		}
	}()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return plugins.PluginResult{Success: true}, nil
	}

	return plugins.PluginResult{Success: false}, nil
}
