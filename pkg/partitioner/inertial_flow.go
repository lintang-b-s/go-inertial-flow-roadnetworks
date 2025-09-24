package partitioner

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"time"

	"github.com/lintang-b-s/navigatorx-partitioner/pkg/datastructure"
	"golang.org/x/exp/rand"
)

// will be used to implement customizable route planning
// ngasal heheh
// idk if this implementation is correct, but the partition result is very good & similiar to metis partition result
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

// floorPow2 return the largest power of 2 that is <= n.
func floorPow2(n int) int {
	if n < 1 {
		return 0
	}
	p := 1
	for p<<1 <= n {
		p <<= 1
	}
	return p
}

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

// https://www.sommer.jp/roadseparator.pdf
func (iflow *InertialFlow) RunInertialFlow(graph *datastructure.Graph) {

	start := time.Now()

	log.Printf("prepartitioning...")
	bigSCCs, smallSCCs := iflow.PrePartitionSCC(graph, int32(math.Pow(2, 12)))

	log.Printf("prepartitioning done... %d big sccs, %d small sccs", len(bigSCCs), len(smallSCCs))

	log.Printf("running inertial flow partitioning..., with at most %d nodes per cell", iflow.regionSize)

	for _, pNodeIDs := range bigSCCs {
		iflow.RecursiveBisection(
			pNodeIDs, graph,
		)
	}

	iflow.partitions = append(iflow.partitions, smallSCCs...)

	iflow.saveCutEdgesToFile(iflow.cutEdges, graph)

	iflow.savePartitionsToFile(iflow.partitions, graph, "inertial_flow")
	log.Printf("regionSize(r): %d, regions created: %d, boundary (min-cut) size: %d, time: %ds", iflow.regionSize,
		iflow.regionsCreated, len(iflow.cutEdges), int(time.Since(start).Seconds()))
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
		a := rand.Float64() // [0,1)
		b := rand.Float64() // [0,1)
		lines = append(lines, []float64{a, b})
	}
	line := lines[rand.Intn(len(lines))]
	var (
		idToIndex   map[int32]int32
		indexToID   []int32
		edmondsKarp *EdmondsKarp
	)

	sort.Slice(nodeIDs, func(i, j int) bool {
		a, b := nodeIDs[i], nodeIDs[j]
		return graph.GetNode(a).Lon*line[0]+
			graph.GetNode(a).Lat*line[1] <
			graph.GetNode(b).Lon*line[0]+
				graph.GetNode(b).Lat*line[1]
	})

	_, curedmondsKarp, currIdToIndex, currIndexToID := iflow.runMaxFlow(nodeIDs, v, graph)
	idToIndex = currIdToIndex
	indexToID = currIndexToID
	edmondsKarp = curedmondsKarp

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

type cutEdge struct {
	FromLat float64 `json:"fromLat"`
	FromLon float64 `json:"fromLon"`
	ToLat   float64 `json:"toLat"`
	ToLon   float64 `json:"toLon"`
}

func (iflow *InertialFlow) saveCutEdgesToFile(cutEdges []datastructure.Edge, graph *datastructure.Graph) error {
	cutEdgesCoord := make([]cutEdge, 0)
	for _, edge := range cutEdges {
		from := graph.GetNode(edge.FromNodeID)
		to := graph.GetNode(edge.ToNodeID)
		cutEdgesCoord = append(cutEdgesCoord, cutEdge{
			FromLat: from.Lat,
			FromLon: from.Lon,
			ToLat:   to.Lat,
			ToLon:   to.Lon,
		})
	}
	buf, err := json.MarshalIndent(cutEdgesCoord, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile("cutEdges.json", buf, 0644); err != nil {
		return err
	}
	return nil
}

func (iflow *InertialFlow) savePartitionsToFile(partitions [][]int32, graph *datastructure.Graph,
	name string) error {
	type partitionType struct {
		Nodes []datastructure.Coordinate `json:"nodes"`
	}
	nodes := make([]int, len(graph.GetNodes()))

	parts := []partitionType{}
	for partitionID, partition := range partitions {
		rand.Seed(uint64(time.Now().UnixNano()))
		rand.Shuffle(len(partition), func(i, j int) { partition[i], partition[j] = partition[j], partition[i] })
		partitionNodes := make([]datastructure.Coordinate, 0)

		for i := 0; i < int(float64(len(partition))*0.3); i++ {
			node := graph.GetNode(partition[i])
			partitionNodes = append(partitionNodes, datastructure.NewCoordinate(
				node.Lat, node.Lon,
			))
		}
		parts = append(parts, partitionType{
			Nodes: partitionNodes,
		})

		for _, nodeID := range partition {
			nodes[nodeID] = partitionID
		}
	}
	buf, err := json.MarshalIndent(parts, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("nodes after partitioning: %v\n", len(nodes))

	if err := os.WriteFile(fmt.Sprintf("nodePerPartitions_%s.json", name), buf, 0644); err != nil {
		return err
	}
	return nil

}
