package services

import (
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
)

var (
	_ manifests.Manifest     = (*Service)(nil)
	_ manifests.Dependencies = (*Service)(nil)
	_ manifests.MetaTable    = (*Service)(nil)
	_ manifests.Defaultable  = (*Service)(nil)
	_ manifests.Prepare      = (*Service)(nil)
	_ manifests.Marshaler    = (*Service)(nil)
	_ manifests.Unmarshaler  = (*Service)(nil)
)

type Service struct {
	kinds.BaseManifest `yaml:",inline"`

	Spec struct {
		Containers []Container `yaml:"containers" valid:"required,length(1|50)"`
	} `yaml:"spec" valid:"required"`

	kinds.Dependencies `yaml:",inline" json:",inline"`
	Meta               *kinds.Meta `yaml:"-" json:"meta"`
}

func (s *Service) GetID() string {
	return kinds.FormManifestID(s.Namespace, s.Kind, s.Name)
}

func (s *Service) GetKind() string {
	return s.Kind
}

func (s *Service) GetName() string {
	return s.Name
}

func (s *Service) GetNamespace() string {
	return s.Namespace
}

func (s *Service) GetDependsOn() []string {
	return s.DependsOn
}

func (s *Service) GetMeta() manifests.Meta {
	return s.Meta
}

func (s *Service) Default() {
	s.Namespace = manifests.DefaultNamespace
	s.Meta = kinds.DefaultMeta
}

func (s *Service) Prepare() {
	if s.Namespace == "" {
		s.Namespace = manifests.DefaultNamespace
	}
}

func (s *Service) MarshalYAML() ([]byte, error) {
	return kinds.BaseMarshalYAML(s)
}

func (s *Service) MarshalJSON() ([]byte, error) {
	return kinds.BaseMarshalJSON(s)
}

func (s *Service) UnmarshalYAML(bytes []byte) error {
	return kinds.BaseUnmarshalYAML(bytes, s)
}

func (s *Service) UnmarshalJSON(bytes []byte) error {
	return kinds.BaseUnmarshalJSON(bytes, s)
}
