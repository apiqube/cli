package depends

// AddRule adds a new rule to the registry
func (gb *GraphBuilderV2) AddRule(rule DependencyRule) {
	gb.registry.Register(rule)
}

// GetSaveRequirement returns save requirement for a manifest
func (gr *GraphResultV2) GetSaveRequirement(manifestID string) (SaveRequirement, bool) {
	req, exists := gr.SaveRequirements[manifestID]
	return req, exists
}

// GetDependenciesFor returns all dependencies for a manifest
func (gr *GraphResultV2) GetDependenciesFor(manifestID string) []Dependency {
	var deps []Dependency
	for _, dep := range gr.Dependencies {
		if dep.From == manifestID {
			deps = append(deps, dep)
		}
	}
	return deps
}

// GetDependentsOf returns dependencies that depend on the given manifest
func (gr *GraphResultV2) GetDependentsOf(manifestID string) []Dependency {
	var dependents []Dependency
	for _, dep := range gr.Dependencies {
		if dep.To == manifestID {
			dependents = append(dependents, dep)
		}
	}
	return dependents
}

// GetDependenciesOf returns dependencies that the given manifest depends on
func (gr *GraphResultV2) GetDependenciesOf(manifestID string) []Dependency {
	var dependencies []Dependency
	for _, dep := range gr.Dependencies {
		if dep.From == manifestID {
			dependencies = append(dependencies, dep)
		}
	}
	return dependencies
}

// GetIntraManifestDependencies returns intra-manifest dependencies for a given manifest
func (gr *GraphResultV2) GetIntraManifestDependencies(manifestID string) []Dependency {
	if deps, exists := gr.IntraManifestDeps[manifestID]; exists {
		return deps
	}
	return []Dependency{}
}

// HasIntraManifestDependencies checks if a manifest has intra-manifest dependencies
func (gr *GraphResultV2) HasIntraManifestDependencies(manifestID string) bool {
	deps, exists := gr.IntraManifestDeps[manifestID]
	return exists && len(deps) > 0
}
