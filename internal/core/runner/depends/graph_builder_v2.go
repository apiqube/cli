package depends

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/api"
)

var priorityOrder = map[string]int{
	manifests.ValuesKind:       1,
	manifests.ServerKind:       10,
	manifests.ServiceKind:      20,
	manifests.HttpTestKind:     30,
	manifests.HttpLoadTestKind: 40,
}

// GraphBuilderV2 builds dependency graphs using rule-based analysis
type GraphBuilderV2 struct {
	registry         *RuleRegistry
	manifestPriority map[string]int
	templateRegex    *regexp.Regexp
}

// GraphResultV2 represents the result of graph building with enhanced metadata
type GraphResultV2 struct {
	Graph              map[string][]string          // Adjacency list representation
	ExecutionOrder     []string                     // Topologically sorted execution order
	Dependencies       []Dependency                 // All inter-manifest dependencies
	IntraManifestDeps  map[string][]Dependency      // Dependencies within manifests
	SaveRequirements   map[string]SaveRequirement   // What data needs to be saved
	ManifestPriorities map[string]int               // Priority of each manifest
	AliasToManifest    map[string]string            // Maps alias to manifest ID
	TestCaseAliases    map[string]TestCaseAliasInfo // Maps alias to test case info
}

// SaveRequirement defines what data needs to be saved from a manifest execution
type SaveRequirement struct {
	Required      bool     // Whether saving is required
	RequiredPaths []string // Specific paths that need to be saved
	Consumers     []string // Which manifests consume this data
	Paths         []string // All paths (for compatibility)
}

// TestCaseAliasInfo contains information about test case aliases
type TestCaseAliasInfo struct {
	ManifestID    string   // Full manifest ID (e.g., "default.HttpTest.http-test-users")
	Alias         string   // The alias name (e.g., "fetch-users")
	TestCaseIndex int      // Index of the test case in the manifest
	RequiredPaths []string // Paths that other test cases need from this alias
	Consumers     []string // Which manifests/test cases consume this alias
}

// NewGraphBuilderV2 creates a new graph builder with rule registry
func NewGraphBuilderV2(registry *RuleRegistry) *GraphBuilderV2 {
	return &GraphBuilderV2{
		registry:         registry,
		manifestPriority: make(map[string]int),
		templateRegex:    regexp.MustCompile(`\{\{\s*([a-zA-Z][a-zA-Z0-9_-]*)\.(.*?)\s*}}`),
	}
}

// BuildGraphWithRules builds dependency graph using registered rules
func (gb *GraphBuilderV2) BuildGraphWithRules(manifests []manifests.Manifest) (*GraphResultV2, error) {
	result := &GraphResultV2{
		Graph:              make(map[string][]string),
		Dependencies:       make([]Dependency, 0),
		IntraManifestDeps:  make(map[string][]Dependency),
		SaveRequirements:   make(map[string]SaveRequirement),
		ManifestPriorities: make(map[string]int),
		AliasToManifest:    make(map[string]string),
		TestCaseAliases:    make(map[string]TestCaseAliasInfo),
	}

	// Step 1: Initialize manifest priorities and collect aliases
	if err := gb.initializeManifests(manifests, result); err != nil {
		return nil, err
	}

	// Step 2: Analyze dependencies using all rules (but ignore explicit dependencies)
	allDependencies, err := gb.analyzeAllDependencies(manifests)
	if err != nil {
		return nil, err
	}

	// Step 3: Separate inter-manifest and intra-manifest dependencies
	gb.categorizeDependencies(allDependencies, result)

	// Step 4: Build adjacency graph from dependencies
	gb.buildAdjacencyGraph(result)

	// Step 5: Calculate save requirements
	gb.calculateSaveRequirements(result)

	// Step 6: Build execution order using topological sort with priorities
	executionOrder, err := gb.buildExecutionOrder(manifests, result.Dependencies)
	if err != nil {
		return nil, err
	}
	result.ExecutionOrder = executionOrder

	return result, nil
}

// initializeManifests sets up manifest priorities and collects alias information
func (gb *GraphBuilderV2) initializeManifests(manifests []manifests.Manifest, result *GraphResultV2) error {
	for _, manifest := range manifests {
		manifestID := manifest.GetID()

		// Set priority
		priority := gb.getManifestPriority(manifest)
		gb.manifestPriority[manifestID] = priority
		result.ManifestPriorities[manifestID] = priority

		// Collect test case aliases for HTTP tests
		if httpTest, ok := manifest.(*api.Http); ok {
			for i, testCase := range httpTest.Spec.Cases {
				if testCase.Alias != nil {
					alias := *testCase.Alias
					result.AliasToManifest[alias] = manifestID
					result.TestCaseAliases[alias] = TestCaseAliasInfo{
						ManifestID:    manifestID,
						Alias:         alias,
						TestCaseIndex: i,
						RequiredPaths: make([]string, 0),
						Consumers:     make([]string, 0),
					}
				}
			}
		}
	}

	return nil
}

// analyzeAllDependencies analyzes dependencies using all registered rules
func (gb *GraphBuilderV2) analyzeAllDependencies(manifests []manifests.Manifest) ([]Dependency, error) {
	var allDependencies []Dependency

	for _, manifest := range manifests {
		for _, rule := range gb.registry.GetRules() {
			if rule.CanHandle(manifest) {
				// Skip explicit dependency rule - we want to build dependencies ourselves
				if rule.Name() == "explicit" {
					continue
				}

				deps, err := rule.AnalyzeDependencies(manifest)
				if err != nil {
					return nil, fmt.Errorf("rule %s failed for manifest %s: %w", rule.Name(), manifest.GetID(), err)
				}
				allDependencies = append(allDependencies, deps...)
			}
		}
	}

	// Add smart template-based dependencies
	smartDeps, err := gb.analyzeSmartTemplateDependencies(manifests, allDependencies)
	if err != nil {
		return nil, err
	}
	allDependencies = append(allDependencies, smartDeps...)

	return allDependencies, nil
}

// analyzeSmartTemplateDependencies creates inter-manifest dependencies based on template analysis
func (gb *GraphBuilderV2) analyzeSmartTemplateDependencies(mans []manifests.Manifest, existingDeps []Dependency) ([]Dependency, error) {
	var smartDeps []Dependency
	aliasToManifest := make(map[string]string)

	// Build alias to manifest mapping
	for _, manifest := range mans {
		if httpTest, ok := manifest.(*api.Http); ok {
			for _, testCase := range httpTest.Spec.Cases {
				if testCase.Alias != nil {
					aliasToManifest[*testCase.Alias] = manifest.GetID()
				}
			}
		}
	}

	// Analyze template references and create inter-manifest dependencies
	for _, manifest := range mans {
		manifestID := manifest.GetID()

		if httpTest, ok := manifest.(*api.Http); ok {
			// Find all template references in this manifest
			templateRefs := gb.extractAllTemplateReferences(httpTest)

			// Group by alias and create dependencies
			aliasGroups := make(map[string][]string)
			for _, ref := range templateRefs {
				aliasGroups[ref.Alias] = append(aliasGroups[ref.Alias], ref.Path)
			}

			for alias, paths := range aliasGroups {
				// Check if this alias refers to another manifest
				if targetManifestID, exists := aliasToManifest[alias]; exists && targetManifestID != manifestID {
					// This is an inter-manifest dependency
					smartDeps = append(smartDeps, Dependency{
						From: manifestID,
						To:   targetManifestID,
						Type: DependencyTypeTemplate,
						Metadata: map[string]any{
							"alias":          alias,
							"required_paths": paths,
							"save_required":  true,
							"smart_detected": true,
						},
					})
				} else if alias == "Values" {
					// Check if this is a reference to Values manifest
					for _, valuesManifest := range mans {
						if valuesManifest.GetKind() == manifests.ValuesKind {
							smartDeps = append(smartDeps, Dependency{
								From: manifestID,
								To:   valuesManifest.GetID(),
								Type: DependencyTypeTemplate,
								Metadata: map[string]any{
									"alias":          alias,
									"required_paths": paths,
									"save_required":  true,
									"smart_detected": true,
								},
							})
							break
						}
					}
				}
			}
		}
	}

	return smartDeps, nil
}

// extractAllTemplateReferences extracts all template references from an HTTP test
func (gb *GraphBuilderV2) extractAllTemplateReferences(httpTest *api.Http) []TemplateReference {
	var references []TemplateReference

	for _, testCase := range httpTest.Spec.Cases {
		// Check endpoint
		refs := gb.findTemplateReferencesInString(testCase.Endpoint)
		references = append(references, refs...)

		// Check URL
		refs = gb.findTemplateReferencesInString(testCase.Url)
		references = append(references, refs...)

		// Check headers
		for _, value := range testCase.Headers {
			refs = gb.findTemplateReferencesInString(value)
			references = append(references, refs...)
		}

		// Check body recursively
		if testCase.Body != nil {
			refs = gb.findTemplateReferencesInValue(testCase.Body)
			references = append(references, refs...)
		}

		// Check assertions
		for _, assert := range testCase.Assert {
			if assert.Template != "" {
				refs = gb.findTemplateReferencesInString(assert.Template)
				references = append(references, refs...)
			}
		}
	}

	return references
}

// findTemplateReferencesInString finds template references in a string
func (gb *GraphBuilderV2) findTemplateReferencesInString(str string) []TemplateReference {
	var references []TemplateReference
	matches := gb.templateRegex.FindAllStringSubmatch(str, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			references = append(references, TemplateReference{
				Alias: match[1],
				Path:  match[2],
			})
		}
	}

	return references
}

// findTemplateReferencesInValue recursively finds template references in any value
func (gb *GraphBuilderV2) findTemplateReferencesInValue(value any) []TemplateReference {
	var references []TemplateReference

	switch v := value.(type) {
	case string:
		references = append(references, gb.findTemplateReferencesInString(v)...)
	case map[string]any:
		for _, val := range v {
			references = append(references, gb.findTemplateReferencesInValue(val)...)
		}
	case []any:
		for _, val := range v {
			references = append(references, gb.findTemplateReferencesInValue(val)...)
		}
	case map[any]any:
		for _, val := range v {
			references = append(references, gb.findTemplateReferencesInValue(val)...)
		}
	}

	return references
}

// categorizeDependencies separates inter-manifest and intra-manifest dependencies
func (gb *GraphBuilderV2) categorizeDependencies(allDependencies []Dependency, result *GraphResultV2) {
	for _, dep := range allDependencies {
		fromManifest := gb.getBaseManifestID(dep.From)
		toManifest := gb.getBaseManifestID(dep.To)

		if fromManifest == toManifest {
			// Intra-manifest dependency
			if result.IntraManifestDeps[fromManifest] == nil {
				result.IntraManifestDeps[fromManifest] = make([]Dependency, 0)
			}
			result.IntraManifestDeps[fromManifest] = append(result.IntraManifestDeps[fromManifest], dep)
		} else {
			// Inter-manifest dependency
			result.Dependencies = append(result.Dependencies, dep)
		}
	}
}

// buildAdjacencyGraph builds the adjacency graph from dependencies
func (gb *GraphBuilderV2) buildAdjacencyGraph(result *GraphResultV2) {
	for _, dep := range result.Dependencies {
		toManifest := gb.getBaseManifestID(dep.To)
		fromManifest := gb.getBaseManifestID(dep.From)

		if result.Graph[toManifest] == nil {
			result.Graph[toManifest] = make([]string, 0)
		}
		result.Graph[toManifest] = append(result.Graph[toManifest], fromManifest)
	}
}

// calculateSaveRequirements determines what data needs to be saved
func (gb *GraphBuilderV2) calculateSaveRequirements(result *GraphResultV2) {
	// Process inter-manifest dependencies
	for _, dep := range result.Dependencies {
		if dep.Type == DependencyTypeTemplate {
			toManifest := gb.getBaseManifestID(dep.To)

			// Update save requirement for the target manifest
			req := result.SaveRequirements[toManifest]
			req.Required = true
			req.Consumers = append(req.Consumers, dep.From)

			if paths, ok := dep.Metadata["required_paths"].([]string); ok {
				req.RequiredPaths = append(req.RequiredPaths, paths...)
				req.Paths = append(req.Paths, paths...)
			}

			result.SaveRequirements[toManifest] = req

			// Update test case alias info if applicable
			var paths []string
			if alias, ok := dep.Metadata["alias"].(string); ok {
				if aliasInfo, exists := result.TestCaseAliases[alias]; exists {
					aliasInfo.Consumers = append(aliasInfo.Consumers, dep.From)
					if paths, ok = dep.Metadata["required_paths"].([]string); ok {
						aliasInfo.RequiredPaths = append(aliasInfo.RequiredPaths, paths...)
					}
					result.TestCaseAliases[alias] = aliasInfo
				}
			}
		}
	}

	// Process intra-manifest dependencies
	for manifestID, deps := range result.IntraManifestDeps {
		req := result.SaveRequirements[manifestID]
		req.Required = true
		req.Consumers = append(req.Consumers, manifestID) // Self-consumer

		for _, dep := range deps {
			if paths, ok := dep.Metadata["required_paths"].([]string); ok {
				req.RequiredPaths = append(req.RequiredPaths, paths...)
				req.Paths = append(req.Paths, paths...)
			}
		}

		result.SaveRequirements[manifestID] = req
	}
}

// buildExecutionOrder creates topologically sorted execution order
func (gb *GraphBuilderV2) buildExecutionOrder(manifests []manifests.Manifest, dependencies []Dependency) ([]string, error) {
	// Initialize in-degree count for each manifest
	inDegree := make(map[string]int)
	manifestIDs := make([]string, 0, len(manifests))
	for _, manifest := range manifests {
		id := manifest.GetID()
		inDegree[id] = 0
		manifestIDs = append(manifestIDs, id)
	}

	// Calculate in-degrees from all inter-manifest dependencies
	for _, dep := range dependencies {
		fromBase := gb.getBaseManifestID(dep.From)
		toBase := gb.getBaseManifestID(dep.To)
		if fromBase != toBase {
			if _, exists := inDegree[fromBase]; exists && dep.Type != DependencyTypeTemplate {
				inDegree[fromBase]++
			}
		}
	}

	// Use a slice as a queue for topological sorting with priorities and deterministic order
	zeroInDegreeNodes := make([]*Node, 0)
	for id, degree := range inDegree {
		if degree == 0 {
			zeroInDegreeNodes = append(zeroInDegreeNodes, &Node{
				ID:       id,
				Priority: gb.manifestPriority[id],
			})
		}
	}

	// Sort for deterministic behavior
	sort.Slice(zeroInDegreeNodes, func(i, j int) bool {
		if zeroInDegreeNodes[i].Priority != zeroInDegreeNodes[j].Priority {
			return zeroInDegreeNodes[i].Priority < zeroInDegreeNodes[j].Priority
		}
		return zeroInDegreeNodes[i].ID < zeroInDegreeNodes[j].ID
	})

	executionOrder := make([]string, 0, len(manifests))
	queue := zeroInDegreeNodes

	for len(queue) > 0 {
		currentNode := queue[0]
		queue = queue[1:]
		executionOrder = append(executionOrder, currentNode.ID)

		newNodes := make([]*Node, 0)
		for _, dep := range dependencies {
			fromBase := gb.getBaseManifestID(dep.From)
			toBase := gb.getBaseManifestID(dep.To)
			if toBase == currentNode.ID && fromBase != toBase {
				inDegree[fromBase]--
				if inDegree[fromBase] == 0 {
					newNodes = append(newNodes, &Node{
						ID:       fromBase,
						Priority: gb.manifestPriority[fromBase],
					})
				}
			}
		}
		// Сортируем кандидатов по приоритету и ID для детерминизма
		sort.Slice(newNodes, func(i, j int) bool {
			if newNodes[i].Priority != newNodes[j].Priority {
				return newNodes[i].Priority < newNodes[j].Priority
			}
			return newNodes[i].ID < newNodes[j].ID
		})
		queue = append(queue, newNodes...)
	}

	// Check for cycles
	if len(executionOrder) != len(manifests) {
		var remaining []string
		for manifestID, degree := range inDegree {
			if degree > 0 {
				remaining = append(remaining, manifestID)
			}
		}
		return nil, fmt.Errorf("cyclic dependency detected among manifests: %v", remaining)
	}

	return executionOrder, nil
}

// getManifestPriority returns priority for a manifest based on its kind
func (gb *GraphBuilderV2) getManifestPriority(manifest manifests.Manifest) int {
	kind := manifest.GetKind()

	// Find kind priority rule
	for _, rule := range gb.registry.GetRules() {
		if kindRule, ok := rule.(*KindPriorityRule); ok {
			return kindRule.GetKindPriority(kind)
		}
	}

	return gb.getManifestPriorityByID(manifest.GetID())
}

// getManifestPriorityByID returns priority for a manifest by its ID
func (gb *GraphBuilderV2) getManifestPriorityByID(manifestID string) int {
	if priority, exists := gb.manifestPriority[manifestID]; exists {
		return priority
	}
	return 0
}

// getBaseManifestID extracts base manifest ID from potentially extended ID
func (gb *GraphBuilderV2) getBaseManifestID(id string) string {
	// Remove any suffix after # (for test case aliases)
	if idx := strings.Index(id, "#"); idx != -1 {
		return id[:idx]
	}
	return id
}
