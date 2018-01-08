package graph

import (
	"reflect"
	"sort"
	"testing"
)

func TestMaxSpanningTree(t *testing.T) {
	cases := []struct {
		nodes       []int
		edges, want []WEdge
	}{{
		nodes: []int{1, 2, 3, 4, 5, 6, 7},
		edges: []WEdge{
			{1, 2, 7},
			{1, 4, 8},
			{4, 2, 6},
			{2, 3, 4},
			{2, 5, 5},
			{3, 5, 9},
			{4, 5, 1},
			{4, 6, 10},
			{6, 5, 11},
			{6, 7, 3},
			{5, 7, 2},
		},
		want: []WEdge{
			{6, 5, 11},
			{4, 6, 10},
			{3, 5, 9},
			{1, 4, 8},
			{1, 2, 7},
			{6, 7, 3},
		},
	}}
	for _, tt := range cases {
		got := MaxSpanningTree(tt.nodes, tt.edges)
		if got == nil {
			t.Fatalf("returned empty graph for edges:\n%v\n", tt.edges)
		}
		sort.Slice(got, func(i int, j int) bool {
			return got[i].Weight > got[j].Weight
		})
		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("wrong tree!\nexpected:\n%v\ngot:\n%v\n", tt.want, got)
		}
	}
}
