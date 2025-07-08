package rules

import "github.com/apiqube/cli/internal/core/manifests"

const (
	DependencyTypeTemplate DependencyType = "template" // From template references
	DependencyTypeValue    DependencyType = "values"   // From value passing
)

// DependencyRule defines interface for dependency analysis rules
type DependencyRule interface {
	// Name returns the rule name for debugging
	Name() string

	// AnalyzeDependencies extracts dependencies from manifest
	AnalyzeDependencies(manifest manifests.Manifest) ([]Dependency, error)

	// GetPriority returns priority for this type of dependency
	GetPriority() int

	// CanHandle checks if this rule can handle the given manifest
	CanHandle(manifest manifests.Manifest) bool
}

// Dependency represents a dependency relationship
type Dependency struct {
	From     string             // Source manifest ID
	To       string             // Target manifest ID (what we depend on)
	Type     DependencyType     // Type of dependency
	Metadata DependencyMetadata // Additional metadata (e.g., what data to save)
}

type DependencyType string

type DependencyMetadata struct {
	Alias        string
	Paths        []string
	Locations    []string
	Save         bool
	CaseName     string
	ManifestKind string
}

// RuleRegistry manages dependency rules
type RuleRegistry struct {
	rules []DependencyRule
}

func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{
		rules: make([]DependencyRule, 0),
	}
}

func (r *RuleRegistry) Register(rule DependencyRule) {
	r.rules = append(r.rules, rule)
}

func (r *RuleRegistry) GetRules() []DependencyRule {
	return r.rules
}

// DefaultRuleRegistry creates a registry with default rules
func DefaultRuleRegistry() *RuleRegistry {
	registry := NewRuleRegistry()

	// Register default rules
	registry.Register(NewKindPriorityRule())
	registry.Register(NewTemplateDependencyRule())
	registry.Register(NewHttpTestDependencyRule())

	return registry
}
