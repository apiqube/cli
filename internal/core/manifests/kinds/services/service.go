package services

import (
	"time"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
)

var (
	_ manifests.Manifest     = (*Service)(nil)
	_ manifests.Dependencies = (*Service)(nil)
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
		kinds.ID:        s.GetID(),
		kinds.Version:   float64(s.Version),
		kinds.Kind:      s.Kind,
		kinds.Name:      s.Name,
		kinds.Namespace: s.Namespace,
		kinds.DependsOn: s.DependsOn,

		kinds.MetaHash:        s.Meta.Hash,
		kinds.MetaVersion:     float64(s.Meta.Version),
		kinds.MetaIsCurrent:   s.Meta.IsCurrent,
		kinds.MetaCreatedAt:   s.Meta.CreatedAt.Format(time.RFC3339Nano),
		kinds.MetaCreatedBy:   s.Meta.CreatedBy,
		kinds.MetaUpdatedAt:   s.Meta.UpdatedAt.Format(time.RFC3339Nano),
		kinds.MetaUpdatedBy:   s.Meta.UpdatedBy,
		kinds.MetaUsedBy:      s.Meta.UsedBy,
		kinds.MetaLastApplied: s.Meta.LastApplied.Format(time.RFC3339Nano),
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
