package server

import (
	"fmt"

	"github.com/apiqube/cli/internal/manifests"
	"github.com/apiqube/cli/internal/manifests/kinds"
)

var (
	_ manifests.Manifest             = (*Server)(nil)
	_ manifests.Defaultable[*Server] = (*Server)(nil)
)

type Server struct {
	kinds.BaseManifest `yaml:",inline"`

	Spec struct {
		BaseUrl string            `yaml:"baseUrl" valid:"required,url"`
		Headers map[string]string `yaml:"headers,omitempty"`
	} `yaml:"spec" valid:"required"`
}

func (s *Server) GetID() string {
	return fmt.Sprintf("%s.%s.%s", s.Namespace, s.Kind, s.Name)
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

func (s *Server) GetDependsOn() []string {
	return s.DependsOn
}

func (s *Server) Default() *Server {
	s.Namespace = manifests.DefaultNamespace
	s.Spec.Headers = map[string]string{
		"Content-Type": "application/json",
	}

	return s
}
