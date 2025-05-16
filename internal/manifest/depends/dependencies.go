package depends

import (
	"container/heap"
	"fmt"
	"github.com/apiqube/cli/internal/collections"
	"strings"

	"github.com/apiqube/cli/internal/manifest"
)

var priorityOrder = map[string]int{
	"Values":    100,
	"ConfigMap": 90,
	"Server":    50,
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

func BuildGraphWithPriority(manifests []manifest.Manifest) (*GraphResult, error) {
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	idToNode := make(map[string]manifest.Manifest)
	nodePriority := make(map[string]int)

	for _, node := range manifests {
		id := node.GetID()
		idToNode[id] = node
		inDegree[id] = 0

		parts := strings.Split(id, ".")
		if len(parts) >= 2 {
			kind := parts[1]
			nodePriority[id] = getPriority(kind)
		}
	}

	for _, node := range manifests {
		id := node.GetID()
		for _, depID := range node.GetDependsOn() {
			if depID == id {
				return nil, fmt.Errorf("цикл: %s зависит от самого себя", id)
			}
			graph[depID] = append(graph[depID], id)
			inDegree[id]++
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

	if len(order) != len(manifests) {
		cyclicNodes := findCyclicNodes(inDegree)
		return nil, fmt.Errorf("циклы в зависимостях: %v", cyclicNodes)
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
