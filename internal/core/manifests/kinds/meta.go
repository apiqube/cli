package kinds

import (
	"math"
	"os/user"
	"time"

	"github.com/apiqube/cli/internal/core/manifests"
)

func DefaultMeta() *Meta {
	var name string

	currentUser, err := user.Current()
	if err != nil {
		name = "qube"
	} else {
		name = currentUser.Name
	}

	return &Meta{
		Hash:        "",
		Version:     1,
		CreatedAt:   time.Now(),
		CreatedBy:   name,
		UpdatedAt:   time.Now(),
		UpdatedBy:   name,
		UsedBy:      name,
		LastApplied: time.Now(),
	}
}

var _ manifests.Meta = (*Meta)(nil)

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

func (m *Meta) GetHash() string {
	return m.Hash
}

func (m *Meta) SetHash(hash string) {
	m.Hash = hash
}

func (m *Meta) GetVersion() uint8 {
	return m.Version
}

func (m *Meta) SetVersion(version uint8) {
	m.Version = version
}

func (m *Meta) IncVersion() {
	if m.Version < math.MaxUint8 {
		m.Version++
	}
}

func (m *Meta) GetCreatedAt() time.Time {
	return m.CreatedAt
}

func (m *Meta) SetCreatedAt(createdAt time.Time) {
	m.CreatedAt = createdAt
}

func (m *Meta) GetCreatedBy() string {
	return m.CreatedBy
}

func (m *Meta) SetCreatedBy(createdBy string) {
	m.CreatedBy = createdBy
}

func (m *Meta) GetUpdatedAt() time.Time {
	return m.UpdatedAt
}

func (m *Meta) SetUpdatedAt(updatedAt time.Time) {
	m.UpdatedAt = updatedAt
}

func (m *Meta) GetUpdatedBy() string {
	return m.UpdatedBy
}

func (m *Meta) SetUpdatedBy(updatedBy string) {
	m.UpdatedBy = updatedBy
}

func (m *Meta) GetUsedBy() string {
	return m.UsedBy
}

func (m *Meta) SetUsedBy(usedBy string) {
	m.UsedBy = usedBy
}

func (m *Meta) GetLastApplied() time.Time {
	return m.LastApplied
}

func (m *Meta) SetLastApplied(lastApplied time.Time) {
	m.LastApplied = lastApplied
}
