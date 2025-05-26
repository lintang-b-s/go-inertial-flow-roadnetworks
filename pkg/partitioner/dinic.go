package partitioner

import (
	"container/list"

	"github.com/lintang-b-s/go-graph-inertial-flow/pkg/datastructure"
)

const (
	inf     = 1e18
	infFlow = 1e10
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
	d, last       []int32
	p             [][2]int32
}

func NewDinicMinCut(v int32) *DinicMinCut {
	dinic := &DinicMinCut{
		v:             v,
		edgeList:      make([]maxFlowEdge, 0),
		adjacencyList: make([][]int32, v),
		d:             make([]int32, 0),
		last:          make([]int32, 0),
		p:             make([][2]int32, 0),
	}

	return dinic
}

func (d *DinicMinCut) bfs(s, t int32) bool { // find augmenting path
	d.d = make([]int32, d.v)
	d.p = make([][2]int32, d.v)
	for i := int32(0); i < d.v; i++ {
		d.d[i] = -1
		d.p[i] = [2]int32{-1, -1} // record bfs sp tree
	}
	d.d[s] = 0

	queue := list.New()
	queue.PushBack(s)

	for queue.Len() != 0 {
		u := queue.Front()
		queue.Remove(u)
		if u.Value == t {
			// stop as sink t reached
			break
		}

		uVal := u.Value.(int32)
		for _, idx := range d.adjacencyList[uVal] {
			// explore neighbors of u
			v, cap, flow := d.edgeList[idx].v, d.edgeList[idx].capacity, d.edgeList[idx].flow
			if (cap-flow) > 0 && d.d[v] == -1 {
				// positive residual edge
				d.d[v] = d.d[uVal] + 1
				queue.PushBack(v)
				d.p[v] = [2]int32{uVal, idx}
			}
		}
	}

	return d.d[t] != -1 // has an augmenting path
}

func (d *DinicMinCut) dfs(u, t int32, flow float64) float64 { // traverse from s->t
	if (u == t) || (flow == 0) {
		return flow
	}

	for i := &d.last[u]; *i < int32(len(d.adjacencyList[u])); *i++ { // from last edge
		edge := &d.edgeList[d.adjacencyList[u][*i]]
		if d.d[edge.v] != d.d[u]+1 {
			continue // not part of layer graph
		}
		if pushed := d.dfs(edge.v, t, min(flow, edge.capacity-edge.flow)); pushed > 0 {
			edge.flow += pushed
			rflow := &d.edgeList[d.adjacencyList[u][*i]^1].flow
			*rflow -= pushed
			return pushed
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
		for from := int32(0); from < d.v; from++ {
			if !visited[from] {
				bisectedGraph[1] = append(bisectedGraph[1], from)
			}
			if visited[from] {
				for _, edgeIDx := range d.adjacencyList[from] {
					edge := d.edgeList[edgeIDx]
					if !visited[edge.v] && edge.capacity > 0 && edge.capacity-edge.flow == 0 {
						*cutEdges = append(*cutEdges, datastructure.NewEdgePlain(0, 0, 0, edge.v, from))
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
		for from := int32(0); from < d.v; from++ {
			if !visited[from] {
				if from != t {
					bisectedGraph[1] = append(bisectedGraph[1], from)
				}
			}
			if visited[from] {
				for _, edgeIDx := range d.adjacencyList[from] {
					edge := d.edgeList[edgeIDx]
					if !visited[edge.v] && edge.capacity > 0 && edge.capacity-edge.flow == 0 &&
						edge.v != t {
						*cutEdges = append(*cutEdges, datastructure.NewEdgePlain(0, 0, 0, edge.v, from))
					}
				}
			}
		}
	}

}
