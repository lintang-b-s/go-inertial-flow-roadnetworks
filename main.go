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
	mapFile    = flag.String("f", "solo_jogja.osm.pbf", "openstreeetmap file buat road network graphnya")
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

	r := int(math.Pow(2, 17))
	partitioner.NewInertialFlow(int32(len(graph.GetNodes())), r).RunInertialFlow(graph)

}
