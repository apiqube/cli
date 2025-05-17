package servers

import (
	"time"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/index"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
)

var (
	_ manifests.Manifest    = (*Server)(nil)
	_ manifests.MetaTable   = (*Server)(nil)
	_ manifests.Defaultable = (*Server)(nil)
	_ manifests.Prepare     = (*Server)(nil)
	_ manifests.Marshaler   = (*Server)(nil)
	_ manifests.Unmarshaler = (*Server)(nil)
)

type Server struct {
	kinds.BaseManifest `yaml:",inline" json:",inline"`

	Spec struct {
		BaseUrl string            `yaml:"baseUrl" json:"baseUrl" valid:"required,url"`
		Headers map[string]string `yaml:"headers,omitempty" json:"headers" valid:"-"`
	} `yaml:"spec" json:"spec" valid:"required"`

	Meta *kinds.Meta `yaml:"-" json:"meta"`
}

func (s *Server) GetID() string {
	return kinds.FormManifestID(s.Namespace, s.Kind, s.Name)
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
		index.Version:   float64(s.Version),
		index.Kind:      s.Kind,
		index.Name:      s.Name,
		index.Namespace: s.Namespace,

		index.MetaHash:        s.Meta.Hash,
		index.MetaCreatedAt:   s.Meta.CreatedAt.Format(time.RFC3339Nano),
		index.MetaCreatedBy:   s.Meta.CreatedBy,
		index.MetaUpdatedAt:   s.Meta.UpdatedAt.Format(time.RFC3339Nano),
		index.MetaUpdatedBy:   s.Meta.UpdatedBy,
		index.MetaUsedBy:      s.Meta.UsedBy,
		index.MetaLastApplied: s.Meta.LastApplied.Format(time.RFC3339Nano),
	}
}

func (s *Server) GetMeta() manifests.Meta {
	return s.Meta
}

func (s *Server) Default() {
	s.Namespace = manifests.DefaultNamespace
	s.Meta = kinds.DefaultMeta
}

func (s *Server) Prepare() {
	if s.Namespace == "" {
		s.Namespace = manifests.DefaultNamespace
	}
}

func (s *Server) MarshalYAML() ([]byte, error) {
	return kinds.BaseMarshalYAML(s)
}

func (s *Server) MarshalJSON() ([]byte, error) {
	return kinds.BaseMarshalJSON(s)
}

func (s *Server) UnmarshalYAML(bytes []byte) error {
	return kinds.BaseUnmarshalYAML(bytes, s)
}

func (s *Server) UnmarshalJSON(bytes []byte) error {
	return kinds.BaseUnmarshalJSON(bytes, s)
}
