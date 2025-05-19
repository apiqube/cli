package plan

import (
	"errors"
	"sync"
)

type depGraph struct {
	edges map[string][]string
	nodes map[string]bool
	lock  sync.Mutex
}

func newDepGraph() *depGraph {
	return &depGraph{
		edges: map[string][]string{},
		nodes: map[string]bool{},
	}
}

func (g *depGraph) addNode(id string) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.nodes[id] = true
}

func (g *depGraph) addEdge(from, to string) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.edges[from] = append(g.edges[from], to)
	g.nodes[from] = true
	g.nodes[to] = true
}

func (g *depGraph) topoSort() ([]string, error) {
	inDegree := map[string]int{}
	for node := range g.nodes {
		inDegree[node] = 0
	}

	for _, toList := range g.edges {
		for _, to := range toList {
			inDegree[to]++
		}
	}

	var queue []string
	for node, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, node)
		}
	}

	var result []string
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		result = append(result, n)

		for _, neighbor := range g.edges[n] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(result) != len(g.nodes) {
		return nil, errors.New("cycle detected in dependency graph")
	}
	return result, nil
}
