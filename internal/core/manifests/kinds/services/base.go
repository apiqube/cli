package services

import "github.com/apiqube/cli/internal/core/manifests/kinds"

type Container struct {
	Name               string            `yaml:"name" valid:"required"`
	ContainerName      string            `yaml:"containerName,omitempty"`
	Dockerfile         string            `yaml:"dockerfile,omitempty"`
	Image              string            `yaml:"image,omitempty"`
	Ports              []string          `yaml:"ports,omitempty"`
	Env                map[string]string `yaml:"env,omitempty"`
	Command            string            `yaml:"command,omitempty"`
	Replicas           int               `yaml:"replicas,omitempty" valid:"length(0|25)"`
	HealthPath         string            `yaml:"healthPath,omitempty"`
	kinds.Dependencies `yaml:",inline,omitempty" json:"dependencies,omitempty"`
}
