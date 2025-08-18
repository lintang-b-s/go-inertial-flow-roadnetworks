package partitioner

import (
	"container/list"
	"math"

	"github.com/lintang-b-s/go-graph-inertial-flow/pkg/datastructure"
)

const (
	inf     = 1e18
	infFlow = 1e10
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

func (d *DinicMinCut) addEdge(u, v int32, w float64, directed bool) {
	if u == v {
		return
	}
	d.edgeList = append(d.edgeList, newMaxFlowEdge(v, w, 0))
	d.adjacencyList[u] = append(d.adjacencyList[u], int32(len(d.edgeList)-1))
	if !directed {
		d.edgeList = append(d.edgeList, newMaxFlowEdge(u, w, 0))
	} else {
		d.edgeList = append(d.edgeList, newMaxFlowEdge(u, 0, 0))
	}
	d.adjacencyList[v] = append(d.adjacencyList[v], int32(len(d.edgeList)-1))
}

func (d *DinicMinCut) bfs(s, t int32) bool { // find augmenting path
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
			if (cap-flow) > 0 && d.level[v] == -1 {
				// positive residual edge
				d.level[v] = d.level[uVal] + 1
				queue.PushBack(v)
			}
		}
	}

	return d.level[t] != -1 // has an augmenting path
}

func (d *DinicMinCut) dfs(u, t int32, flow float64) float64 { // traverse from s->t
	if (u == t) || (flow == 0) {
		return flow
	}

	for ; d.last[u] < int32(len(d.adjacencyList[u])); d.last[u]++ { // from last edge
		edgeID := d.adjacencyList[u][d.last[u]]
		edge := &d.edgeList[edgeID]
		v := edge.v
		if d.level[v] != d.level[u]+1 {
			continue // not part of layer graph
		}
		if pushed := d.dfs(edge.v, t, min(flow, edge.capacity-edge.flow)); pushed > 0 {
			edge.flow += pushed
			d.edgeList[edgeID^1].flow -= pushed // subtract flow from reverse edge
			// we add reverse edge in consecutive order, so we can use edgeID^1 to get the reverse edge
			return pushed
		}

	}
	return 0
}

func (d *DinicMinCut) dinic(s, t int32) float64 {
	maxflow := 0.0
	for {
		if !d.bfs(s, t) {
			break
		}
		d.last = make([]int32, d.v)
		for {
			pushed := d.dfs(s, t, math.Inf(1))
			if pushed <= 0 {
				break
			}
			maxflow += pushed
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
