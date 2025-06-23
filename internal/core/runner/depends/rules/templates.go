package rules

import (
	"fmt"
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/utils"
	"regexp"
)

const TemplateRuleName = "template"

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
	return TemplateRuleName
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
			Metadata: DependencyMetadata{
				Alias: alias,
				Paths: paths,
				Save:  true,
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
	// Format: namespace.kind.name#alias
	namespace, kind, name := utils.ParseManifestID(manifest.GetID())
	return utils.FormManifestID(namespace, kind, fmt.Sprintf("%s#%s", name, alias))
}
