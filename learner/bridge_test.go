package learner

import (
	"reflect"
	"testing"

	"github.com/britojr/lkbn/vars"
)

type fakeMICalc struct {
	mat [][]float64
}

func (f fakeMICalc) Get(i, j int) float64 {
	if j < i {
		return f.mat[i][j]
	}
	return f.mat[j][i]
}

func TestGroupVariables(t *testing.T) {
	cases := []struct {
		vs      vars.VarList
		k       int
		mutInfo mutInfCalc
		want    []vars.VarList
	}{{
		vars.NewList([]int{0, 1, 2, 3, 4}, nil), 1, nil,
		[]vars.VarList{
			vars.NewList([]int{0}, nil), vars.NewList([]int{1}, nil),
			vars.NewList([]int{2}, nil), vars.NewList([]int{3}, nil), vars.NewList([]int{4}, nil),
		},
	}, {
		vars.NewList([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, nil), 3, fakeMICalc{[][]float64{
			{100},
			{99, 100},
			{1, 1, 100},
			{1, 1, 98, 100},
			{1, 1, 1, 1, 100},
			{1, 1, 1, 1, 97, 100},
			{2, 2, 2, 2, 2, 2, 100},
			{1, 1, 1, 1, 96, 95, 1, 100},
			{1, 1, 94, 93, 1, 1, 1, 1, 100},
			{92, 91, 1, 1, 1, 1, 1, 1, 1, 100},
		}},
		[]vars.VarList{
			vars.NewList([]int{0, 1, 9}, nil), vars.NewList([]int{2, 3, 8}, nil),
			vars.NewList([]int{4, 5, 7}, nil), vars.NewList([]int{6}, nil),
		},
	}, {
		vars.NewList([]int{0, 1, 2, 3, 4, 5, 6, 7}, nil), 3, fakeMICalc{[][]float64{
			{100},
			{99, 100},
			{1, 1, 100},
			{1, 1, 98, 100},
			{1, 1, 1, 1, 100},
			{1, 1, 1, 1, 97, 100},
			{1, 1, 94, 93, 1, 1, 100},
			{96, 95, 1, 1, 1, 1, 1, 100},
		}},
		[]vars.VarList{
			vars.NewList([]int{0, 1, 7}, nil), vars.NewList([]int{2, 3, 6}, nil),
			vars.NewList([]int{4, 5}, nil),
		},
	}}
	for _, tt := range cases {
		got := groupVariables(tt.vs, tt.k, tt.mutInfo)
		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("wrong grops,  want:\n%v\ngot:\n%v\n", tt.want, got)
		}
	}
}

func TestComputeGroupedMI(t *testing.T) {
	cases := []struct {
		mutInfo mutInfCalc
		gs      []vars.VarList
		want    map[string]map[string]float64
	}{{
		fakeMICalc{[][]float64{
			{100},
			{99, 100},
			{1, 1, 100},
			{1, 20, 98, 100},
			{1, 1, 1, 21, 100},
			{1, 1, 1, 10, 97, 100},
			{1, 10, 94, 93, 1, 1, 100},
			{96, 95, 1, 1, 1, 22, 1, 100},
		}},
		[]vars.VarList{
			vars.NewList([]int{0, 1, 7}, nil), vars.NewList([]int{2, 3, 6}, nil),
			vars.NewList([]int{4, 5}, nil),
		}, map[string]map[string]float64{
			"X0[2] X1[2] X7[2]": {
				"X2[2] X3[2] X6[2]": 20,
				"X4[2] X5[2]":       22,
			},
			"X2[2] X3[2] X6[2]": {
				"X0[2] X1[2] X7[2]": 20,
				"X4[2] X5[2]":       21,
			},
			"X4[2] X5[2]": {
				"X0[2] X1[2] X7[2]": 22,
				"X2[2] X3[2] X6[2]": 21,
			},
		},
	}}
	for _, tt := range cases {
		got := computeGroupedMI(tt.gs, tt.mutInfo)
		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("wrong gropMI,  want:\n%v\ngot:\n%v\n", tt.want, got)
		}
	}
}
