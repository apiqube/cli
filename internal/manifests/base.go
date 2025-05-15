package manifests

type BaseManifest struct {
	Kind      string `yaml:"kind" valid:"required,alpha,in(Server|Service|TestHttp|TestLoad)"`
	Metadata  `yaml:"metadata,inline"`
	DependsOn []string `yaml:"dependsOn,omitempty"`
}
