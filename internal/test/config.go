package test

type Case struct {
	Name             string                 `yaml:"name"`
	Description      string                 `yaml:"description"`
	Type             string                 `yaml:"type"`
	Method           string                 `yaml:"method"`
	URL              string                 `yaml:"url"`
	Headers          map[string]string      `yaml:"headers"`
	Body             map[string]interface{} `yaml:"body"`
	ExpectedResponse ExpectedResponse       `yaml:"expected_response"`
	Flags            []string               `yaml:"flags"`
	UseCase          []string               `yaml:"use_case"`
}

type ExpectedResponse struct {
	StatusCode int                    `yaml:"status_code"`
	Body       map[string]interface{} `yaml:"body"`
}
