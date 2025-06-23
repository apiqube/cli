package depends

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/apiqube/cli/internal/core/manifests"
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
	From     string         // Source manifest ID
	To       string         // Target manifest ID (what we depend on)
	Type     DependencyType // Type of dependency
	Metadata map[string]any // Additional metadata (e.g., what data to save)
}

type DependencyType string

const (
	DependencyTypeExplicit DependencyType = "explicit" // From dependsOn field
	DependencyTypeTemplate DependencyType = "template" // From template references
	DependencyTypeValue    DependencyType = "value"    // From value passing
)

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

// ExplicitDependencyRule handles explicit dependsOn declarations
type ExplicitDependencyRule struct{}

func NewExplicitDependencyRule() *ExplicitDependencyRule {
	return &ExplicitDependencyRule{}
}

func (r *ExplicitDependencyRule) Name() string {
	return "explicit"
}

func (r *ExplicitDependencyRule) CanHandle(manifest manifests.Manifest) bool {
	_, ok := manifest.(manifests.Dependencies)
	return ok
}

func (r *ExplicitDependencyRule) AnalyzeDependencies(manifest manifests.Manifest) ([]Dependency, error) {
	dep, ok := manifest.(manifests.Dependencies)
	if !ok {
		return nil, nil
	}

	var dependencies []Dependency
	fromID := manifest.GetID()

	for _, toID := range dep.GetDependsOn() {
		if toID == fromID {
			return nil, fmt.Errorf("manifest %s cannot depend on itself", fromID)
		}

		dependencies = append(dependencies, Dependency{
			From: fromID,
			To:   toID,
			Type: DependencyTypeExplicit,
		})
	}

	return dependencies, nil
}

func (r *ExplicitDependencyRule) GetPriority() int {
	return 100 // Highest priority for explicit dependencies
}

// TemplateDependencyRule handles template-based dependencies ({{ alias.path }})
type TemplateDependencyRule struct {
	templateRegex *regexp.Regexp
}

func NewTemplateDependencyRule() *TemplateDependencyRule {
	// Regex to match {{ alias.path }} patterns
	regex := regexp.MustCompile(`\{\{\s*([a-zA-Z][a-zA-Z0-9_-]*)\.(.*?)\s*}}`)
	return &TemplateDependencyRule{
		templateRegex: regex,
	}
}

func (r *TemplateDependencyRule) Name() string {
	return "template"
}

func (r *TemplateDependencyRule) CanHandle(_ manifests.Manifest) bool {
	// This rule can handle any manifest, we'll check content during analysis
	return true
}

func (r *TemplateDependencyRule) AnalyzeDependencies(manifest manifests.Manifest) ([]Dependency, error) {
	var dependencies []Dependency
	fromID := manifest.GetID()

	// Extract all template references from the manifest
	references := r.extractTemplateReferences(manifest)

	// Group by alias to avoid duplicates and collect required paths
	aliasData := make(map[string][]string)
	for _, ref := range references {
		aliasData[ref.Alias] = append(aliasData[ref.Alias], ref.Path)
	}

	// Create dependencies with metadata about what data is needed
	for alias, paths := range aliasData {
		// Convert alias to full manifest ID (assuming same namespace for now)
		// This might need to be more sophisticated based on your ID scheme
		toID := r.resolveAliasToID(manifest, alias)

		dependencies = append(dependencies, Dependency{
			From: fromID,
			To:   toID,
			Type: DependencyTypeTemplate,
			Metadata: map[string]any{
				"alias":          alias,
				"required_paths": paths,
				"save_required":  true,
			},
		})
	}

	return dependencies, nil
}

func (r *TemplateDependencyRule) GetPriority() int {
	return 50 // Medium priority for template dependencies
}

// TemplateReference represents a parsed template reference
type TemplateReference struct {
	Alias string // The alias part (e.g., "users-list")
	Path  string // The path part (e.g., "response.body.data[0].id")
}

func (r *TemplateDependencyRule) extractTemplateReferences(manifest manifests.Manifest) []TemplateReference {
	var references []TemplateReference

	// Convert manifest to string representation for parsing
	// This is a simplified approach - in real implementation you might want
	// to traverse the structure more carefully
	manifestStr := fmt.Sprintf("%+v", manifest)

	matches := r.templateRegex.FindAllStringSubmatch(manifestStr, -1)
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

func (r *TemplateDependencyRule) resolveAliasToID(manifest manifests.Manifest, alias string) string {
	// Simple resolution: assume same namespace and kind as current manifest
	// Format: namespace.kind.alias
	parts := strings.Split(manifest.GetID(), ".")
	if len(parts) >= 2 {
		return fmt.Sprintf("%s.%s.%s", parts[0], parts[1], alias)
	}
	return alias
}

// KindPriorityRule handles kind-based priorities
type KindPriorityRule struct {
	priorities map[string]int
}

var priorities = map[string]int{
	manifests.ValuesKind:       1,
	manifests.ServerKind:       10,
	manifests.ServiceKind:      20,
	manifests.HttpTestKind:     30,
	manifests.HttpLoadTestKind: 40,
}

func NewKindPriorityRule() *KindPriorityRule {
	return &KindPriorityRule{
		priorities: priorities,
	}
}

func (r *KindPriorityRule) Name() string {
	return "kind_priority"
}

func (r *KindPriorityRule) CanHandle(_ manifests.Manifest) bool {
	return true // Can handle any manifest for priority assignment
}

func (r *KindPriorityRule) AnalyzeDependencies(_ manifests.Manifest) ([]Dependency, error) {
	// This rule doesn't create dependencies, just provides priority info
	return nil, nil
}

func (r *KindPriorityRule) GetPriority() int {
	return 0 // Lowest priority as this is just for ordering
}

func (r *KindPriorityRule) GetKindPriority(kind string) int {
	if priority, ok := r.priorities[kind]; ok {
		return priority
	}
	return 0
}

// DefaultRuleRegistry creates a registry with default rules
func DefaultRuleRegistry() *RuleRegistry {
	registry := NewRuleRegistry()

	// Register default rules
	registry.Register(NewExplicitDependencyRule())
	registry.Register(NewTemplateDependencyRule())
	registry.Register(NewKindPriorityRule())
	registry.Register(NewHttpTestDependencyRule())

	return registry
}
