package partitioner

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/lintang-b-s/navigatorx-partitioner/pkg/datastructure"
)

type KaffpaPartitioner struct {
	nodeIds                        []int32
	kaffpaNodeIdsToOriginalNodeIds []int32
	graph                          *datastructure.Graph
}

func newKaffpaPartitioner(graph *datastructure.Graph, parentCellNodeIds []int32) *KaffpaPartitioner {
	return &KaffpaPartitioner{
		graph:   graph,
		nodeIds: parentCellNodeIds,
	}
}

func (kp *KaffpaPartitioner) partitionCell(level, cellId int, name string, cellSize int) ([][]int32, error) {

	err := kp.saveGraphToFile(fmt.Sprintf("./data/%s_level_%d_cell_%d", name, level, cellId))
	if err != nil {
		return nil, err
	}
	err = kp.runKaffpa(fmt.Sprintf("./data/%s_level_%d_cell_%d.graph", name, level, cellId), fmt.Sprintf("./data/%s_level_%d_cell_%d_part", name, level, cellId), cellSize)
	if err != nil {
		return [][]int32{}, err
	}
	return kp.readPartitionResult(fmt.Sprintf("./data/%s_level_%d_cell_%d_part", name, level, cellId))
}

func (kp *KaffpaPartitioner) runKaffpa(inputName, outputName string, cellSize int) error {
	k := int(math.Ceil(float64(len(kp.nodeIds)) / float64(cellSize)))
	log.Printf("running kaffpa with k=%d, cellSize=%d", k, cellSize)
	os, err := exec.Command("/home/lintangbs/KaHIP/deploy/kaffpa", inputName, fmt.Sprintf("--output=%s", outputName),
		fmt.Sprintf("--k=%v", k), fmt.Sprintf("--preconfiguration=strong")).CombinedOutput()
	if err != nil || len(os) == 0 {
		return err
	}

	fmt.Printf("%s\n", os)
	return nil
}

func (kp *KaffpaPartitioner) readPartitionResult(filename string) ([][]int32, error) {
	f, err := os.Open(filename)
	if err != nil {
		return [][]int32{}, err
	}

	defer f.Close()

	br := bufio.NewReader(f)

	readLine := func() (string, error) {
		line, err := br.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) && len(line) > 0 {
			} else {
				return "", err
			}
		}
		return strings.TrimRight(line, "\r\n"), nil
	}

	partitionResult := make([][]int32, 0)

	for i := 0; i < len(kp.nodeIds); i++ {
		line, err := readLine()
		if err != nil {
			return [][]int32{}, err
		}
		partIdStr := strings.TrimSpace(line)
		partId, err := strconv.Atoi(partIdStr)
		if err != nil {
			return [][]int32{}, err
		}

		for len(partitionResult) <= partId {
			// extend partitionResult slice if current partId exceed current length
			partitionResult = append(partitionResult, []int32{})
		}

		partitionResult[partId] = append(partitionResult[partId], kp.kaffpaNodeIdsToOriginalNodeIds[i])
	}

	return partitionResult, nil
}

func (kp *KaffpaPartitioner) saveGraphToFile(filename string) error {

	file, err := os.Create(fmt.Sprintf(`%v.graph`, filename))
	if err != nil {
		return err
	}
	defer file.Close()
	nodeEdges := make([][]edge, len(kp.nodeIds))
	adjs := make([]map[int32]bool, len(kp.nodeIds))
	inCell := make(map[int32]bool, len(kp.nodeIds))

	kp.kaffpaNodeIdsToOriginalNodeIds = make([]int32, len(kp.nodeIds))
	originalNodeIdToKaffpaNodeId := make(map[int32]int32, len(kp.nodeIds))
	for idx, id := range kp.nodeIds {
		inCell[id] = true
		originalNodeIdToKaffpaNodeId[id] = int32(idx + 1)
		kp.kaffpaNodeIdsToOriginalNodeIds[idx] = id
	}

	for id, nodeId := range kp.nodeIds {
		adjs[id] = make(map[int32]bool)

		for _, outEdgeIDx := range kp.graph.GetNodeFirstOutEdges(nodeId) {
			outEdge := kp.graph.GetOutEdge(outEdgeIDx)
			adjs[id][outEdge.ToNodeID] = true
			if !inCell[outEdge.ToNodeID] || nodeId == outEdge.ToNodeID {
				continue
			}
			nodeEdges[id] = append(nodeEdges[id], newEdge(originalNodeIdToKaffpaNodeId[outEdge.ToNodeID], outEdge.Weight))
		}

		for _, inEdgeIDx := range kp.graph.GetNodeFirstInEdges(nodeId) {
			inEdge := kp.graph.GetInEdge(inEdgeIDx)
			if !inCell[inEdge.ToNodeID] || nodeId == inEdge.ToNodeID {
				continue
			}
			if _, found := adjs[id][inEdge.ToNodeID]; found {
				continue
			}
			adjs[id][inEdge.ToNodeID] = true
			nodeEdges[id] = append(nodeEdges[id], newEdge(originalNodeIdToKaffpaNodeId[inEdge.ToNodeID], inEdge.Weight))
		}
	}

	edgeCounter := 0
	for _, edges := range nodeEdges {
		edgeCounter += len(edges)
	}

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(fmt.Sprintf("%d %d %d\n", len(kp.nodeIds), int(edgeCounter/2), 1))
	if err != nil {
		return err
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	for vId, edges := range nodeEdges {

		for eIdx, edge := range edges {
			_, err = writer.WriteString(fmt.Sprintf("%d %v", edge.to, int(edge.weight*100+1)))
			if err != nil {
				return err
			}
			if eIdx < len(edges)-1 {
				_, err = writer.WriteString(" ")
				if err != nil {
					return err
				}
			}
		}
		if vId < len(nodeEdges)-1 || (vId == len(nodeEdges)-1 && len(edges) == 0) {
			_, err = writer.WriteString("\n")
			if err != nil {
				return err
			}
		}
	}

	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}
