package depends

import (
	"fmt"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
	"github.com/apiqube/cli/internal/core/runner/depends/rules"
	"regexp"
	"sort"
	"strings"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/api"
)

// GraphBuilder builds dependency graphs using rule-based analysis
type GraphBuilder struct {
	registry         *rules.RuleRegistry
	manifestPriority map[string]int
	templateRegex    *regexp.Regexp
}

// GraphResultV2 represents the result of graph building with enhanced metadata
type GraphResultV2 struct {
	Graph             map[string][]string           // Adjacency list representation
	ExecutionOrder    []string                      // Topologically sorted execution order
	Dependencies      []rules.Dependency            // All inter-manifest dependencies
	IntraManifestDeps map[string][]rules.Dependency // Dependencies within manifests
	SaveRequirements  map[string]SaveRequirement    // What data needs to be saved
	AliasToManifest   map[string]string             // Maps alias to manifest ID
	TestCaseAliases   map[string]TestCaseAliasInfo  // Maps alias to test case info
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

type Node struct {
	ID       string
	Priority int
}

// NewGraphBuilder creates a new graph builder with rule registry
func NewGraphBuilder(registry *rules.RuleRegistry) *GraphBuilder {
	if registry == nil {
		registry = rules.DefaultRuleRegistry()
	}

	return &GraphBuilder{
		registry:         registry,
		manifestPriority: make(map[string]int),
		templateRegex:    regexp.MustCompile(`\{\{\s*([a-zA-Z][a-zA-Z0-9_-]*)\.(.*?)\s*}}`),
	}
}

// Build builds dependency graph using registered rules
func (gb *GraphBuilder) Build(manifests ...manifests.Manifest) (*GraphResultV2, error) {
	result := &GraphResultV2{
		Graph:             make(map[string][]string),
		Dependencies:      make([]rules.Dependency, 0),
		IntraManifestDeps: make(map[string][]rules.Dependency),
		SaveRequirements:  make(map[string]SaveRequirement),
		AliasToManifest:   make(map[string]string),
		TestCaseAliases:   make(map[string]TestCaseAliasInfo),
	}

	// Step 1: Initialize manifest priorities and collect aliases
	if err := gb.initializeManifests(manifests, result); err != nil {
		return nil, err
	}

	// Step 2: Analyze dependencies using all rules
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
func (gb *GraphBuilder) initializeManifests(manifests []manifests.Manifest, result *GraphResultV2) error {
	for _, manifest := range manifests {
		manifestID := manifest.GetID()

		// Set priority
		priority := gb.getManifestPriority(manifest)
		gb.manifestPriority[manifestID] = priority

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
func (gb *GraphBuilder) analyzeAllDependencies(manifests []manifests.Manifest) ([]rules.Dependency, error) {
	var allDependencies []rules.Dependency

	for _, manifest := range manifests {
		for _, rule := range gb.registry.GetRules() {
			if rule.CanHandle(manifest) {
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
func (gb *GraphBuilder) analyzeSmartTemplateDependencies(mans []manifests.Manifest, _ []rules.Dependency) ([]rules.Dependency, error) {
	var smartDeps []rules.Dependency
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
					smartDeps = append(smartDeps, rules.Dependency{
						From: manifestID,
						To:   targetManifestID,
						Type: rules.DependencyTypeTemplate,
						Metadata: rules.DependencyMetadata{
							Alias: alias,
							Paths: paths,
							Save:  true,
						},
					})
				} else if alias == "Values" {
					// Check if this is a reference to Values manifest
					for _, valuesManifest := range mans {
						if valuesManifest.GetKind() == manifests.ValuesKind {
							smartDeps = append(smartDeps, rules.Dependency{
								From: manifestID,
								To:   valuesManifest.GetID(),
								Type: rules.DependencyTypeValue,
								Metadata: rules.DependencyMetadata{
									Alias: alias,
									Paths: paths,
									Save:  true,
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
func (gb *GraphBuilder) extractAllTemplateReferences(httpTest *api.Http) []rules.TemplateReference {
	var references []rules.TemplateReference

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
func (gb *GraphBuilder) findTemplateReferencesInString(str string) []rules.TemplateReference {
	var references []rules.TemplateReference
	matches := gb.templateRegex.FindAllStringSubmatch(str, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			references = append(references, rules.TemplateReference{
				Alias: match[1],
				Path:  match[2],
			})
		}
	}

	return references
}

// findTemplateReferencesInValue recursively finds template references in any value
func (gb *GraphBuilder) findTemplateReferencesInValue(value any) []rules.TemplateReference {
	var references []rules.TemplateReference

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
func (gb *GraphBuilder) categorizeDependencies(allDependencies []rules.Dependency, result *GraphResultV2) {
	for _, dep := range allDependencies {
		fromManifest := gb.getBaseManifestID(dep.From)
		toManifest := gb.getBaseManifestID(dep.To)

		if fromManifest == toManifest {
			// Intra-manifest dependency
			if result.IntraManifestDeps[fromManifest] == nil {
				result.IntraManifestDeps[fromManifest] = make([]rules.Dependency, 0)
			}
			result.IntraManifestDeps[fromManifest] = append(result.IntraManifestDeps[fromManifest], dep)
		} else {
			// Inter-manifest dependency
			result.Dependencies = append(result.Dependencies, dep)
		}
	}
}

// buildAdjacencyGraph builds the adjacency graph from dependencies
func (gb *GraphBuilder) buildAdjacencyGraph(result *GraphResultV2) {
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
func (gb *GraphBuilder) calculateSaveRequirements(result *GraphResultV2) {
	// Process inter-manifest dependencies
	for _, dep := range result.Dependencies {
		if dep.Type == rules.DependencyTypeTemplate {
			toManifest := gb.getBaseManifestID(dep.To)

			// Update save requirement for the target manifest
			req := result.SaveRequirements[toManifest]
			req.Required = true
			req.Consumers = append(req.Consumers, dep.From)

			req.RequiredPaths = append(req.RequiredPaths, dep.Metadata.Paths...)
			req.Paths = append(req.Paths, dep.Metadata.Paths...)

			result.SaveRequirements[toManifest] = req

			// Update test case alias info if applicable
			if aliasInfo, exists := result.TestCaseAliases[dep.Metadata.Alias]; exists {
				aliasInfo.Consumers = append(aliasInfo.Consumers, dep.From)
				aliasInfo.RequiredPaths = append(aliasInfo.RequiredPaths, dep.Metadata.Paths...)
				result.TestCaseAliases[dep.Metadata.Alias] = aliasInfo
			}
		}
	}

	// Process intra-manifest dependencies
	for manifestID, deps := range result.IntraManifestDeps {
		req := result.SaveRequirements[manifestID]
		req.Required = true
		req.Consumers = append(req.Consumers, manifestID) // Self-consumer

		for _, dep := range deps {
			req.RequiredPaths = append(req.RequiredPaths, dep.Metadata.Paths...)
			req.Paths = append(req.Paths, dep.Metadata.Paths...)
		}

		result.SaveRequirements[manifestID] = req
	}
}

// buildExecutionOrder creates topologically sorted execution order
func (gb *GraphBuilder) buildExecutionOrder(manifests []manifests.Manifest, dependencies []rules.Dependency) ([]string, error) {
	// Initialize in-degree count for each manifest
	inDegree := make(map[string]int)
	for _, manifest := range manifests {
		id := manifest.GetID()
		inDegree[id] = 0
	}

	// Calculate in-degrees from all inter-manifest dependencies
	for _, dep := range dependencies {
		fromBase := gb.getBaseManifestID(dep.From)
		toBase := gb.getBaseManifestID(dep.To)
		if fromBase != toBase {
			if _, exists := inDegree[fromBase]; exists && dep.Type != rules.DependencyTypeTemplate {
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
func (gb *GraphBuilder) getManifestPriority(manifest manifests.Manifest) int {
	kind := manifest.GetKind()

	// Find kind priority rule
	for _, rule := range gb.registry.GetRules() {
		if rule.Name() == rules.KindPriorityRuleName {
			return rule.(*rules.KindPriorityRule).GetKindPriority(kind)
		}
	}

	return kinds.PriorityMap[kind]
}

// getBaseManifestID extracts base manifest ID from potentially extended ID
func (gb *GraphBuilder) getBaseManifestID(id string) string {
	// Remove any suffix after # (for test case aliases)
	if idx := strings.Index(id, "#"); idx != -1 {
		return id[:idx]
	}
	return id
}
