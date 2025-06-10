package executors

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/api"
	"github.com/apiqube/cli/internal/core/runner/assert"
	"github.com/apiqube/cli/internal/core/runner/form"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
	"github.com/apiqube/cli/internal/core/runner/metrics"
	"github.com/apiqube/cli/internal/core/runner/save"
)

const (
	httpExecutorOutputPrefix = "HTTP Executor:"
	httpExecutorRunTimeout   = time.Second * 30
)

var _ interfaces.Executor = (*HTTPExecutor)(nil)

type HTTPExecutor struct {
	client    *http.Client
	extractor *save.Extractor
	assertor  *assert.Runner
	passer    *form.Runner
}

func NewHTTPExecutor() *HTTPExecutor {
	return &HTTPExecutor{
		client:    &http.Client{Timeout: httpExecutorRunTimeout},
		extractor: save.NewExtractor(),
		assertor:  assert.NewRunner(),
		passer:    form.NewRunner(),
	}
}

func (e *HTTPExecutor) Run(ctx interfaces.ExecutionContext, manifest manifests.Manifest) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%s run cancelled, run context was canceled", httpExecutorOutputPrefix)
	default:
	}

	httpMan, ok := manifest.(*api.Http)
	if !ok {
		return fmt.Errorf("%s manifest %s is not a %s kind", httpExecutorOutputPrefix, manifest.GetID(), manifests.HttpTestKind)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(httpMan.Spec.Cases))

	for _, c := range httpMan.Spec.Cases {
		testCase := c
		if testCase.Parallel {
			wg.Add(1)
			go func(tc api.HttpCase) {
				defer wg.Done()
				if err := e.runCase(ctx, httpMan, tc); err != nil {
					errCh <- err
				}
			}(testCase)
		} else {
			if err := e.runCase(ctx, httpMan, testCase); err != nil {
				return err
			}
		}
	}

	wg.Wait()
	close(errCh)

	var rErr error
	for er := range errCh {
		rErr = errors.Join(rErr, er)
	}
	if rErr != nil {
		return rErr
	}
	return nil
}

func (e *HTTPExecutor) runCase(ctx interfaces.ExecutionContext, man *api.Http, c api.HttpCase) error {
	output := ctx.GetOutput()

	caseResult := &interfaces.CaseResult{
		Name:    c.Name,
		Values:  make(map[string]any),
		Details: make(map[string]any),
	}

	var (
		req      *http.Request
		resp     *http.Response
		reqBody  = &bytes.Buffer{}
		respBody = &bytes.Buffer{}
		err      error
	)

	output.StartCase(man, c.Name)
	defer func() {
		metrics.CollectHTTPMetrics(req, resp, c.Details, caseResult)
		e.extractor.Extract(ctx, man, c.HttpCase, resp, reqBody.Bytes(), respBody.Bytes(), caseResult)
		output.EndCase(man, c.Name, caseResult)
	}()

	url := buildHttpURL(c.Url, man.Spec.Target, c.Endpoint)
	url = e.passer.Apply(ctx, url, c.Pass)
	headers := e.passer.MapHeaders(ctx, c.Headers, c.Pass)
	body := e.passer.ApplyBody(ctx, c.Body, c.Pass)

	if body != nil {
		if err = json.NewEncoder(reqBody).Encode(body); err != nil {
			caseResult.Errors = append(caseResult.Errors, fmt.Sprintf("failed to encode request body: %s", err.Error()))
			return fmt.Errorf("encode body failed: %w", err)
		}
	}

	req, err = http.NewRequest(c.Method, url, reqBody)
	if err != nil {
		caseResult.Errors = append(caseResult.Errors, fmt.Sprintf("failed to create request: %s", err.Error()))
		return fmt.Errorf("create request failed: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := e.client
	if c.Timeout > 0 {
		client = &http.Client{Timeout: c.Timeout}
		caseResult.Details["timeout"] = c.Timeout
	}

	start := time.Now()
	resp, err = client.Do(req)
	caseResult.Duration = time.Since(start)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			caseResult.Errors = append(caseResult.Errors, "request timed out")
			return fmt.Errorf("request to %s timed out", url)
		}
		return fmt.Errorf("http request failed: %w", err)
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			caseResult.Errors = append(caseResult.Errors, fmt.Sprintf("failed to close response body: %s", err.Error()))
			output.Logf(interfaces.ErrorLevel, "%s %s response body close failed\nTarget: %s\nName: %s\nMethod: %s\nReason: %s", httpExecutorOutputPrefix, man.GetName(), man.Spec.Target, c.Name, c.Method, err.Error())
		}
	}()

	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		caseResult.Errors = append(caseResult.Errors, fmt.Sprintf("failed to read response body: %s", err.Error()))
		return fmt.Errorf("read response body failed: %w", err)
	}

	if c.Assert != nil {
		output.Logf(interfaces.InfoLevel, "%s response asserting for %s %s", httpExecutorOutputPrefix, man.GetName(), c.Name)
		if err = e.assertor.Assert(ctx, c.Assert, resp, respBody.Bytes()); err != nil {
			caseResult.Assert = "no"
			caseResult.Errors = append(caseResult.Errors, fmt.Sprintf("assertion failed: %s", err.Error()))
			return fmt.Errorf("assert failed: %w", err)
		}
		caseResult.Assert = "yes"
	}

	caseResult.StatusCode = resp.StatusCode
	caseResult.Success = true
	output.Logf(interfaces.InfoLevel, "%s HTTP Test %s passed", httpExecutorOutputPrefix, c.Name)
	return nil
}

func buildHttpURL(url, target, endpoint string) string {
	if url == "" {
		baseUrl := strings.TrimRight(target, "/")
		ep := strings.TrimLeft(endpoint, "/")
		if baseUrl != "" && ep != "" {
			url = baseUrl + "/" + ep
		} else if baseUrl != "" {
			url = baseUrl
		} else {
			url = ep
		}
	}
	return url
}
