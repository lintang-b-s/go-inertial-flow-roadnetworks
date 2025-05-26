package partitioner

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/lintang-b-s/go-graph-inertial-flow/pkg/datastructure"
)

func SaveMetisGraphFile(graph *datastructure.Graph) {
	file, err := os.Create("file.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(fmt.Sprintf("%d %d %d\n", len(graph.GetNodes()), graph.GetOutEdgeCount()/2, 1))
	if err != nil {
		log.Fatal(err)
	}

	err = writer.Flush()
	if err != nil {
		panic(err)
	}
	edgeCounter := 0
	for _, node := range graph.GetNodes() {

		for ei, outEdgeIDx := range graph.GetNodeFirstOutEdges(node.ID) {
			outEdge := graph.GetOutEdge(outEdgeIDx)
			if ei < len(graph.GetNodeFirstOutEdges(node.ID))-1 {
				writer.WriteString(fmt.Sprintf("%d %v ", outEdge.ToNodeID+1, int(outEdge.Weight*1e6)))
			} else {
				writer.WriteString(fmt.Sprintf("%d %v", outEdge.ToNodeID+1, int(outEdge.Weight*1e6)))
			}
			edgeCounter++
		}
		writer.WriteString("\n")

	}

	fmt.Printf("edge counter: %v\n", edgeCounter)

	err = writer.Flush()
	if err != nil {
		panic(err)
	}

	// read txt file

	dat, err := os.ReadFile("file.txt")
	if err != nil {
		log.Fatal(err)
	}

	newEdgeCounter := 0
	datLines := strings.Split(string(dat), "\n")
	for i, line := range datLines {
		if i == 0 {
			continue
		}
		edges := strings.Split(line, "   ")
		for _ = range edges {
			newEdgeCounter++
		}
	}
	fmt.Printf("new edge counter: %v\n", newEdgeCounter)
}

func ReadMetisGraphFile(filePath string, graph *datastructure.Graph,
	nPartition int) {
	// read txt file

	dat, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	partitions := make([][]int32, nPartition)
	datLines := strings.Split(string(dat), "\n")
	for i, partition := range datLines {
		nodeID := i
		partitionID, err := strconv.Atoi(strings.TrimSpace(partition))
		if err != nil {
			panic(err)
		}
		partitions[partitionID] = append(partitions[partitionID], int32(nodeID))

	}

	iflow := NewInertialFlow(int32(len(graph.GetNodes())), nPartition)
	iflow.savePartitionsToFile(partitions, graph, "metis")
}
