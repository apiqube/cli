package depends

// GraphBuilderV2 is the new modular graph builder
type GraphBuilderV2 struct {
	ruleRegistry *RuleRegistry
}

func NewGraphBuilderV2(registry *RuleRegistry) *GraphBuilderV2 {
	if registry == nil {
		registry = DefaultRuleRegistry()
	}
	return &GraphBuilderV2{
		ruleRegistry: registry,
	}
}

// GraphResultV2 contains enhanced graph information
type GraphResultV2 struct {
	Graph             map[string][]string        // adjacency list (inter-manifest only)
	ExecutionOrder    []string                   // topologically sorted order
	Dependencies      []Dependency               // inter-manifest dependencies only
	AllDependencies   []Dependency               // all discovered dependencies (inter + intra)
	SaveRequirements  map[string]SaveRequirement // what each manifest needs to save
	Metadata          map[string]map[string]any  // additional metadata per manifest
	IntraManifestDeps map[string][]Dependency    // intra-manifest dependencies grouped by manifest
}

// SaveRequirement defines what data a manifest should save for others
type SaveRequirement struct {
	Required      bool     // whether saving is required
	ManifestID    string   // ID of the manifest that provides data
	RequiredPaths []string // specific paths to save (renamed from Paths for consistency)
	Paths         []string // alias for RequiredPaths for backward compatibility
	UsedBy        []string // which manifests will use this data (alias for Consumers)
	Consumers     []string // which manifests will consume this data
}

// AddRule adds a new rule to the registry
func (gb *GraphBuilderV2) AddRule(rule DependencyRule) {
	gb.ruleRegistry.Register(rule)
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
