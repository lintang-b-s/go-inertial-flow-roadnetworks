package partitioner

type edge struct {
	to     int32
	weight float64
}

func newEdge(to int32, weight float64) edge {
	return edge{
		to:     to,
		weight: weight,
	}
}
