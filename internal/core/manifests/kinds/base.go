package kinds

import "time"

type Metadata struct {
	Name      string `yaml:"name" json:"name" valid:"required,alpha"`
	Namespace string `yaml:"namespace" json:"namespace" valid:"required,alpha"`
}

type BaseManifest struct {
	Version   uint8  `yaml:"version" json:"version" valid:"required,numeric"`
	Kind      string `yaml:"kind" json:"kind" valid:"required,alpha,in(Server|Service|HttpTest|HttpLoadTest)"`
	Metadata  `yaml:"metadata" json:"metadata"`
	DependsOn []string `yaml:"dependsOn,omitempty" json:"dependsOn,omitempty"`
}

type Meta struct {
	Hash        string    `yaml:"-" json:"hash"`
	Version     uint8     `yaml:"-" json:"version"`
	CreatedAt   time.Time `yaml:"-" json:"createdAt"`
	CreatedBy   string    `yaml:"-" json:"createdBy"`
	UpdatedAt   time.Time `yaml:"-" json:"updatedAt"`
	UpdatedBy   string    `yaml:"-" json:"updatedBy"`
	UsedBy      string    `yaml:"-" json:"usedBy"`
	LastApplied time.Time `yaml:"-" json:"lastApplied"`
}
