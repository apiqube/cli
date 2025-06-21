package depends

import (
	"container/heap"
	"fmt"
	"sort"
	"strings"

	"github.com/apiqube/cli/internal/collections"

	"github.com/apiqube/cli/internal/core/manifests"
)

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

	// Initialize all manifests
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

	// Build dependency graph
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

	// Use priority queue for topological sorting with priorities
	// Lower priority number = higher execution priority (executes first)
	priorityQueue := collections.NewPriorityQueue[*Node](func(a, b *Node) bool {
		// First compare by priority (lower number = higher priority)
		if a.Priority != b.Priority {
			return a.Priority < b.Priority
		}
		// If priorities are equal, sort by ID for deterministic behavior
		return a.ID < b.ID
	})

	// Add all nodes with zero in-degree to the queue
	var zeroInDegreeNodes []*Node
	for id, degree := range inDegree {
		if degree == 0 {
			zeroInDegreeNodes = append(zeroInDegreeNodes, &Node{
				ID:       id,
				Priority: nodePriority[id],
			})
		}
	}

	// Sort for deterministic behavior
	sort.Slice(zeroInDegreeNodes, func(i, j int) bool {
		if zeroInDegreeNodes[i].Priority != zeroInDegreeNodes[j].Priority {
			return zeroInDegreeNodes[i].Priority < zeroInDegreeNodes[j].Priority
		}
		return zeroInDegreeNodes[i].ID < zeroInDegreeNodes[j].ID
	})

	// Add to priority queue
	for _, node := range zeroInDegreeNodes {
		heap.Push(priorityQueue, node)
	}

	var order []string
	for priorityQueue.Len() > 0 {
		current := heap.Pop(priorityQueue).(*Node).ID
		order = append(order, current)

		// Process neighbors in sorted order for deterministic behavior
		neighbors := graph[current]
		sort.Strings(neighbors)

		for _, neighbor := range neighbors {
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
	return 100 // Default low priority for unknown kinds
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
