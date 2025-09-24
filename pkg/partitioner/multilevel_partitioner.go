package partitioner

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/lintang-b-s/navigatorx-partitioner/pkg/datastructure"
	"golang.org/x/exp/rand"
)

type MulitlevelPartitioner struct {
	u []int //  cell size for  each cell levels. from biggest to smallest.
	// best parameter for customizable route planning by delling et al:
	// [2^8, 2^11, 2^14, 2^17, 2^20]
	l            int         // max level of overlay graph
	overlayNodes [][][]int32 // nodes in each cells in each level
	graph        *datastructure.Graph
}

func NewMultilevelPartitioner(u []int, l int, graph *datastructure.Graph) *MulitlevelPartitioner {
	if len(u) != l {
		panic(fmt.Sprintf("cell levels %d and cell array size %d must be the same", l, len(u)))
	}
	return &MulitlevelPartitioner{
		u:            u,
		l:            l,
		overlayNodes: make([][][]int32, l),
		graph:        graph,
	}
}

func (mp *MulitlevelPartitioner) Run(paramName string) error {
	// start from highest level
	nodeIDs := mp.graph.GetNodeIDs()

	// partitions original graph into cells with size <= u[l-1]
	log.Printf("partitioning level %d with max cell size %d", mp.l-1, mp.u[mp.l-1])
	if len(nodeIDs) > mp.u[mp.l-1] {
		iflow := NewInertialFlow(mp.u[mp.l-1])
		iflow.RecursiveBisection(nodeIDs, mp.graph)
		mp.overlayNodes[mp.l-1] = iflow.partitions
	} else {
		mp.overlayNodes[mp.l-1] = [][]int32{nodeIDs}
	}
	log.Printf("level %d done, total cells: %d", mp.l-1, len(mp.overlayNodes[mp.l-1]))

	// next partition each cell in previous level
	for level := mp.l - 2; level >= 0; level-- {
		log.Printf("partitioning level %d with max cell size %d", level, mp.u[level])
		for _, cell := range mp.overlayNodes[level+1] {
			// TODO: make each cell partitioning run concurently using goroutine pool
			iflow := NewInertialFlow(mp.u[level])
			iflow.RecursiveBisection(cell, mp.graph)
			mp.overlayNodes[level] = append(mp.overlayNodes[level], iflow.partitions...)
		}
		log.Printf("level %d done, total cells: %d", level, len(mp.overlayNodes[level]))
	}

	err := mp.writeMLPToFile(paramName)
	if err != nil {
		return err
	}
	return mp.writeMLPToMLPFile(fmt.Sprintf("multilevel_partitioning_%s.mlp", paramName))
}

func (mp *MulitlevelPartitioner) RunMLPKaffpa(name string) error {
	// start from highest level
	nodeIDs := mp.graph.GetNodeIDs()

	// partitions original graph into cells with size <= u[l-1]
	log.Printf("partitioning level %d with max cell size %d", mp.l-1, mp.u[mp.l-1])
	if len(nodeIDs) > mp.u[mp.l-1] {
		kaffpa := newKaffpaPartitioner(mp.graph, nodeIDs)
		partitions, err := kaffpa.partitionCell(mp.l-1, 0, name, mp.u[mp.l-1])
		if err != nil {
			return err
		}
		mp.overlayNodes[mp.l-1] = partitions
	} else {
		mp.overlayNodes[mp.l-1] = [][]int32{nodeIDs}
	}
	log.Printf("level %d done, total cells: %d", mp.l-1, len(mp.overlayNodes[mp.l-1]))

	// next partition each cell in previous level
	for level := mp.l - 2; level >= 0; level-- {
		log.Printf("partitioning level %d with max cell size %d", level, mp.u[level])
		for cellId, cell := range mp.overlayNodes[level+1] {
			log.Printf("partitioning cell %d in level %d", cellId, level+1)
			kaffpa := newKaffpaPartitioner(mp.graph, cell)
			partitions, err := kaffpa.partitionCell(level, cellId, name, mp.u[level])
			if err != nil {
				return err
			}
			mp.overlayNodes[level] = append(mp.overlayNodes[level], partitions...)
			log.Printf("level %d, cell %d done, total cells: %d", level, cellId, len(mp.overlayNodes[level]))

		}

		log.Printf("level %d done, total cells: %d", level, len(mp.overlayNodes[level]))
		mp.savePartitionsToFile(mp.overlayNodes[level], mp.graph, name, level)
	}
	return mp.writeMLPToMLPFile(fmt.Sprintf("kaffpa_%s.mlp", name))
}

// writeMLPToFile. write each level in separate txt file
func (mp *MulitlevelPartitioner) writeMLPToFile(paramName string) error {

	for i := 0; i < mp.l; i++ {
		filename := fmt.Sprintf("multilevel_partitioning_level_%d_u_%d_%s.txt", i, mp.u[i], paramName)

		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		nodeIDCellMap := make(map[int32]int)
		for cellID, cell := range mp.overlayNodes[i] {
			for _, nodeID := range cell {
				nodeIDCellMap[nodeID] = cellID
			}
		}

		log.Printf("level %d, total nodes: %d", i, len(nodeIDCellMap))
		for _, nodeID := range mp.graph.GetNodeIDs() {
			cellID, exists := nodeIDCellMap[nodeID]
			if !exists {
				return err
			}

			_, err := f.WriteString(fmt.Sprintf("%d\n", cellID))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (mp *MulitlevelPartitioner) writeMLPToMLPFile(filename string) error {

	numCells := make([]int, mp.l)
	for i := 0; i < mp.l; i++ {
		numCells[i] = len(mp.overlayNodes[i])
	}

	pvOffset := make([]int, mp.l+1)
	for i := 0; i < mp.l; i++ {
		pvOffset[i+1] = pvOffset[i] + int(math.Ceil(math.Log2(float64(numCells[i])))) // ceil(log2(numCells[i])) = number of bits needed to represent cell id in level-i
	}

	cellNumbers := make([]uint64, len(mp.graph.GetNodeIDs())) // 64 bit integer. rightmost contain level 0 cellId, leftmost contain level l-1 cellId

	for l := 0; l < mp.l; l++ {
		for cellId, vertexIds := range mp.overlayNodes[l] {
			for _, vertexId := range vertexIds {
				cellNumbers[vertexId] |= uint64(cellId) << uint64(pvOffset[l])
			}
		}
	}

	overlayEdgeCount := 0
	nonOverlayEdgeCount := 0
	for _, u := range mp.graph.ContractedNodes {

		for _, eId := range mp.graph.GetNodeFirstOutEdges(u.ID) {
			e := mp.graph.GetOutEdge(eId)
			v := e.ToNodeID

			if cellNumbers[u.ID] != cellNumbers[v] {
				overlayEdgeCount++
			} else {
				nonOverlayEdgeCount++
			}
		}
	}

	log.Printf("overlayEdgeCount: %d, nonOverlayEdgeCount: %d", overlayEdgeCount, nonOverlayEdgeCount)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%d\n", len(numCells)))
	if err != nil {
		return err
	}

	for i := 0; i < len(numCells); i++ {
		_, err := f.WriteString(fmt.Sprintf("%d\n", numCells[i]))
		if err != nil {
			return err
		}
	}

	_, err = f.WriteString(fmt.Sprintf("%d\n", len(mp.graph.GetNodeIDs())))
	if err != nil {
		return err
	}

	for _, vertexID := range mp.graph.GetNodeIDs() {
		_, err := f.WriteString(fmt.Sprintf("%d\n", cellNumbers[vertexID]))
		if err != nil {
			return err
		}
	}
	return nil
}

func (mp *MulitlevelPartitioner) savePartitionsToFile(partitions [][]int32, graph *datastructure.Graph,
	name string, level int) error {
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

	if err := os.WriteFile(fmt.Sprintf("nodePerPartitions_%s_level_%v.json", name, level), buf, 0644); err != nil {
		return err
	}
	return nil
}
