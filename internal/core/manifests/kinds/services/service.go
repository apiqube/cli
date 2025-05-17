package services

import (
	"time"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/index"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
)

var (
	_ manifests.Manifest     = (*Service)(nil)
	_ manifests.Dependencies = (*Service)(nil)
	_ manifests.MetaTable    = (*Service)(nil)
	_ manifests.Defaultable  = (*Service)(nil)
	_ manifests.Prepare      = (*Service)(nil)
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

func (s *Service) Index() any {
	return map[string]any{
		index.Version:   float64(s.Version),
		index.Kind:      s.Kind,
		index.Name:      s.Name,
		index.Namespace: s.Namespace,
		index.DependsOn: s.DependsOn,

		index.MetaHash:        s.Meta.Hash,
		index.MetaCreatedAt:   s.Meta.CreatedAt.Format(time.RFC3339Nano),
		index.MetaCreatedBy:   s.Meta.CreatedBy,
		index.MetaUpdatedAt:   s.Meta.UpdatedAt.Format(time.RFC3339Nano),
		index.MetaUpdatedBy:   s.Meta.UpdatedBy,
		index.MetaUsedBy:      s.Meta.UsedBy,
		index.MetaLastApplied: s.Meta.LastApplied.Format(time.RFC3339Nano),
	}
}

func (s *Service) GetMeta() manifests.Meta {
	return s.Meta
}

func (s *Service) Default() {
	if s.Namespace == "" {
		s.Namespace = manifests.DefaultNamespace
	}

	if s.Meta == nil {
		s.Meta = kinds.DefaultMeta()
	}
}

func (s *Service) Prepare() {
	if s.Namespace == "" {
		s.Namespace = manifests.DefaultNamespace
	}
}
