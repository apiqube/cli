package save

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goccy/go-json"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
	"github.com/tidwall/gjson"
)

const (
	KeyPrefix       = "Save"
	ResultKeySuffix = "Result"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(ctx interfaces.ExecutionContext, man manifests.Manifest, c tests.HttpCase, resp *http.Response, reqBody, respBody []byte, caseResult *interfaces.CaseResult) {
	key := FormSaveKey(man.GetID(), c.Name, ResultKeySuffix)

	result := &Result{
		ManifestID: man.GetID(),
		CaseName:   c.Name,
		Target:     resp.Request.URL.String(),
		Method:     resp.Request.Method,
		StatusCode: resp.StatusCode,
		Duration:   caseResult.Duration,
		Request: &Entry{
			Headers: make(map[string]string),
			Body:    make(map[string]any),
		},
		Response: &Entry{
			Headers: make(map[string]string),
			Body:    make(map[string]any),
		},
	}

	defer func() {
		var builder strings.Builder

		builder.WriteString("\nExtractor:")
		builder.WriteString(fmt.Sprintf("\nID: %s\nCase: %s\nTarget: %s\n Status: %d", result.ManifestID, result.CaseName, result.Target, result.StatusCode))

		reqData, _ := json.MarshalIndent(result.Request, "", " ")
		builder.WriteString(fmt.Sprintf("\n\tRequest: %v", string(reqData)))

		resData, _ := json.MarshalIndent(result.Response, "", " ")
		builder.WriteString(fmt.Sprintf("\n\tResponse: %v", string(resData)))

		ctx.GetOutput().Logf(interfaces.DebugLevel, builder.String())

		ctx.Set(key, result)
	}()

	if c.Save != nil {
		if c.Save.Request != nil {
			result.Request.Headers = e.extractHeaders(c.Save.Request.Headers, resp.Request.Header, result.Request.Headers)
			result.Request.Body = e.extractBody(c.Save.Request.Body, reqBody, result.Response.Body)
		}

		if c.Save.Response != nil {
			result.Response.Headers = e.extractHeaders(c.Save.Response.Headers, resp.Header, result.Response.Headers)
			result.Response.Body = e.extractBody(c.Save.Response.Body, respBody, result.Response.Body)
		}
	}
}

func (e *Extractor) extractHeaders(list []string, origin http.Header, source map[string]string) map[string]string {
	var value string
	for _, l := range list {
		if value = origin.Get(l); value == "" {
			value = "NONE"
		}

		source[l] = value
	}

	return source
}

func (e *Extractor) extractBody(mapList map[string]string, origin []byte, source map[string]any) map[string]any {
	var value any
	var once bool

	for key, path := range mapList {
		if path == "*" && !once {
			if err := json.Unmarshal(origin, &value); err != nil {
				continue
			}
			once = true
		} else {
			result := gjson.GetBytes(origin, path)
			if !result.Exists() {
				continue
			}

			value = e.convertJsonResult(result)
		}

		if value != nil {
			source[key] = value
		}
	}

	return source
}

func (e *Extractor) convertJsonResult(result gjson.Result) any {
	switch {
	case result.IsArray():
		var arr []any
		result.ForEach(func(_, val gjson.Result) bool {
			arr = append(arr, e.convertJsonResult(val))
			return true
		})
		return arr

	case result.IsObject():
		obj := make(map[string]any)
		result.ForEach(func(k, val gjson.Result) bool {
			obj[k.String()] = e.convertJsonResult(val)
			return true
		})
		return obj

	default:
		return result.Value() // string, number, bool, null
	}
}
