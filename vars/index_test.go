package vars

import (
	"reflect"
	"testing"
)

func TestNewIndexFor(t *testing.T) {
	cases := []struct {
		indexVars, forVars VarList
		seq                []int
	}{{
		[]*Var{New(0, 2, "", false), New(5, 3, "", false)},
		[]*Var{New(0, 2, "", false), New(3, 2, "", false), New(5, 3, "", false)},
		[]int{0, 1, 0, 1, 2, 3, 2, 3, 4, 5, 4, 5},
	}, {
		NewList([]int{0, 3, 5}, []int{2, 2, 3}),
		NewList([]int{0, 3, 5}, []int{2, 2, 3}),
		[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	}}
	for _, tt := range cases {
		ix := NewIndexFor(tt.indexVars, tt.forVars)
		for i, v := range tt.seq {
			if ix.Ended() {
				t.Errorf("index ended at %v iterations, should have %v", i, len(tt.seq))
			}
			if v != ix.I() {
				t.Errorf("wrong index i[%v]=%v, should be i[%v]%v", i, ix.I(), i, tt.seq[i])
			}
			ix.Next()
		}
		ix.Reset()
		for i, v := range tt.seq {
			if ix.Ended() {
				t.Errorf("RESET index ended at %v iterations, should have %v", i, len(tt.seq))
			}
			if v != ix.I() {
				t.Errorf("wrong RESET index i[%v]=%v, should be i[%v]%v", i, ix.I(), i, tt.seq[i])
			}
			ix.Next()
		}
	}
}

func TestNextRight(t *testing.T) {
	cases := []struct {
		indexVars, forVars VarList
		seq                []int
	}{{
		NewList([]int{0}, []int{2}),
		NewList([]int{0}, []int{2}),
		[]int{0, 1},
	}, {
		NewList([]int{0, 3}, []int{2, 3}),
		NewList([]int{0, 3}, []int{2, 3}),
		[]int{0, 2, 4, 1, 3, 5},
	}}
	for _, tt := range cases {
		ix := NewIndexFor(tt.indexVars, tt.forVars)
		for i, v := range tt.seq {
			if ix.Ended() {
				t.Errorf("index ended at %v iterations, should have %v", i, len(tt.seq))
			}
			if v != ix.I() {
				t.Errorf("wrong index i[%v]=%v, should be i[%v]%v", i, ix.I(), i, tt.seq[i])
			}
			ix.NextRight()
		}
		ix.Reset()
		for i, v := range tt.seq {
			if ix.Ended() {
				t.Errorf("RESET index ended at %v iterations, should have %v", i, len(tt.seq))
			}
			if v != ix.I() {
				t.Errorf("wrong RESET index i[%v]=%v, should be i[%v]%v", i, ix.I(), i, tt.seq[i])
			}
			ix.NextRight()
		}
	}
}

func TestAttribution(t *testing.T) {
	cases := []struct {
		indexVars, forVars VarList
		attrbMaps          []map[int]int
	}{{
		NewList([]int{0, 2}, []int{2, 3}),
		NewList([]int{0, 2}, []int{2, 3}),
		[]map[int]int{
			map[int]int{0: 0, 2: 0},
			map[int]int{0: 1, 2: 0},
			map[int]int{0: 0, 2: 1},
			map[int]int{0: 1, 2: 1},
			map[int]int{0: 0, 2: 2},
			map[int]int{0: 1, 2: 2},
		},
	}}
	for _, tt := range cases {
		ix := NewIndexFor(tt.indexVars, tt.forVars)
		for i, m := range tt.attrbMaps {
			got := ix.Attribution()
			if !reflect.DeepEqual(got, m) {
				t.Errorf("wrong attribution map i[%v]=%v, should be %v", i, got, m)
			}
			ix.Next()
		}
	}
}
