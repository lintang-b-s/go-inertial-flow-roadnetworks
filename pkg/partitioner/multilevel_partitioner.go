package partitioner

import (
	"fmt"
	"log"
	"os"

	"github.com/lintang-b-s/go-graph-inertial-flow/pkg/datastructure"
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

func (mp *MulitlevelPartitioner) Run(paramName string) {
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

	mp.writeMLPToFile(paramName)
}

// writeMLPToFile. write each level in separate txt file
func (mp *MulitlevelPartitioner) writeMLPToFile(paramName string) {

	for i := 0; i < mp.l; i++ {
		filename := fmt.Sprintf("multilevel_partitioning_level_%d_u_%d_%s.txt", i, mp.u[i], paramName)

		f, err := os.Create(filename)
		if err != nil {
			panic(err)
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
				panic(fmt.Sprintf("nodeID %d not found in cell map", nodeID))
			}

			_, err := f.WriteString(fmt.Sprintf("%d\n", cellID))
			if err != nil {
				panic(err)
			}
		}
	}
}
