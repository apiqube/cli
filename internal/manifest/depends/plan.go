package depends

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"

	"github.com/apiqube/cli/internal/manifest"
)

type ExecutionPlan struct {
	Order []string `json:"order"`
}

func GeneratePlan(manifests []manifest.Manifest) ([]string, error) {
	if len(manifests) == 0 {
		return nil, fmt.Errorf("no manifests to generate plan")
	}

	graph, _, err := BuildDependencyGraph(manifests)
	if err != nil {
		return nil, err
	}

	order, err := TopoSort(graph)
	if err != nil {
		return nil, err
	}

	return order, nil
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

	filePath, err := xdg.DataFile(manifest.ExecutionPlansDirPath + fileName)
	if err != nil {
		return fmt.Errorf("failed to build path: %w", err)
	}

	if err = os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0o644)
}
