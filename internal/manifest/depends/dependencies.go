package depends

import (
	"fmt"

	"github.com/apiqube/cli/internal/manifest"
)

type Node struct {
	ID       string
	Manifest manifest.Manifest
	Depends  []string
}

func BuildDependencyGraph(mans []manifest.Manifest) (map[string][]string, map[string]manifest.Manifest, error) {
	graph := make(map[string][]string)
	idToManifest := make(map[string]manifest.Manifest)

	for _, m := range mans {
		id := m.GetID()
		idToManifest[id] = m
		graph[id] = []string{}
	}

	for id, m := range idToManifest {
		for _, dep := range m.GetDependsOn() {
			if dep == id {
				return nil, nil, fmt.Errorf("m %s cannot depend on itself", id)
			}

			if _, ok := idToManifest[dep]; !ok {
				return nil, nil, fmt.Errorf("m %s depends on unknown m %s", id, dep)
			}

			graph[id] = append(graph[id], dep)
		}
	}

	return graph, idToManifest, nil
}
