package partitioner

import (
	"container/list"

	"github.com/lintang-b-s/go-graph-inertial-flow/pkg/datastructure"
)

const (
	inf     = 1e18
	infFlow = 1e18
	eps     = 1e-12
)

type maxFlowEdge struct {
	v        int32
	capacity float64
	flow     float64
}

func newMaxFlowEdge(v int32, capacity, flow float64) maxFlowEdge {
	return maxFlowEdge{v: v, capacity: capacity, flow: flow}
}

type DinicMinCut struct {
	v             int32
	edgeList      []maxFlowEdge
	adjacencyList [][]int32
	level, last   []int32
}

func NewDinicMinCut(v int32) *DinicMinCut {
	dinic := &DinicMinCut{
		v:             v,
		edgeList:      make([]maxFlowEdge, 0),
		adjacencyList: make([][]int32, v),
		level:         make([]int32, 0),
		last:          make([]int32, 0),
	}

	return dinic
}


func (d *DinicMinCut) addEdge(u, v int32, w float64, _ bool) {
	if u == v {
		return
	}
	// forward edge
	d.edgeList = append(d.edgeList, newMaxFlowEdge(v, w, 0))
	d.adjacencyList[u] = append(d.adjacencyList[u], int32(len(d.edgeList)-1))

	d.edgeList = append(d.edgeList, newMaxFlowEdge(u, 0, 0))
	d.adjacencyList[v] = append(d.adjacencyList[v], int32(len(d.edgeList)-1))
	// notes: we add reverse edge in consecutive order, so we can use edgeID^1 to get the reverse edge
}

func (d *DinicMinCut) bfs(s, t int32) bool { // build level graph
	d.level = make([]int32, d.v)
	for i := int32(0); i < d.v; i++ {
		d.level[i] = -1
	}
	d.level[s] = 0

	queue := list.New()
	queue.PushBack(s)

	for queue.Len() != 0 {
		u := queue.Front()
		queue.Remove(u)

		uVal := u.Value.(int32)
		for _, idx := range d.adjacencyList[uVal] {
			// explore neighbors of u
			v, cap, flow := d.edgeList[idx].v, d.edgeList[idx].capacity, d.edgeList[idx].flow
			residual := (cap - flow)
			if residual > 0 && d.level[v] == -1 {
				// unvisited edge with positive residual capacity
				d.level[v] = d.level[uVal] + 1
				queue.PushBack(v)
			}
		}
	}

	return d.level[t] != -1 // return false if sink is not reachable from source (graph is saturated)
}

func (d *DinicMinCut) dfs(u, t int32, minFlow float64) float64 { // send flow along the level graph
	if (u == t) || (minFlow == 0) {
		return minFlow
	}

	for ; d.last[u] < int32(len(d.adjacencyList[u])); d.last[u]++ { // from last edge to avoid dead-ends
		edgeID := d.adjacencyList[u][d.last[u]]
		edge := &d.edgeList[edgeID]
		v := edge.v
		if d.level[v] != d.level[u]+1 {
			continue // not part of level graph
		}
		residual := edge.capacity - edge.flow
		if flow := d.dfs(edge.v, t, min(minFlow, residual)); flow > 0 {
			// augment flow
			edge.flow += flow
			d.edgeList[edgeID^1].flow -= flow // subtract flow from reverse edge
			// we add reverse edge in consecutive order, so we can use edgeID^1 to get the reverse edge
			return flow
		}

	}
	return 0
}

func (d *DinicMinCut) dinic(s, t int32) float64 {
	maxflow := 0.0
	for d.bfs(s, t) {
		d.last = make([]int32, d.v)
		flow := d.dfs(s, t, inf)
		for flow > 0 {
			maxflow += flow
			flow = d.dfs(s, t, inf)
		}
	}
	return maxflow
}

func (d *DinicMinCut) dfsMinCut(u int32, bisectedGraph *[2][]int32, visited []bool, s int32,
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
			} else {
				for _, edgeIDx := range d.adjacencyList[otherNode] {
					edge := d.edgeList[edgeIDx]
					residualCapacity := edge.capacity - edge.flow
					if !visited[edge.v] && edge.capacity > 0 && residualCapacity == 0 {
						*cutEdges = append(*cutEdges, datastructure.NewEdgePlain(0, 0, 0, edge.v, otherNode))
					}
				}
			}
		}
	}

}

func (d *DinicMinCut) dfsMinCutMultipleSourcesSinks(u int32, bisectedGraph *[2][]int32, visited []bool,
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
			} else {
				for _, edgeIDx := range d.adjacencyList[otherNode] {
					edge := d.edgeList[edgeIDx]
					residualCapacity := edge.capacity - edge.flow
					if !visited[edge.v] && edge.capacity > 0 && residualCapacity == 0 &&
						edge.v != t {
						*cutEdges = append(*cutEdges, datastructure.NewEdgePlain(0, 0, 0, edge.v, otherNode))
					}
				}
			}
		}
	}

}
