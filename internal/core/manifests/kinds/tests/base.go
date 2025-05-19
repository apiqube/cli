package tests

import (
	"time"
)

type HttpCase struct {
	Name     string            `yaml:"name" json:"name" valid:"required"`
	Method   string            `yaml:"method" json:"method" valid:"required,uppercase,in(GET|POST|PUT|DELETE)"`
	Endpoint string            `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	Url      string            `yaml:"url,omitempty" json:"url,omitempty"`
	Headers  map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body     map[string]any    `yaml:"body,omitempty" json:"body,omitempty"`
	Assert   Assert            `yaml:"assert,omitempty" json:"assert,omitempty"`
	Save     Save              `yaml:"save,omitempty" json:"save,omitempty"`
	Pass     Pass              `yaml:"pass,omitempty" json:"pass,omitempty"`
	Timeout  time.Duration     `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Parallel bool              `yaml:"async,omitempty" json:"async,omitempty"`
}

type Assert struct {
	Assertions []*AssertElement `yaml:",inline,omitempty" json:",inline,omitempty"`
}

type AssertElement struct {
	Target   string `yaml:"target,omitempty" json:"target,omitempty"`
	Equals   any    `yaml:"equals,omitempty" json:"equals,omitempty"`
	Contains string `yaml:"contains,omitempty" json:"contains,omitempty"`
	Exists   bool   `yaml:"exists,omitempty" json:"exists,omitempty"`
	Template string `yaml:"template,omitempty" json:"template,omitempty"`
}

type Save struct {
	Json    map[string]string `yaml:"json,omitempty" json:"json,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Status  bool              `yaml:"status,omitempty" json:"status,omitempty"`
	Body    bool              `yaml:"body,omitempty" json:"body,omitempty"`
	All     bool              `yaml:"all,omitempty" json:"all,omitempty"`
	Group   string            `yaml:"group,omitempty" json:"group,omitempty"`
}

type Pass struct {
	From string            `yaml:"from" json:"from"`
	Map  map[string]string `yaml:"map,omitempty" json:"map,omitempty"`
}
