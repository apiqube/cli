package api

import (
	"fmt"

	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
)

var (
	_ manifests.Manifest     = (*Http)(nil)
	_ manifests.Dependencies = (*Http)(nil)
	_ manifests.MetaTable    = (*Http)(nil)
	_ manifests.Defaultable  = (*Http)(nil)
	_ manifests.Prepare      = (*Http)(nil)
	_ manifests.Marshaler    = (*Http)(nil)
	_ manifests.Unmarshaler  = (*Http)(nil)
)

type Http struct {
	kinds.BaseManifest `yaml:",inline" json:",inline"`

	Spec struct {
		Server string     `yaml:"server,omitempty" json:"server,omitempty"`
		Cases  []HttpCase `yaml:"cases" valid:"required,length(1|100)" json:"cases"`
	} `yaml:"spec" json:"spec" valid:"required"`

	kinds.Dependencies `yaml:",inline" json:",inline"`
	Meta               kinds.Meta `yaml:"-" json:"meta"`
}

type HttpCase struct {
	tests.HttpCase `yaml:",inline" json:",inline"`
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

func (h *Http) Default() {
	h.Namespace = manifests.DefaultNamespace
	h.Meta = kinds.DefaultMeta
}

func (h *Http) GetMeta() manifests.Meta {
	return h.Meta
}

func (h *Http) Prepare() {
	if h.Namespace == "" {
		h.Namespace = manifests.DefaultNamespace
	}
}

func (h *Http) MarshalYAML() ([]byte, error) {
	return kinds.BaseMarshalYAML(h)
}

func (h *Http) MarshalJSON() ([]byte, error) {
	return kinds.BaseMarshalJSON(h)
}

func (h *Http) UnmarshalYAML(bytes []byte) error {
	return kinds.BaseUnmarshalYAML(bytes, h)
}

func (h *Http) UnmarshalJSON(bytes []byte) error {
	return kinds.BaseUnmarshalJSON(bytes, h)
}
