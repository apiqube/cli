package load

import (
	"time"

	"github.com/apiqube/cli/internal/core/manifests/utils"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
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

	Type      string        `yaml:"type,omitempty" json:"type,omitempty" validate:"omitempty,oneof=wave ramp step"`
	Repeats   int           `yaml:"repeats,omitempty" json:"repeats,omitempty" validate:"omitempty,min=1,max=1000"`
	Agents    int           `yaml:"agents,omitempty" json:"agents,omitempty" validate:"omitempty,min=1,max=1000"`
	RPS       int           `yaml:"rps,omitempty" json:"rps,omitempty" validate:"omitempty,min=1,max=100000"`
	Ramp      *RampConfig   `yaml:"ramp,omitempty" json:"ramp,omitempty" validate:"omitempty"`
	Wave      *WaveConfig   `yaml:"wave,omitempty" json:"wave,omitempty" validate:"omitempty"`
	Step      *StepConfig   `yaml:"step,omitempty" json:"step,omitempty" validate:"omitempty"`
	Duration  time.Duration `yaml:"duration,omitempty" json:"duration,omitempty" validate:"omitempty,duration"`
	SaveEvery int           `yaml:"saveEvery,omitempty" json:"saveEvery,omitempty" validate:"omitempty,min=1,max=1000"`
}

type WaveConfig struct {
	Low   int `yaml:"low,omitempty" json:"low,omitempty" validate:"required,min=1,max=1000"`
	High  int `yaml:"high,omitempty" json:"high,omitempty" validate:"required,min=1,max=100000"`
	Delta int `yaml:"delta,omitempty" json:"delta,omitempty" validate:"omitempty,min=1,max=100000"`
}

type RampConfig struct {
	Start int `yaml:"start,omitempty" json:"start,omitempty" validate:"required,min=1,max=1000"`
	End   int `yaml:"end,omitempty" json:"end,omitempty" validate:"required,min=1,max=100000"`
	Delta int `yaml:"delta,omitempty" json:"delta,omitempty" validate:"omitempty,min=1,max=100000"`
}

type StepConfig struct {
	Pause time.Duration `yaml:"pause,omitempty" json:"pause,omitempty" validate:"required,duration"`
}

func (h *Http) GetID() string {
	return utils.FormManifestID(h.Namespace, h.Kind, h.Name)
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
		kinds.ID:        h.GetID(),
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

func (h *Http) GetMeta() manifests.Meta {
	return h.Meta
}

func (h *Http) Default() {
	if h.Namespace == "" {
		h.Namespace = manifests.DefaultNamespace
	}

	if h.Meta == nil {
		h.Meta = kinds.DefaultMeta()
	}
}

func (h *Http) Prepare() {
	if h.Namespace == "" {
		h.Namespace = manifests.DefaultNamespace
	}
}
