package plan

import (
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
)

var (
	_ manifests.Manifest    = (*Plan)(nil)
	_ manifests.MetaTable   = (*Plan)(nil)
	_ manifests.Defaultable = (*Plan)(nil)
	_ manifests.Prepare     = (*Plan)(nil)
	_ manifests.Marshaler   = (*Plan)(nil)
	_ manifests.Unmarshaler = (*Plan)(nil)
)

type Plan struct {
	kinds.BaseManifest `yaml:",inline" json:",inline"`

	Spec struct {
		Stages Stages `yaml:"stages" json:"stages"`
		Hooks  Hooks  `yaml:"hooks" json:"hooks"`
	} `yaml:"spec" json:"spec"`

	Meta kinds.Meta `yaml:"meta" json:"meta"`
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

func (p *Plan) GetMeta() manifests.Meta {
	return p.Meta
}

func (p *Plan) Default() {
	p.Namespace = manifests.DefaultNamespace
	p.Kind = manifests.PlanManifestKind
	p.Meta = kinds.DefaultMeta
}

func (p *Plan) Prepare() {
	if p.Namespace == "" {
		p.Namespace = manifests.DefaultNamespace
	}
}

func (p *Plan) MarshalYAML() ([]byte, error) {
	return kinds.BaseMarshalYAML(p)
}

func (p *Plan) MarshalJSON() ([]byte, error) {
	return kinds.BaseMarshalJSON(p)
}

func (p *Plan) UnmarshalYAML(bytes []byte) error {
	return kinds.BaseUnmarshalYAML(bytes, p)
}

func (p *Plan) UnmarshalJSON(bytes []byte) error {
	return kinds.BaseUnmarshalJSON(bytes, p)
}
