package api

import (
	"fmt"
	"time"

	"github.com/apiqube/cli/internal/core/manifests/index"

	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
)

var (
	_ manifests.Manifest     = (*Http)(nil)
	_ manifests.Dependencies = (*Http)(nil)
	_ manifests.Defaultable  = (*Http)(nil)
	_ manifests.Prepare      = (*Http)(nil)
)

type Http struct {
	kinds.BaseManifest `yaml:",inline" json:",inline"`

	Spec struct {
		Server string     `yaml:"server,omitempty" json:"server,omitempty"`
		Cases  []HttpCase `yaml:"cases" valid:"required,length(1|100)" json:"cases"`
	} `yaml:"spec" json:"spec" valid:"required"`

	kinds.Dependencies `yaml:",inline" json:",inline"`
	Meta               *kinds.Meta `yaml:"-" json:"meta"`
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

func (h *Http) Index() any {
	return map[string]any{
		index.Version:   float64(h.Version),
		index.Kind:      h.Kind,
		index.Name:      h.Name,
		index.Namespace: h.Namespace,
		index.DependsOn: h.DependsOn,

		index.MetaHash:        h.Meta.Hash,
		index.MetaVersion:     float64(h.Meta.Version),
		index.MetaIsCurrent:   h.Meta.IsCurrent,
		index.MetaCreatedAt:   h.Meta.CreatedAt.Format(time.RFC3339Nano),
		index.MetaCreatedBy:   h.Meta.CreatedBy,
		index.MetaUpdatedAt:   h.Meta.UpdatedAt.Format(time.RFC3339Nano),
		index.MetaUpdatedBy:   h.Meta.UpdatedBy,
		index.MetaUsedBy:      h.Meta.UsedBy,
		index.MetaLastApplied: h.Meta.LastApplied.Format(time.RFC3339Nano),
	}
}

func (h *Http) GetDependsOn() []string {
	return h.DependsOn
}

func (h *Http) Default() {
	if h.Namespace == "" {
		h.Namespace = manifests.DefaultNamespace
	}

	if h.Meta == nil {
		h.Meta = kinds.DefaultMeta()
	}
}

func (h *Http) GetMeta() manifests.Meta {
	return h.Meta
}

func (h *Http) Prepare() {
	if h.Namespace == "" {
		h.Namespace = manifests.DefaultNamespace
	}

	if h.Meta == nil {
		h.Meta = kinds.DefaultMeta()
	}
}
