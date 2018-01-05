package graph

import (
	"reflect"
	"sort"
	"testing"
)

func TestMaxSpanningTree(t *testing.T) {
	cases := []struct {
		edges, want []WEdge
	}{{
		edges: []WEdge{
			{"A", "B", 7},
			{"A", "D", 8},
			{"D", "B", 6},
			{"B", "C", 4},
			{"B", "E", 5},
			{"C", "E", 9},
			{"D", "E", 1},
			{"D", "F", 10},
			{"F", "E", 11},
			{"F", "G", 3},
			{"E", "G", 2},
		},
		want: []WEdge{
			{"A", "B", 7},
			{"A", "D", 8},
			{"C", "E", 9},
			{"D", "F", 10},
			{"F", "E", 11},
			{"F", "G", 3},
		},
	}}
	for _, tt := range cases {
		got := MaxSpanningTree(tt.edges)
		if got == nil {
			t.Fatalf("returned empty graph for edges:\n%v\n", tt.edges)
		}
		sort.Slice(got, func(i int, j int) bool {
			return got[i].Weight > got[j].Weight
		})
		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("wrong tree, expected:\n%v\ngot:\n%v\n", tt.want, got)
		}
	}
}
