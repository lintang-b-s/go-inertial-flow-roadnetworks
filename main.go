package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/lintang-b-s/navigatorx-partitioner/pkg/datastructure"
	"github.com/lintang-b-s/navigatorx-partitioner/pkg/osmparser"
	"github.com/lintang-b-s/navigatorx-partitioner/pkg/partitioner"
)

var (
	mapFile    = flag.String("f", "solo_jogja.osm.pbf", "openstreeetmap file buat road network graphnya")
	npartition = flag.Int("n", 16, "jumlah partisi yang diinginkan, default 16")
)

func main() {
	dir := "data"
	if _, err := os.Stat("dir"); os.IsNotExist(err) {
		// Create directory if not exists
		err := os.MkdirAll(dir, 0755) // 0755 = rwxr-xr-x
		if err != nil {
			panic(err)
		}
		fmt.Println("Created directory:", dir)
	} else {
		fmt.Println("Directory already exists:", dir)
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

	err := mlp.RunMLPKaffpa("kaffpa_solo_jogja")
	if err != nil {
		panic(err)
	}
}
