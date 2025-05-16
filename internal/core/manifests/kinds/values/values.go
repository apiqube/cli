package values

import (
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
)

var (
	_ manifests.Manifest    = (*Values)(nil)
	_ manifests.Defaultable = (*Values)(nil)
	_ manifests.Marshaler   = (*Values)(nil)
	_ manifests.Unmarshaler = (*Values)(nil)
	_ manifests.MetaTable   = (*Values)(nil)
	_ manifests.Prepare     = (*Values)(nil)
)

type Values struct {
	kinds.BaseManifest `yaml:",inline" json:",inline"`

	Spec struct {
		Content `yaml:",inline" json:",inline"`
	} `yaml:"spec" valid:"required"`

	Meta kinds.Meta `yaml:"-" json:"meta"`
}

type Content struct {
	Values map[string]any `yaml:",inline" json:",inline"`
}

func (v *Values) GetID() string {
	return kinds.FormManifestID(v.Namespace, v.Kind, v.Name)
}

func (v *Values) GetKind() string {
	return v.Kind
}

func (v *Values) GetName() string {
	return v.Name
}

func (v *Values) GetNamespace() string {
	return v.Namespace
}

func (v *Values) GetMeta() manifests.Meta {
	return v.Meta
}

func (v *Values) Default() {
	v.Namespace = manifests.DefaultNamespace
	v.Meta = kinds.DefaultMeta
}

func (v *Values) Prepare() {
	if v.Namespace == "" {
		v.Namespace = manifests.DefaultNamespace
	}
}

func (v *Values) MarshalYAML() ([]byte, error) {
	return kinds.BaseMarshalYAML(v)
}

func (v *Values) MarshalJSON() ([]byte, error) {
	return kinds.BaseMarshalJSON(v)
}

func (v *Values) UnmarshalYAML(bytes []byte) error {
	return kinds.BaseUnmarshalYAML(bytes, v)
}

func (v *Values) UnmarshalJSON(bytes []byte) error {
	return kinds.BaseUnmarshalJSON(bytes, v)
}
