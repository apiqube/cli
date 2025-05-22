package executors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/api"
	"github.com/apiqube/cli/internal/core/runner/assert"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
	"github.com/apiqube/cli/internal/core/runner/pass"
	"github.com/apiqube/cli/internal/core/runner/values"
)

const httpExecutorOutputPrefix = "HTTP Executor:"

var _ interfaces.Executor = (*HTTPExecutor)(nil)

type HTTPExecutor struct {
	client    *http.Client
	extractor *values.Extractor
	assertor  *assert.Runner
	passer    *pass.Runner
}

func NewHTTPExecutor() *HTTPExecutor {
	return &HTTPExecutor{
		client:    &http.Client{Timeout: 30 * time.Second},
		extractor: values.NewExtractor(),
		assertor:  assert.NewRunner(),
		passer:    pass.NewRunner(),
	}
}

func (e *HTTPExecutor) Run(ctx interfaces.ExecutionContext, manifest manifests.Manifest) error {
	_ = ctx.GetOutput()
	var err error

	select {
	case <-ctx.Done():
		return fmt.Errorf("%s run cancelled, run context was canceled", httpExecutorOutputPrefix)
	default:
	}

	httpMan, ok := manifest.(*api.Http)
	if !ok {
		return fmt.Errorf("%s manifest %s is not a %s kind", httpExecutorOutputPrefix, manifest.GetID(), manifests.HttpTestManifestKind)
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(httpMan.Spec.Cases))

	for _, c := range httpMan.Spec.Cases {
		testCase := c
		if testCase.Parallel {
			wg.Add(1)
			go func() {
				defer wg.Done()
				var caseErr error
				if caseErr = e.runCase(ctx, httpMan, testCase); err != nil {
					errs <- caseErr
				}
			}()
		} else {
			if err = e.runCase(ctx, httpMan, testCase); err != nil {
				return err
			}
		}
	}

	wg.Wait()
	close(errs)

	var rErr error
	if len(errs) > 0 {
		for er := range errs {
			rErr = errors.Join(rErr, er)
		}

		return rErr
	}

	return nil
}

func (e *HTTPExecutor) runCase(ctx interfaces.ExecutionContext, man *api.Http, c api.HttpCase) error {
	output := ctx.GetOutput()
	url := c.Url

	if url == "" {
		url = strings.TrimRight(man.Spec.Target, "/") + "/" + strings.TrimLeft(c.Endpoint, "/")
	}

	url = e.passer.Apply(ctx, url, c.Pass)
	headers := e.passer.MapHeaders(ctx, c.Headers, c.Pass)
	body := e.passer.ApplyBody(ctx, c.Body, c.Pass)

	reqBody := &bytes.Buffer{}

	if body != nil {
		if err := json.NewEncoder(reqBody).Encode(body); err != nil {
			return fmt.Errorf("encode body failed: %w", err)
		}
	}

	req, err := http.NewRequest(c.Method, url, reqBody)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	timeout := c.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			output.Logf(interfaces.ErrorLevel, "%s %s response body closed failed\nTarget: %s\nName: %s\nMathod: %s\nReason: %s", httpExecutorOutputPrefix, man.GetName(), man.Spec.Target, c.Name, c.Method, err.Error())
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}

	var jsonBody any
	if err = json.Unmarshal(respBody, &jsonBody); err != nil {
		return fmt.Errorf("parse response body failed: %w", err)
	}

	if c.Save != nil {
		e.extractor.Extract(ctx, man.GetID(), c.HttpCase, resp, respBody, jsonBody)
	}

	if c.Assert != nil {
		if err = e.assertor.Assert(ctx, c.Assert, resp, respBody, jsonBody); err != nil {
			return fmt.Errorf("assert failed: %w", err)
		}
	}

	output.Logf(interfaces.InfoLevel, "%s HTTP Test %s passed", httpExecutorOutputPrefix, c.Name)

	return nil
}
