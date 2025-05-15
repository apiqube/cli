package depends

import (
	"fmt"

	"github.com/apiqube/cli/internal/manifests"
)

func TopoSort(graph map[string][]string) ([]string, error) {
	visited := make(map[string]bool)
	temp := make(map[string]bool)
	var result []string

	var visit func(string) error
	visit = func(n string) error {
		if temp[n] {
			return fmt.Errorf("circular dependency detected at %s", n)
		}
		if !visited[n] {
			temp[n] = true
			for _, dep := range graph[n] {
				if err := visit(dep); err != nil {
					return err
				}
			}
			visited[n] = true
			temp[n] = false
			result = append(result, n)
		}
		return nil
	}

	for node := range graph {
		if err := visit(node); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func SortManifestsByExecutionOrder(mans []manifests.Manifest, order []string) ([]manifests.Manifest, error) {
	idMap := make(map[string]manifests.Manifest)
	for _, m := range mans {
		idMap[m.GetID()] = m
	}

	sorted := make([]manifests.Manifest, 0, len(order))

	for _, id := range order {
		m, ok := idMap[id]
		if !ok {
			return nil, fmt.Errorf("manifest %s not found in loaded manifests", id)
		}

		sorted = append(sorted, m)
	}

	return sorted, nil
}
