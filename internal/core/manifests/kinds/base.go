package kinds

type Metadata struct {
	Name      string `yaml:"name" json:"name" validate:"required,min=3"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

type BaseManifest struct {
	Version  string `yaml:"version" json:"version" validate:"required,eq=v1"`
	Kind     string `yaml:"kind" json:"kind" validate:"required,oneof=Plan Values Server Service HttpTest HttpLoadTest"`
	Metadata `yaml:"metadata" json:"metadata"`
}

type Dependencies struct {
	DependsOn []string `yaml:"dependsOn" json:"dependsOn" validate:"omitempty,min=1,max=25"`
}
