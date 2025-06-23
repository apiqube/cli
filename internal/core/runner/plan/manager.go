package plan

import (
	"fmt"
	"sort"
	"strings"

	"github.com/apiqube/cli/internal/operations"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds/plan"
	"github.com/apiqube/cli/internal/core/manifests/utils"
	"github.com/apiqube/cli/internal/core/runner/depends"
)

var kindPriority = map[string]int{
	"Values":       0,
	"Server":       10,
	"Service":      20,
	"HttpTest":     30,
	"HttpLoadTest": 40,
}

const defaultPriority = 10_000

type Manager interface {
	Generate() (*plan.Plan, error)
	GenerateV2() (*plan.Plan, *depends.Result, error)
	CheckPlan(*plan.Plan) error
}

type basicManager struct {
	manifests  map[string]manifests.Manifest
	mode       string
	stableSort bool
	parallel   bool
}

func (g *basicManager) CheckPlan(pln *plan.Plan) error {
	stageOrder := make(map[string]int)
	seen := make(map[string]bool)

	for i, stage := range pln.Spec.Stages {
		for j, id := range stage.Manifests {
			namespace, kind, name, err := utils.ParseManifestIDWithError(id)
			if err != nil {
				return err
			}

			id = utils.FormManifestID(namespace, kind, name)
			stage.Manifests[j] = id

			if seen[id] {
				return fmt.Errorf("manifest %s occurs more than once in plans", id)
			}

			if _, ok := g.manifests[id]; !ok {
				return fmt.Errorf("in the privided plan lacks the stated manifest: %s", id)
			}

			stageOrder[id] = i
			seen[id] = true
		}
	}

	for id, m := range g.manifests {
		dep, ok := m.(manifests.Dependencies)
		if !ok {
			continue
		}

		deps := dep.GetDependsOn()
		for _, depID := range deps {
			i, okI := stageOrder[id]
			j, okJ := stageOrder[depID]

			if !okI || !okJ {
				return fmt.Errorf("manifest '%s' or its dependency '%s' not found in the plan", id, depID)
			}

			if j >= i {
				return fmt.Errorf(
					"dependency order is broken: '%s' (stage %d) depends on '%s' (stage %d), but is located earlier or in the same stage",
					id, i, depID, j,
				)
			}
		}
	}

	return nil
}

func (g *basicManager) Generate() (*plan.Plan, error) {
	if len(g.manifests) == 0 {
		return nil, fmt.Errorf("manifests not provided for generating the plan")
	}

	graph := newDepGraph()

	for id, m := range g.manifests {
		if m.GetKind() == manifests.PlanKind {
			delete(g.manifests, id)
			continue
		}

		graph.addNode(id)

		if depend, ok := m.(manifests.Dependencies); ok {
			deps := depend.GetDependsOn()
			for _, depId := range deps {
				if _, found := g.manifests[depId]; !found {
					return nil, fmt.Errorf("manifest '%s' depends on '%s', but it was not found in the manifest set", id, depId)
				}

				graph.addEdge(depId, id)
			}
		}
	}

	sorted, err := graph.topoSort()
	if err != nil {
		return nil, err
	}

	stages := groupByLayers(sorted, g.manifests, g.mode, g.stableSort, g.parallel)

	var newPlan plan.Plan
	newPlan.Default()

	newPlan.Spec.Stages = stages

	planData, err := operations.NormalizeYAML(&newPlan)
	if err != nil {
		return nil, fmt.Errorf("fail while generating plan hash: %v", err)
	}

	planHash, err := utils.CalculateContentHash(planData)
	if err != nil {
		return nil, fmt.Errorf("fail while calculation plan hash: %v", err)
	}

	meta := newPlan.GetMeta()
	meta.SetHash(planHash)
	meta.SetCreatedBy("plan-generator")

	return &newPlan, nil
}

// GenerateV2 generates plan using the new V2 dependency system
func (g *basicManager) GenerateV2() (*plan.Plan, *depends.Result, error) {
	if len(g.manifests) == 0 {
		return nil, nil, fmt.Errorf("manifests not provided for generating the plan")
	}

	// Convert map to slice for V2 system
	var manifestSlice []manifests.Manifest
	for _, m := range g.manifests {
		if m.GetKind() != manifests.PlanKind {
			manifestSlice = append(manifestSlice, m)
		}
	}

	// Create rule registry with default rules
	registry := depends.DefaultRuleRegistry()

	// Build graph using V2 system
	builder := depends.NewGraphBuilder(registry)
	graphResult, err := builder.Build(manifestSlice)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// Convert execution order to stages
	stages := g.createStagesFromExecutionOrder(graphResult.ExecutionOrder, g.manifests)

	// Create plan
	var newPlan plan.Plan
	newPlan.Default()
	newPlan.Spec.Stages = stages

	// Generate hash
	planData, err := operations.NormalizeYAML(&newPlan)
	if err != nil {
		return nil, nil, fmt.Errorf("fail while generating plan hash: %v", err)
	}

	planHash, err := utils.CalculateContentHash(planData)
	if err != nil {
		return nil, nil, fmt.Errorf("fail while calculation plan hash: %v", err)
	}

	meta := newPlan.GetMeta()
	meta.SetHash(planHash)
	meta.SetCreatedBy("plan-generator-v2")

	return &newPlan, graphResult, nil
}

// createStagesFromExecutionOrder creates stages from V2 execution order
func (g *basicManager) createStagesFromExecutionOrder(executionOrder []string, manifests map[string]manifests.Manifest) []plan.Stage {
	// Group manifests by kind and dependency level
	var stages []plan.Stage
	var current []string
	var currentKind string

	for _, id := range executionOrder {
		m, exists := manifests[id]
		if !exists {
			continue
		}

		kind := m.GetKind()

		// If kind changes, create a new stage
		if currentKind != "" && kind != currentKind && len(current) > 0 {
			stages = append(stages, makeStage(current, manifests, g.mode, g.stableSort, g.parallel))
			current = []string{}
		}

		current = append(current, id)
		currentKind = kind
	}

	// Add remaining manifests
	if len(current) > 0 {
		stages = append(stages, makeStage(current, manifests, g.mode, g.stableSort, g.parallel))
	}

	return stages
}

func groupByLayers(sorted []string, mans map[string]manifests.Manifest, mode string, stable, parallel bool) []plan.Stage {
	sort.SliceStable(sorted, func(i, j int) bool {
		ki := mans[sorted[i]].GetKind()
		kj := mans[sorted[j]].GetKind()

		pi := kindPriorityOrDefault(ki)
		pj := kindPriorityOrDefault(kj)

		if pi == pj && stable {
			return sorted[i] < sorted[j]
		}
		return pi < pj
	})

	var stages []plan.Stage
	var current []string
	var currentKind string
	prevDeps := map[string]bool{}

	for _, id := range sorted {
		m := mans[id]
		kind := m.GetKind()
		ready := true

		if depend, ok := m.(manifests.Dependencies); ok {
			for _, dep := range depend.GetDependsOn() {
				if !prevDeps[dep] {
					ready = false
					break
				}
			}
		}

		if (!ready || currentKind != "" && kind != currentKind) && len(current) > 0 {
			stages = append(stages, makeStage(current, mans, mode, stable, parallel))
			for _, cid := range current {
				prevDeps[cid] = true
			}
			current = []string{}
		}

		current = append(current, id)
		currentKind = kind
	}

	if len(current) > 0 {
		stages = append(stages, makeStage(current, mans, mode, stable, parallel))
	}

	return stages
}

func kindPriorityOrDefault(kind string) int {
	if p, ok := kindPriority[kind]; ok {
		return p
	}
	return defaultPriority
}

func makeStage(ids []string, mans map[string]manifests.Manifest, mode string, stable, parallel bool) plan.Stage {
	if stable {
		sort.Strings(ids)
	}

	var nameParts []string
	seenKinds := make(map[string]bool)

	for _, id := range ids {
		m := mans[id]
		k := m.GetKind()
		if !seenKinds[k] {
			nameParts = append(nameParts, k)
			seenKinds[k] = true
		}
	}

	stageName := fmt.Sprintf("stage-%s", strings.Join(nameParts, "_"))

	return plan.Stage{
		Name:      stageName,
		Manifests: ids,
		Parallel:  parallel,
		Mode:      mode,
	}
}
