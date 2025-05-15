package depends

import (
	"encoding/json"
	"fmt"
	"github.com/adrg/xdg"
	"os"
	"path/filepath"

	"github.com/apiqube/cli/internal/manifests"
)

type ExecutionPlan struct {
	Order []string `json:"order"`
}

func GeneratePlan(manifests []manifests.Manifest) error {
	graph, _, err := BuildDependencyGraph(manifests)
	if err != nil {
		return err
	}

	order, err := TopoSort(graph)
	if err != nil {
		return err
	}

	return SaveExecutionPlan(manifests[0].GetNamespace(), order)
}

func SaveExecutionPlan(namespace string, order []string) error {
	if namespace == "" {
		namespace = "default"
	}

	plan := ExecutionPlan{Order: order}

	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plan: %w", err)
	}

	fileName := fmt.Sprintf("/plan-%s.json", namespace)

	filePath, err := xdg.DataFile(manifests.ExecutionPlansDirPath + fileName)
	if err != nil {
		return fmt.Errorf("failed to build path: %w", err)
	}

	if err = os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0o644)
}
