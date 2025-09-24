package partitioner

import (
	"testing"

	"github.com/lintang-b-s/navigatorx-partitioner/pkg/datastructure"
	"github.com/lintang-b-s/navigatorx-partitioner/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestTarjanSCC(t *testing.T) {
	edges := make([]datastructure.Edge, 0)
	edges = append(edges, datastructure.Edge{FromNodeID: 0, ToNodeID: 1})
	edges = append(edges, datastructure.Edge{FromNodeID: 1, ToNodeID: 2})
	edges = append(edges, datastructure.Edge{FromNodeID: 1, ToNodeID: 4})
	edges = append(edges, datastructure.Edge{FromNodeID: 2, ToNodeID: 3})
	edges = append(edges, datastructure.Edge{FromNodeID: 3, ToNodeID: 2})
	edges = append(edges, datastructure.Edge{FromNodeID: 4, ToNodeID: 0})

	nodes := make([]datastructure.CHNode, 5)
	for i := 0; i < 5; i++ {
		nodes[i] = datastructure.CHNode{
			ID: int32(i),
		}
	}

	graphStorage := datastructure.NewGraphStorage()
	for _, edge := range edges {

		graphStorage.AppendEdgeStorage(edge)
	}

	graph := datastructure.NewGraph()
	graph.InitGraph(nodes, graphStorage, nil, util.NewIdMap())

	tarjan := NewTarjanSCC(len(graph.GetNodes()))
	tarjan.run(graph)

	scc, _ := tarjan.GetSCC()

	assert.Equal(t, 2, len(scc))
	assert.Equal(t, 3, len(scc[0]))
	assert.Equal(t, 2, len(scc[1]))

}
