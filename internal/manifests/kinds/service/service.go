package service

import (
	"fmt"
	"github.com/apiqube/cli/internal/manifests"
	"github.com/apiqube/cli/internal/manifests/kinds"
)

var _ manifests.Manifest = (*Service)(nil)
var _ manifests.Defaultable[*Service] = (*Service)(nil)

type Service struct {
	kinds.BaseManifest `yaml:",inline"`

	Spec struct {
		Containers []Container `yaml:"containers" valid:"required,length(1|50)"`
	} `yaml:"spec" valid:"required"`
}

type Container struct {
	Name          string            `yaml:"name" valid:"required"`
	ContainerName string            `yaml:"containerName,omitempty"`
	Dockerfile    string            `yaml:"dockerfile,omitempty"`
	Image         string            `yaml:"image,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Env           map[string]string `yaml:"env,omitempty"`
	Command       string            `yaml:"command,omitempty"`
	Depends       *ContainerDepend  `yaml:"depends,omitempty"`
	Replicas      int               `yaml:"replicas,omitempty" valid:"length(0|25)"`
	HealthPath    string            `yaml:"healthPath,omitempty"`
}

type ContainerDepend struct {
	Depends []string `yaml:"depends,omitempty" valid:"required,length(1|25)"`
}

func (s *Service) GetID() string {
	return fmt.Sprintf("%s.%s.%s", s.Namespace, s.Kind, s.Name)
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

func (s *Service) Default() *Service {
	s.Namespace = manifests.DefaultNamespace

	return s
}
