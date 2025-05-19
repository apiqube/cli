package plan

import "github.com/apiqube/cli/internal/core/manifests"

type RunMode string

const (
	StrictMode RunMode = "strict"
	Parallel   RunMode = "parallel"
)

func (m RunMode) String() string {
	return string(m)
}

type Builder interface {
	WithManifests(manifests ...manifests.Manifest) Builder
	WithManifestsMap(manifests map[string]manifests.Manifest) Builder
	WithRunMode(mode string) Builder
	WithStableSort(enable bool) Builder
	WithParallel(parallel bool) Builder
	Build() Manager
}

type managerBuilder struct {
	manifests  map[string]manifests.Manifest
	mode       string
	stableSort bool
	parallel   bool
}

func NewPlanManagerBuilder() Builder {
	return &managerBuilder{
		manifests:  make(map[string]manifests.Manifest),
		mode:       StrictMode.String(),
		stableSort: true,
		parallel:   false,
	}
}

func (b *managerBuilder) WithManifests(manifests ...manifests.Manifest) Builder {
	for _, manifest := range manifests {
		b.manifests[manifest.GetID()] = manifest
	}
	return b
}

func (b *managerBuilder) WithManifestsMap(set map[string]manifests.Manifest) Builder {
	b.manifests = set
	return b
}

func (b *managerBuilder) WithRunMode(mode string) Builder {
	b.mode = mode
	return b
}

func (b *managerBuilder) WithStableSort(enable bool) Builder {
	b.stableSort = enable
	return b
}

func (b *managerBuilder) WithParallel(parallel bool) Builder {
	b.parallel = parallel
	return b
}

func (b *managerBuilder) Build() Manager {
	return &basicManager{
		manifests:  b.manifests,
		mode:       b.mode,
		stableSort: b.stableSort,
	}
}
