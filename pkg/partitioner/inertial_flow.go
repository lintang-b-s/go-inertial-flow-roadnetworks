package partitioner

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/lintang-b-s/go-graph-inertial-flow/pkg/datastructure"
	"golang.org/x/exp/rand"
)

// will be used to implement customizable route planning & GTree
// idk if this implementation is correct, but the partition result is very good & similiar to metis partition result
type InertialFlow struct {
	regionSize     int
	regionsCreated int
	partitions     [][]int32
	cutEdges       []datastructure.Edge
	wg             sync.WaitGroup
	mu             sync.Mutex
}

func NewInertialFlow(v int32, regionSize int) *InertialFlow {
	return &InertialFlow{
		regionSize:     regionSize,
		regionsCreated: 1,
		partitions:     make([][]int32, 0),
		cutEdges:       make([]datastructure.Edge, 0),
	}
}

var (
	lines = [4][]int{{1, 0}, {0, 1}, {1, 1}, {-1, 1}}
)

const (
	b = 0.25
)

func (iflow *InertialFlow) RunInertialFlow(graph *datastructure.Graph) {

	start := time.Now()

	nodesCount := len(graph.GetNodes())
	initialNodeIDs := make([]int32, nodesCount)
	for i, node := range graph.GetNodes() {
		initialNodeIDs[i] = node.ID
	}
	targetRegions := int(math.Ceil(float64(nodesCount / iflow.regionSize)))
	log.Printf("target regions: %d", targetRegions)

	iflow.wg.Add(1)
	go iflow.recursiveBisection(
		initialNodeIDs, graph,
	)
	iflow.wg.Wait()

	iflow.partitions = iflow.partitions[len(iflow.partitions)-targetRegions:]
	iflow.saveCutEdgesToFile(iflow.cutEdges, graph)
	iflow.savePartitionsToFile(iflow.partitions, graph, "inertial_flow")
	log.Printf("regionSize(r): %d, regions created: %d, boundary (min-cut) size: %d, time: %ds", iflow.regionSize,
		iflow.regionsCreated, len(iflow.cutEdges), int(time.Since(start).Seconds()))
}

func (iflow *InertialFlow) recursiveBisection(nodeIDs []int32, graph *datastructure.Graph,
) {
	v := len(nodeIDs)
	defer iflow.wg.Done()

	iflow.mu.Lock()
	targetRegions := int(math.Ceil(float64(len(graph.GetNodes()) / iflow.regionSize)))
	done := iflow.regionsCreated >= targetRegions
	iflow.mu.Unlock()
	if done || len(nodeIDs) <= 5 {
		return
	}

	line := lines[rand.Intn(len(lines))]
	sort.Slice(nodeIDs, func(i, j int) bool {
		a, b := nodeIDs[i], nodeIDs[j]

		return graph.GetNode(a).Lon*float64(line[0])+
			graph.GetNode(a).Lat*float64(line[1]) <
			graph.GetNode(b).Lon*float64(line[0])+
				graph.GetNode(b).Lat*float64(line[1])
	})

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

	artificialSource := datastructure.NewCHNode(
		0, 0, 0, int32(v),
	)
	artificialSink := datastructure.NewCHNode(
		0, 0, 0, int32(v+1),
	)

	dinic := NewDinicMinCut(int32(v + 2))
	for _, nodeID := range nodeIDs {
		u := idToIndex[nodeID]
		for _, edgeIDx := range graph.GetNodeFirstOutEdges(nodeID) {
			edge := graph.GetOutEdge(edgeIDx)
			v := idToIndex[edge.ToNodeID]
			dinic.addEdge(u, v, edge.Weight, edge.Directed)
		}
	}

	for _, sourceID := range sources {
		dinic.addEdge(artificialSource.ID, idToIndex[sourceID], infFlow, true)
	}

	for _, sinkID := range sinks {
		dinic.addEdge(idToIndex[sinkID], artificialSink.ID, infFlow, true)
	}

	// run dinic algorithm
	dinic.dinic(artificialSource.ID, artificialSink.ID)
	bisectedGraph := [2][]int32{}
	visited := make([]bool, v+2)

	// get min-cut
	localMinCuts := make([]datastructure.Edge, 0)
	dinic.dfsMinCutMultipleSourcesSinks(int32(artificialSource.ID), &bisectedGraph, visited,
		int32(artificialSource.ID), int32(artificialSink.ID), &localMinCuts)

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

	iflow.mu.Lock()
	regionsCreated := iflow.regionsCreated
	regionsCreated++
	continueBisection := regionsCreated < targetRegions
	if continueBisection {
		iflow.regionsCreated = regionsCreated
		iflow.partitions = append(iflow.partitions, part0, part1)
		iflow.cutEdges = append(iflow.cutEdges, localMinCuts...)
		log.Printf("recursiveBisection: %d nodes, %d regions created", v, iflow.regionsCreated)
		iflow.wg.Add(2)
		go iflow.recursiveBisection(part0, graph)
		go iflow.recursiveBisection(part1, graph)
	}
	iflow.mu.Unlock()

}

type cutEdge struct {
	FromLat float64 `json:"fromLat"`
	FromLon float64 `json:"fromLon"`
	ToLat   float64 `json:"toLat"`
	ToLon   float64 `json:"toLon"`
}

func (iflow *InertialFlow) saveCutEdgesToFile(cutEdges []datastructure.Edge, graph *datastructure.Graph) {
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
		panic(err)
	}

	if err := os.WriteFile("cutEdges.json", buf, 0644); err != nil {
		panic(err)
	}
}

func (iflow *InertialFlow) savePartitionsToFile(partitions [][]int32, graph *datastructure.Graph,
	name string) {
	type partitionType struct {
		Nodes []datastructure.Coordinate `json:"nodes"`
	}
	parts := []partitionType{}
	for _, partition := range partitions {
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
	}
	buf, err := json.MarshalIndent(parts, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(fmt.Sprintf("nodePerPartitions_%s.json", name), buf, 0644); err != nil {
		panic(err)
	}
}
