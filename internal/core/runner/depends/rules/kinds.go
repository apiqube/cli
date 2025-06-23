package rules

import (
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/manifests/kinds"
)

const KindPriorityRuleName = "kind_priority"

// KindPriorityRule handles kind-based priorities
type KindPriorityRule struct {
	priorities map[string]int
}

func NewKindPriorityRule() *KindPriorityRule {
	return &KindPriorityRule{
		priorities: kinds.PriorityMap,
	}
}

func (r *KindPriorityRule) Name() string {
	return KindPriorityRuleName
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
	return 1_000
}
