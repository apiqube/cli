package manifests

type BaseManifest struct {
	Version   uint8  `yaml:"version" valid:"required,numeric"`
	Kind      string `yaml:"kind" valid:"required,alpha,in(Server|Service|HttpTest|HttpLoadTest)"`
	Metadata  `yaml:"metadata"`
	DependsOn []string `yaml:"dependsOn,omitempty"`
}
