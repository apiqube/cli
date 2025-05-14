package menifests

import "time"

type TestHttp struct {
	Kind     string   `yaml:"kind"`
	Metadata Metadata `yaml:"metadata"`
	Spec     struct {
		Cases []TestHttpCase `yaml:"cases,flow"`
	} `yaml:"spec"`
	Depends []string `yaml:"depends,omitempty"`
}

type TestHttpCase struct {
	Method   string                 `yaml:"method"`
	Endpoint string                 `yaml:"endpoint,omitempty"`
	Url      string                 `yaml:"url,omitempty"`
	Headers  map[string]string      `yaml:"headers,omitempty"`
	Body     map[string]interface{} `yaml:"body,omitempty"`
	Expected *TestHttpCaseExpect    `yaml:"expected,omitempty"`
	Extract  *TestHttpExtractRule   `yaml:"extract,omitempty"`
	Timeout  time.Duration          `yaml:"timeout,omitempty"`
	Async    bool                   `yaml:"async,omitempty"`
	Repeats  int                    `yaml:"repeats,omitempty"`
}

type TestHttpCaseExpect struct {
	Code    int                    `yaml:"code"`
	Message string                 `yaml:"message,omitempty"`
	Data    map[string]interface{} `yaml:"data,omitempty"`
}

type TestHttpExtractRule struct {
	Path  string `yaml:"path,omitempty"`
	Value string `yaml:"value,omitempty"`
}
