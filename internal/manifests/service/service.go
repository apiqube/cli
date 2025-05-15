package service

import "github.com/apiqube/cli/internal/manifests"

var _ manifests.Manifest = (*Service)(nil)
var _ manifests.Defaultable[*Service] = (*Service)(nil)

type Service struct {
	manifests.BaseManifest `yaml:",inline"`

	Spec struct {
		Containers []Container `yaml:"containers" valid:"required,length(1|50)"`
	} `yaml:"spec" valid:"required"`

	DependsOn []string `yaml:"dependsOn,omitempty"`
}

type Container struct {
	Name          string            `yaml:"name" valid:"required"`
	ContainerName string            `yaml:"container_name,omitempty"`
	Dockerfile    string            `yaml:"dockerfile,omitempty"`
	Image         string            `yaml:"image,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Env           map[string]string `yaml:"env,omitempty"`
	Command       string            `yaml:"command,omitempty"`
	Depends       *ContainerDepend  `yaml:"depends,omitempty"`
	Replicas      int               `yaml:"replicas,omitempty" valid:"length(0|25)"`
	HealthPath    string            `yaml:"health_path,omitempty"`
}

type ContainerDepend struct {
	Depends []string `yaml:"depends,omitempty" valid:"required,length(1|25)"`
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
