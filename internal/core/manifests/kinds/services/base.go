package services

import "github.com/apiqube/cli/internal/core/manifests/kinds"

type Container struct {
	Name               string            `yaml:"name" json:"name" validate:"required,min=5,max=25"`
	ContainerName      string            `yaml:"containerName,omitempty" json:"containerName" validate:"omitempty,min=3,max=25"`
	Dockerfile         string            `yaml:"dockerfile,omitempty" json:"dockerfile" validate:"excluded_with=Image"`
	Image              string            `yaml:"image,omitempty" json:"image" validate:"excluded_with=Dockerfile"`
	Ports              []string          `yaml:"ports,omitempty" json:"ports"`
	Env                map[string]string `yaml:"env,omitempty" json:"env" validate:"omitempty,max=100,dive"`
	Command            string            `yaml:"command,omitempty" json:"command" validate:"omitempty,min=1,max=25"`
	Replicas           int               `yaml:"replicas,omitempty" json:"replicas" validate:"omitempty,min=1,max=25"`
	HealthPath         string            `yaml:"healthPath,omitempty" json:"healthPath"`
	kinds.Dependencies `yaml:",inline,omitempty" json:"dependencies,omitempty" validate:"omitempty"`
}
