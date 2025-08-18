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

	// r = regionSize
	// r := int(math.Pow(2, 15)) 
	r := int(math.Pow(2, 12))
	partitioner.NewInertialFlow(int32(len(graph.GetNodes())), r).RunInertialFlow(graph)

	// if *useMetis {
	// 	// partitioner.SaveMetisGraphFile(graph, strings.Split(*mapFile, ".")[0])
	// 	// run gpmetis ...
	// 	// partitioner.ReadMetisGraphFile("nyc.graph.part.64", graph, 64)
	// }
}
