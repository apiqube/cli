package servers

import (
	"github.com/apiqube/cli/internal/core/manifests"
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
		"version":   s.Version,
		"kind":      s.Kind,
		"name":      s.Name,
		"namespace": s.Namespace,

		"hash":        s.Meta.Hash,
		"createdAt":   s.Meta.CreatedAt,
		"createdBy":   s.Meta.CreatedBy,
		"updatedAt":   s.Meta.UpdatedAt,
		"updatedBy":   s.Meta.UpdatedBy,
		"userBy":      s.Meta.UsedBy,
		"lastApplied": s.Meta.LastApplied,
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
