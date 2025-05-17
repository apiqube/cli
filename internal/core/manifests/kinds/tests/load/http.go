package load

import (
	"time"

	"github.com/apiqube/cli/internal/core/manifests/index"

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
	kinds.BaseManifest `yaml:",inline" json:",inline"`

	Spec struct {
		Server string     `yaml:"server,omitempty" json:"server,omitempty"`
		Cases  []HttpCase `yaml:"cases" json:"cases" valid:"required,length(1|100)"`
	} `yaml:"spec" json:"spec" valid:"required"`

	kinds.Dependencies `yaml:",inline" json:",inline"`
	Meta               *kinds.Meta `yaml:"-" json:"meta"`
}

type HttpCase struct {
	tests.HttpCase `yaml:",inline" json:",inline" valid:"in(constant|ramp|wave|step)"`

	Type      string        `yaml:"type,omitempty" json:"type,omitempty"`
	Repeats   int           `yaml:"repeats,omitempty" json:"repeats,omitempty"`
	Agents    int           `yaml:"agents,omitempty" json:"agents,omitempty"`
	RPS       int           `yaml:"rps,omitempty" json:"rps,omitempty"`
	Ramp      *RampConfig   `yaml:"ramp,omitempty" json:"ramp,omitempty"`
	Wave      *WaveConfig   `yaml:"wave,omitempty" json:"wave,omitempty"`
	Step      *StepConfig   `yaml:"step,omitempty" json:"step,omitempty"`
	Duration  time.Duration `yaml:"duration,omitempty" json:"duration,omitempty"`
	SaveEvery int           `yaml:"saveEvery,omitempty" json:"saveEvery,omitempty"`
}

type WaveConfig struct {
	Low   int `yaml:"low,omitempty" json:"low,omitempty"`
	High  int `yaml:"high,omitempty" json:"high,omitempty"`
	Delta int `yaml:"delta,omitempty" json:"delta,omitempty"`
}

type RampConfig struct {
	Start int `yaml:"start,omitempty" json:"start,omitempty"`
	End   int `yaml:"end,omitempty" json:"end,omitempty"`
}

type StepConfig struct {
	Pause time.Duration `yaml:"pause,omitempty" json:"pause,omitempty"`
}

func (h *Http) GetID() string {
	return kinds.FormManifestID(h.Namespace, h.Kind, h.Name)
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
		index.ID:        h.GetID(),
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
