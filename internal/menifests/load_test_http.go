package menifests

import "time"

type LoadTestHttp struct {
	Kind     string   `yaml:"kind"`
	Metadata Metadata `yaml:"metadata"`

	Spec struct {
		Cases []LoadTestHttpCase `yaml:"cases"`
	} `yaml:"spec"`

	Depends []string `yaml:"depends,omitempty"`
}

type LoadTestHttpCase struct {
	Method   string                 `yaml:"method"`
	Endpoint string                 `yaml:"endpoint,omitempty"`
	Url      string                 `yaml:"url,omitempty"`
	Headers  map[string]string      `yaml:"headers,omitempty"`
	Body     map[string]interface{} `yaml:"body,omitempty"`

	Expected *LoadTestHttpCaseExpect  `yaml:"expected,omitempty"`
	Extract  *LoadTestHttpExtractRule `yaml:"extract,omitempty"`

	Timeout time.Duration `yaml:"timeout,omitempty"`
	Async   bool          `yaml:"async,omitempty"`
	Repeats int           `yaml:"repeats,omitempty"`
}

type LoadTestHttpCaseExpect struct {
	Code    int                    `yaml:"code"`
	Message string                 `yaml:"message,omitempty"`
	Data    map[string]interface{} `yaml:"data,omitempty"`
}

type LoadTestHttpExtractRule struct {
	Path  string `yaml:"path"`
	Value string `yaml:"value"`
}
