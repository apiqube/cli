package kinds

type Metadata struct {
	Name      string `yaml:"name" json:"name" valid:"required,alpha"`
	Namespace string `yaml:"namespace" json:"namespace" valid:"required,alpha"`
}

type BaseManifest struct {
	Version  uint8  `yaml:"version" json:"version" valid:"required,numeric"`
	Kind     string `yaml:"kind" json:"kind" valid:"required,alpha,in(Server|Service|HttpTest|HttpLoadTest)"`
	Metadata `yaml:"metadata" json:"metadata"`
}

type Dependencies struct {
	DependsOn []string `yaml:"dependsOn" json:"dependsOn"`
}
