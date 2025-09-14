package partitioner

import (
	"sort"
	"testing"

	"github.com/lintang-b-s/go-graph-inertial-flow/pkg/datastructure"
	"github.com/stretchr/testify/assert"
)

func Test_EdmondsKarp_MaxFlow(t *testing.T) {
	t.Run("test case 1", func(t *testing.T) {
		v := 4
		source := 0
		target := 3
		edmondsKarp := NewEdmondsKarp(int32(v))

		edmondsKarp.addEdge(0, 1, 8, true)
		edmondsKarp.addEdge(0, 2, 8, true)
		edmondsKarp.addEdge(1, 2, 1, true)
		edmondsKarp.addEdge(1, 3, 8, true)
		edmondsKarp.addEdge(2, 3, 8, true)
		maxFlow := edmondsKarp.run(int32(source), int32(target))
		if maxFlow != 16 {
			t.Errorf("Expected max flow of 16, got %f", maxFlow)
		}
	})

	t.Run("test case 2", func(t *testing.T) {
		// https://cp-algorithms.com/graph/edmonds_karp.html
		v := 6
		source := 0
		target := 5
		edmondsKarp := NewEdmondsKarp(int32(v))

		edmondsKarp.addEdge(0, 1, 7, true)
		edmondsKarp.addEdge(0, 4, 4, true)
		edmondsKarp.addEdge(1, 2, 5, true)
		edmondsKarp.addEdge(1, 3, 3, true)

		// B
		edmondsKarp.addEdge(2, 5, 8, true)

		// C
		edmondsKarp.addEdge(3, 5, 5, true)
		edmondsKarp.addEdge(3, 2, 3, true)

		// D
		edmondsKarp.addEdge(4, 3, 2, true)
		edmondsKarp.addEdge(4, 1, 3, true)
		maxFlow := edmondsKarp.run(int32(source), int32(target))
		if maxFlow != 10.0 {
			t.Errorf("Expected max flow of 10, got %f", maxFlow)
		}
	})

	t.Run("test case 3", func(t *testing.T) {
		// https://cp-algorithms.com/graph/edmonds_karp.html
		//  multiple sources-sinks
		v := 10 // original 6 + 1 additional source and 1 additional sink + 2 artificial nodes
		source := 8
		target := 9
		edmondsKarp := NewEdmondsKarp(int32(v))

		// A
		edmondsKarp.addEdge(0, 1, 7, true)
		edmondsKarp.addEdge(0, 4, 4, true)
		edmondsKarp.addEdge(1, 2, 5, true)
		edmondsKarp.addEdge(1, 3, 3, true)

		// B
		edmondsKarp.addEdge(2, 5, 8, true)

		// C
		edmondsKarp.addEdge(3, 5, 5, true)
		edmondsKarp.addEdge(3, 2, 3, true)

		// D
		edmondsKarp.addEdge(4, 3, 2, true)
		edmondsKarp.addEdge(4, 1, 3, true)

		// E
		// E edges: E->A (2)
		edmondsKarp.addEdge(6, 1, 2, true)

		// F
		// F edges: B-> F = 2
		edmondsKarp.addEdge(2, 7, 2, true)

		// artifiicial source
		edmondsKarp.addEdge(8, 6, infFlow, true)
		edmondsKarp.addEdge(8, 0, infFlow, true)

		// artifiicial sink
		edmondsKarp.addEdge(2, 9, infFlow, true)
		edmondsKarp.addEdge(3, 9, infFlow, true)

		maxFlow := edmondsKarp.run(int32(source), int32(target))
		if maxFlow != 10.0 {
			t.Errorf("Expected max flow of 10, got %f", maxFlow)
		}
	})

	t.Run("test case 4", func(t *testing.T) {
		// https://cp-algorithms.com/graph/edmonds_karp.html
		v := 6
		source := 0
		target := 5
		edmondsKarp := NewEdmondsKarp(int32(v))

		edmondsKarp.addEdge(0, 1, 7, true)
		edmondsKarp.addEdge(0, 4, 4, true)
		edmondsKarp.addEdge(1, 2, 5, true)
		edmondsKarp.addEdge(1, 3, 3, true)

		// B
		edmondsKarp.addEdge(2, 5, 8, true)

		// C
		edmondsKarp.addEdge(3, 5, 5, true)
		edmondsKarp.addEdge(3, 2, 3, true)

		// D
		edmondsKarp.addEdge(4, 3, 2, true)
		edmondsKarp.addEdge(4, 1, 3, true)
		maxFlow := edmondsKarp.run(int32(source), int32(target))
		if maxFlow != 10 {
			t.Errorf("Expected max flow of 10, got %f", maxFlow)
		}
		// get min cut
		bisectedGraph := [2][]int32{}
		visited := make([]bool, v)
		cutEdges := []datastructure.Edge{}
		edmondsKarp.dfsMinCut(int32(source), &bisectedGraph, visited, int32(source), &cutEdges)
		expectedMinCut := [2][]int32{
			{0, 1, 4},
			{2, 3, 5},
		}
		if len(bisectedGraph) != len(expectedMinCut) {
			t.Errorf("Expected min cut of length %d, got %d", len(expectedMinCut), len(bisectedGraph))
		}
		sort.Slice(bisectedGraph[0], func(i, j int) bool {
			return bisectedGraph[0][i] < bisectedGraph[0][j]
		})
		sort.Slice(bisectedGraph[1], func(i, j int) bool {
			return bisectedGraph[1][i] < bisectedGraph[1][j]
		})
		if assert.Equal(t, bisectedGraph, expectedMinCut) == false {
			t.Errorf("Expected min cut %v, got %v", expectedMinCut, bisectedGraph)
		}
	})

	t.Run("test case 5", func(t *testing.T) {
		// https://cp-algorithms.com/graph/edmonds_karp.html
		//  multiple sources-sinks
		v := 10 // original 6 + 1 additional source and 1 additional sink + 2 artificial nodes
		source := 8
		target := 9
		edmondsKarp := NewEdmondsKarp(int32(v))

		// A
		edmondsKarp.addEdge(0, 1, 7, true)
		edmondsKarp.addEdge(0, 4, 4, true)
		edmondsKarp.addEdge(1, 2, 5, true)
		edmondsKarp.addEdge(1, 3, 3, true)

		// B
		edmondsKarp.addEdge(2, 5, 8, true)

		// C
		edmondsKarp.addEdge(3, 5, 5, true)
		edmondsKarp.addEdge(3, 2, 3, true)

		// D
		edmondsKarp.addEdge(4, 3, 2, true)
		edmondsKarp.addEdge(4, 1, 3, true)

		// E
		// E edges: E->A (2)
		edmondsKarp.addEdge(6, 1, 2, true)

		// F
		// F edges: B-> F = 2
		edmondsKarp.addEdge(2, 7, 2, true)

		// artifiicial source
		edmondsKarp.addEdge(8, 6, infFlow, true)
		edmondsKarp.addEdge(8, 0, infFlow, true)

		// artifiicial sink
		edmondsKarp.addEdge(7, 9, infFlow, true)
		edmondsKarp.addEdge(5, 9, infFlow, true)

		maxFlow := edmondsKarp.run(int32(source), int32(target))
		if maxFlow != 10 {
			t.Errorf("Expected max flow of 10, got %f", maxFlow)
		}
		// get min cut
		bisectedGraph := [2][]int32{}
		visited := make([]bool, v)
		cutEdges := &[]datastructure.Edge{}
		edmondsKarp.dfsMinCutMultipleSourcesSinks(int32(source), &bisectedGraph, visited,
			int32(source), int32(target), cutEdges)
		expectedMinCut := [2][]int32{
			{0, 1, 4, 6},
			{2, 3, 5, 7},
		}
		if len(bisectedGraph) != len(expectedMinCut) {
			t.Errorf("Expected min cut of length %d, got %d", len(expectedMinCut), len(bisectedGraph))
		}
		sort.Slice(bisectedGraph[0], func(i, j int) bool {
			return bisectedGraph[0][i] < bisectedGraph[0][j]
		})
		sort.Slice(bisectedGraph[1], func(i, j int) bool {
			return bisectedGraph[1][i] < bisectedGraph[1][j]
		})
		if assert.Equal(t, bisectedGraph, expectedMinCut) == false {
			t.Errorf("Expected min cut %v, got %v", expectedMinCut, bisectedGraph)
		}
	})
}
