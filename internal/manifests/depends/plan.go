package depends

import (
	"encoding/json"
	"fmt"
	"github.com/apiqube/cli/internal/manifests"
	"os"
	"path/filepath"
)

type ExecutionPlan struct {
	Order []string `json:"order"`
}

func GeneratePlan(outPath string, manifests []manifests.Manifest) error {
	graph, _, err := BuildDependencyGraph(manifests)
	if err != nil {
		return err
	}

	order, err := TopoSort(graph)
	if err != nil {
		return err
	}

	return SaveExecutionPlan(outPath, order)
}

func SaveExecutionPlan(path string, order []string) error {
	plan := ExecutionPlan{Order: order}

	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plan: %w", err)
	}

	if err = os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create dir: %w", err)
	}

	return os.WriteFile(filepath.Join(path, "execution-plan.json"), data, 0644)
}
