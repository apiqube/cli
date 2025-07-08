package depends

import "github.com/apiqube/cli/internal/core/runner/depends/rules"

// AddRule adds a new rule to the registry
func (b *Builder) AddRule(rule rules.DependencyRule) {
	b.registry.Register(rule)
}

// GetSaveRequirement returns save requirement for a manifest
func (r *Result) GetSaveRequirement(manifestID string) (SaveRequirement, bool) {
	req, exists := r.SaveRequirements[manifestID]
	return req, exists
}

// GetDependenciesFor returns all dependencies for a manifest
func (r *Result) GetDependenciesFor(manifestID string) []rules.Dependency {
	var deps []rules.Dependency
	for _, dep := range r.Dependencies {
		if dep.From == manifestID {
			deps = append(deps, dep)
		}
	}
	return deps
}

// GetDependentsOf returns dependencies that depend on the given manifest
func (r *Result) GetDependentsOf(manifestID string) []rules.Dependency {
	var dependents []rules.Dependency
	for _, dep := range r.Dependencies {
		if dep.To == manifestID {
			dependents = append(dependents, dep)
		}
	}
	return dependents
}

// GetIntraManifestDependencies returns intra-manifest dependencies for a given manifest
func (r *Result) GetIntraManifestDependencies(manifestID string) []rules.Dependency {
	if deps, exists := r.IntraManifestDeps[manifestID]; exists {
		return deps
	}
	return []rules.Dependency{}
}

// HasIntraManifestDependencies checks if a manifest has intra-manifest dependencies
func (r *Result) HasIntraManifestDependencies(manifestID string) bool {
	deps, exists := r.IntraManifestDeps[manifestID]
	return exists && len(deps) > 0
}
