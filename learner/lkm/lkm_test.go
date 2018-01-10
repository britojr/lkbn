package lkm

import (
	"testing"

	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/emlearner"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/ioutl"
)

func TestCreateLKMStruct(t *testing.T) {
	vs := vars.NewList([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil)
	vstr := `
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
- {name: X9,  card: 2}
- {name: X10, card: 2}
`
	lvs := []*vars.Var{vars.New(11, 3, "Y11", true), vars.New(12, 5, "Y12", true)}
	lv0 := `
- {name: Y11,  card: 3, latent: true}
`
	lv1 := `
- {name: Y12,  card: 5, latent: true}
`
	nodes1L := `
nodes:
- clqvars: [Y11]
- clqvars: [X1, X5, X7, Y11]
  parent: [Y11]
- clqvars: [X2, X3, Y11]
  parent: [Y11]
- clqvars: [X4, X6, X8, Y11]
  parent: [Y11]
- clqvars: [X0, X9, X10, Y11]
  parent: [Y11]
`
	nodes2L := `
nodes:
- clqvars: [Y11, Y12]
- clqvars: [X1, X5, X7, Y12]
  parent: [Y11, Y12]
- clqvars: [X2, X3, Y12]
  parent: [Y11, Y12]
- clqvars: [X4, X6, X8, Y11]
  parent: [Y11, Y12]
- clqvars: [X0, X9, X10, Y12]
  parent: [Y11, Y12]
`
	struct1L := vstr + lv0 + nodes1L
	struct2L := vstr + lv0 + lv1 + nodes2L
	cases := []struct {
		gs1, gs2 []vars.VarList
		lvs      vars.VarList
		reloc    int
		ct       *model.CTree
	}{{
		[]vars.VarList{
			[]*vars.Var{vs[1], vs[5], vs[7]},
			[]*vars.Var{vs[2], vs[3]},
			[]*vars.Var{vs[4], vs[6], vs[8]},
			[]*vars.Var{vs[0], vs[9], vs[10]},
		}, nil,
		lvs[:1], -1,
		model.CTreeFromString(struct1L),
	}, {
		[]vars.VarList{
			[]*vars.Var{vs[4], vs[6], vs[8]},
			[]*vars.Var{vs[2], vs[3]},
		},
		[]vars.VarList{
			[]*vars.Var{vs[1], vs[5], vs[7]},
			[]*vars.Var{vs[0], vs[9], vs[10]},
		},
		lvs, 1,
		model.CTreeFromString(struct2L),
	}}
	for _, tt := range cases {
		ct := createLKMStruct(tt.lvs, tt.gs1, tt.gs2, tt.reloc)
		if !tt.ct.EqualStruct(ct) {
			t.Errorf("different tree:\n%v\n-----\n%v\n", tt.ct, ct)
		}
	}
}

type fakeLearner struct {
	maxcard int
}

func (f fakeLearner) SetProperties(props map[string]string) {}
func (f fakeLearner) PrintProperties()                      {}
func (f fakeLearner) Copy() emlearner.EMLearner             { return f }
func (f fakeLearner) Run(m *model.CTree, evset []map[int]int) (*model.CTree, float64, int) {
	vs := m.Variables()
	c := 1
	for _, v := range vs {
		if v.Latent() {
			c = v.NState()
			break
		}
	}
	if c > f.maxcard {
		c = f.maxcard
	}
	m.SetScore(-1000 / float64(c))
	return m, m.Score(), 1
}

func TestLearnLKM1L(t *testing.T) {
	content := `A,B,C,D,E,F,G,H,I,J,K
0,0,0,0,0,0,0,0,0,0,0
1,0,1,1,1,1,1,0,1,1,1
1,1,1,1,1,1,1,1,1,1,1
0,0,0,0,0,0,0,0,0,0,0
0,0,0,0,0,0,0,0,0,0,0
0,0,0,0,1,0,0,0,0,0,0
0,0,0,0,0,0,0,0,0,0,0
0,0,0,0,0,0,0,0,0,0,0
0,0,0,0,0,0,0,0,0,0,0
`
	vs := vars.NewList([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil)
	fname := ioutl.TempFile("lkm_test", content)
	cases := []struct {
		gs      []vars.VarList
		lv      *vars.Var
		ds      *data.Dataset
		maxcard int
	}{{
		[]vars.VarList{
			[]*vars.Var{vs[1], vs[5], vs[7]},
			[]*vars.Var{vs[2], vs[3]},
			[]*vars.Var{vs[4], vs[6], vs[8]},
			[]*vars.Var{vs[0], vs[9], vs[10]},
		},
		vars.New(len(vs), 2, "", true),
		data.NewDataset(fname),
		5,
	}}
	for _, tt := range cases {
		ct, lv := LearnLKM1L(tt.gs, tt.lv, tt.ds, fakeLearner{tt.maxcard})
		vs := ct.Variables()
		if len(vs) != len(tt.ds.Variables())+1 {
			t.Errorf("latent variable wasn't created: %v", vs)
		}
		if lv.NState() != tt.maxcard {
			t.Errorf("stoped at wrong cardinality: %v != %v", tt.maxcard, lv.NState())
		}
		if lv != ct.Variables()[len(tt.ds.Variables())] {
			t.Errorf("wrong return: %v != %v", lv, ct.Variables()[len(tt.ds.Variables())])
		}
	}
}

func TestComputeModelSize(t *testing.T) {
	struct1 := `
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
- {name: X9,  card: 2}
- {name: X10, card: 2}
- {name: Y11,  card: 3, latent: true}
- {name: Y12,  card: 5, latent: true}
nodes:
- clqvars: [Y11, Y12]
- clqvars: [X1, X5, X7, Y12]
  parent: [Y11, Y12]
- clqvars: [X2, X3, Y12]
  parent: [Y11, Y12]
- clqvars: [X4, X6, X8, Y11]
  parent: [Y11, Y12]
- clqvars: [X0, X9, X10, Y12]
  parent: [Y11, Y12]
`
	cases := []struct {
		ct   *model.CTree
		want int
	}{{
		model.CTreeFromString(struct1),
		120,
	}}
	for _, tt := range cases {
		got := computeModelSize(tt.ct)
		if got != tt.want {
			t.Errorf("wrong model size: %v != %v", tt.want, got)
		}
	}
}
