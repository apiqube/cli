package depends

import (
	"container/heap"
	"fmt"
	"strings"

	"github.com/apiqube/cli/internal/collections"

	"github.com/apiqube/cli/internal/core/manifests"
)

var priorityOrder = map[string]int{
	"Values":    100,
	"ConfigMap": 90,
	"Target":    50,
	"Service":   30,
}

type GraphResult struct {
	Graph          map[string][]string
	ExecutionOrder []string
}

type Node struct {
	ID       string
	Priority int
}

func BuildGraphWithPriority(mans []manifests.Manifest) (*GraphResult, error) {
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	idToNode := make(map[string]manifests.Manifest)
	nodePriority := make(map[string]int)

	for _, node := range mans {
		id := node.GetID()
		idToNode[id] = node
		inDegree[id] = 0

		parts := strings.Split(id, ".")
		if len(parts) >= 2 {
			kind := parts[1]
			nodePriority[id] = getPriority(kind)
		}
	}

	for _, man := range mans {
		if dep, has := man.(manifests.Dependencies); has {
			id := man.GetID()
			for _, depID := range dep.GetDependsOn() {
				if depID == id {
					return nil, fmt.Errorf("dependency error: %s manifest cannot depend on itself", id)
				}
				graph[depID] = append(graph[depID], id)
				inDegree[id]++
			}
		}
	}

	priorityQueue := collections.NewPriorityQueue[*Node](func(a, b *Node) bool {
		return a.Priority > b.Priority
	})

	for id, degree := range inDegree {
		if degree == 0 {
			heap.Push(priorityQueue, &Node{
				ID:       id,
				Priority: nodePriority[id],
			})
		}
	}

	var order []string
	for priorityQueue.Len() > 0 {
		current := heap.Pop(priorityQueue).(*Node).ID
		order = append(order, current)

		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				heap.Push(priorityQueue, &Node{
					ID:       neighbor,
					Priority: nodePriority[neighbor],
				})
			}
		}
	}

	if len(order) != len(mans) {
		cyclicNodes := findCyclicNodes(inDegree)
		return nil, fmt.Errorf("dependency error: сyclic dependency: %v", cyclicNodes)
	}

	return &GraphResult{
		Graph:          graph,
		ExecutionOrder: order,
	}, nil
}

func getPriority(kind string) int {
	if p, ok := priorityOrder[kind]; ok {
		return p
	}
	return 0
}

func findCyclicNodes(inDegree map[string]int) []string {
	cyclicNodes := make([]string, 0)
	for id, degree := range inDegree {
		if degree > 0 {
			cyclicNodes = append(cyclicNodes, id)
		}
	}
	return cyclicNodes
}
