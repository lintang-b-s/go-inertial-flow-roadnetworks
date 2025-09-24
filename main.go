package main

import (
	"flag"
	"math"
	"os"

	"github.com/lintang-b-s/navigatorx-partitioner/pkg/datastructure"
	"github.com/lintang-b-s/navigatorx-partitioner/pkg/osmparser"
	"github.com/lintang-b-s/navigatorx-partitioner/pkg/partitioner"
)

var (
	mapFile = flag.String("f", "solo_jogja.osm.pbf", "openstreeetmap file buat road network graphnya")
)

func main() {
	dir := "data"
	if _, err := os.Stat("dir"); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
	flag.Parse()
	osmParser := osmparser.NewOSMParserV2()
	processedNodes, graphStorage, streetDirection := osmParser.Parse(*mapFile)

	graph := datastructure.NewGraph()
	graph.InitGraph(processedNodes, graphStorage, streetDirection, osmParser.GetTagStringIdMap())

	mlp := partitioner.NewMultilevelPartitioner(
		[]int{int(math.Pow(2, 8)), int(math.Pow(2, 11)), int(math.Pow(2, 14)), int(math.Pow(2, 17)), int(math.Pow(2, 20))},
		5,
		graph,
	)

	err := mlp.RunMLPKaffpa("kaffpa_test_5_level_crp")
	if err != nil {
		panic(err)
	}
}
