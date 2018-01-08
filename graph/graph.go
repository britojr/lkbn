package graph

import "sort"

// WEdge defines a weighted edge
type WEdge struct {
	Head, Tail int
	Weight     float64
}

// MaxSpanningTree receives a list of weighted edges of a graph
// and returns the list of weighted edges of a maximum spanning tree
func MaxSpanningTree(nodes []int, edges []WEdge) (mst []WEdge) {
	// sort edges by descending weight
	sort.Slice(edges, func(i int, j int) bool {
		return edges[i].Weight > edges[j].Weight
	})
	// initialize each node as a separate component
	components := make(map[int]*[]int)
	for _, nd := range nodes {
		components[nd] = &[]int{nd}
	}
	// connect components using Kruskal's algorithm
	for _, e := range edges {
		aComp, bComp := components[e.Head], components[e.Tail]
		if aComp != bComp {
			// add this edge if it connects two different components
			mst = append(mst, e)
			// finish if n-1 edges have been added
			if len(mst) == len(nodes)-1 {
				break
			}
			// merge connected component
			*aComp = append(*aComp, *bComp...)
			for _, c := range *bComp {
				components[c] = aComp
			}
		}
	}
	return
}

// RootedTree returns a set of edges ordered according the given root node
func RootedTree(root int, edges []WEdge) (rt []WEdge) {
	// TODO: find a way to keep the weight of the edges
	// create adj list
	adj := make(map[int][]int)
	for _, e := range edges {
		adj[e.Head] = append(adj[e.Head], e.Tail)
		adj[e.Tail] = append(adj[e.Tail], e.Head)
	}
	// start visit from root
	visit := make(map[int]struct{})
	visit[root] = struct{}{}
	queue := []int{root}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		for _, u := range adj[v] {
			if _, ok := visit[u]; !ok {
				visit[u] = struct{}{}
				rt = append(rt, WEdge{Head: v, Tail: u})
				queue = append(queue, u)
			}
		}
	}
	return
}
