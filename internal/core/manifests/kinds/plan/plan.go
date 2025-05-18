package plan

import (
	"github.com/apiqube/cli/internal/core/manifests/utils"
	"strings"
	"time"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/index"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
)

var (
	_ manifests.Manifest    = (*Plan)(nil)
	_ manifests.Defaultable = (*Plan)(nil)
	_ manifests.Prepare     = (*Plan)(nil)
)

type Plan struct {
	kinds.BaseManifest `yaml:",inline" json:",inline"`

	Spec struct {
		Stages []Stage `yaml:"stages" json:"stages"`
		Hooks  Hooks   `yaml:"hooks,omitempty" json:"hooks,omitempty"`
	} `yaml:"spec" json:"spec"`

	Meta *kinds.Meta `yaml:",inline" json:"meta"`
}

type Stage struct {
	Name        string         `yaml:"name" json:"name"`
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Manifests   []string       `yaml:"manifests" json:"manifests"`
	Parallel    bool           `yaml:"parallel,omitempty" json:"parallel,omitempty"`
	Params      map[string]any `yaml:"params,omitempty" json:"params,omitempty"`
	Mode        string         `yaml:"mode,omitempty" json:"mode,omitempty"` // eg parallel mode, (strict|lite)
	Hooks       Hooks          `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

type Hooks struct {
	BeforeStart []Action `yaml:"beforeStart,omitempty" json:"beforeStart,omitempty"`
	AfterFinish []Action `yaml:"afterFinish,omitempty" json:"afterFinish,omitempty"`
	OnSuccess   []Action `yaml:"onSuccess,omitempty" json:"onSuccess,omitempty"`
	OnFailure   []Action `yaml:"onFailure,omitempty" json:"onFailure,omitempty"`
}

type Action struct {
	Type   string         `yaml:"type" json:"type"` // eg log/save/skip/fail/exec/notify
	Params map[string]any `yaml:"params" json:"params"`
}

func (p *Plan) GetID() string {
	return kinds.FormManifestID(p.Namespace, p.Kind, p.Name)
}

func (p *Plan) GetKind() string {
	return p.Kind
}

func (p *Plan) GetName() string {
	return p.Name
}

func (p *Plan) GetNamespace() string {
	return p.Namespace
}

func (p *Plan) Index() any {
	return map[string]any{
		index.ID:        p.GetID(),
		index.Version:   float64(p.Version),
		index.Kind:      p.Kind,
		index.Name:      p.Name,
		index.Namespace: p.Namespace,

		index.MetaHash:        p.Meta.Hash,
		index.MetaVersion:     float64(p.Meta.Version),
		index.MetaIsCurrent:   p.Meta.IsCurrent,
		index.MetaCreatedAt:   p.Meta.CreatedAt.Format(time.RFC3339Nano),
		index.MetaCreatedBy:   p.Meta.CreatedBy,
		index.MetaUpdatedAt:   p.Meta.UpdatedAt.Format(time.RFC3339Nano),
		index.MetaUpdatedBy:   p.Meta.UpdatedBy,
		index.MetaUsedBy:      p.Meta.UsedBy,
		index.MetaLastApplied: p.Meta.LastApplied.Format(time.RFC3339Nano),
	}
}

func (p *Plan) GetMeta() manifests.Meta {
	return p.Meta
}

func (p *Plan) Default() {
	if p.Kind == "" {
		p.Kind = manifests.PlanManifestKind
	}

	if p.Namespace == "" {
		p.Namespace = manifests.DefaultNamespace
	}

	if p.Meta == nil {
		p.Meta = kinds.DefaultMeta()
	}
}

func (p *Plan) Prepare() {
	if p.Namespace == "" {
		p.Namespace = manifests.DefaultNamespace
	}

	if p.Kind == "" {
		p.Kind = manifests.PlanManifestKind
	}

	if p.Meta == nil {
		p.Meta = kinds.DefaultMeta()
	}

	for i, stage := range p.Spec.Stages {
		if stage.Mode == "" {
			stage.Mode = "lite"
		}

		for j, m := range stage.Manifests {
			namespace, kind, name := utils.ParseManifestID(m)
			m = strings.Join([]string{namespace, kind, name}, ".")
			stage.Manifests[j] = m
		}

		p.Spec.Stages[i] = stage
	}
}
