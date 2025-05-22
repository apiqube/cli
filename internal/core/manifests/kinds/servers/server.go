package servers

import (
	"time"

	"github.com/apiqube/cli/internal/core/manifests/utils"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
)

var (
	_ manifests.Manifest    = (*Server)(nil)
	_ manifests.Defaultable = (*Server)(nil)
	_ manifests.Prepare     = (*Server)(nil)
)

type Server struct {
	kinds.BaseManifest `yaml:",inline" json:",inline" validate:"required"`

	Spec struct {
		BaseURL string            `yaml:"baseUrl" json:"baseUrl" validate:"required,url"`
		Health  string            `yaml:"health" json:"health" validate:"omitempty,max=100"`
		Headers map[string]string `yaml:"headers,omitempty" json:"headers"`
	} `yaml:"spec" json:"spec" validate:"required"`

	Meta *kinds.Meta `yaml:"-" json:"meta"`
}

func (s *Server) GetID() string {
	return utils.FormManifestID(s.Namespace, s.Kind, s.Name)
}

func (s *Server) GetKind() string {
	return s.Kind
}

func (s *Server) GetName() string {
	return s.Name
}

func (s *Server) GetNamespace() string {
	return s.Namespace
}

func (s *Server) Index() any {
	return map[string]any{
		kinds.ID:        s.GetID(),
		kinds.Version:   s.Version,
		kinds.Kind:      s.Kind,
		kinds.Name:      s.Name,
		kinds.Namespace: s.Namespace,

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

func (s *Server) GetMeta() manifests.Meta {
	return s.Meta
}

func (s *Server) Default() {
	if s.Namespace == "" {
		s.Namespace = manifests.DefaultNamespace
	}

	if s.Meta == nil {
		s.Meta = kinds.DefaultMeta()
	}
}

func (s *Server) Prepare() {
	if s.Namespace == "" {
		s.Namespace = manifests.DefaultNamespace
	}

	if s.Meta == nil {
		s.Meta = kinds.DefaultMeta()
	}
}
