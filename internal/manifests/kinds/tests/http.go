package tests

import (
	"fmt"
	"github.com/apiqube/cli/internal/manifests"
	"github.com/apiqube/cli/internal/manifests/kinds"
	"time"
)

var _ manifests.Manifest = (*Http)(nil)
var _ manifests.Defaultable[*Http] = (*Http)(nil)

type Http struct {
	kinds.BaseManifest `yaml:",inline"`

	Spec struct {
		Server string     `yaml:"server,omitempty"`
		Cases  []HttpCase `yaml:"cases" valid:"required,length(1|100)"`
	} `yaml:"spec" valid:"required"`
}

type HttpCase struct {
	Name     string                 `yaml:"name" valid:"required"`
	Method   string                 `yaml:"method" valid:"required,uppercase,in(GET|POST|PUT|DELETE)"`
	Endpoint string                 `yaml:"endpoint,omitempty"`
	Url      string                 `yaml:"url,omitempty"`
	Headers  map[string]string      `yaml:"headers,omitempty"`
	Body     map[string]interface{} `yaml:"body,omitempty"`
	Expected *HttpExpect            `yaml:"expected,omitempty"`
	Extract  *HttpExtractRule       `yaml:"extract,omitempty"`
	Timeout  time.Duration          `yaml:"timeout,omitempty"`
	Async    bool                   `yaml:"async,omitempty"`
	Repeats  int                    `yaml:"repeats,omitempty"`
}

type HttpExpect struct {
	Code    int                    `yaml:"code" valid:"required,range(0|599)"`
	Message string                 `yaml:"message,omitempty"`
	Data    map[string]interface{} `yaml:"data,omitempty"`
}

type HttpExtractRule struct {
	Path  string `yaml:"path,omitempty"`
	Value string `yaml:"value,omitempty"`
}

func (h *Http) GetID() string {
	return fmt.Sprintf("%s.%s.%s", h.Namespace, h.Kind, h.Name)
}

func (h *Http) GetKind() string {
	return h.Kind
}

func (h *Http) GetName() string {
	return h.Name
}

func (h *Http) GetNamespace() string {
	return h.Namespace
}

func (h *Http) GetDependsOn() []string {
	return h.DependsOn
}

func (h *Http) Default() *Http {
	h.Namespace = manifests.DefaultNamespace

	return h
}
