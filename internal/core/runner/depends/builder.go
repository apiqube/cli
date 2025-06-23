package depends

import "github.com/apiqube/cli/internal/core/runner/depends/rules"

// AddRule adds a new rule to the registry
func (gb *GraphBuilder) AddRule(rule rules.DependencyRule) {
	gb.registry.Register(rule)
}

// GetSaveRequirement returns save requirement for a manifest
func (gr *GraphResultV2) GetSaveRequirement(manifestID string) (SaveRequirement, bool) {
	req, exists := gr.SaveRequirements[manifestID]
	return req, exists
}

// GetDependenciesFor returns all dependencies for a manifest
func (gr *GraphResultV2) GetDependenciesFor(manifestID string) []rules.Dependency {
	var deps []rules.Dependency
	for _, dep := range gr.Dependencies {
		if dep.From == manifestID {
			deps = append(deps, dep)
		}
	}
	return deps
}

// GetDependentsOf returns dependencies that depend on the given manifest
func (gr *GraphResultV2) GetDependentsOf(manifestID string) []rules.Dependency {
	var dependents []rules.Dependency
	for _, dep := range gr.Dependencies {
		if dep.To == manifestID {
			dependents = append(dependents, dep)
		}
	}
	return dependents
}

// GetIntraManifestDependencies returns intra-manifest dependencies for a given manifest
func (gr *GraphResultV2) GetIntraManifestDependencies(manifestID string) []rules.Dependency {
	if deps, exists := gr.IntraManifestDeps[manifestID]; exists {
		return deps
	}
	return []rules.Dependency{}
}

// HasIntraManifestDependencies checks if a manifest has intra-manifest dependencies
func (gr *GraphResultV2) HasIntraManifestDependencies(manifestID string) bool {
	deps, exists := gr.IntraManifestDeps[manifestID]
	return exists && len(deps) > 0
}
