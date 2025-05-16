package collections

import "container/heap"

type PriorityQueue[T any] struct {
	nodes []T
	less  func(a, b T) bool
}

func NewPriorityQueue[T any](less func(a, b T) bool) *PriorityQueue[T] {
	return &PriorityQueue[T]{less: less}
}

func (pq *PriorityQueue[T]) Len() int { return len(pq.nodes) }

func (pq *PriorityQueue[T]) Less(i, j int) bool {
	return pq.less(pq.nodes[i], pq.nodes[j])
}

func (pq *PriorityQueue[T]) Swap(i, j int) {
	pq.nodes[i], pq.nodes[j] = pq.nodes[j], pq.nodes[i]
}

func (pq *PriorityQueue[T]) Push(x any) {
	pq.nodes = append(pq.nodes, x)
	heap.Fix(pq, len(pq.nodes)-1)
}

func (pq *PriorityQueue[T]) Pop() any {
	item := pq.nodes[0]
	pq.nodes = pq.nodes[1:]
	heap.Fix(pq, 0)
	return item
}
