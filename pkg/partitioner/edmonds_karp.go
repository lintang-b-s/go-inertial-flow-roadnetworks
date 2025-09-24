package partitioner

import (
	"container/list"

	"github.com/lintang-b-s/navigatorx-partitioner/pkg/datastructure"
)

type edmondsMaxFlowEdge struct {
	v        int32
	u        int32
	capacity float64
	flow     float64
	id       int
}

type EdmondsKarp struct {
	v             int32
	edgeList      []edmondsMaxFlowEdge
	adjacencyList [][]int32
	parent        []int32
	visited       []bool
	maxflow       float64
}

func NewEdmondsKarp(v int32) *EdmondsKarp {
	ek := &EdmondsKarp{
		v:             v,
		edgeList:      make([]edmondsMaxFlowEdge, 0),
		adjacencyList: make([][]int32, v),
		parent:        make([]int32, v),
		visited:       make([]bool, v),
		maxflow:       0,
	}

	return ek

}

func newEdmondsMaxFlowEdge(u, v int32, capacity, flow float64, id int) edmondsMaxFlowEdge {
	return edmondsMaxFlowEdge{v: v, u: u, capacity: capacity, flow: flow, id: id}
}

func (d *EdmondsKarp) addEdge(u, v int32, w float64, _ bool) {
	if u == v {
		return
	}
	// forward edge
	d.edgeList = append(d.edgeList, newEdmondsMaxFlowEdge(u, v, w, 0.0, len(d.edgeList)))
	d.adjacencyList[u] = append(d.adjacencyList[u], int32(len(d.edgeList)-1))

	d.edgeList = append(d.edgeList, newEdmondsMaxFlowEdge(v, u, 0.0, 0.0, len(d.edgeList)))
	d.adjacencyList[v] = append(d.adjacencyList[v], int32(len(d.edgeList)-1))
	// notes: we add reverse edge in consecutive order, so we can use edgeID^1 to get the reverse edge
}

func (d *EdmondsKarp) bfs(s, t int32) float64 {
	prev := make([]*edmondsMaxFlowEdge, d.v)

	queue := list.New()

	d.visited[s] = true
	queue.PushBack(s)

	for queue.Len() > 0 {
		node := queue.Remove(queue.Front()).(int32)
		if node == t {
			break
		}

		for _, edgeID := range d.adjacencyList[node] {
			edge := &d.edgeList[edgeID]
			residual := edge.capacity - edge.flow
			if !d.visited[edge.v] && residual > 0.0 {
				d.visited[edge.v] = true
				prev[edge.v] = edge
				queue.PushBack(edge.v)
			}
		}
	}

	if prev[t] == nil {
		return 0.0
	}
	bottleneck := inf

	for edge := prev[t]; edge != nil; edge = prev[edge.u] {
		residual := edge.capacity - edge.flow
		bottleneck = min(bottleneck, residual)
	}

	for edge := prev[t]; edge != nil; edge = prev[edge.u] {
		edge.flow += bottleneck
		d.edgeList[(*edge).id^1].flow -= bottleneck
	}

	return bottleneck
}

func (d *EdmondsKarp) run(s, t int32) float64 {
	flow := 0.0
	for {
		for i := range d.visited {
			d.visited[i] = false
		}

		flow = d.bfs(s, t)
		d.maxflow += flow

		if flow == 0.0 {
			break
		}
	}

	return d.maxflow
}

func (d *EdmondsKarp) dfsMinCutMultipleSourcesSinks(u int32, bisectedGraph *[2][]int32, visited []bool,
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

func (d *EdmondsKarp) dfsMinCut(u int32, bisectedGraph *[2][]int32, visited []bool, s int32,
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
