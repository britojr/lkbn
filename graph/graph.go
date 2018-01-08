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
