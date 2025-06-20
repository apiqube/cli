package depends

import "strings"

// GetSaveRequirement returns save requirement for a manifest
func (r *GraphResultV2) GetSaveRequirement(manifestID string) (SaveRequirement, bool) {
	req, exists := r.SaveRequirements[manifestID]
	return req, exists
}

// GetDependenciesFor returns all dependencies for a manifest
func (r *GraphResultV2) GetDependenciesFor(manifestID string) []Dependency {
	var deps []Dependency
	for _, dep := range r.Dependencies {
		if dep.From == manifestID {
			deps = append(deps, dep)
		}
	}
	return deps
}

// GetDependenciesOf returns dependencies that the given manifest depends on
func (r *GraphResultV2) GetDependenciesOf(manifestID string) []Dependency {
	var dependencies []Dependency
	for _, dep := range r.Dependencies {
		if dep.From == manifestID {
			dependencies = append(dependencies, dep)
		}
	}
	return dependencies
}

// GetIntraManifestDependencies returns intra-manifest dependencies for a given manifest
func (r *GraphResultV2) GetIntraManifestDependencies(manifestID string) []Dependency {
	if deps, exists := r.IntraManifestDeps[manifestID]; exists {
		return deps
	}
	return []Dependency{}
}

// HasIntraManifestDependencies checks if a manifest has intra-manifest dependencies
func (r *GraphResultV2) HasIntraManifestDependencies(manifestID string) bool {
	deps, exists := r.IntraManifestDeps[manifestID]
	return exists && len(deps) > 0
}

// GetDependentsOf returns all dependencies that depend on the given manifest
func (r *GraphResultV2) GetDependentsOf(manifestID string) []Dependency {
	var dependents []Dependency

	for _, dep := range r.Dependencies {
		if dep.To == manifestID || strings.HasPrefix(dep.To, manifestID+"#") {
			dependents = append(dependents, dep)
		}
	}

	return dependents
}
