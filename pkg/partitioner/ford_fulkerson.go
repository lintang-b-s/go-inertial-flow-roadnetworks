package partitioner

import "github.com/lintang-b-s/navigatorx-partitioner/pkg/datastructure"

type FordFulkerson struct {
	v             int32
	edgeList      []maxFlowEdge
	adjacencyList [][]int32
	parent        []int32
	visited       []bool
	maxflow       float64
}

func NewFordFulkerson(v int32) *FordFulkerson {
	ff := &FordFulkerson{
		v:             v,
		edgeList:      make([]maxFlowEdge, 0),
		adjacencyList: make([][]int32, v),
		parent:        make([]int32, v),
		visited:       make([]bool, v),
		maxflow:       0.0,
	}

	return ff
}

func (d *FordFulkerson) addEdge(u, v int32, w float64, _ bool) {
	if u == v {
		return
	}
	// forward edge
	d.edgeList = append(d.edgeList, newMaxFlowEdge(v, w, 0.0))
	d.adjacencyList[u] = append(d.adjacencyList[u], int32(len(d.edgeList)-1))

	d.edgeList = append(d.edgeList, newMaxFlowEdge(u, 0.0, 0.0))
	d.adjacencyList[v] = append(d.adjacencyList[v], int32(len(d.edgeList)-1))
	// notes: we add reverse edge in consecutive order, so we can use edgeID^1 to get the reverse edge
}

func (d *FordFulkerson) dfs(u, t int32, flow float64) float64 {
	if u == t {
		return flow
	}
	d.visited[u] = true
	for _, edgeID := range d.adjacencyList[u] {
		edge := &d.edgeList[edgeID]
		residual := edge.capacity - edge.flow
		if !d.visited[edge.v] && residual > 0.0 {
			bottleneck := d.dfs(edge.v, t, min(flow, residual))

			if bottleneck > 0.0 {
				edge.flow += bottleneck
				d.edgeList[edgeID^1].flow -= bottleneck
				return bottleneck
			}
		}
	}
	return 0.0
}

func (d *FordFulkerson) run(s, t int32) float64 {

	for flow := d.dfs(s, t, inf); flow > 0.0; flow = d.dfs(s, t, inf) {
		for i := range d.visited {
			d.visited[i] = false
		}
		d.maxflow += flow
	}
	return d.maxflow
}

func (d *FordFulkerson) dfsMinCutMultipleSourcesSinks(u int32, bisectedGraph *[2][]int32, visited []bool,
	s, t int32, cutEdges *[]datastructure.Edge) {
	if u != s && u != t {
		bisectedGraph[0] = append(bisectedGraph[0], u)
	}

	visited[u] = true
	for _, edgeIdx := range d.adjacencyList[u] {
		edge := d.edgeList[edgeIdx]
		residualCapacity := edge.capacity - edge.flow
		if residualCapacity > 0 && !visited[edge.v] {
			d.dfsMinCutMultipleSourcesSinks(edge.v, bisectedGraph, visited,
				s, t, cutEdges)
		}
	}

	if u == s {
		for otherNode := int32(0); otherNode < d.v; otherNode++ {
			if !visited[otherNode] && otherNode != t && otherNode != s {
				bisectedGraph[1] = append(bisectedGraph[1], otherNode)
			}
		}
	}
}


func (d *FordFulkerson) dfsMinCut(u int32, bisectedGraph *[2][]int32, visited []bool, s int32,
	cutEdges *[]datastructure.Edge) {

	bisectedGraph[0] = append(bisectedGraph[0], u)
	visited[u] = true
	for _, edgeIdx := range d.adjacencyList[u] {
		edge := d.edgeList[edgeIdx]
		residualCapacity := edge.capacity - edge.flow
		if residualCapacity > 0 && !visited[edge.v] {
			d.dfsMinCut(edge.v, bisectedGraph, visited, s, cutEdges)
		}
	}

	if u == s {
		for otherNode := int32(0); otherNode < d.v; otherNode++ {
			if !visited[otherNode] {
				bisectedGraph[1] = append(bisectedGraph[1], otherNode)
			}
		}
	}

}
