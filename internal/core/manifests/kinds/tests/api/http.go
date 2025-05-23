package api

import (
	"fmt"
	"time"

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
	kinds.BaseManifest `yaml:",inline" json:",inline" validate:"required"`

	Spec struct {
		Target string     `yaml:"target,omitempty" json:"target,omitempty" validate:"required"`
		Cases  []HttpCase `yaml:"cases" json:"cases" validate:"required,min=1,max=100,dive"`
	} `yaml:"spec" json:"spec" validate:"required"`

	kinds.Dependencies `yaml:",inline" json:",inline" validate:"omitempty"`
	Meta               *kinds.Meta `yaml:"-" json:"meta"`
}

type HttpCase struct {
	tests.HttpCase `yaml:",inline" json:",inline" validate:"required"`
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
		kinds.Version:   h.Version,
		kinds.Kind:      h.Kind,
		kinds.Name:      h.Name,
		kinds.Namespace: h.Namespace,
		kinds.DependsOn: h.DependsOn,

		kinds.MetaHash:        h.Meta.Hash,
		kinds.MetaVersion:     float64(h.Meta.Version),
		kinds.MetaIsCurrent:   h.Meta.IsCurrent,
		kinds.MetaCreatedAt:   h.Meta.CreatedAt.Format(time.RFC3339Nano),
		kinds.MetaCreatedBy:   h.Meta.CreatedBy,
		kinds.MetaUpdatedAt:   h.Meta.UpdatedAt.Format(time.RFC3339Nano),
		kinds.MetaUpdatedBy:   h.Meta.UpdatedBy,
		kinds.MetaUsedBy:      h.Meta.UsedBy,
		kinds.MetaLastApplied: h.Meta.LastApplied.Format(time.RFC3339Nano),
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
