package rules

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/api"
)

const HttpTestDependencyRuleName = "Http Test"

// HttpTestDependencyRule handles HTTP test specific dependencies
type HttpTestDependencyRule struct {
	templateRegex *regexp.Regexp
}

func NewHttpTestDependencyRule() *HttpTestDependencyRule {
	// Enhanced regex to capture various template patterns
	regex := regexp.MustCompile(`\{\{\s*([a-zA-Z][a-zA-Z0-9_-]*)\.(.*?)\s*}}`)
	return &HttpTestDependencyRule{
		templateRegex: regex,
	}
}

func (r *HttpTestDependencyRule) Name() string {
	return HttpTestDependencyRuleName
}

func (r *HttpTestDependencyRule) CanHandle(manifest manifests.Manifest) bool {
	return manifest.GetKind() == manifests.HttpTestKind
}

func (r *HttpTestDependencyRule) AnalyzeDependencies(manifest manifests.Manifest) ([]Dependency, error) {
	httpTest, ok := manifest.(*api.Http)
	var dependencies []Dependency
	if !ok {
		return dependencies, nil
	}

	fromID := manifest.GetID()

	// Analyze each test case
	for _, testCase := range httpTest.Spec.Cases {
		caseDeps := r.analyzeTestCase(fromID, testCase, httpTest)
		dependencies = append(dependencies, caseDeps...)
	}

	return dependencies, nil
}

func (r *HttpTestDependencyRule) GetPriority() int {
	return 60 // Higher than generic template rule
}

// analyzeTestCase analyzes a single HTTP test case for dependencies
func (r *HttpTestDependencyRule) analyzeTestCase(manifestID string, testCase api.HttpCase, httpTest *api.Http) []Dependency {
	var dependencies []Dependency

	// Collect all template references from the test case
	references := r.extractAllReferences(testCase)

	// Group references by alias
	aliasRefs := make(map[string][]HttpTemplateReference)
	for _, ref := range references {
		aliasRefs[ref.Alias] = append(aliasRefs[ref.Alias], ref)
	}

	// Create dependencies for each referenced alias
	for alias, refs := range aliasRefs {
		toID := r.resolveAliasToManifestID(httpTest, alias)

		// Collect all required paths for this alias
		var requiredPaths []string
		var locations []string

		for _, ref := range refs {
			requiredPaths = append(requiredPaths, ref.Path)
			locations = append(locations, ref.Location)
		}

		dependency := Dependency{
			From: manifestID,
			To:   toID,
			Type: DependencyTypeTemplate,
			Metadata: DependencyMetadata{
				Alias:        alias,
				Paths:        requiredPaths,
				Locations:    locations,
				Save:         true,
				CaseName:     testCase.Name,
				ManifestKind: httpTest.GetKind(),
			},
		}

		dependencies = append(dependencies, dependency)
	}

	return dependencies
}

// HttpTemplateReference represents a template reference in HTTP test
type HttpTemplateReference struct {
	Alias    string // e.g., "users-list"
	Path     string // e.g., "response.body.data[0].id"
	Location string // where it was found (e.g., "body.user_id", "endpoint")
}

// extractAllReferences finds all template references in a test case
func (r *HttpTestDependencyRule) extractAllReferences(testCase api.HttpCase) []HttpTemplateReference {
	var references []HttpTemplateReference

	// Check endpoint
	if refs := r.findReferencesInString(testCase.Endpoint, "endpoint"); len(refs) > 0 {
		references = append(references, refs...)
	}

	// Check URL
	if refs := r.findReferencesInString(testCase.Url, "url"); len(refs) > 0 {
		references = append(references, refs...)
	}

	// Check headers
	for key, value := range testCase.Headers {
		location := fmt.Sprintf("headers.%s", key)
		if refs := r.findReferencesInString(value, location); len(refs) > 0 {
			references = append(references, refs...)
		}
	}

	// Check body (recursively)
	if testCase.Body != nil {
		bodyRefs := r.findReferencesInValue(testCase.Body, "body")
		references = append(references, bodyRefs...)
	}

	// Check assertions
	for i, assert := range testCase.Assert {
		location := fmt.Sprintf("assert[%d]", i)
		if assert.Template != "" {
			if refs := r.findReferencesInString(assert.Template, location+".template"); len(refs) > 0 {
				references = append(references, refs...)
			}
		}
	}

	return references
}

// findReferencesInString extracts template references from a string
func (r *HttpTestDependencyRule) findReferencesInString(str, location string) []HttpTemplateReference {
	var references []HttpTemplateReference

	matches := r.templateRegex.FindAllStringSubmatch(str, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			references = append(references, HttpTemplateReference{
				Alias:    match[1],
				Path:     match[2],
				Location: location,
			})
		}
	}

	return references
}

// findReferencesInValue recursively finds references in any value type
func (r *HttpTestDependencyRule) findReferencesInValue(value any, location string) []HttpTemplateReference {
	var references []HttpTemplateReference

	switch v := value.(type) {
	case string:
		references = append(references, r.findReferencesInString(v, location)...)
	case map[string]any:
		for key, val := range v {
			newLocation := fmt.Sprintf("%s.%s", location, key)
			references = append(references, r.findReferencesInValue(val, newLocation)...)
		}
	case []any:
		for i, val := range v {
			newLocation := fmt.Sprintf("%s[%d]", location, i)
			references = append(references, r.findReferencesInValue(val, newLocation)...)
		}
	case map[any]any:
		// Handle interface{} maps
		for key, val := range v {
			keyStr := fmt.Sprintf("%v", key)
			newLocation := fmt.Sprintf("%s.%s", location, keyStr)
			references = append(references, r.findReferencesInValue(val, newLocation)...)
		}
	}

	return references
}

// resolveAliasToManifestID converts test case alias to full manifest ID
func (r *HttpTestDependencyRule) resolveAliasToManifestID(httpTest *api.Http, alias string) string {
	// For HTTP tests, we assume the alias refers to another test case in the same manifest
	// Format: namespace.kind.name#alias
	baseID := httpTest.GetID()
	return fmt.Sprintf("%s#%s", baseID, alias)
}

// IntraManifestDependencyRule handles dependencies within the same manifest
type IntraManifestDependencyRule struct{}

func NewIntraManifestDependencyRule() *IntraManifestDependencyRule {
	return &IntraManifestDependencyRule{}
}

func (r *IntraManifestDependencyRule) Name() string {
	return "intra_manifest"
}

func (r *IntraManifestDependencyRule) CanHandle(manifest manifests.Manifest) bool {
	_, ok := manifest.(*api.Http)
	return ok
}

func (r *IntraManifestDependencyRule) AnalyzeDependencies(manifest manifests.Manifest) ([]Dependency, error) {
	httpTest, ok := manifest.(*api.Http)
	if !ok {
		return nil, nil
	}

	var dependencies []Dependency
	manifestID := manifest.GetID()

	// Create a map of aliases to test cases
	aliasToCase := make(map[string]api.HttpCase)
	for _, testCase := range httpTest.Spec.Cases {
		if testCase.Alias != nil {
			aliasToCase[*testCase.Alias] = testCase
		}
	}

	// Analyze dependencies between test cases
	for i, testCase := range httpTest.Spec.Cases {
		caseID := fmt.Sprintf("%s#case_%d", manifestID, i)
		if testCase.Alias != nil {
			caseID = fmt.Sprintf("%s#%s", manifestID, *testCase.Alias)
		}

		// Find references to other test cases
		references := r.findIntraManifestReferences(testCase)

		for _, ref := range references {
			if _, exists := aliasToCase[ref.Alias]; exists {
				depID := fmt.Sprintf("%s#%s", manifestID, ref.Alias)

				dependency := Dependency{
					From: caseID,
					To:   depID,
					Type: DependencyTypeValue,
					Metadata: DependencyMetadata{
						Alias: ref.Alias,
						Paths: []string{ref.Path},
						Save:  true,
					},
				}

				dependencies = append(dependencies, dependency)
			}
		}
	}

	return dependencies, nil
}

func (r *IntraManifestDependencyRule) GetPriority() int {
	return 70 // Higher priority for intra-manifest dependencies
}

// findIntraManifestReferences finds references to other test cases within the same manifest
func (r *IntraManifestDependencyRule) findIntraManifestReferences(testCase api.HttpCase) []HttpTemplateReference {
	var references []HttpTemplateReference

	// Use reflection to traverse all string fields
	r.findReferencesInStruct(reflect.ValueOf(testCase), "", &references)

	return references
}

// findReferencesInStruct recursively finds template references in struct fields
func (r *IntraManifestDependencyRule) findReferencesInStruct(v reflect.Value, path string, references *[]HttpTemplateReference) {
	templateRegex := regexp.MustCompile(`\{\{\s*([a-zA-Z][a-zA-Z0-9_-]*)\.(.*?)\s*}}`)
	var newPath string

	switch v.Kind() {
	case reflect.String:
		str := v.String()
		matches := templateRegex.FindAllStringSubmatch(str, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				*references = append(*references, HttpTemplateReference{
					Alias:    match[1],
					Path:     match[2],
					Location: path,
				})
			}
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			if path != "" {
				newPath = fmt.Sprintf("%s.%s", path, keyStr)
			} else {
				newPath = keyStr
			}
			r.findReferencesInStruct(v.MapIndex(key), newPath, references)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			newPath = fmt.Sprintf("%s[%d]", path, i)
			r.findReferencesInStruct(v.Index(i), newPath, references)
		}
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if field.IsExported() {
				if path != "" {
					newPath = fmt.Sprintf("%s.%s", path, field.Name)
				} else {
					newPath = field.Name
				}
				r.findReferencesInStruct(v.Field(i), newPath, references)
			}
		}
	case reflect.Ptr, reflect.Interface:
		if !v.IsNil() {
			r.findReferencesInStruct(v.Elem(), path, references)
		}
	default:
		// Skip unsupported types
	}
}
