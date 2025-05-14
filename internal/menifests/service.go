package menifests

type Service struct {
	Kind     string   `yaml:"kind"`
	Metadata Metadata `yaml:"metadata"`
	Spec     struct {
		Containers []Container `yaml:"containers"`
	} `yaml:"spec,flow"`
}

type Container struct {
	Name          string            `yaml:"name"`
	ContainerName string            `yaml:"container_name"`
	Dockerfile    string            `yaml:"dockerfile"`
	Image         string            `yaml:"image"`
	Ports         []string          `yaml:"ports,omitempty"`
	Env           map[string]string `yaml:"env,omitempty"`
	Command       string            `yaml:"command"`
	Depends       *ContainerDepend  `yaml:"depends,omitempty"`
	Replicas      int               `yaml:"replicas"`
	HealthPath    string            `yaml:"health_path"`
}

type ContainerDepend struct {
	Depends []string `yaml:"depends,omitempty"`
}
