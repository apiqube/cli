package depends

import (
	"fmt"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
	"github.com/apiqube/cli/internal/core/manifests/kinds/servers"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/tests/api"
	"github.com/apiqube/cli/internal/core/manifests/kinds/values"
	"github.com/apiqube/cli/internal/core/manifests/utils"
	"net/http"
	"strings"
	"testing"

	"github.com/apiqube/cli/internal/core/manifests"
)

// TestGraphBuilder tests the graph building functionality
func TestGraphBuilder(t *testing.T) {
	fmt.Println("🧪 Running Graph Builder Tests")
	fmt.Println(strings.Repeat("=", 10))

	// Test Case 1: Simple manifests without dependencies
	t.Run("Simple manifests without dependencies", func(t *testing.T) {
		mans := []manifests.Manifest{
			&api.Http{
				BaseManifest: kinds.BaseManifest{
					Version: "v1",
					Kind:    manifests.HttpTestKind,
					Metadata: kinds.Metadata{
						Name:      "http-test",
						Namespace: manifests.DefaultNamespace,
					},
				},
				Spec: struct {
					Target string         `yaml:"target,omitempty" json:"target,omitempty" validate:"required"`
					Cases  []api.HttpCase `yaml:"cases" json:"cases" validate:"required,min=1,max=100,dive"`
				}{
					Target: "http://127.0.0.1:8080",
					Cases: []api.HttpCase{
						{
							HttpCase: tests.HttpCase{
								Name:     "Simple HTTP Test Case 1",
								Method:   http.MethodGet,
								Endpoint: "/api",
							},
						},
					},
				},
			},
		}

		registry := DefaultRuleRegistry()
		builder := NewGraphBuilderV2(registry)

		result, err := builder.BuildGraphWithRules(mans)
		if err != nil {
			t.Fatalf("Failed to build graph: %v", err)
		}

		// Print results
		printDependencyGraph(builder, result)

		// Assertions
		if len(result.ExecutionOrder) != 1 {
			t.Errorf("Expected 1 manifest in execution order, got %d", len(result.ExecutionOrder))
		}

		// Values should come first due to higher priority
		if result.ExecutionOrder[0] != mans[0].GetID() {
			t.Errorf("Expected %s manifest id, got %s", mans[0].GetID(), result.ExecutionOrder[0])
		}
	})

	// Test Case 2: Manifests with explicit dependencies
	t.Run("Manifests with explicit dependencies", func(t *testing.T) {
		fmt.Println("\n📝 Test Case 2: Manifests with explicit dependencies")

		mans := []manifests.Manifest{
			&values.Values{
				BaseManifest: kinds.BaseManifest{
					Version: "v1",
					Kind:    manifests.ValuesKind,
					Metadata: kinds.Metadata{
						Name:      "values-test",
						Namespace: manifests.DefaultNamespace,
					},
				},
				Spec: struct {
					Data map[string]any `yaml:",inline" json:",inline" validate:"required,min=1,dive"`
				}{
					Data: map[string]any{
						"data_1": 1,
						"data_2": []int{1, 2, 3},
					},
				},
			},
			&api.Http{
				BaseManifest: kinds.BaseManifest{
					Version: "v1",
					Kind:    manifests.HttpTestKind,
					Metadata: kinds.Metadata{
						Name:      "http-test",
						Namespace: manifests.DefaultNamespace,
					},
				},
				Spec: struct {
					Target string         `yaml:"target,omitempty" json:"target,omitempty" validate:"required"`
					Cases  []api.HttpCase `yaml:"cases" json:"cases" validate:"required,min=1,max=100,dive"`
				}{
					Target: "http://127.0.0.1:8080",
					Cases: []api.HttpCase{
						{
							HttpCase: tests.HttpCase{
								Name:     "Simple HTTP Test Case 1",
								Method:   http.MethodGet,
								Endpoint: "/api",
							},
						},
					},
				},
				Dependencies: kinds.Dependencies{
					DependsOn: []string{
						utils.FormManifestID(manifests.DefaultNamespace, manifests.ValuesKind, "values-test"),
					},
				},
			},
		}

		registry := DefaultRuleRegistry()
		builder := NewGraphBuilderV2(registry)

		result, err := builder.BuildGraphWithRules(mans)
		if err != nil {
			t.Fatalf("Failed to build graph: %v", err)
		}

		// Print results
		printDependencyGraph(builder, result)

		// Assertions
		if len(result.Dependencies) != 1 {
			t.Errorf("Expected 1 dependency, got %d", len(result.Dependencies))
		}

		// Values should come first
		if result.ExecutionOrder[0] != mans[0].GetID() {
			t.Errorf("Expected Values manifest first, got %s", result.ExecutionOrder[0])
		}
	})

	// Test Case 3: Single manifest with intra-manifest dependencies (your case)
	t.Run("Single manifest with intra-manifest dependencies", func(t *testing.T) {
		fmt.Println("\n📝 Test Case 3: Single manifest with intra-manifest dependencies")

		// Create a mock HTTP test with internal test case dependencies
		mans := []manifests.Manifest{
			&api.Http{
				BaseManifest: kinds.BaseManifest{
					Version: "v1",
					Kind:    manifests.HttpTestKind,
					Metadata: kinds.Metadata{
						Name:      "http-test",
						Namespace: manifests.DefaultNamespace,
					},
				},
				Spec: struct {
					Target string         `yaml:"target,omitempty" json:"target,omitempty" validate:"required"`
					Cases  []api.HttpCase `yaml:"cases" json:"cases" validate:"required,min=1,max=100,dive"`
				}{
					Target: "http://127.0.0.1:8080",
					Cases: []api.HttpCase{
						{
							HttpCase: tests.HttpCase{
								Name:     "Simple HTTP Test Case",
								Alias:    stringPtr("fetch-users"),
								Method:   http.MethodGet,
								Endpoint: "/users",
							},
						},
						{
							HttpCase: tests.HttpCase{
								Name:     "HTTP Test Case With Internal Dependencies",
								Method:   http.MethodDelete,
								Endpoint: "/users/{{ fetch-users.response.body.users.0.id }}",
								Body: map[string]any{
									"name": "{{ fetch-users.response.body.users.0.name }}",
								},
							},
						},
					},
				},
			},
		}

		registry := DefaultRuleRegistry()
		builder := NewGraphBuilderV2(registry)

		result, err := builder.BuildGraphWithRules(mans)
		if err != nil {
			t.Fatalf("Failed to build graph: %v", err)
		}

		// Print results
		printDependencyGraph(builder, result)

		// Assertions
		if len(result.ExecutionOrder) != 1 {
			t.Errorf("Expected 1 manifest in execution order, got %d", len(result.ExecutionOrder))
		}

		// Should have intra-manifest dependencies
		if len(result.IntraManifestDeps) == 0 {
			t.Errorf("Expected intra-manifest dependencies, got none")
		}

		// Should have no inter-manifest dependencies (no cycles)
		if len(result.Dependencies) != 1 {
			t.Errorf("Expected no inter-manifest dependencies, got %d", len(result.Dependencies))
		}
	})

	// Test Case 4: Multiple manifests with mixed dependencies
	t.Run("Multiple manifests with mixed dependencies", func(t *testing.T) {
		fmt.Println("\n📝 Test Case 4: Multiple manifests with mixed dependencies")

		mans := []manifests.Manifest{
			&values.Values{
				BaseManifest: kinds.BaseManifest{
					Version: "v1",
					Kind:    manifests.ValuesKind,
					Metadata: kinds.Metadata{
						Name:      "values-test",
						Namespace: manifests.DefaultNamespace,
					},
				},
				Spec: struct {
					Data map[string]any `yaml:",inline" json:",inline" validate:"required,min=1,dive"`
				}{
					Data: map[string]any{
						"data_1": 1,
						"data_2": []int{1, 2, 3},
						"user": struct {
							Name  string
							Email string
						}{
							Name:  "user_1",
							Email: "user_1@example.com",
						},
					},
				},
			},
			&servers.Server{
				BaseManifest: kinds.BaseManifest{
					Version: "v1",
					Kind:    manifests.ServerKind,
					Metadata: kinds.Metadata{
						Name:      "http-test-server",
						Namespace: manifests.DefaultNamespace,
					},
				},
				Spec: struct {
					BaseURL string            `yaml:"baseUrl" json:"baseUrl" validate:"required,url"`
					Health  string            `yaml:"health" json:"health" validate:"omitempty,max=100"`
					Headers map[string]string `yaml:"headers,omitempty" json:"headers" validate:"omitempty,max=20"`
				}{
					BaseURL: "http://127.0.0.1:8080",
					Health:  "",
					Headers: map[string]string{
						"Content-Type": "application/json",
					}},
			},
			&api.Http{
				BaseManifest: kinds.BaseManifest{
					Version: "v1",
					Kind:    manifests.HttpTestKind,
					Metadata: kinds.Metadata{
						Name:      "http-test-roles",
						Namespace: manifests.DefaultNamespace,
					},
				},
				Spec: struct {
					Target string         `yaml:"target,omitempty" json:"target,omitempty" validate:"required"`
					Cases  []api.HttpCase `yaml:"cases" json:"cases" validate:"required,min=1,max=100,dive"`
				}{
					Target: "http://127.0.0.1:8080",
					Cases: []api.HttpCase{
						{
							HttpCase: tests.HttpCase{
								Name:     "Simple HTTP Test Case",
								Alias:    stringPtr("users-roles"),
								Method:   http.MethodGet,
								Endpoint: "/roles",
							},
						},
					},
				},
				Dependencies: kinds.Dependencies{
					DependsOn: []string{
						utils.FormManifestID(manifests.DefaultNamespace, manifests.ServerKind, "http-test-server"),
					},
				},
			},
			&api.Http{
				BaseManifest: kinds.BaseManifest{
					Version: "v1",
					Kind:    manifests.HttpTestKind,
					Metadata: kinds.Metadata{
						Name:      "http-test-users",
						Namespace: manifests.DefaultNamespace,
					},
				},
				Spec: struct {
					Target string         `yaml:"target,omitempty" json:"target,omitempty" validate:"required"`
					Cases  []api.HttpCase `yaml:"cases" json:"cases" validate:"required,min=1,max=100,dive"`
				}{
					Target: "http://127.0.0.1:8080",
					Cases: []api.HttpCase{
						{
							HttpCase: tests.HttpCase{
								Name:     "Simple HTTP Test Case",
								Alias:    stringPtr("fetch-users"),
								Method:   http.MethodGet,
								Endpoint: "/users",
							},
						},
						{
							HttpCase: tests.HttpCase{
								Name:     "HTTP Test Case With Internal Dependencies",
								Method:   http.MethodDelete,
								Endpoint: "/users/{{ fetch-users.response.body.users.0.id }}",
								Body: map[string]any{
									"name": "{{ fetch-users.response.body.users.0.name }}",
								},
							},
						},
						{
							HttpCase: tests.HttpCase{
								Name:     "HTTP Test Case With Explicit Values Dependencies",
								Method:   http.MethodPost,
								Endpoint: "/users",
								Body: map[string]any{
									"user": "{{ Values.values-test.user }}",
									"role": "{{ users.roles.roles.3 }}",
								},
							},
						},
					},
				},
				Dependencies: kinds.Dependencies{
					DependsOn: []string{
						utils.FormManifestID(manifests.DefaultNamespace, manifests.ValuesKind, "values-test"),
						utils.FormManifestID(manifests.DefaultNamespace, manifests.ServerKind, "http-test-server"),
					},
				},
			},
		}

		registry := DefaultRuleRegistry()
		builder := NewGraphBuilderV2(registry)

		result, err := builder.BuildGraphWithRules(mans)
		if err != nil {
			t.Fatalf("Failed to build graph: %v", err)
		}

		// Print results
		printDependencyGraph(builder, result)

		// Assertions
		if len(result.ExecutionOrder) != 3 {
			t.Errorf("Expected 3 manifests in execution order, got %d", len(result.ExecutionOrder))
		}

		// Check execution order (Values -> Server -> HttpTest & HttpTest)
		expectedOrder := []string{
			mans[0].GetID(),
			mans[1].GetID(),
			mans[2].GetID(),
			mans[3].GetID(),
		}

		for i, expected := range expectedOrder {
			if result.ExecutionOrder[i] != expected {
				t.Errorf("Expected %s at position %d, got %s", expected, i, result.ExecutionOrder[i])
			}
		}
	})
}

/*
// BenchmarkGraphBuilder benchmarks the graph building performance
func BenchmarkGraphBuilder(b *testing.B) {
	// Create a large set of manifests for benchmarking
	var manifests []manifests.Manifest

	for i := 0; i < 100; i++ {
		manifests = append(manifests, &MockManifest{
			ID:   fmt.Sprintf("test.HttpTest.test-%d", i),
			Kind: manifests.HttpTestKind,
		})
	}

	registry := DefaultRuleRegistry()
	builder := NewGraphBuilderV2(registry)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := builder.BuildGraphWithRules(manifests)
		if err != nil {
			b.Fatalf("Failed to build graph: %v", err)
		}
	}
}
*/

// PrintDependencyGraph prints a beautiful visualization of the dependency graph
func printDependencyGraph(gb *GraphBuilderV2, result *GraphResultV2) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🔗 DEPENDENCY GRAPH VISUALIZATION")
	fmt.Println(strings.Repeat("=", 80))

	// Print execution order
	fmt.Println("\n📋 EXECUTION ORDER:")
	fmt.Println(strings.Repeat("-", 40))
	for i, manifestID := range result.ExecutionOrder {
		priority := gb.getManifestPriorityByID(manifestID)
		fmt.Printf("  %d. %s (priority: %d)\n", i+1, manifestID, priority)
	}

	// Print inter-manifest dependencies
	fmt.Println("\n🔄 INTER-MANIFEST DEPENDENCIES:")
	fmt.Println(strings.Repeat("-", 40))
	if len(result.Dependencies) == 0 {
		fmt.Println("  ✅ No inter-manifest dependencies found")
	} else {
		for _, dep := range result.Dependencies {
			fmt.Printf("  %s ──(%s)──> %s\n", dep.From, dep.Type, dep.To)
			if dep.Metadata != nil {
				for key, value := range dep.Metadata {
					fmt.Printf("    └─ %s: %v\n", key, value)
				}
			}
		}
	}

	// Print intra-manifest dependencies
	fmt.Println("\n🏠 INTRA-MANIFEST DEPENDENCIES:")
	fmt.Println(strings.Repeat("-", 40))
	if len(result.IntraManifestDeps) == 0 {
		fmt.Println("  ✅ No intra-manifest dependencies found")
	} else {
		for manifestID, deps := range result.IntraManifestDeps {
			fmt.Printf("  📦 %s:\n", manifestID)
			for _, dep := range deps {
				fmt.Printf("    %s ──(%s)──> %s\n", dep.From, dep.Type, dep.To)
				if dep.Metadata != nil {
					for key, value := range dep.Metadata {
						fmt.Printf("      └─ %s: %v\n", key, value)
					}
				}
			}
		}
	}

	// Print save requirements
	fmt.Println("\n💾 SAVE REQUIREMENTS:")
	fmt.Println(strings.Repeat("-", 40))
	hasRequirements := false
	for manifestID, req := range result.SaveRequirements {
		if req.Required {
			hasRequirements = true
			fmt.Printf("  📁 %s:\n", manifestID)
			fmt.Printf("    └─ Paths: %v\n", req.RequiredPaths)
			fmt.Printf("    └─ Used by: %v\n", req.Consumers)
		}
	}
	if !hasRequirements {
		fmt.Println("  ✅ No save requirements found")
	}

	// Print graph adjacency list
	fmt.Println("\n🕸️  ADJACENCY GRAPH:")
	fmt.Println(strings.Repeat("-", 40))
	for head, tails := range result.Graph {
		if len(tails) > 0 {
			fmt.Printf("  %s:\n", head)
			for _, tail := range tails {
				fmt.Printf("    └─ %s\n", tail)
			}
		} else {
			fmt.Printf("  %s: (no dependencies)\n", head)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
