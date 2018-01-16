package learner

import (
	"reflect"
	"testing"

	"github.com/britojr/lkbn/graph"
	"github.com/britojr/lkbn/model"
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

func TestIncreaseCluster(t *testing.T) {
	gs := []vars.VarList{
		vars.NewList([]int{0, 1}, nil),
		vars.NewList([]int{2, 3}, nil),
		vars.NewList([]int{4, 5}, nil),
		vars.NewList([]int{6, 7}, nil),
		vars.NewList([]int{8, 9}, nil),
	}
	gpmi := make(map[string]map[string]float64)
	for i := range gs {
		gpmi[groupKey(gs[i])] = make(map[string]float64)
	}
	gpmi[groupKey(gs[1])][groupKey(gs[0])], gpmi[groupKey(gs[0])][groupKey(gs[1])] = 90, 90
	gpmi[groupKey(gs[2])][groupKey(gs[0])], gpmi[groupKey(gs[0])][groupKey(gs[2])] = 10, 10
	gpmi[groupKey(gs[2])][groupKey(gs[1])], gpmi[groupKey(gs[1])][groupKey(gs[2])] = 25, 25
	gpmi[groupKey(gs[3])][groupKey(gs[0])], gpmi[groupKey(gs[0])][groupKey(gs[3])] = 80, 80
	gpmi[groupKey(gs[3])][groupKey(gs[1])], gpmi[groupKey(gs[1])][groupKey(gs[3])] = 15, 15
	gpmi[groupKey(gs[3])][groupKey(gs[2])], gpmi[groupKey(gs[2])][groupKey(gs[3])] = 50, 50
	gpmi[groupKey(gs[4])][groupKey(gs[0])], gpmi[groupKey(gs[0])][groupKey(gs[4])] = 20, 20
	gpmi[groupKey(gs[4])][groupKey(gs[1])], gpmi[groupKey(gs[1])][groupKey(gs[4])] = 85, 85
	gpmi[groupKey(gs[4])][groupKey(gs[2])], gpmi[groupKey(gs[2])][groupKey(gs[4])] = 50, 50
	gpmi[groupKey(gs[4])][groupKey(gs[3])], gpmi[groupKey(gs[3])][groupKey(gs[4])] = 50, 50

	cases := []struct {
		gpMI   map[string]map[string]float64
		cl, gs []vars.VarList
		want   [][]vars.VarList
	}{{
		gpmi, append([]vars.VarList(nil), gs[0], gs[1]),
		append([]vars.VarList(nil), gs[2], gs[3], gs[4]),
		[][]vars.VarList{
			append([]vars.VarList(nil), gs[0]),
			append([]vars.VarList(nil), gs[1], gs[4]),
			append([]vars.VarList(nil), gs[0], gs[1], gs[4]),
			append([]vars.VarList(nil), gs[2], gs[3]),
		},
	}, {
		gpmi, append([]vars.VarList(nil), gs[0], gs[1], gs[4]),
		append([]vars.VarList(nil), gs[2], gs[3]),
		[][]vars.VarList{
			append([]vars.VarList(nil), gs[1], gs[4]),
			append([]vars.VarList(nil), gs[0], gs[3]),
			append([]vars.VarList(nil), gs[0], gs[1], gs[4], gs[3]),
			append([]vars.VarList(nil), gs[2]),
		},
	}}
	for _, tt := range cases {
		got := make([][]vars.VarList, 4)
		got[0], got[1] = increaseCluster(&tt.cl, &tt.gs, tt.gpMI)
		got[2], got[3] = tt.cl, tt.gs
		for i := range got {
			if !reflect.DeepEqual(tt.want[i], got[i]) {
				t.Errorf("wrong cl%v,  want:\n%v\ngot:\n%v\n", i, tt.want[i], got[i])
			}
		}
	}
}

func TestChooseGroupOne(t *testing.T) {
	gs := []vars.VarList{
		vars.NewList([]int{0, 1}, nil),
		vars.NewList([]int{2, 3}, nil),
		vars.NewList([]int{4, 5}, nil),
		vars.NewList([]int{6, 7}, nil),
		vars.NewList([]int{8, 9}, nil),
	}
	cases := []struct {
		cl, gs1, gs2 []vars.VarList
		want         bool
	}{
		{gs, gs[:2], gs[2:], true},
		{gs, gs[2:], gs[:2], false},
		{gs, append([]vars.VarList(nil), gs[0], gs[2], gs[4]), append([]vars.VarList(nil), gs[1], gs[3]), true},
		{gs, append([]vars.VarList(nil), gs[1], gs[3]), append([]vars.VarList(nil), gs[0], gs[2], gs[4]), false},
	}
	for _, tt := range cases {
		got := chooseGroupOne(tt.cl, tt.gs1, tt.gs2)
		if got != tt.want {
			t.Errorf("wrong choice, want:%v got: %v\n", tt.want, got)
		}
	}
}

func TestConnectSubtrees(t *testing.T) {
	varStr := `
variables:
- {name: X0,  card: 2}
- {name: X1,  card: 2}
- {name: X2,  card: 2}
- {name: X3,  card: 2}
- {name: X4,  card: 2}
- {name: X5,  card: 2}
- {name: X6,  card: 2}
- {name: X7,  card: 2}
- {name: X8,  card: 2}
- {name: Y0,  card: 3, latent: true}
- {name: Y1,  card: 4, latent: true}
- {name: Y2,  card: 2, latent: true}
- {name: Y3,  card: 2, latent: true}
`
	subtreeStr := []string{
		varStr + `
nodes:
- clqvars: [Y0]
- clqvars: [X0, Y0]
  parent: [Y0]
- clqvars: [X1, Y0]
  parent: [Y0]
`, varStr + `
nodes:
- clqvars: [Y1]
- clqvars: [X2, Y1]
  parent: [Y1]
- clqvars: [X3, X4, Y1]
  parent: [Y1]
`, varStr + `
nodes:
- clqvars: [Y2]
- clqvars: [X5, Y2]
  parent: [Y2]
- clqvars: [X6, Y2]
  parent: [Y2]
`, varStr + `
nodes:
- clqvars: [Y3]
- clqvars: [X7, Y3]
  parent: [Y3]
- clqvars: [X8, Y3]
  parent: [Y3]
`,
	}
	treeStr := varStr + `
nodes:
- clqvars: [Y0]
- clqvars: [X0, Y0]
  parent: [Y0]
- clqvars: [X1, Y0]
  parent: [Y0]

- clqvars: [Y0,Y3]
  parent: [Y0]
- clqvars: [X7, Y3]
  parent: [Y0,Y3]
- clqvars: [X8, Y3]
  parent: [Y0,Y3]

- clqvars: [Y1,Y3]
  parent: [Y0,Y3]
- clqvars: [X2, Y1]
  parent: [Y1,Y3]
- clqvars: [X3, X4, Y1]
  parent: [Y1,Y3]

- clqvars: [Y2,Y3]
  parent: [Y0,Y3]
- clqvars: [X5, Y2]
  parent: [Y2,Y3]
- clqvars: [X6, Y2]
  parent: [Y2,Y3]
`
	var subtrees []*model.CTree
	for _, str := range subtreeStr {
		subtrees = append(subtrees, model.CTreeFromString(str))
	}
	ct := model.CTreeFromString(treeStr)
	edges := []graph.WEdge{
		{Head: 0, Tail: 3},
		{Head: 3, Tail: 1},
		{Head: 3, Tail: 2},
	}
	cases := []struct {
		edges    []graph.WEdge
		subtrees []*model.CTree
		ct       *model.CTree
	}{{edges, subtrees, ct}}
	for _, tt := range cases {
		got := connectSubtrees(tt.edges, tt.subtrees)
		if !got.EqualStruct(tt.ct) {
			t.Errorf("wrong structure, want:\n%v\ngot:\n%v\n", tt.ct, got)
		}
	}
}
