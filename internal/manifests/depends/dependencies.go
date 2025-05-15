package depends

import (
	"fmt"
	"github.com/apiqube/cli/internal/manifests"
)

type Node struct {
	ID       string
	Manifest manifests.Manifest
	Depends  []string
}

func BuildDependencyGraph(mans []manifests.Manifest) (map[string][]string, map[string]manifests.Manifest, error) {
	graph := make(map[string][]string)
	idToManifest := make(map[string]manifests.Manifest)

	for _, m := range mans {
		id := m.GetID()
		idToManifest[id] = m
		graph[id] = []string{}
	}

	for id, manifest := range idToManifest {
		for _, dep := range manifest.GetDependsOn() {
			if dep == id {
				return nil, nil, fmt.Errorf("manifest %s cannot depend on itself", id)
			}

			if _, ok := idToManifest[dep]; !ok {
				return nil, nil, fmt.Errorf("manifest %s depends on unknown manifest %s", id, dep)
			}

			graph[id] = append(graph[id], dep)
		}
	}

	return graph, idToManifest, nil
}
