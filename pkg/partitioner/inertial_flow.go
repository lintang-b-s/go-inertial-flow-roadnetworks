package partitioner

import (
	"log"
	"sort"

	"github.com/lintang-b-s/navigatorx-partitioner/pkg/datastructure"
	"golang.org/x/exp/rand"
)

type InertialFlow struct {
	regionSize     int
	regionsCreated int
	partitions     [][]int32
	cutEdges       []datastructure.Edge
}

func NewInertialFlow(regionSize int) *InertialFlow {
	return &InertialFlow{
		regionSize:     regionSize,
		regionsCreated: 0,
		partitions:     make([][]int32, 0),
		cutEdges:       make([]datastructure.Edge, 0),
	}
}

const (
	b                  = 0.25
	artificialSourceID = int32(2147483646)
	artificialSinkID   = int32(2147483647)
)

func (iflow *InertialFlow) PrePartitionSCC(graph *datastructure.Graph, smallComponentSize int32) ([][]int32, [][]int32) {
	tarjan := NewTarjanSCC(len(graph.GetNodes()))
	tarjan.run(graph)
	scc, sccSizes := tarjan.GetSCC()

	// only bisect the non-small scc
	bigSCCs := make([][]int32, 0, len(scc))
	smallSCCs := make([][]int32, 0, len(scc))

	for i, size := range sccSizes {
		if size > smallComponentSize {
			bigSCCs = append(bigSCCs, scc[i])
		} else {
			smallSCCs = append(smallSCCs, scc[i])
		}
	}

	return bigSCCs, smallSCCs
}

// RecursiveBisection. recursively bisect the graph until the number of nodes in the partition is <= regionSize
// TODO: make this function run concurrently for each recursion using goroutine pool
func (iflow *InertialFlow) RecursiveBisection(nodeIDs []int32, graph *datastructure.Graph,
) {
	continueBisection := len(nodeIDs) > iflow.regionSize
	if !continueBisection {
		iflow.regionsCreated += 1
		log.Printf("recursiveBisection: %d partitions created", iflow.regionsCreated)
		iflow.partitions = append(iflow.partitions, nodeIDs)
		return
	}

	v := len(nodeIDs)
	var lines = [][]float64{{1, 0}, {0, 1}, {1, 1}, {-1, 1}}
	for i := 0; i < 2; i++ {
		a := rand.Float64()
		b := rand.Float64()
		lines = append(lines, []float64{a, b})
	}
	line := lines[rand.Intn(len(lines))]

	sort.Slice(nodeIDs, func(i, j int) bool {
		a, b := nodeIDs[i], nodeIDs[j]
		return graph.GetNode(a).Lon*line[0]+
			graph.GetNode(a).Lat*line[1] <
			graph.GetNode(b).Lon*line[0]+
				graph.GetNode(b).Lat*line[1]
	})

	_, curedmondsKarp, currIdToIndex, currIndexToID := iflow.runMaxFlow(nodeIDs, v, graph)
	idToIndex := currIdToIndex
	indexToID := currIndexToID
	edmondsKarp := curedmondsKarp

	bisectedGraph := [2][]int32{}
	visited := make([]bool, v+2)

	// get min-cut
	localMinCuts := make([]datastructure.Edge, 0)
	edmondsKarp.dfsMinCutMultipleSourcesSinks(int32(idToIndex[artificialSourceID]), &bisectedGraph, visited,
		int32(idToIndex[artificialSourceID]), int32(idToIndex[artificialSinkID]), &localMinCuts)

	for i := 0; i < len(localMinCuts); i++ {
		localMinCuts[i].FromNodeID = indexToID[localMinCuts[i].FromNodeID]
		localMinCuts[i].ToNodeID = indexToID[localMinCuts[i].ToNodeID]
	}

	part0 := make([]int32, len(bisectedGraph[0]))
	for i, idx := range bisectedGraph[0] {
		part0[i] = indexToID[idx]
	}
	part1 := make([]int32, len(bisectedGraph[1]))
	for i, idx := range bisectedGraph[1] {
		part1[i] = indexToID[idx]
	}

	iflow.RecursiveBisection(part0, graph)
	iflow.RecursiveBisection(part1, graph)
}

func (iflow *InertialFlow) runMaxFlow(nodeIDs []int32, v int, graph *datastructure.Graph) (float64, *EdmondsKarp, map[int32]int32, []int32) {
	indexToID := make([]int32, v)         // map for index to node ID
	idToIndex := make(map[int32]int32, v) // map for node ID to index
	for i, id := range nodeIDs {
		indexToID[i] = id
		idToIndex[id] = int32(i)
	}

	sources := make([]int32, int(b*float64(v)))
	sinks := make([]int32, int(b*float64(v)))
	for i := 0; i < len(sources); i++ {
		sources[i] = nodeIDs[i]
	}
	for i := 0; i < len(sinks); i++ {
		sinks[i] = nodeIDs[v-1-i]
	}

	idToIndex[artificialSourceID] = int32(v)
	idToIndex[artificialSinkID] = int32(v + 1)

	edmondsKarp := NewEdmondsKarp(int32(v + 2))
	for _, nodeID := range nodeIDs {
		u := idToIndex[nodeID]
		for _, edgeIDx := range graph.GetNodeFirstOutEdges(nodeID) {
			edge := graph.GetOutEdge(edgeIDx)
			v := idToIndex[edge.ToNodeID]
			_, ok := idToIndex[edge.ToNodeID]
			if u == v || nodeID == edge.ToNodeID || !ok {
				continue // skip self-loop & edge that point to node that is not in the current partition
			}
			edmondsKarp.addEdge(u, v, edge.Weight, edge.Directed)

		}
	}

	for _, sourceID := range sources {
		if idToIndex[sourceID] == idToIndex[artificialSourceID] {
			continue
		}
		edmondsKarp.addEdge(idToIndex[artificialSourceID], idToIndex[sourceID], infFlow, true)
	}

	for _, sinkID := range sinks {
		if idToIndex[artificialSinkID] == idToIndex[sinkID] {
			continue
		}
		edmondsKarp.addEdge(idToIndex[sinkID], idToIndex[artificialSinkID], infFlow, true)
	}
	minCut := edmondsKarp.run(idToIndex[artificialSourceID], idToIndex[artificialSinkID])
	return minCut, edmondsKarp, idToIndex, indexToID
}
