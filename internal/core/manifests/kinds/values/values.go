package values

import (
	"time"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/index"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
)

var (
	_ manifests.Manifest    = (*Values)(nil)
	_ manifests.Defaultable = (*Values)(nil)
	_ manifests.Prepare     = (*Values)(nil)
)

type Values struct {
	kinds.BaseManifest `yaml:",inline" json:",inline"`

	Spec struct {
		Content `yaml:",inline" json:",inline"`
	} `yaml:"spec" valid:"required"`

	Meta *kinds.Meta `yaml:"-" json:"meta"`
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

func (v *Values) Index() any {
	return map[string]any{
		index.ID:        v.GetID(),
		index.Version:   float64(v.Version),
		index.Kind:      v.Kind,
		index.Name:      v.Name,
		index.Namespace: v.Namespace,

		index.MetaHash:        v.Meta.Hash,
		index.MetaVersion:     float64(v.Meta.Version),
		index.MetaIsCurrent:   v.Meta.IsCurrent,
		index.MetaCreatedAt:   v.Meta.CreatedAt.Format(time.RFC3339Nano),
		index.MetaCreatedBy:   v.Meta.CreatedBy,
		index.MetaUpdatedAt:   v.Meta.UpdatedAt.Format(time.RFC3339Nano),
		index.MetaUpdatedBy:   v.Meta.UpdatedBy,
		index.MetaUsedBy:      v.Meta.UsedBy,
		index.MetaLastApplied: v.Meta.LastApplied.Format(time.RFC3339Nano),
	}
}

func (v *Values) GetMeta() manifests.Meta {
	return v.Meta
}

func (v *Values) Default() {
	if v.Namespace == "" {
		v.Namespace = manifests.DefaultNamespace
	}

	if v.Meta == nil {
		v.Meta = kinds.DefaultMeta()
	}
}

func (v *Values) Prepare() {
	if v.Namespace == "" {
		v.Namespace = manifests.DefaultNamespace
	}
}
