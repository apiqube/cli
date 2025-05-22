package plan

import (
	"fmt"
	"github.com/apiqube/cli/internal/core/runner/hooks"
	"time"

	"github.com/apiqube/cli/internal/core/manifests/utils"
	"github.com/google/uuid"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
)

var (
	_ manifests.Manifest    = (*Plan)(nil)
	_ manifests.Defaultable = (*Plan)(nil)
	_ manifests.Prepare     = (*Plan)(nil)
)

type Plan struct {
	kinds.BaseManifest `yaml:",inline" json:",inline" validate:"required"`

	Spec struct {
		Stages []Stage `yaml:"stages" json:"stages" validate:"required,min=1,dive"`
		Hooks  *Hooks  `yaml:"hooks,omitempty" json:"hooks,omitempty" validate:"omitempty,dive"`
	} `yaml:"spec" json:"spec" validate:"required"`

	Meta *kinds.Meta `yaml:"-" json:"meta"`
}

type Stage struct {
	Name        string         `yaml:"name" json:"name" validate:"required,min=3,max=50"`
	Description string         `yaml:"description,omitempty" json:"description,omitempty" validate:"omitempty,max=255"`
	Manifests   []string       `yaml:"manifests" json:"manifests" validate:"required,min=1,dive"`
	Parallel    bool           `yaml:"parallel,omitempty" json:"parallel,omitempty"`
	Params      map[string]any `yaml:"params,omitempty" json:"params,omitempty" validate:"omitempty"`
	Mode        string         `yaml:"mode,omitempty" json:"mode,omitempty" validate:"omitempty,oneof=strict parallel"` // (strict|parallel)
	Hooks       *Hooks         `yaml:"hooks,omitempty" json:"hooks,omitempty" validate:"omitempty,dive"`
}

type Hooks struct {
	BeforeRun []hooks.Action `yaml:"beforeRun,omitempty" json:"beforeRun,omitempty" validate:"omitempty,dive"`
	AfterRun  []hooks.Action `yaml:"afterRun,omitempty" json:"afterRun,omitempty" validate:"omitempty,dive"`
	OnSuccess []hooks.Action `yaml:"onSuccess,omitempty" json:"onSuccess,omitempty" validate:"omitempty,dive"`
	OnFailure []hooks.Action `yaml:"onFailure,omitempty" json:"onFailure,omitempty" validate:"omitempty,dive"`
}

func (p *Plan) GetID() string {
	return utils.FormManifestID(p.Namespace, p.Kind, p.Name)
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
		kinds.ID:        p.GetID(),
		kinds.Version:   p.Version,
		kinds.Kind:      p.Kind,
		kinds.Name:      p.Name,
		kinds.Namespace: p.Namespace,

		kinds.MetaHash:        p.Meta.Hash,
		kinds.MetaVersion:     float64(p.Meta.Version),
		kinds.MetaIsCurrent:   p.Meta.IsCurrent,
		kinds.MetaCreatedAt:   p.Meta.CreatedAt.Format(time.RFC3339Nano),
		kinds.MetaCreatedBy:   p.Meta.CreatedBy,
		kinds.MetaUpdatedAt:   p.Meta.UpdatedAt.Format(time.RFC3339Nano),
		kinds.MetaUpdatedBy:   p.Meta.UpdatedBy,
		kinds.MetaUsedBy:      p.Meta.UsedBy,
		kinds.MetaLastApplied: p.Meta.LastApplied.Format(time.RFC3339Nano),
	}
}

func (p *Plan) GetMeta() manifests.Meta {
	return p.Meta
}

func (p *Plan) Default() {
	if p.Version != manifests.V1Version {
		p.Version = manifests.V1Version
	}

	if p.Name == "" {
		p.Name = fmt.Sprintf("%s-%s", "generated", uuid.NewString()[:8])
	}

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
			stage.Mode = Strict.String()
		}

		for j, m := range stage.Manifests {
			namespace, kind, name := utils.ParseManifestID(m)
			stage.Manifests[j] = utils.FormManifestID(namespace, kind, name)
		}

		p.Spec.Stages[i] = stage
	}
}

func (p *Plan) GetAllManifests() []string {
	var results []string

	for _, stage := range p.Spec.Stages {
		results = append(results, stage.Manifests...)
	}

	return results
}
