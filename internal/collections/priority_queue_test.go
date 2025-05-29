package collections

import (
	"github.com/stretchr/testify/require"
	"slices"
	"sort"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue(func(a, b int) bool {
		return a > b
	})

	values := []int{
		1, 2, 3, 10, 3, 145, 94, 173, 833,
	}

	for _, v := range values {
		pq.Push(v)
	}

	sort.Ints(values)
	slices.Reverse(values)

	require.Equal(t, len(values), pq.Len())

	for _, x := range values {
		require.Equal(t, x, pq.Pop())
	}
}
