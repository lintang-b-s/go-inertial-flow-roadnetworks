package partitioner

import (
	"log"

	"github.com/lintang-b-s/navigatorx-partitioner/pkg/datastructure"
	"github.com/lintang-b-s/navigatorx-partitioner/pkg/util"
)

const (
	UNVISITED = -1
)

type TarjanSCC struct {
	dfsLow           []int
	dfsNum           []int
	onStack          []bool
	dfsNumberCounter int
	numSCC           int
	stack            []int32
	scc              [][]int32
	sccSizes         []int32
}

func NewTarjanSCC(V int) *TarjanSCC {
	return &TarjanSCC{
		dfsLow:           make([]int, V),
		dfsNum:           make([]int, V),
		onStack:          make([]bool, V),
		stack:            make([]int32, 0, V),
		dfsNumberCounter: 0,
		numSCC:           0,
	}
}

func (tj *TarjanSCC) run(graph *datastructure.Graph) {
	for i := 0; i < len(graph.GetNodes()); i++ {
		tj.dfsNum[i] = UNVISITED
	}

	for i := 0; i < len(graph.GetNodes()); i++ {
		if tj.dfsNum[i] == UNVISITED {
			tj.tarjanDFS(graph, int32(i))
		}
	}
	log.Printf("found %d strongly connected components", tj.numSCC)
}

func (tj *TarjanSCC) tarjanDFS(graph *datastructure.Graph, u int32) {
	tj.dfsLow[u] = tj.dfsNumberCounter
	tj.dfsNum[u] = tj.dfsNumberCounter
	tj.dfsNumberCounter++
	tj.stack = append(tj.stack, u)
	tj.onStack[u] = true

	for _, eID := range graph.GetNodeFirstOutEdges(u) {
		edge := graph.GetOutEdge(eID)
		v := edge.ToNodeID
		if tj.dfsNum[v] == UNVISITED {
			tj.tarjanDFS(graph, v)
		}
		if tj.onStack[v] { // v is on stack (should be if not else if, after callback check if v is on the stack)
			tj.dfsLow[u] = util.Min(tj.dfsLow[u], tj.dfsLow[v])
		}
	}

	if tj.dfsLow[u] == tj.dfsNum[u] {
		tj.numSCC++
		tj.scc = append(tj.scc, make([]int32, 0))
		tj.sccSizes = append(tj.sccSizes, 0)
		for {
			v := tj.stack[len(tj.stack)-1]
			tj.stack = tj.stack[:len(tj.stack)-1]
			tj.onStack[v] = false // marks as removed from stack
			tj.scc[tj.numSCC-1] = append(tj.scc[tj.numSCC-1], v)
			tj.sccSizes[tj.numSCC-1]++
			if u == v {
				break
			}
		}
	}

}

func (tj *TarjanSCC) GetSCC() ([][]int32, []int32) {

	return tj.scc, tj.sccSizes
}
