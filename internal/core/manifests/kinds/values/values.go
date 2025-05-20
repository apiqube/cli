package values

import (
	"time"

	"github.com/apiqube/cli/internal/core/manifests"
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
		Data map[string]any `yaml:",inline" json:",inline"`
	} `yaml:"spec" valid:"required"`

	Meta *kinds.Meta `yaml:"-" json:"meta"`
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
		kinds.ID:        v.GetID(),
		kinds.Version:   float64(v.Version),
		kinds.Kind:      v.Kind,
		kinds.Name:      v.Name,
		kinds.Namespace: v.Namespace,

		kinds.MetaHash:        v.Meta.Hash,
		kinds.MetaVersion:     float64(v.Meta.Version),
		kinds.MetaIsCurrent:   v.Meta.IsCurrent,
		kinds.MetaCreatedAt:   v.Meta.CreatedAt.Format(time.RFC3339Nano),
		kinds.MetaCreatedBy:   v.Meta.CreatedBy,
		kinds.MetaUpdatedAt:   v.Meta.UpdatedAt.Format(time.RFC3339Nano),
		kinds.MetaUpdatedBy:   v.Meta.UpdatedBy,
		kinds.MetaUsedBy:      v.Meta.UsedBy,
		kinds.MetaLastApplied: v.Meta.LastApplied.Format(time.RFC3339Nano),
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
