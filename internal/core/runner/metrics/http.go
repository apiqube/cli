package metrics

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

// Request
const (
	RequestGeneral = "Request-General"
	RequestHeaders = "Request-Headers"
	RequestTiming  = "Request-Timing"

	RequestContentLength = "Request-Content-Length"
	RequestContentType   = "Request-Content-Type"
	RequestHost          = "Request-Host"
	RequestMethod        = "Request-Method"
	RequestPost          = "Request-Post"
	RequestProtocol      = "Request-Protocol"
	RequestQuery         = "Request-Query"
	RequestRemoteAddr    = "Request-Remote-Addr"
	RequestURI           = "Request-URI"
	RequestAuthorization = "Request-Authorization"
	RequestCookies       = "Request-Cookies"
	RequestReferer       = "Request-Referer"
	RequestUserAgent     = "Request-User-Agent"
	RequestXForwardedFor = "Request-X-Forwarded-For"
	RequestScheme        = "Request-Scheme"
	RequestPath          = "Request-Path"
	RequestTime          = "Request-Time"
)

// Response
const (
	ResponseGeneral = "Response-General"
	ResponseHeaders = "Response-Headers"
	ResponseTiming  = "Response-Timing"

	ResponseContentLength   = "Response-Content-Length"
	ResponseProtocol        = "Response-Protocol"
	ResponseStatus          = "Response-Status"
	ResponseStatusCode      = "Response-StatusCode"
	ResponseContentType     = "Response-ContentType"
	ResponseTime            = "Response-Time"
	ResponseLatency         = "Response-Latency"
	ResponseLocation        = "Response-Location"
	ResponseCacheControl    = "Response-Cache-Control"
	ResponseContentEncoding = "Response-Content-Encoding"
	ResponseSetCookies      = "Response-Set-Cookies"
	ResponseTrailers        = "Response-Trailers"
)

var detailsMap = map[string]struct{}{
	RequestGeneral:       {},
	RequestHeaders:       {},
	RequestTiming:        {},
	RequestContentLength: {},
	RequestContentType:   {},
	RequestHost:          {},
	RequestMethod:        {},
	RequestPost:          {},
	RequestProtocol:      {},
	RequestQuery:         {},
	RequestRemoteAddr:    {},
	RequestURI:           {},
	RequestAuthorization: {},
	RequestCookies:       {},
	RequestReferer:       {},
	RequestUserAgent:     {},
	RequestXForwardedFor: {},
	RequestScheme:        {},
	RequestPath:          {},
	RequestTime:          {},

	ResponseGeneral:         {},
	ResponseHeaders:         {},
	ResponseTiming:          {},
	ResponseContentLength:   {},
	ResponseProtocol:        {},
	ResponseStatus:          {},
	ResponseStatusCode:      {},
	ResponseContentType:     {},
	ResponseTime:            {},
	ResponseLatency:         {},
	ResponseLocation:        {},
	ResponseCacheControl:    {},
	ResponseContentEncoding: {},
	ResponseSetCookies:      {},
	ResponseTrailers:        {},
}

const unset = "<unset>"

func CollectHTTPMetrics(req *http.Request, resp *http.Response, details []string, result *interfaces.CaseResult) {
	if result.Details == nil {
		result.Details = make(map[string]any)
	}

	for _, detail := range details {
		if _, ok := detailsMap[detail]; !ok {
			continue
		}

		var value any

		// Request metrics
		if strings.HasPrefix(detail, "Request") && req != nil {
			switch detail {
			case RequestGeneral:
				value = fmt.Sprintf("%s %s %s", req.Method, req.URL.RequestURI(), req.Proto)
			case RequestContentLength:
				value = req.ContentLength
				if value == 0 {
					value = unset
				}
			case RequestContentType:
				value = req.Header.Get("Content-Type")
				if value == "" {
					value = unset
				}
			case RequestHost:
				value = req.Host
				if value == "" {
					value = unset
				}
			case RequestMethod:
				value = req.Method
			case RequestPost:
				if req.Method == "POST" {
					if body, err := io.ReadAll(req.Body); err == nil {
						value = string(body)
						req.Body = io.NopCloser(bytes.NewReader(body))
					} else {
						value = "<error>"
					}
				} else {
					value = unset
				}
			case RequestProtocol:
				value = req.Proto
			case RequestQuery:
				value = req.URL.Query().Encode()
				if value == "" {
					value = unset
				}
			case RequestRemoteAddr:
				value = req.RemoteAddr
			case RequestURI:
				value = req.URL.RequestURI()
			case RequestAuthorization:
				value = req.Header.Get("Authorization")
				if value == "" {
					value = unset
				}
			case RequestCookies:
				if cookies := req.Cookies(); len(cookies) > 0 {
					value = cookies
				} else {
					value = unset
				}
			case RequestReferer:
				value = req.Header.Get("Referer")
				if value == "" {
					value = unset
				}
			case RequestUserAgent:
				value = req.Header.Get("User-Agent")
				if value == "" {
					value = unset
				}
			case RequestXForwardedFor:
				value = req.Header.Get("X-Forwarded-For")
				if value == "" {
					value = unset
				}
			case RequestScheme:
				if req.URL != nil {
					value = req.URL.Scheme
				}
				if value == "" {
					value = unset
				}
			case RequestPath:
				value = req.URL.Path
			case RequestTime:
				value = time.Now().Format(time.RFC3339)
			}
		}

		// Response metrics
		if strings.HasPrefix(detail, "Response") && resp != nil {
			switch detail {
			case ResponseGeneral:
				value = fmt.Sprintf("%s %d %s", resp.Proto, resp.StatusCode, resp.Status)
			case ResponseContentLength:
				value = resp.ContentLength
				if value == 0 {
					value = unset
				}
			case ResponseProtocol:
				value = resp.Proto
			case ResponseStatus:
				value = resp.Status
			case ResponseStatusCode:
				value = resp.StatusCode
			case ResponseContentType:
				value = resp.Header.Get("Content-Type")
				if value == "" {
					value = unset
				}
			case ResponseTime:
				value = time.Now().Format(time.RFC3339)
			case ResponseLatency:
				value = result.Duration.String()
			case ResponseLocation:
				value = resp.Header.Get("Location")
				if value == "" {
					value = unset
				}
			case ResponseCacheControl:
				value = resp.Header.Get("Cache-Control")
				if value == "" {
					value = unset
				}
			case ResponseContentEncoding:
				value = resp.Header.Get("Content-Encoding")
				if value == "" {
					value = unset
				}
			case ResponseSetCookies:
				if len(resp.Cookies()) > 0 {
					value = resp.Cookies()
				} else {
					value = unset
				}
			case ResponseTrailers:
				if len(resp.Trailer) > 0 {
					value = resp.Trailer
				} else {
					value = unset
				}
			}
		}

		result.Details[detail] = value
	}
}
