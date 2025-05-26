package main

import (
	"flag"
	"log"
	"math"

	"github.com/lintang-b-s/go-graph-inertial-flow/pkg/datastructure"
	"github.com/lintang-b-s/go-graph-inertial-flow/pkg/osmparser"
	"github.com/lintang-b-s/go-graph-inertial-flow/pkg/partitioner"
)

var (
	mapFile    = flag.String("f", "nyc.osm.pbf", "openstreeetmap file buat road network graphnya")
	npartition = flag.Int("n", 16, "jumlah partisi yang diinginkan, default 16")
	useMetis   = flag.Bool("metis", false, "metis")
)

func main() {
	flag.Parse()
	log.Printf("use metis: %v", *useMetis)
	osmParser := osmparser.NewOSMParserV2(*useMetis)
	processedNodes, graphStorage, streetDirection := osmParser.Parse(*mapFile)

	graph := datastructure.NewGraph()
	graph.InitGraph(processedNodes, graphStorage, streetDirection, osmParser.GetTagStringIdMap())
	// using npartitons
	// r := int(math.Ceil(float64(len(graph.GetNodes())) / float64(*npartition))) // regionSize (number of nodes in each partition)

	// using balance
	// balance := 0.43                                               //  number of nodes in the smaller subgraph divided by the total number of nodes
	// r := int(math.Ceil(balance * float64(len(graph.GetNodes())))) // regionSize (number of nodes in each partition)
	r := int(math.Pow(2, 15))
	partitioner.NewInertialFlow(int32(len(graph.GetNodes())), r).RunInertialFlow(graph)

	// if *useMetis {
	// 	// partitioner.SaveMetisGraphFile(graph)
	// 	partitioner.ReadMetisGraphFile("nyc.graph.part.7", graph, 7)
	// }

}
