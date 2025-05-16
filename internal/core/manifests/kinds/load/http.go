package load

import (
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"gopkg.in/yaml.v3"
	"math"
	"time"

	"github.com/apiqube/cli/internal/manifests"
	"github.com/apiqube/cli/internal/manifests/kinds"
)

var (
	_ manifests.Manifest           = (*Http)(nil)
	_ manifests.Defaultable[*Http] = (*Http)(nil)
	_ manifests.Marshaler          = (*Http)(nil)
	_ manifests.Unmarshaler        = (*Http)(nil)
	_ manifests.Meta               = (*Http)(nil)
)

type Http struct {
	kinds.BaseManifest `yaml:",inline" json:",inline"`

	Spec struct {
		Server string     `yaml:"server,omitempty" json:"server,omitempty"`
		Cases  []HttpCase `yaml:"cases" json:"cases" valid:"required,length(1|100)"`
	} `yaml:"spec" json:"spec" valid:"required"`

	Meta kinds.Meta `yaml:"-" json:"meta"`
}

type HttpCase struct {
	Name     string                 `yaml:"name" json:"name" valid:"required"`
	Method   string                 `yaml:"method" json:"method" valid:"required,uppercase,in(GET|POST|PUT|DELETE)"`
	Endpoint string                 `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	Url      string                 `yaml:"url,omitempty" json:"url,omitempty"`
	Headers  map[string]string      `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body     map[string]interface{} `yaml:"body,omitempty" json:"body,omitempty"`
	Expected *HttpExpect            `yaml:"expected,omitempty" json:"expected,omitempty"`
	Extract  *HttpExtractRule       `yaml:"extract,omitempty" json:"extract,omitempty"`
	Timeout  time.Duration          `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Async    bool                   `yaml:"async,omitempty" json:"async,omitempty"`
	Repeats  int                    `yaml:"repeats,omitempty" json:"repeats,omitempty"`
}

type HttpExpect struct {
	Code    int                    `yaml:"code" json:"code" valid:"required,range(0|599)"`
	Message string                 `yaml:"message,omitempty" json:"message,omitempty"`
	Data    map[string]interface{} `yaml:"data,omitempty" json:"data,omitempty"`
}

type HttpExtractRule struct {
	Path  string `yaml:"path,omitempty" json:"path,omitempty"`
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
}

func (h *Http) GetID() string {
	return fmt.Sprintf("%s.%s.%s", h.Namespace, h.Kind, h.Name)
}

func (h *Http) GetKind() string {
	return h.Kind
}

func (h *Http) GetName() string {
	return h.Name
}

func (h *Http) GetNamespace() string {
	return h.Namespace
}

func (h *Http) GetDependsOn() []string {
	return h.DependsOn
}

func (h *Http) Default() *Http {
	h.Namespace = manifests.DefaultNamespace

	return h
}

func (h *Http) MarshalJSON() ([]byte, error) {
	return json.Marshal(h)
}

func (h *Http) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, h)
}

func (h *Http) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(h)
}

func (h *Http) UnmarshalYAML(bytes []byte) error {
	return yaml.Unmarshal(bytes, h)
}

func (h *Http) GetHash() string {
	return h.Meta.Hash
}

func (h *Http) SetHash(hash string) {
	h.Meta.Hash = hash
}

func (h *Http) GetVersion() uint8 {
	return h.Meta.Version
}

func (h *Http) SetVersion(version uint8) {
	h.Meta.Version = version
}

func (h *Http) IncVersion() {
	if h.Meta.Version < math.MaxUint8 {
		h.Meta.Version++
	}
}

func (h *Http) GetCreatedAt() time.Time {
	return h.Meta.CreatedAt
}

func (h *Http) SetCreatedAt(createdAt time.Time) {
	h.Meta.CreatedAt = createdAt
}

func (h *Http) GetCreatedBy() string {
	return h.Meta.CreatedBy
}

func (h *Http) SetCreatedBy(createdBy string) {
	h.Meta.CreatedBy = createdBy
}

func (h *Http) GetUpdatedAt() time.Time {
	return h.Meta.UpdatedAt
}

func (h *Http) SetUpdatedAt(updatedAt time.Time) {
	h.Meta.UpdatedAt = updatedAt
}

func (h *Http) GetUpdatedBy() string {
	return h.Meta.CreatedBy
}

func (h *Http) SetUpdatedBy(updatedBy string) {
	h.Meta.UpdatedBy = updatedBy
}

func (h *Http) GetUsedBy() string {
	return h.Meta.UsedBy
}

func (h *Http) SetUsedBy(usedBy string) {
	h.Meta.UsedBy = usedBy
}

func (h *Http) GetLastApplied() time.Time {
	return h.Meta.LastApplied
}

func (h *Http) SetLastApplied(lastApplied time.Time) {
	h.Meta.LastApplied = lastApplied
}

func foo() {
	gofakeit.Password()
}
