package values

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
	"github.com/tidwall/gjson"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(ctx interfaces.ExecutionContext, manifestID string, c tests.HttpCase, resp *http.Response, body []byte) {
	group := c.Save.Group

	saveKey := func(key string) string {
		var builder strings.Builder

		if group != "" {
			builder.WriteString(fmt.Sprintf("Save.%s.%s", group, key))
		} else {
			builder.WriteString(fmt.Sprintf("Save.%s.%s.%s", manifestID, c.Name, key))
		}

		return builder.String()
	}

	// Json
	for key, jsonPath := range c.Save.Json {
		val := gjson.GetBytes(body, jsonPath).Value()
		ctx.SetTyped(saveKey(key), val, reflect.TypeOf(val).Kind())
	}

	// Headers
	for key, headerName := range c.Save.Headers {
		if val := resp.Header.Get(headerName); val != "" {
			ctx.SetTyped(saveKey(fmt.Sprintf(".headers.%s", key)), val, reflect.String)
		}
	}

	if c.Save.All || c.Save.Status && c.Save.Body {
		ctx.SetTyped(saveKey("status"), resp.StatusCode, reflect.Int)
		ctx.Set(saveKey("body"), string(body))
	} else {
		if c.Save.Status {
			ctx.SetTyped(saveKey("status"), resp.StatusCode, reflect.Int)
		} else if c.Save.Body {
			ctx.Set(saveKey("body"), string(body))
		}
	}
}
