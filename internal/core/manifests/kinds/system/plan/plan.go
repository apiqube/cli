package plan

import (
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
		Stages Stages `yaml:"stages" json:"stages"`
		Hooks  Hooks  `yaml:"hooks,omitempty" json:"hooks"`
	} `yaml:"spec" json:"spec"`

	Meta *kinds.Meta `yaml:",inline" json:"meta"`
}

type Stages struct {
	Stages []Stage `yaml:",inline" json:",inline"`
}

type Stage struct {
	Name      string   `yaml:"name" json:"name"`
	Manifests []string `yaml:"manifests" json:"manifests" `
	Parallel  bool     `yaml:"parallel" json:"parallel"`
}

type Hooks struct {
	OnSuccess []string `yaml:"onSuccess" json:"onSuccess"`
	OnFailure []string `yaml:"onFailure" json:"onFailure"`
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
}
