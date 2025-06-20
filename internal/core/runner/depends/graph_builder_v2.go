package depends

import (
	"container/heap"
	"fmt"
	"github.com/apiqube/cli/internal/core/manifests/utils"
	"strings"

	"github.com/apiqube/cli/internal/collections"
	"github.com/apiqube/cli/internal/core/manifests"
)

var (
	priorities = map[string]int{
		manifests.ValuesKind:       1,
		manifests.ServerKind:       10,
		manifests.ServiceKind:      20,
		manifests.HttpTestKind:     30,
		manifests.HttpLoadTestKind: 40,
	}
)

// BuildGraphWithRules builds a dependency graph using registered rules
func (gb *GraphBuilderV2) BuildGraphWithRules(mans []manifests.Manifest) (*GraphResultV2, error) {
	if len(mans) == 0 {
		return &GraphResultV2{
			Graph:             make(map[string][]string),
			ExecutionOrder:    []string{},
			Dependencies:      []Dependency{},
			AllDependencies:   []Dependency{},
			SaveRequirements:  make(map[string]SaveRequirement),
			Metadata:          make(map[string]map[string]any),
			IntraManifestDeps: make(map[string][]Dependency),
		}, nil
	}

	// Step 1: Analyze dependencies using all rules
	allDependencies, err := gb.analyzeDependencies(mans)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze dependencies: %w", err)
	}

	// Step 2: Separate inter-manifest and intra-manifest dependencies
	manifestMap := make(map[string]manifests.Manifest)
	for _, manifest := range mans {
		manifestMap[manifest.GetID()] = manifest
	}

	interManifestDeps, intraManifestDeps := gb.separateDependencies(allDependencies, manifestMap)

	// Step 3: Build graph from inter-manifest dependencies only
	graph := gb.buildAdjacencyGraph(interManifestDeps, mans)

	// Step 4: Build execution order with topological sort
	executionOrder, err := gb.buildExecutionOrder(mans, graph, interManifestDeps)
	if err != nil {
		return nil, fmt.Errorf("failed to build execution order: %w", err)
	}

	// Step 5: Build save requirements (using all dependencies)
	saveRequirements := gb.calculateSaveRequirements(allDependencies, mans)

	// Step 6: Build metadata
	metadata := gb.collectMetadata(allDependencies, mans)

	// Step 7: Group intra-manifest dependencies by manifest
	intraManifestDepsByManifest := gb.groupIntraManifestDeps(intraManifestDeps, manifestMap)

	return &GraphResultV2{
		Graph:             graph,
		ExecutionOrder:    executionOrder,
		Dependencies:      interManifestDeps,
		AllDependencies:   allDependencies,
		SaveRequirements:  saveRequirements,
		Metadata:          metadata,
		IntraManifestDeps: intraManifestDepsByManifest,
	}, nil
}

// separateDependencies separates inter-manifest and intra-manifest dependencies
func (gb *GraphBuilderV2) separateDependencies(dependencies []Dependency, manifestMap map[string]manifests.Manifest) ([]Dependency, []Dependency) {
	var interManifestDeps []Dependency
	var intraManifestDeps []Dependency

	for _, dep := range dependencies {
		if gb.isIntraManifestDependency(dep, manifestMap) {
			// This is an intra-manifest dependency (e.g., test case aliases)
			intraManifestDeps = append(intraManifestDeps, dep)
		} else {
			// This is an inter-manifest dependency
			interManifestDeps = append(interManifestDeps, dep)
		}
	}

	return interManifestDeps, intraManifestDeps
}

// groupIntraManifestDeps groups intra-manifest dependencies by base manifest ID
func (gb *GraphBuilderV2) groupIntraManifestDeps(intraManifestDeps []Dependency, manifestMap map[string]manifests.Manifest) map[string][]Dependency {
	grouped := make(map[string][]Dependency)

	for _, dep := range intraManifestDeps {
		baseID := gb.getBaseManifestID(dep.From)
		grouped[baseID] = append(grouped[baseID], dep)
	}

	return grouped
}

// isIntraManifestDependency checks if a dependency is within the same manifest
func (gb *GraphBuilderV2) isIntraManifestDependency(dep Dependency, manifestMap map[string]manifests.Manifest) bool {
	// Extract base manifest ID (without alias/fragment)
	fromBase := gb.getBaseManifestID(dep.From)
	toBase := gb.getBaseManifestID(dep.To)

	// If both refer to the same base manifest, it's intra-manifest
	return fromBase == toBase
}

// getBaseManifestID extracts the base manifest ID without alias/fragment
func (gb *GraphBuilderV2) getBaseManifestID(id string) string {
	// Remove fragment part (after #)
	if idx := strings.Index(id, "#"); idx != -1 {
		return id[:idx]
	}
	return id
}

// buildAdjacencyGraph creates adjacency list from dependencies
func (gb *GraphBuilderV2) buildAdjacencyGraph(dependencies []Dependency, manifests []manifests.Manifest) map[string][]string {
	graph := make(map[string][]string)

	// Initialize all manifests in the graph
	for _, manifest := range manifests {
		id := manifest.GetID()
		graph[id] = []string{}
	}

	// Add edges from dependencies
	for _, dep := range dependencies {
		// Only add edges for dependencies where both nodes exist as manifests
		fromBase := gb.getBaseManifestID(dep.From)
		toBase := gb.getBaseManifestID(dep.To)

		// Check if both base IDs exist in our manifest map
		fromExists := false
		toExists := false

		for _, manifest := range manifests {
			if manifest.GetID() == fromBase {
				fromExists = true
			}
			if manifest.GetID() == toBase {
				toExists = true
			}
		}

		if fromExists && toExists {
			// Add edge: To -> From (dependency direction)
			graph[toBase] = append(graph[toBase], fromBase)
		}
	}

	return graph
}

// buildExecutionOrder creates topologically sorted execution order
func (gb *GraphBuilderV2) buildExecutionOrder(mans []manifests.Manifest, graph map[string][]string, dependencies []Dependency) ([]string, error) {
	// Calculate in-degrees
	inDegree := make(map[string]int)
	idToManifest := make(map[string]manifests.Manifest)
	nodePriority := make(map[string]int)

	// Initialize
	for _, manifest := range mans {
		id := manifest.GetID()
		idToManifest[id] = manifest
		inDegree[id] = 0
		nodePriority[id] = gb.getManifestPriority(manifest)
	}

	// Calculate in-degrees from dependencies
	for _, dep := range dependencies {
		fromBase := gb.getBaseManifestID(dep.From)
		if _, exists := inDegree[fromBase]; exists && dep.Type != DependencyTypeTemplate {
			inDegree[fromBase]++
		}
	}

	// Priority queue for topological sort
	priorityQueue := collections.NewPriorityQueue[*Node](func(a, b *Node) bool {
		return a.Priority > b.Priority
	})

	// Add nodes with no dependencies
	for id, degree := range inDegree {
		if degree == 0 {
			heap.Push(priorityQueue, &Node{
				ID:       id,
				Priority: nodePriority[id],
			})
		}
	}

	var order []string
	for priorityQueue.Len() > 0 {
		current := heap.Pop(priorityQueue).(*Node).ID
		order = append(order, current)

		// Process neighbors
		for _, neighbor := range graph[current] {
			if inDegree[neighbor] > 0 {
				inDegree[neighbor]--
				if inDegree[neighbor] == 0 {
					heap.Push(priorityQueue, &Node{
						ID:       neighbor,
						Priority: nodePriority[neighbor],
					})
				}
			}
		}
	}

	// Check for cycles
	if len(order) != len(mans) {
		cyclicNodes := gb.findCyclicNodes(inDegree)
		return nil, fmt.Errorf("cyclic dependency detected: %v", cyclicNodes)
	}

	return order, nil
}

// getManifestPriorityByID gets priority by manifest ID
func (gb *GraphBuilderV2) getManifestPriorityByID(manifestID string) int {
	// Extract kind from ID (assuming format: namespace.kind.name)
	_, kind, _ := utils.ParseManifestID(manifestID)
	return gb.getKindPriority(kind)
}

// getKindPriority returns priority for a manifest kind
func (gb *GraphBuilderV2) getKindPriority(kind string) int {
	if priority, ok := priorities[kind]; ok {
		return priority
	}

	return 1_000
}

// analyzeDependencies runs all rules to discover dependencies
func (gb *GraphBuilderV2) analyzeDependencies(manifests []manifests.Manifest) ([]Dependency, error) {
	var allDependencies []Dependency

	for _, manifest := range manifests {
		for _, rule := range gb.ruleRegistry.GetRules() {
			if !rule.CanHandle(manifest) {
				continue
			}

			dependencies, err := rule.AnalyzeDependencies(manifest)
			if err != nil {
				return nil, fmt.Errorf("rule %s failed for manifest %s: %w",
					rule.Name(), manifest.GetID(), err)
			}

			allDependencies = append(allDependencies, dependencies...)
		}
	}

	return gb.deduplicateDependencies(allDependencies), nil
}

// deduplicateDependencies removes duplicate dependencies
func (gb *GraphBuilderV2) deduplicateDependencies(deps []Dependency) []Dependency {
	seen := make(map[string]bool)
	var result []Dependency

	for _, dep := range deps {
		key := fmt.Sprintf("%s->%s:%s", dep.From, dep.To, dep.Type)
		if !seen[key] {
			seen[key] = true
			result = append(result, dep)
		}
	}

	return result
}

// calculateSaveRequirements determines what each manifest needs to save
func (gb *GraphBuilderV2) calculateSaveRequirements(dependencies []Dependency, manifests []manifests.Manifest) map[string]SaveRequirement {
	requirements := make(map[string]SaveRequirement)

	// Initialize all manifests with no save requirement
	for _, manifest := range manifests {
		requirements[manifest.GetID()] = SaveRequirement{
			Required:      false,
			ManifestID:    manifest.GetID(),
			RequiredPaths: []string{},
			Paths:         []string{},
			UsedBy:        []string{},
			Consumers:     []string{},
		}
	}

	// Process template dependencies to determine save requirements
	for _, dep := range dependencies {
		if dep.Type == DependencyTypeTemplate {
			toBase := gb.getBaseManifestID(dep.To)
			req := requirements[toBase]
			req.Required = true
			req.UsedBy = append(req.UsedBy, dep.From)
			req.Consumers = append(req.Consumers, dep.From)

			// Add required paths from metadata
			if paths, ok := dep.Metadata["required_paths"].([]string); ok {
				req.RequiredPaths = append(req.RequiredPaths, paths...)
				req.Paths = append(req.Paths, paths...) // for backward compatibility
			}

			requirements[toBase] = req
		}
	}

	// Remove duplicates from paths and consumers
	for id, req := range requirements {
		req.RequiredPaths = gb.removeDuplicateStrings(req.RequiredPaths)
		req.Paths = gb.removeDuplicateStrings(req.Paths)
		req.UsedBy = gb.removeDuplicateStrings(req.UsedBy)
		req.Consumers = gb.removeDuplicateStrings(req.Consumers)
		requirements[id] = req
	}

	return requirements
}

// collectMetadata gathers metadata from all dependencies
func (gb *GraphBuilderV2) collectMetadata(dependencies []Dependency, manifests []manifests.Manifest) map[string]map[string]any {
	metadata := make(map[string]map[string]any)

	// Initialize metadata for all manifests
	for _, manifest := range manifests {
		metadata[manifest.GetID()] = make(map[string]any)
	}

	// Collect metadata from dependencies
	for _, dep := range dependencies {
		fromBase := gb.getBaseManifestID(dep.From)
		if dep.Metadata != nil {
			if metadata[fromBase] == nil {
				metadata[fromBase] = make(map[string]any)
			}
			manifestMeta := metadata[fromBase]
			for key, value := range dep.Metadata {
				manifestMeta[key] = value
			}
		}
	}

	return metadata
}

// getManifestPriority calculates priority for a manifest
func (gb *GraphBuilderV2) getManifestPriority(manifest manifests.Manifest) int {
	// Try to find KindPriorityRule
	for _, rule := range gb.ruleRegistry.GetRules() {
		if kindRule, ok := rule.(*KindPriorityRule); ok {
			return kindRule.GetKindPriority(manifest.GetKind())
		}
	}

	// Fallback to direct priority calculation
	return gb.getKindPriority(manifest.GetKind())
}

// findCyclicNodes finds nodes involved in cycles
func (gb *GraphBuilderV2) findCyclicNodes(inDegree map[string]int) []string {
	var cyclicNodes []string
	for id, degree := range inDegree {
		if degree > 0 {
			cyclicNodes = append(cyclicNodes, id)
		}
	}
	return cyclicNodes
}

// removeDuplicateStrings removes duplicate strings from a slice
func (gb *GraphBuilderV2) removeDuplicateStrings(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// Legacy methods for backward compatibility
func (gb *GraphBuilderV2) filterIntraManifestDependencies(dependencies []Dependency, manifestMap map[string]manifests.Manifest) []Dependency {
	interManifestDeps, _ := gb.separateDependencies(dependencies, manifestMap)
	return interManifestDeps
}

func (gb *GraphBuilderV2) buildSaveRequirements(dependencies []Dependency, manifestMap map[string]manifests.Manifest) map[string]SaveRequirement {
	requirements := make(map[string]SaveRequirement)

	// Group dependencies by target (what provides the data)
	providerMap := make(map[string][]Dependency)
	for _, dep := range dependencies {
		if dep.Type == DependencyTypeTemplate {
			providerMap[dep.To] = append(providerMap[dep.To], dep)
		}
	}

	// Build save requirements
	for providerID, deps := range providerMap {
		var requiredPaths []string
		var consumers []string

		for _, dep := range deps {
			consumers = append(consumers, dep.From)

			// Extract required paths from metadata
			if paths, ok := dep.Metadata["required_paths"].([]string); ok {
				requiredPaths = append(requiredPaths, paths...)
			}
		}

		// Remove duplicates
		requiredPaths = gb.removeDuplicateStrings(requiredPaths)
		consumers = gb.removeDuplicateStrings(consumers)

		requirements[providerID] = SaveRequirement{
			Required:      true,
			ManifestID:    providerID,
			RequiredPaths: requiredPaths,
			Paths:         requiredPaths, // for backward compatibility
			UsedBy:        consumers,     // for backward compatibility
			Consumers:     consumers,
		}
	}

	return requirements
}

func (gb *GraphBuilderV2) buildMetadata(dependencies []Dependency, manifestMap map[string]manifests.Manifest) map[string]map[string]any {
	metadata := make(map[string]map[string]any)

	for _, dep := range dependencies {
		if metadata[dep.From] == nil {
			metadata[dep.From] = make(map[string]any)
		}

		// Add dependency metadata
		depKey := fmt.Sprintf("dep_%s", dep.To)
		metadata[dep.From][depKey] = dep.Metadata
	}

	return metadata
}
