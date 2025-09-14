package datastructure

import (
	"log"

	"github.com/lintang-b-s/go-graph-inertial-flow/pkg/util"
)

type Metadata struct {
	MeanDegree       float64
	ShortcutsCount   int64
	degrees          []int
	OutEdgeOrigCount []int
	EdgeCount        int
	NodeCount        int
}
type Graph struct {
	GraphStorage           *GraphStorage
	ContractedNodes        []CHNode
	Metadata               Metadata
	ContractedFirstOutEdge [][]int32

	SCC                []int32 // map for nodeID -> sccID
	SCCNodesCount      []int32 // map for sccID -> nodes count in scc
	SCCCondensationAdj [][]int32

	nextEdgeID int32

	StreetDirection map[int][2]bool // 0 = forward, 1 = backward
	TagStringIDMap  util.IDMap
}

func NewGraph() *Graph {

	return &Graph{
		ContractedNodes: make([]CHNode, 0),

		StreetDirection: make(map[int][2]bool),
		TagStringIDMap:  util.NewIdMap(),
	}

}

func (ch *Graph) InitGraph(processedNodes []CHNode,
	graphStorage *GraphStorage, streetDirections map[string][2]bool,
	tagStringIdMap util.IDMap) {

	ch.TagStringIDMap = tagStringIdMap

	gLen := len(processedNodes)

	for streetName, direction := range streetDirections {
		ch.StreetDirection[ch.TagStringIDMap.GetID(streetName)] = direction
	}
	ch.GraphStorage = graphStorage

	ch.ContractedNodes = make([]CHNode, gLen)

	copy(ch.ContractedNodes, processedNodes)

	ch.Metadata.degrees = make([]int, gLen)
	ch.Metadata.OutEdgeOrigCount = make([]int, gLen)
	ch.Metadata.ShortcutsCount = 0

	edgeID := int32(0)
	ch.ContractedFirstOutEdge = make([][]int32, len(ch.ContractedNodes))

	log.Printf("intializing original osm graph...")

	// init graph original
	for _, edge := range ch.GraphStorage.EdgeStorage {

		ch.ContractedFirstOutEdge[edge.FromNodeID] = append(ch.ContractedFirstOutEdge[edge.FromNodeID], int32(edgeID))

		ch.Metadata.OutEdgeOrigCount[edge.FromNodeID]++
		edgeID++
	}

	log.Printf("initializing osm graph done... \n total nodes: %d", gLen)

	ch.Metadata.EdgeCount = len(ch.GraphStorage.EdgeStorage)
	ch.Metadata.NodeCount = gLen
}

func (ch *Graph) GetNodes() []CHNode {
	return ch.ContractedNodes
}

func (ch *Graph) GetNode(nodeID int32) CHNode {
	return ch.ContractedNodes[nodeID]
}

func (ch *Graph) GetNodeFirstOutEdges(nodeID int32) []int32 {
	return ch.ContractedFirstOutEdge[nodeID]
}

func (ch *Graph) GetOutEdge(edgeID int32) Edge {
	return ch.GraphStorage.GetOutEdge(edgeID)
}

func (ch *Graph) GetOutEdgeCount() int {
	return len(ch.GraphStorage.EdgeStorage)
}
